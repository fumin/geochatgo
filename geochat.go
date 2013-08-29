package geochat

import (
  "fmt"
  "html/template"
  "encoding/json"
  "net/http"
  "regexp"
)

func init() {
  http.HandleFunc("/say", say)
  http.HandleFunc("/chatlogs/", chatlogs)
  http.Handle("/static/",
    http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
  http.HandleFunc("/", index)
}

func index(w http.ResponseWriter, r *http.Request) {
  fmt.Println(LatLngToInt(LatLng{12, 12}, 25))
  t, _ := template.ParseFiles("tmpl/index.html")
  t.Execute(w, nil)
}

func say(w http.ResponseWriter, r *http.Request) {
  msg := r.FormValue("msg")
  latitude := r.FormValue("latitude")
  longitude := r.FormValue("longitude")
  fmt.Println(msg, latitude, longitude)

  jsonResp(w, map[string]interface{}{"ok": true})
}

func chatlogs(w http.ResponseWriter, r *http.Request) {
  re := regexp.MustCompile(`/(\d+)/(\d+)/(\d+)`)
  matches := re.FindStringSubmatch(r.URL.Path)
  x := matches[1]
  y := matches[2]
  z := matches[3]
  fmt.Println(x, y, z)
  fmt.Fprintf(w, "hello")
}

// Helpers
func jsonResp(w http.ResponseWriter, o map[string]interface{}) {
  w.Header().Set("Content-Type", "application/json")
  b, err := json.Marshal(o)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  w.Write(b)
}
