package eveapi

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/oauth2"
)

const apiURL = "https://esi.evetech.net/"

// Eve holds state for the Eve API
type Eve struct {
	conf  Config
	oauth *oauth2.Config
	users *UserCache
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

func writeError(w http.ResponseWriter, msg string, err error) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(500)

	fmt.Fprintf(w, "%s\n%v", msg, err)
}

func apiFetch(client *http.Client, path string) (json.RawMessage, error) {
	log.Printf("EVE API Fetching %s", path)

	resp, err := client.Get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var j json.RawMessage
	if err = json.Unmarshal(body, &j); err != nil {
		return nil, err
	}

	return j, nil
}

// NewEve creates a new eve
func NewEve() *Eve {
	e := Eve{}
	err := e.readConfig()
	if err != nil {
		log.Fatal("Can't read eve config:", err)
	}
	e.oauth = &oauth2.Config{
		ClientID:     e.conf.ClientID,
		ClientSecret: e.conf.Secret,
		Scopes:       []string{"esi-assets.read_assets.v1"},
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

func (e *Eve) makeClient(tok *oauth2.Token) *http.Client {
	return e.oauth.Client(context.Background(), tok)
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
	w.Header().Set("Content-Type", "text/html;charset=utf-8")
	user, err := e.getUser(r)
	if err == nil {

		t := template.New("hello2.html")
		t, err := t.ParseFiles("../eveAPI/hello2.html")
		if err != nil {
			writeError(w, "Can't parse template", err)
			return
		}

		if err = t.Execute(w, user); err != nil {
			writeError(w, "Can't use template", err)
			return
		}
	} else {

		t := template.New("hello.html")
		t, err := t.ParseFiles("../eveAPI/hello.html")
		if err != nil {
			writeError(w, "Can't parse template", err)
			return
		}

		state := randStr(40)

		data := struct {
			AuthURL string
			State   string
		}{
			AuthURL: e.getAuthURL(state),
			State:   state,
		}

		if err = t.Execute(w, data); err != nil {
			writeError(w, "Can't use template", err)
			return
		}
	}
}

func (e *Eve) handleAuthCallback(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	authCode := query.Get("code")
	state := query.Get("state")

	tok, err := e.oauth.Exchange(context.Background(), authCode)
	if err != nil {
		writeError(w, "Can't exchange auth code for token", err)
		return
	}

	client := e.makeClient(tok)
	j, err := apiFetch(client, fmt.Sprintf("%sverify", apiURL))
	if err != nil {
		writeError(w, "Can't verify user", err)
		return
	}

	var v Verify
	if err = json.Unmarshal(j, &v); err != nil {
		writeError(w, "Can't parse user data", err)
		return
	}

	user := User{
		Token: tok,
		ID:    v.CharacterID,
		Name:  v.CharacterName,
	}

	e.users.add(state, &user)

	http.SetCookie(w, &http.Cookie{
		Name:     "state",
		Value:    state,
		Secure:   true,
		HttpOnly: true,
		MaxAge:   60 * 60 * 24,
	})

	w.Header().Set("Location", "/eve")
	w.WriteHeader(302)
}
func (e *Eve) handleAPI(w http.ResponseWriter, r *http.Request) {

	user, err := e.getUser(r)
	if err != nil {
		writeError(w, "Can't get current user", err)
		return
	}

	target := fmt.Sprintf("%slatest/characters/%d/assets/", apiURL, user.ID)

	raw, err := apiFetch(e.makeClient(user.Token), target)
	if err != nil {
		log.Fatal("Can't get from api", err)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(200)
	w.Write(raw)

}

func (e *Eve) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/eve/auth2") {
		log.Print("EVE: Handing to auth callback")
		e.handleAuthCallback(w, r)
	} else if strings.HasPrefix(r.URL.Path, "/eve/api") {
		log.Print("EVE: Handing to API")
		e.handleAPI(w, r)
	} else {
		log.Print("EVE: Showing login link")
		e.handleLogin(w, r)
	}
}
