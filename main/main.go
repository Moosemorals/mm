// Run the webserver
package main

import (
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/moosemorals/mm/linkshare"
	"github.com/moosemorals/mm/server"
)

func logRequest(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			log.Printf("%s %s %s", req.RemoteAddr, req.Method, req.RequestURI)
		}()
		h.ServeHTTP(w, req)
	}
}

func main() {
	opts := server.Options{}

	wwwroot := flag.String("wwwroot", ".", "Directory to serve static files from")
	debug := flag.Bool("debug", false, "Use debug certificates")

	flag.Parse()

	if *debug {
		log.Println("Debug enabled")
		opts.SetDebug()
	}

	for _, a := range flag.Args() {
		parts := strings.Split(a, ",")
		if len(parts) == 2 {
			opts.AddAddr(parts[0], parts[1])
		}
	}


	s := server.Create(opts)
	// Static handler
	s.Handle("/", logRequest(http.FileServer(http.Dir(*wwwroot))))

	// Linkshare
	hub := linkshare.NewHub()
	s.Handle("/ws", hub)

	s.Start()
}
