// Run the webserver
package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/moosemorals/mm/server"
)

func logRequest(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		h.ServeHTTP(w, req)
		log.Printf("%s %s %s", req.RemoteAddr, req.Method, req.RequestURI)
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

	opts.AddAddr(":8081", ":8443")

	s := server.Create(opts)
	s.Handle("/", logRequest(http.FileServer(http.Dir(*wwwroot))))
	s.Start()
}
