package eveapi

import (
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"log"
	"net/http"
	"os"
	"time"
)

const cacheBase = "cache/"

type cacheEntry struct {
	responseTime time.Time
	maxAge       time.Duration
}

func (e *cacheEntry) fresh() bool {
	return time.Since(e.responseTime) < e.maxAge
}

func (e *cacheEntry) tilStale() time.Duration {
	return time.Duration(e.maxAge - time.Since(e.responseTime))
}

type apiCache struct {
	store map[string]*cacheEntry
}

func newAPICache() *apiCache {
	return &apiCache{
		store: make(map[string]*cacheEntry),
	}
}

func stringHash(in string) string {
	sha := sha256.New()
	sha.Write([]byte(in))
	return base64.URLEncoding.EncodeToString(sha.Sum(nil))
}

func cachePath(target string) string {
	return cacheBase + stringHash(target)
}

func calcMaxAge(resp *http.Response) time.Duration {
	date, err := http.ParseTime(resp.Header.Get("Date"))
	if err != nil {
		return time.Duration(0)
	}

	expires, err := http.ParseTime(resp.Header.Get("Expires"))
	if err != nil {
		return time.Duration(0)
	}

	return expires.Sub(date)
}

func (c *apiCache) put(target string, resp *http.Response) (*cacheEntry, error) {
	e := &cacheEntry{
		responseTime: time.Now(),
		maxAge:       calcMaxAge(resp),
	}

	if !e.fresh() {
		return nil, errors.New("Already stale")
	}

	path := cachePath(target)
	log.Printf("Saving body of %s to %s", target, path)
	out, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if err := resp.Write(out); err != nil {
		return nil, err
	}

	c.store[target] = e
	return e, nil
}

func constructResponse(target string, entry *cacheEntry) (*http.Response, error) {
	path := cachePath(target)
	log.Printf("Getting body of %s from %s", target, path)
	body, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return http.ReadResponse(bufio.NewReader(body), nil)
}

func (c *apiCache) get(client *http.Client, target string) (*http.Response, error) {
	entry, ok := c.store[target]
	if !ok || !entry.fresh() {

		resp, err := client.Get(target)
		if err != nil {
			return resp, err
		}
		entry, err = c.put(target, resp)
		if err != nil {
			log.Print("WARN: Couldn't store to cache: ", err)
			return resp, nil
		}
	}

	log.Printf("Serving %s from cache (%s left)", target, entry.tilStale())
	return constructResponse(target, entry)

}
