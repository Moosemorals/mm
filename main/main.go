// Run the webserver
package main

import "github.com/moosemorals/mm/server"
import "net/http"

type hello struct{}

func (h *hello) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "text/plain")
	res.WriteHeader(200)
	res.Write([]byte("Hello, world!\n"))
}

func main() {
	s := server.Create(":8081", ":8443")
	s.Handle("/", &hello{})
	s.Start()
}
