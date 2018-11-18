// Package server is the webserver for moosemorals.com
package server

import (
	"crypto/tls"
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

// Server is a wrapper around net.httpd
type Server struct {
	http, https *http.Server
	mux         *http.ServeMux
}

// Create creates a new server
func Create(httpAddr, httpsAddr string) *Server {
	s := &Server{
		mux: http.NewServeMux(),
	}

	s.http = &http.Server{
		Addr:         httpAddr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Connection", "close")
			url := "https://" + req.Host + req.URL.String()
			http.Redirect(w, req, url, http.StatusMovedPermanently)
		}),
	}

	m := &autocert.Manager{
		Cache:      autocert.DirCache("~/moosemorals.com/tls"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist("moosemorals.com", "www.moosemorals.com"),
	}

	tlsConfig := &tls.Config{
		GetCertificate:           m.GetCertificate,
		NextProtos:               m.TLSConfig().NextProtos,
		PreferServerCipherSuites: true,
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.X25519,
		},
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305, // Go 1.8 only
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,   // Go 1.8 only
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}

	s.https = &http.Server{
		Addr:         httpsAddr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		TLSConfig:    tlsConfig,
		Handler:      s.mux,
	}
	return s
}

// Handle a http request to a path
func (s *Server) Handle(path string, h http.Handler) {
	s.mux.Handle(path, h)

}

// Start starts a server
func (s *Server) Start() {
	go func() { log.Fatal(s.http.ListenAndServe()) }()
	log.Fatal(s.https.ListenAndServeTLS("", ""))
}
