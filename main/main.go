package main

import (
	_ "geochat"
	"net/http"
)

func main() {
	http.ListenAndServe(":3000", nil)
}
