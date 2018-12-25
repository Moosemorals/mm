package eveapi

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"os"
)

const cacheBase = "cache/"

type cacheEntry struct {
	target     string
	headers    http.Header
	status     string
	statusCode int
	proto      string
	protoMajor int
	protoMinor int
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

func (c *apiCache) put(target string, resp *http.Response) (*cacheEntry, error) {
	e := &cacheEntry{
		target:     target,
		headers:    resp.Header,
		status:     resp.Status,
		statusCode: resp.StatusCode,
		proto:      resp.Proto,
		protoMajor: resp.ProtoMajor,
		protoMinor: resp.ProtoMinor,
	}

	path := cachePath(target)
	log.Printf("Saving body of %s to %s", target, path)
	out, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if _, err := io.Copy(out, resp.Body); err != nil {
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

	return &http.Response{
		Header:     entry.headers,
		Status:     entry.status,
		StatusCode: entry.statusCode,
		Proto:      entry.proto,
		ProtoMajor: entry.protoMajor,
		ProtoMinor: entry.protoMinor,
		Body:       body,
	}, nil
}

func (c *apiCache) get(client *http.Client, target string) (*http.Response, error) {
	entry, ok := c.store[target]
	if !ok {
		resp, err := client.Get(target)
		if err != nil {
			return resp, err
		}
		entry, err = c.put(target, resp)
		if err != nil {
			log.Print("WARN: Couldn't store to cache", err)
			return resp, nil
		}
	}

	log.Printf("Serving %s from cache", target)
	return constructResponse(target, entry)

}
