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

// TypeAttr holds a pointer to an attribute, and a value
type TypeAttr struct {
	Attribute *EveAttribute `json:"attribute"`
	Value     float64       `json:"value"`
}

// TypeValue is an id/value pair
type TypeValue struct {
	ID    int32 `json:"id"`
	Value int64 `json:"v"`
}

// EveType holds basic data about things in eve
type EveType struct {
	BasePrice     float64           `json:"basePrice"`
	Description   map[string]string `json:"description"`
	GroupID       int32             `json:"groupID"`
	IconID        int32             `json:"iconID"`
	MarketGroupID int32             `json:"marketGroupID"`
	Name          map[string]string `json:"name"`
	PortionSize   int32             `json:"portionSize"`
	Published     bool              `json:"published"`
	RaceID        int32             `json:"raceID"`
	Volume        float64           `json:"volume"`
	Mass          float64           `json:"mass"`
	Radius        float64           `json:"radius"`
	FactionID     int32             `json:"factionID"`
	Attributes    []TypeAttr        `json:"attributes"`
	Materials     []TypeValue       `json:"materials"`
}

type eveTypes map[int32]*EveType

// EveAttribute defines an attribute on a type
type EveAttribute struct {
	ID           int32   `json:"attributeID"`
	Name         string  `json:"attributeName"`
	Description  string  `json:"description"`
	DisplayName  string  `json:"displayName"`
	CategoryID   int32   `json:"categoryID"`
	DefaultValue float64 `json:"defaultValue"`
	HighIsGood   bool    `json:"highIsGood"`
	Published    bool    `json:"published"`
	Stackable    bool    `json:"stackable"`
	UnitID       int32   `json:"unitID"`
}

type eveAttributes map[int32]*EveAttribute
