// Run the webserver
package main

import (
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/moosemorals/mm/eveapi"
	_ "github.com/moosemorals/mm/linkshare"
	"github.com/moosemorals/mm/server"
)

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

	// Make the server
	s := server.Create(opts)

	// Add a static handler
	s.Handle("/", http.FileServer(http.Dir(*wwwroot)))

	// Add the eve handler
	s.Handle("/eveapi/", eveapi.NewEve())

	/*
	// Add the linkshare handler
	hub := linkshare.NewHub()
	s.Handle("/ws", hub)
	s.OnShutdown(hub.Shutdown)
*/

	s.Start()
}
