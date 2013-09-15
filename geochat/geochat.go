package main

import (
	"flag"
	"fmt"
	_ "geochat"
	"net/http"
)

const (
	defaultHost = "127.0.0.1"
	hostUsage   = "host to bind to"
	defaultPort = 3000
	portUsage   = "port to bind to"
)

var host string
var port int

func init() {
	flag.StringVar(&host, "host", defaultHost, hostUsage)
	flag.StringVar(&host, "h", defaultHost, hostUsage)
	flag.IntVar(&port, "port", defaultPort, portUsage)
	flag.IntVar(&port, "p", defaultPort, portUsage)
}

func main() {
	flag.Parse()
	err := http.ListenAndServe(fmt.Sprintf("%v:%d", host, port), nil)
	if err != nil {
		fmt.Printf(err.Error())
	}
}
