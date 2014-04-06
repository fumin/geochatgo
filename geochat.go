package geochat

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/glog"
	"html/template"
	"net/http"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/fumin/webutil"
)

func init() {
	initConfig()

	http.HandleFunc("/numgoroutine", numgoroutine)

	// Experimental
	http.HandleFunc("/webrtc", webrtc)
	http.HandleFunc("/webrtc/join", webrtcJoin)
	http.HandleFunc("/webrtc/signal/transmitter", webrtcTransmitter)
	http.HandleFunc("/webrtc/signal/source", webrtcSource)
	http.HandleFunc("/webrtc/signal/leave_source", webrtcLeaveSource)

	// Production
	http.HandleFunc("/open_popup", openPopup)
	http.HandleFunc("/close_popup", closePopup)
	http.HandleFunc("/update_mapbounds", updateMapbounds)
	http.HandleFunc("/stream", stream)
	http.HandleFunc("/say", say)
	http.HandleFunc("/chatlogs/", chatlogs)
	http.Handle("/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", index)
}

func openPopup(w http.ResponseWriter, r *http.Request) {
	parser := &webutil.ParamParser{R: r, W: w}
	username := parser.RequiredStringParam("username")
	south, west, north, east := requiredLatLngBoundsParam(parser)
	if parser.Err != nil {
		return
	}

	point := [2]float64{south, west}
	lengths := [2]float64{north - south, east - west}
	popupId, err := rTree.insertPopup(username, point, lengths)
	if err != nil {
		glog.Warningf("%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	webutil.JsonResp(w, map[string]interface{}{"popupId": popupId})
}

func closePopup(w http.ResponseWriter, r *http.Request) {
	parser := webutil.ParamParser{R: r, W: w}
	username := parser.RequiredStringParam("username")
	popupId := parser.RequiredStringParam("popupId")
	if parser.Err != nil {
		return
	}

	rTree.delPopup(username, popupId)
	webutil.JsonResp(w, map[string]interface{}{"ok": true})
}

func updateMapbounds(w http.ResponseWriter, r *http.Request) {
	parser := &webutil.ParamParser{R: r, W: w}
	username := parser.RequiredStringParam("username")
	south, west, north, east := requiredLatLngBoundsParam(parser)
	if parser.Err != nil {
		return
	}

	// Users' presences should be handled solely by the endpoint "/stream".
	// Therefore we "update", which requires that the key exist in the Rtree,
	// instead of "insert" the bounds of the user's map here.
	lengths := [2]float64{north - south, east - west}
	err := rTree.update(username, [2]float64{south, west}, lengths)
	if err != nil {
		glog.Warningf("%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	webutil.JsonResp(w, map[string]interface{}{"ok": true})
}

func index(w http.ResponseWriter, r *http.Request) {
	// fmt.Println(LatLngToInt(LatLng{12, 12}, 25))
	t, _ := template.ParseFiles("tmpl/index.html")
	t.Execute(w, nil)
}

func stream(w http.ResponseWriter, r *http.Request) {
	username := string(webutil.RandByteSlice())

	// Insert client into Rtree with an map bounds [0, 0, 0.1, 0.1].
	// The client will update the correct bounds after we inform her/his username.
	channel := make(chan recvMsg_t, 32)
	rTree.insert(username, channel, [2]float64{0.0, 0.0}, [2]float64{0.1, 0.1})
	defer rTree.del(username) // when the EventSource connection is lost.

	// Openshift proxy's keep-alive has a timeout of 15, we need to be shorter.
	sse := webutil.NewServerSideEventWriter(w, "heartbeat", 10*time.Second)
	defer sse.Close()
	// Inform client what her/his username is. Throughout the entire session,
	// clients should use this string as the identifier of themselves.
	err := sse.EventWrite("username", []byte(username))
	if err != nil {
		return
	}

L:
	for {
		select {
		case msg := <-channel:
			err = sse.EventWrite(msg.kind, msg.content)
			if err != nil {
				break L
			}
		case <-sse.ConnClosed:
			break L
		}
	}
}

func say(w http.ResponseWriter, r *http.Request) {
	parser := &webutil.ParamParser{R: r, W: w}
	msg := parser.RequiredStringParam("msg")
	lat, lng := requiredLatLngParams(parser, "latitude", "longitude")
	skipSelf := r.FormValue("skipSelf")
	if parser.Err != nil {
		return
	}
	data := map[string]interface{}{
		"msg":        msg,
		"created_at": time.Now().Unix(),
		"latitude":   lat,
		"longitude":  lng,
	}

	// Store the message into the chatlogs which may be retrieved later on.
	conn := redisPool.Get()
	err := maptileStore(rediskeyTileChatlog, data, conn)
	conn.Close()
	if err != nil {
		glog.Warningf("%v", err)
	}

	// Broadcast message to others.
	neighbors := rTree.nearestNeighbors(100, [2]float64{lat, lng})
	b, _ := json.Marshal(data)
	for _, neighbor := range neighbors {
		if skipSelf == "true" && neighbor.receiver.key == r.FormValue("username") {
			continue
		}
		select {
		case neighbor.receiver.channel <- recvMsg_t{"custom", b}:
		default:
		}
	}
	popups := rTree.searchContaining(100, [2]float64{lat, lng})
	for _, popup := range popups {
		select {
		case popup.channel <- recvMsg_t{popup.key, b}:
		default:
		}
	}

	webutil.JsonResp(w, map[string]interface{}{"ok": true})
}

func chatlogs(w http.ResponseWriter, r *http.Request) {
	re := regexp.MustCompile(`/(\d+)/(\d+)/(\d+)`)
	matches := re.FindStringSubmatch(r.URL.Path)
	if len(matches) != 4 {
		errMsg := fmt.Sprintf("Wrong path format: %v", r.URL.Path)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	z, _ := strconv.Atoi(matches[1])
	x, _ := strconv.Atoi(matches[2])
	y, _ := strconv.Atoi(matches[3])

	parser := webutil.ParamParser{R: r, W: w}
	limit := parser.OptionalIntParam("limit", 16)
	if parser.Err != nil {
		return
	}

	conn := redisPool.Get()
	defer conn.Close()
	v, err := maptileRead(rediskeyTileChatlog, z, x, y, 0, limit, conn)
	if err != nil {
		glog.Warningf("%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// These no cache headers do not work for openshift...
	// w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	// w.Header().Set("Pragma", "no-cache")
	// w.Header().Set("Expires", "0")
	// w.Header().Set("ETag", string(webutil.RandByteSlice()))
	//
	// Openshift's cache is a bastard, it accepts only UTC time,
	// when we explicitly stated that the local machine time is in EDT.
	// w.Header().Set("Last-Modified", time.Now().UTC().Format(time.RFC1123))

	// Concatenate json strings by ourselves, should be fast than json.Marshall?
	w.Header().Set("Content-Type", "application/json")
	bw := &webutil.ByteWriter{RespWriter: w}
	bw.Write([]byte{'['})
	bw.Write([]byte(strings.Join(v, ",")))
	bw.Write([]byte{']'})
}

func numgoroutine(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf(`{"numgoroutines":%d}`, runtime.NumGoroutine())))
}

func requiredLatLngBoundsParam(parser *webutil.ParamParser) (float64, float64, float64, float64) {
	south, west := requiredLatLngParams(parser, "south", "west")
	north, east := requiredLatLngParams(parser, "north", "east")
	if parser.Err != nil {
		return -200, -200, -200, -200
	}

	if west >= east {
		errMsg := fmt.Sprintf("west %v >= east %v", west, east)
		http.Error(parser.W, errMsg, http.StatusBadRequest)
		parser.Err = errors.New(errMsg)
		return -200, -200, -200, -200
	}
	if south >= north {
		errMsg := fmt.Sprintf("south %v >= north %v", south, north)
		http.Error(parser.W, errMsg, http.StatusBadRequest)
		parser.Err = errors.New(errMsg)
		return -200, -200, -200, -200
	}
	return south, west, north, east
}

func requiredLatLngParams(parser *webutil.ParamParser, latKey, lngKey string) (float64, float64) {
	lat := parser.RequiredFloatParam(latKey)
	lng := parser.RequiredFloatParam(lngKey)
	if parser.Err != nil {
		return -200, -200
	}
	if lat < -90 || lat > 90 {
		errMsg := fmt.Sprint("Wrong latitude: %v", lat)
		http.Error(parser.W, errMsg, http.StatusBadRequest)
		parser.Err = errors.New(errMsg)
		return -200, -200
	}
	if lng < -180 || lng > 180 {
		errMsg := fmt.Sprint("Wrong longitude: %v", lng)
		http.Error(parser.W, errMsg, http.StatusBadRequest)
		parser.Err = errors.New(errMsg)
		return -200, -200
	}
	return lat, lng
}
