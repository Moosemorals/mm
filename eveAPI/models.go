package eveapi

// This file contains types suitable for decoding JSON from the eve api

// Verify holds basic character info
type Verify struct {
	CharacterID           int32
	CharacterName         string
	CharacterOwnerHash    string
	ExpiresOn             string
	IntelllectualProperty string // Probably actally mean IP address, I'll check
	Scopes                string
	TokenType             string
}
