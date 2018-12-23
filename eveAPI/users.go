package eveapi

import (
	"encoding/json"
	"log"
	"os"

	"golang.org/x/oauth2"
)

// User holds details about an Eve user
type User struct {
	Token *oauth2.Token
	Name  string
	ID    int32
	state string
}

// UserCache Holds a cache of users that can be written to disk
type UserCache struct {
	Users map[string]*User
	path  string
}

func readUserCache(path string) (*UserCache, error) {
	u := &UserCache{
		path: path,
	}

	in, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Thats fine, return an empty cache
			log.Print("INFO: Cache not found on disk, creating")
			u.Users = make(map[string]*User)
			return u, nil
		}
		return nil, err
	}

	if err = json.NewDecoder(in).Decode(u); err != nil {
		return nil, err
	}

	return u, nil
}

func (u *UserCache) write() {

	out, err := os.Create(u.path)
	if err != nil {
		log.Print("WARN: Can't save user cache: ", err)
	}

	if err = json.NewEncoder(out).Encode(u); err != nil {
		log.Print("WARN: Can't save user cache: ", err)
	}
}

func (u *UserCache) user(id string) (*User, bool) {
	user, ok := u.Users[id]
	return user, ok
}

func (u *UserCache) add(id string, user *User) {
	u.Users[id] = user
	go u.write()
}
