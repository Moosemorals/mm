package eveapi

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/oauth2"
)

const apiURL = "https://esi.evetech.net"

var interestingHeaders = []string{"Content-Type", "Content-Length", "Cache-Control", "ETag", "Expires", "Last-Modified", "X-Pages"}

// Eve holds state for the Eve API
type Eve struct {
	conf       Config
	oauth      *oauth2.Config
	users      *UserCache
	apiCache   *apiCache
	types      eveTypes
	attributes eveAttributes
	blueprints eveBlueprints
	groups     eveGroups
}

// Config holds configuration details
type Config struct {
	ClientID    string
	Secret      string
	RedirectURL string
}

// Originally from https://stackoverflow.com/a/50581165/195833
func randStr(len int) string {
	flatten := regexp.MustCompile(`[^a-zA-Z0-9]`)
	buff := make([]byte, len)
	rand.Read(buff)
	str := base64.StdEncoding.EncodeToString(buff)

	// Base 64 can be longer than len
	return flatten.ReplaceAllString(str[:len], "")
}

func writeError(w http.ResponseWriter, status int, msg string, err error) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)

	if err != nil {
		fmt.Fprintf(w, "%s\n%v", msg, err)
	} else {
		fmt.Fprintf(w, "%s", msg)
	}
}

func getAPIPath(path string) string {
	return fmt.Sprintf("%s%s", apiURL, path)
}

func (e *Eve) apiGet(u *User, path string) (*http.Response, error) {
	log.Printf("EVE API GET %s", path)
	return e.apiCache.get(e.makeClient(u), getAPIPath(path))
}

func (e *Eve) apiPost(u *User, path string, body io.ReadCloser) (*http.Response, error) {
	log.Printf("EVE API POST %s", path)
	return e.makeClient(u).Post(getAPIPath(path), "application/json", body)
}

// NewEve creates a new eve
func NewEve() *Eve {
	e := Eve{}
	err := e.readConfig()
	if err != nil {
		log.Fatal("Can't read eve config:", err)
	}

	if err := e.loadStatic(); err != nil {
		log.Fatal("Can't load eve static data")
	}

	e.apiCache = newAPICache()

	e.oauth = &oauth2.Config{
		ClientID:     e.conf.ClientID,
		ClientSecret: e.conf.Secret,
		Scopes:       []string{"publicData", "esi-assets.read_assets.v1", "esi-industry.read_character_jobs.v1", "esi-characters.read_blueprints.v1"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://login.eveonline.com/v2/oauth/authorize/",
			TokenURL: "https://login.eveonline.com/v2/oauth/token",
		},
		RedirectURL: e.conf.RedirectURL,
	}

	u, err := readUserCache("users/cache")
	if err == nil {
		e.users = u
	} else {
		log.Printf("Problem reading the user cache")
	}
	return &e
}

func (e *Eve) readConfig() error {

	raw, err := ioutil.ReadFile("../eveAPI/config.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(raw, &e.conf)
	if err != nil {
		return err
	}

	return nil
}

func (e *Eve) makeClient(u *User) *http.Client {
	return e.oauth.Client(context.Background(), u.Token)
}

func (e *Eve) getAuthURL(state string) string {
	return e.oauth.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (e *Eve) getUser(r *http.Request) (*User, error) {
	authCookie, err := r.Cookie("state")
	if err != nil {
		return nil, err
	}

	user, ok := e.users.user(authCookie.Value)
	if !ok {
		return nil, errors.New("User not found")
	}

	return user, nil
}

func (e *Eve) handleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	user, err := e.getUser(r)
	if err == nil {
		json.NewEncoder(w).Encode(struct {
			Name string
			ID   int32
		}{
			Name: user.Name,
			ID:   user.ID,
		})
	} else {
		state := randStr(42)
		json.NewEncoder(w).Encode(struct {
			AuthURL string
		}{
			AuthURL: e.getAuthURL(state),
		})
	}
}

func (e *Eve) handleAuthCallback(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	authCode := query.Get("code")
	state := query.Get("state")

	tok, err := e.oauth.Exchange(context.Background(), authCode)
	if err != nil {
		writeError(w, 500, "Can't exchange auth code for token", err)
		return
	}

	user := &User{
		Token: tok,
	}

	response, err := e.apiGet(user, "/verify")
	if err != nil {
		writeError(w, 500, "Can't verify user", err)
		return
	}
	defer response.Body.Close()
	var v Verify
	if err = json.NewDecoder(response.Body).Decode(&v); err != nil {
		writeError(w, 500, "Can't parse verify respone", err)
		return
	}

	user.ID = v.CharacterID
	user.Name = v.CharacterName

	e.users.add(state, user)

	http.SetCookie(w, &http.Cookie{
		Name:     "state",
		Value:    state,
		Secure:   true,
		HttpOnly: true,
		MaxAge:   60 * 60 * 24,
	})

	w.Header().Set("Location", "/eve/index.html")
	w.WriteHeader(302)
}

func (e *Eve) handleAPI(w http.ResponseWriter, r *http.Request) {
	user, err := e.getUser(r)
	if err != nil {
		writeError(w, 500, "Can't get current user", err)
		return
	}

	param := r.URL.Query()

	target := param.Get("p")
	if len(target) == 0 {
		writeError(w, 400, "Missing path", nil)
		return
	}
	method := param.Get("m")

	var resp *http.Response
	if len(method) == 0 || method == "GET" {
		resp, err = e.apiGet(user, target)
	} else if method == "POST" {
		log.Println("EVE: Starting post")
		resp, err = e.apiPost(user, target, r.Body)
		log.Println("EVE: Post complete")
	}

	if err != nil {
		writeError(w, 500, "Failed to fetch from API", nil)
		return
	}

	defer resp.Body.Close()

	for _, name := range interestingHeaders {
		w.Header().Set(name, resp.Header.Get(name))
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)

}

func (e *Eve) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/eveapi/auth2") {
		log.Print("EVE: Handing to auth callback")
		e.handleAuthCallback(w, r)
	} else if strings.HasPrefix(r.URL.Path, "/eveapi/api") {
		log.Print("EVE: Handing to API")
		e.handleAPI(w, r)
	} else if strings.HasPrefix(r.URL.Path, "/eveapi/static") {
		log.Print("EVE: Handing to Static")
		e.handleStatic(w, r)
	} else {
		log.Print("EVE: Sending user data")
		e.handleLogin(w, r)
	}
}
