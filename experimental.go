package geochat

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"sync"
)

var g_webrtcMap = struct {
	sync.RWMutex
	m map[string]chan recvMsg_t
}{m: make(map[string]chan recvMsg_t)}

type webrtcDataJs_t struct {
	RoomSize int
}
type webrtcData_t struct {
	Js webrtcDataJs_t
}

func webrtc(w http.ResponseWriter, r *http.Request) {
	g_webrtcMap.RLock()
	l := len(g_webrtcMap.m)
	g_webrtcMap.RUnlock()

	t, _ := template.ParseFiles("tmpl/webrtc.html")
	data := webrtcData_t{
		Js: webrtcDataJs_t{
			RoomSize: l,
		},
	}
	t.Execute(w, data)
}

func webrtcTransmitter(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	g_webrtcMap.RLock()
	for _, channel := range g_webrtcMap.m {
		select {
		case channel <- recvMsg_t{"", body}:
		default:
		}
	}
	g_webrtcMap.RUnlock()
}

const leaveMsg = "leave"

func webrtcLeaveSource(w http.ResponseWriter, r *http.Request) {
	parser := paramParser{R: r, W: w}
	username := parser.RequiredStringParam("username")
	if parser.Err != nil {
		return
	}

	g_webrtcMap.RLock()
	channel, ok := g_webrtcMap.m[username]
	g_webrtcMap.RUnlock()
	if ok {
		select {
		case channel <- recvMsg_t{leaveMsg, []byte{}}:
		default:
		}
	}
}
func webrtcSource(w http.ResponseWriter, r *http.Request) {
	username := string(randByteSlice())
	channel := make(chan recvMsg_t, 32)
	g_webrtcMap.Lock()
	g_webrtcMap.m[username] = channel
	g_webrtcMap.Unlock()
	defer func() {
		g_webrtcMap.Lock()
		delete(g_webrtcMap.m, username)
		g_webrtcMap.Unlock()
	}()

	sse := NewServerSideEventWriter(w)
	sse.Write([]byte(`{"type":"username", "username":"` + username + `"}`))
L:
	for {
		select {
		case msg := <-channel:
			if msg.kind == leaveMsg {
				sse.StopTicker <- true
				continue
			}
			err := sse.Write(msg.content)
			if err != nil {
				break L
			}
		case <-sse.ConnClosed:
			break L
		}
	}
}
