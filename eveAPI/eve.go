package eveapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// Eve holds state for the Eve API
type Eve struct {
	conf Config
}

// Config holds configuration details
type Config struct {
	ClientID string
	Secret   string
}

// NewEve creates a new eve
func NewEve() *Eve {
	var e Eve
	err := e.readConfig()
	if err != nil {
		log.Fatal("Can't read eve config:", err)
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

func (e *Eve) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	w.WriteHeader(200)

	fmt.Fprintln(w, "Hello, eve!")
}
