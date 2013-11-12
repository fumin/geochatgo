package geochat

import (
	"encoding/json"
	"html/template"
	"net/http"
	"sync"
)

var g_webrtcMap = struct {
	sync.RWMutex
	m map[string]chan recvMsg_t
}{m: make(map[string]chan recvMsg_t)}

func webrtc(w http.ResponseWriter, r *http.Request) {
	g_webrtcMap.RLock()
	members := make([]string, 0)
	for token, _ := range g_webrtcMap.m {
		members = append(members, token)
	}
	g_webrtcMap.RUnlock()

	t, _ := template.ParseFiles("tmpl/webrtc.html")
	data := struct {
		Members []string
	}{
		Members: members,
	}
	t.Execute(w, data)
}

func webrtcTransmitter(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	data := struct {
		Destinations []string
		Body         string
	}{}
	err := decoder.Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	g_webrtcMap.RLock()
	for _, destination := range data.Destinations {
		channel, ok := g_webrtcMap.m[destination]
		if !ok {
			continue
		}
		select {
		case channel <- recvMsg_t{"", []byte(data.Body)}:
		default:
		}
	}
	g_webrtcMap.RUnlock()
}

const leaveMsg = "leave"

func webrtcLeaveSource(w http.ResponseWriter, r *http.Request) {
	parser := paramParser{R: r, W: w}
	token := parser.RequiredStringParam("token")
	if parser.Err != nil {
		return
	}

	g_webrtcMap.RLock()
	channel, ok := g_webrtcMap.m[token]
	g_webrtcMap.RUnlock()
	if ok {
		select {
		case channel <- recvMsg_t{leaveMsg, []byte{}}:
		default:
		}
	}
}
func webrtcSource(w http.ResponseWriter, r *http.Request) {
	token := string(randByteSlice())
	channel := make(chan recvMsg_t, 32)
	g_webrtcMap.Lock()
	g_webrtcMap.m[token] = channel
	g_webrtcMap.Unlock()
	defer func() {
		g_webrtcMap.Lock()
		delete(g_webrtcMap.m, token)
		byeMsg := `{"type": "bye", "from":"` + token + `"}`
		for _, c := range g_webrtcMap.m {
			select {
			case c <- recvMsg_t{"", []byte(byeMsg)}:
			default:
			}
		}
		g_webrtcMap.Unlock()
	}()

	sse := NewServerSideEventWriter(w)
	sse.Write([]byte(`{"type":"token", "token":"` + token + `"}`))
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
