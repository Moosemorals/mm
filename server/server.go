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
	Options
	servers []*http.Server
	mux     *http.ServeMux
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
func Create(opts Options) *Server {
	s := &Server{
		Options: opts,
		mux:     http.NewServeMux(),
	}
	for i, a := range s.httpAddr {
		s.servers = append(s.servers, &http.Server{
			Addr:         a,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Connection", "close")
				url := buildRedirect(s.httpsAddr[i], req)
				log.Printf("Redirecting to %s", url)
				http.Redirect(w, req, url, http.StatusMovedPermanently)
			}),
		})
	}

	tlsConfig := &tls.Config{
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
	if !s.debug {
		m := &autocert.Manager{
			Cache:      autocert.DirCache("~/moosemorals.com/tls"),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist("moosemorals.com", "www.moosemorals.com"),
		}
		tlsConfig.GetCertificate = m.GetCertificate
		tlsConfig.NextProtos = m.TLSConfig().NextProtos
	}

	for _, a := range s.httpsAddr {
		s.servers = append(s.servers, &http.Server{
			Addr:         a,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
			TLSConfig:    tlsConfig,
			Handler:      s.mux,
		})
	}
	return s
}

// Handle a http request to a path
func (s *Server) Handle(path string, h http.Handler) {
	s.mux.Handle(path, h)
}

// OnShutdown passes f to the http.Server.RegisterShutdown function
func (s *Server) OnShutdown(f func()) {
	for _, h := range s.servers {
		if getProto(h) == "HTTPS" {
			h.RegisterOnShutdown(f)
		}
	}
}

func getProto(h *http.Server) string {
	if h.TLSConfig != nil {
		return "HTTPS"
	}
	return "HTTP"
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
		for _, server := range s.servers {
			if err := server.Shutdown(context.Background()); err != nil {
				log.Printf("%s server %s shutdown error: %v", getProto(server), server.Addr, err)
			}
		}
		close(idleConnsClosed)
	}()

	for _, x := range s.servers {
		go func(server *http.Server) {
			proto := getProto(server)
			if proto == "HTTP" {
				log.Printf("HTTP server listening on %s", server.Addr)
				if err := server.ListenAndServe(); err != http.ErrServerClosed {
					log.Printf("HTTP server %s error ListenAndServe: %v", server.Addr, err)
				}
			} else {
				log.Printf("HTTPS server listening on %s", server.Addr)
				var err error
				if !s.debug {
					err = server.ListenAndServeTLS("", "")
				} else {
					err = server.ListenAndServeTLS("cert.pem", "key.pem")
				}
				if err != http.ErrServerClosed {
					log.Printf("HTTPS server %s error ListenAndServeTLS: %v", server.Addr, err)
				}

			}
		}(x)
	}
	// Wait for shutdown
	<-idleConnsClosed
}
