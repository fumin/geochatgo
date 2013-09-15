package main

import (
	"flag"
	"fmt"
	_ "geochat"
	"net/http"
)

const (
	defaultPort = 3000
	portUsage   = "port to bind to"
)

var port int

func init() {
	flag.IntVar(&port, "port", defaultPort, portUsage)
	flag.IntVar(&port, "p", defaultPort, portUsage)
}

func main() {
	flag.Parse()
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
