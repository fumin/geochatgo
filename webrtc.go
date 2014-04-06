package geochat

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sync"
	"time"

	"github.com/fumin/webutil"
)

type roomMap struct {
	sync.RWMutex
	tokenToRoom  map[string]string
	roomToTokens map[string](map[string]bool)
}

// Leave current room and enter new room
func (r *roomMap) enterRoom(token string, room string) map[string]bool {
	r.Lock()
	defer r.Unlock()

	currentRoom := r.tokenToRoom[token]
	if currentRoom == room {
		return r.roomToTokens[room]
	} else {
		r._leaveRoom(token)
	}

	r.tokenToRoom[token] = room
	if r.roomToTokens[room] == nil {
		r.roomToTokens[room] = make(map[string]bool)
	}
	r.roomToTokens[room][token] = true
	return r.roomToTokens[room]
}
func (r *roomMap) leaveRoom(token string) {
	r.Lock()
	r._leaveRoom(token)
	r.Unlock()
}
func (r *roomMap) _leaveRoom(token string) {
	room := r.tokenToRoom[token]
	delete(r.tokenToRoom, token)
	delete(r.roomToTokens[room], token)
}
func (r *roomMap) tokensInRoom(room string) map[string]bool {
	r.RLock()
	defer r.RUnlock()
	return r.roomToTokens[room]
}

var g_roomMap = &roomMap{
	tokenToRoom:  make(map[string]string),
	roomToTokens: make(map[string](map[string]bool)),
}

var g_webrtcMap = struct {
	sync.RWMutex
	m map[string]chan recvMsg_t
}{m: make(map[string]chan recvMsg_t)}

func webrtc(w http.ResponseWriter, r *http.Request) {
	fmt.Println("kkk")
	room := r.FormValue("room")
	if room == "" {
		room = string(webutil.RandByteSlice())
	}

	t, _ := template.ParseFiles("tmpl/webrtc.html")
	data := struct {
		Room string
	}{
		Room: room,
	}
	t.Execute(w, data)
}

func webrtcJoin(w http.ResponseWriter, r *http.Request) {
	parser := webutil.ParamParser{R: r, W: w}
	token := parser.RequiredStringParam("token")
	room := parser.RequiredStringParam("room")
	if parser.Err != nil {
		return
	}

	members := g_roomMap.enterRoom(token, room)

	webutil.JsonResp(w, map[string]interface{}{"members": members})
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
	parser := webutil.ParamParser{R: r, W: w}
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
	token := string(webutil.RandByteSlice())
	channel := make(chan recvMsg_t, 32)
	g_webrtcMap.Lock()
	g_webrtcMap.m[token] = channel
	g_webrtcMap.Unlock()
	defer func() {
		g_roomMap.leaveRoom(token)

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

	// Openshift proxy's keep-alive has a timeout of 15, we need to be shorter.
	sse := webutil.NewServerSideEventWriter(w, "heartbeat", 10*time.Second)
	defer sse.Close()
	sse.Write([]byte(`{"type":"token", "token":"` + token + `"}`))
L:
	for {
		select {
		case msg := <-channel:
			if msg.kind == leaveMsg {
				break L
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
