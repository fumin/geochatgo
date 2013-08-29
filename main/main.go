package main

import (
  "net/http"
  _ "geochat"
)

func main() {
  http.ListenAndServe(":3000", nil)
}
