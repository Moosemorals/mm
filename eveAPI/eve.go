package eveapi

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"golang.org/x/oauth2"
)

const apiURL = "https://esi.evetech.net/"

// Eve holds state for the Eve API
type Eve struct {
	conf    Config
	oauth   *oauth2.Config
	token   *oauth2.Token
	context context.Context
	client  *http.Client
	mux     *http.ServeMux
	state   string
}

// Config holds configuration details
type Config struct {
	ClientID string
	Secret   string
	MyID     int
}

// From https://stackoverflow.com/a/50581165/195833
func randStr(len int) string {
	buff := make([]byte, len)
	rand.Read(buff)
	str := base64.StdEncoding.EncodeToString(buff)
	// Base 64 can be longer than len
	return str[:len]
}

// NewEve creates a new eve
func NewEve() *Eve {
	e := Eve{
		context: context.Background(),
		mux:     http.NewServeMux(),
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
	}

	e.mux.HandleFunc("/", e.handleLogin)
	e.mux.HandleFunc("/auth2", e.handleAuthCallback)
	e.mux.HandleFunc("/api", e.handleAPI)
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
	target := fmt.Sprintf("%scharacters/%d/assets/", apiURL, e.conf.MyID)
	resp, err := e.client.Get(target)
	if err != nil {
		log.Fatal("Can't get assets", err)
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(200)
	_, err = io.Copy(w, resp.Body)

	if err != nil {
		log.Fatal("Can't write body", err)
	}
}

func (e *Eve) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.mux.ServeHTTP(w, r)
}
