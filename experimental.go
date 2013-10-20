package geochat

import (
	"html/template"
	"net/http"
)

func webrtc(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("tmpl/webrtc.html")
	t.Execute(w, nil)
}
