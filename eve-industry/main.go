package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

// Config holds configuration details
type Config struct {
	ClientID string
	Secret   string
}

func readConfig() (c *Config, err error) {

	raw, err := ioutil.ReadFile("config.json")
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(raw, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func main() {
	c, err := readConfig()
	if err != nil {
		log.Fatal("Can't read configuration", err)
	}

}
