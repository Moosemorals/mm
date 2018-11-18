// Package server is the webserver for moosemorals.com
package server

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

// Server is a wrapper around net.httpd
type Server struct {
	http, https *http.Server
	mux         *http.ServeMux
}

func buildRedirect(httpsAddr string, req *http.Request) string {
	var host string
	if strings.Contains(req.Host, ":") {
		host, _, _ = net.SplitHostPort(req.Host)
	} else {
		host = req.Host
	}

	if strings.Contains(httpsAddr, ":") {
		_, port, _ := net.SplitHostPort(httpsAddr)
		return "https://" + host + ":" + port + req.URL.String()
	}
	return "https://" + host + req.URL.String()
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
			url := buildRedirect(httpsAddr, req)
			log.Printf("Redirecting to %s", url)
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
	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)

		// Wait for signals
		<-sigint

		log.Printf("Shutting down")
		if err := s.http.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
		if err := s.https.Shutdown(context.Background()); err != nil {
			log.Printf("HTTPS server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	go func() {
		log.Printf("Listining on %s", s.http.Addr)
		if err := s.http.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("HTTP server ListenAndServe: %v", err)
		}
	}()

	go func() {
		log.Printf("Listening on %s", s.https.Addr)
		if err := s.https.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
			log.Printf("HTTPs server ListenAndServeTLS: %v", err)
		}
	}()

	// Wait for shutdown
	<-idleConnsClosed
}
