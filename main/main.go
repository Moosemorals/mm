// Run the webserver
package main

import "github.com/moosemorals/mm/server"

func main() {
	s := server.Create(":8081", ":8443")
	s.Start()
}