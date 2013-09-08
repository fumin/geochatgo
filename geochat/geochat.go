package main

import (
	"flag"
	_ "geochat"
	"net/http"
)

func main() {
	flag.Parse()
	http.ListenAndServe(":3000", nil)
}
