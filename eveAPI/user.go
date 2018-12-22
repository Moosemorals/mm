package eveapi

import (
	"golang.org/x/oauth2"
)

// User holds details about an Eve user
type User struct {
	Token *oauth2.Token
	Name  string
	ID    int32
	state string
}
