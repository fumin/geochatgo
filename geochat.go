package geochat

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/glog"
	"html/template"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func init() {
	initConfig()
	http.HandleFunc("/update_mapbounds", updateMapbounds)
	http.HandleFunc("/stream", stream)
	http.HandleFunc("/say", say)
	http.HandleFunc("/chatlogs/", chatlogs)
	http.Handle("/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", index)
}

func updateMapbounds(w http.ResponseWriter, r *http.Request) {
	username, err := requiredStringParam("username", r, w)
	if err != nil {
		return
	}
	south, west, north, east, err := requiredLatLngBoundsParam(r, w)
	if err != nil {
		return
	}

	// Users' presences should be handled solely by the endpoint "/stream".
	// Therefore we "update", which requires that the key exist in the Rtree,
	// instead of "insert" the bounds of the user's map here.
	err = rTree.update(username, [2]float64{south, west}, [2]float64{north - south, east - west})
	if err != nil {
		glog.Warningf("%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResp(w, map[string]interface{}{"ok": true})
}

func index(w http.ResponseWriter, r *http.Request) {
	// fmt.Println(LatLngToInt(LatLng{12, 12}, 25))
	t, _ := template.ParseFiles("tmpl/index.html")
	t.Execute(w, nil)
}

func stream(w http.ResponseWriter, r *http.Request) {
	username := string(randByteSlice())

	// Insert client into Rtree with an map bounds [0, 0, 0.1, 0.1].
	// The client will update the correct bounds after we inform her/his username.
	msg := make(chan []byte, 32)
	rTree.insert(username, msg, [2]float64{0.0, 0.0}, [2]float64{0.1, 0.1})
	defer rTree.del(username) // when the EventSource connection is lost.

	// Inform client what her/his username is. Throughout the entire session,
	// clients should use this string as the identifier of themselves.
	sse := NewServerSideEventWriter(w)
	err := sse.EventWrite("username", []byte(username))
	if err != nil {
		return
	}

L:
	for {
		select {
		case contents := <-msg:
			err = sse.EventWrite("custom", contents)
			if err != nil {
				break L
			}
		case <-sse.ConnClosed:
			break L
		}
	}
}

func say(w http.ResponseWriter, r *http.Request) {
	msg, err := requiredStringParam("msg", r, w)
	if err != nil {
		return
	}
	lat, lng, err := requiredLatLngParams("latitude", "longitude", r, w)
	if err != nil {
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
	err = maptileStore(rediskeyTileChatlog, data, conn)
	conn.Close()
	if err != nil {
		glog.Warningf("%v", err)
	}

	// Broadcast message to others using the Redis pipeline feature.
	neighbors := rTree.nearestNeighbors(100, [2]float64{lat, lng})
	b, _ := json.Marshal(data)
	for _, neighbor := range neighbors {
		if neighbor.key == r.FormValue("username") {
			continue
		}
		select {
		case neighbor.channel <- b:
		default:
		}
	}

	jsonResp(w, map[string]interface{}{"ok": true})
}

func chatlogs(w http.ResponseWriter, r *http.Request) {
	re := regexp.MustCompile(`/(\d+)/(\d+)/(\d+)`)
	matches := re.FindStringSubmatch(r.URL.Path)
	z, _ := strconv.Atoi(matches[1])
	x, _ := strconv.Atoi(matches[2])
	y, _ := strconv.Atoi(matches[3])
	conn := redisPool.Get()
	defer conn.Close()

	v, err := maptileRead(rediskeyTileChatlog, z, x, y, 0, 16, conn)
	if err != nil {
		glog.Warningf("%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Concatenate json strings by ourselves, should be fast than json.Marshall?
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write([]byte{'['})
	if err != nil {
		return
	}
	w.Write([]byte(strings.Join(v, ",")))
	if err != nil {
		return
	}
	w.Write([]byte{']'})
	if err != nil {
		return
	}
}

func requiredLatLngBoundsParam(r *http.Request, w http.ResponseWriter) (float64, float64, float64, float64, error) {
	south, west, err := requiredLatLngParams("south", "west", r, w)
	if err != nil {
		return -200, -200, -200, -200, err
	}
	north, east, err := requiredLatLngParams("north", "east", r, w)
	if err != nil {
		return -200, -200, -200, -200, err
	}

	if west >= east {
		errMsg := fmt.Sprintf("west %v >= east %v", west, east)
		err = errors.New(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return -200, -200, -200, -200, err
	}
	if south >= north {
		errMsg := fmt.Sprintf("south %v >= north %v", south, north)
		err = errors.New(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return -200, -200, -200, -200, err
	}
	return south, west, north, east, nil
}

func requiredLatLngParams(latKey, lngKey string, r *http.Request, w http.ResponseWriter) (float64, float64, error) {
	lat, err := requiredFloatParam(latKey, r, w)
	if err != nil {
		return -200, -200, err
	}
	if lat < -90 || lat > 90 {
		http.Error(w, fmt.Sprint("Wrong latitude: %v", lat), http.StatusBadRequest)
		return -200, -200, err
	}
	lng, err := requiredFloatParam(lngKey, r, w)
	if err != nil {
		return -200, -200, err
	}
	if lng < -180 || lng > 180 {
		http.Error(w, fmt.Sprint("Wrong longitude: %v", lng), http.StatusBadRequest)
		return -200, -200, err
	}
	return lat, lng, nil
}
