package server

// Options holds server options
type Options struct {
	debug bool

	httpAddr  []string
	httpsAddr []string
}

// SetDebug enables the debug option on the server.
func (o *Options) SetDebug() {
	o.debug = true
}

// AddAddr adds a pair of http/https addresses.
// The http address is only used to redirect to https
func (o *Options) AddAddr(http, https string) {
	o.httpAddr = append(o.httpAddr, http)
	o.httpsAddr = append(o.httpsAddr, https)
}
