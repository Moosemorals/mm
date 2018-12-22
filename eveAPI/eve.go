package eveapi

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/oauth2"
)

const apiURL = "https://esi.evetech.net/latest"

// Eve holds state for the Eve API
type Eve struct {
	conf    Config
	oauth   *oauth2.Config
	token   *oauth2.Token
	context context.Context
	client  *http.Client
	state   string
}

// Config holds configuration details
type Config struct {
	ClientID    string
	Secret      string
	MyID        int
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

// NewEve creates a new eve
func NewEve() *Eve {
	e := Eve{
		context: context.Background(),
	}
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

func (e *Eve) apiFetch(path string) (json.RawMessage, error) {
	log.Printf("EVE API Fetching %s", path)

	resp, err := e.client.Get(path)
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

func (e *Eve) getAuthURL() string {
	e.state = randStr(32)
	return e.oauth.AuthCodeURL(e.state, oauth2.AccessTypeOffline)
}

func (e *Eve) handleAuthCallback(w http.ResponseWriter, r *http.Request) {
	authCode := r.URL.Query().Get("code")

	tok, err := e.oauth.Exchange(e.context, authCode)

	if err != nil {
		log.Fatal("Can't exchange auth code for token", err)
	}

	e.token = tok
	e.client = e.oauth.Client(e.context, tok)

	t := template.New("hello2.html")
	t, err = t.ParseFiles("../eveapi/hello2.html")
	if err != nil {
		log.Fatal("Can't parse template", err)
	}

	w.Header().Set("Content-Type", "text/html;charset=utf-8")
	w.WriteHeader(200)
	if err := t.Execute(w, nil); err != nil {
		log.Fatal("Can't execute template", err)
	}
}

func (e *Eve) handleLogin(w http.ResponseWriter, r *http.Request) {

	t := template.New("hello.html")
	t, err := t.ParseFiles("../eveAPI/hello.html")
	if err != nil {
		log.Fatal("Can't parse template", err)
	}

	data := struct {
		AuthURL string
		State   string
	}{
		AuthURL: e.getAuthURL(),
		State:   e.state,
	}

	err = t.Execute(w, data)

	if err != nil {
		log.Fatal("Can't use template: ", err)
	}
}

func (e *Eve) handleAPI(w http.ResponseWriter, r *http.Request) {
	target := fmt.Sprintf("%s/characters/%d/assets/", apiURL, e.conf.MyID)

	raw, err := e.apiFetch(target)
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
