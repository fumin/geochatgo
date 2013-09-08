package geochat

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
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
	west, south, east, north, err := requiredLatLngBoundsParam(r, w)
	if err != nil {
		return
	}

	// The presence of a user should be handled solely by
	// the EventSource endpoint "/stream". Therefore we "update",
	// which returns an error if the username does not exist in the Rtree yet,
	// instead of "insert" the bounds of the user's map here.
	err = rtreeClient.RtreeUpdate(rtreekeyUser, username,
		[]float64{west, south}, []float64{east, north})
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
	west, south, east, north, err := requiredLatLngBoundsParam(r, w)
	if err != nil {
		return
	}
	username := string(randByteSlice())

	// We create a new record in Rtree to store a user's map bounds, and
	// delete that record when the EventSource connection is lost.
	err = rtreeClient.RtreeInsert(rtreekeyUser, username,
		[]float64{west, south}, []float64{east, north})
	if err != nil {
		glog.Warningf("%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		err = rtreeClient.RtreeDelete(rtreekeyUser, username)
		if err != nil {
			glog.Errorf("%v", err)
		}
	}()

	// Use Redis' PubSub feature to pass messages around.
	c, err := NewRedisConn()
	if err != nil {
		glog.Warningf("%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer c.Close()
	subscriber := NewRedisSubscriber(c, username)

	// Inform client what her/his username is. Throughout the entire session,
	// clients should use this string as the identifier of themselves.
	sse := NewServerSideEventWriter(w)
	err = sse.EventWrite("username", []byte(username))
	if err != nil {
		return
	}

L:
	for {
		select {
		case msg := <-subscriber:
			switch v := msg.(type) {
			case redis.Message:
				sse.EventWrite("custom", v.Data)
			case error:
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
	lat, err := requiredFloatParam("latitude", r, w)
	if err != nil {
		return
	}
	lng, err := requiredFloatParam("longitude", r, w)
	if err != nil {
		return
	}
	data := map[string]interface{}{
		"msg":        msg,
		"created_at": time.Now().Unix(),
		"latitude":   lat,
		"longitude":  lng,
	}
	conn := redisPool.Get()
	defer conn.Close()

	// Store the message into the chatlogs which may be retrieved later on.
	err = maptileStore(rediskeyTileChatlog, data, conn)
	if err != nil {
		glog.Warningf("%v", err)
	}

	// Broadcast message to others using the Redis pipeline feature.
	b, _ := json.Marshal(data)
	neighbors, err := rtreeClient.RtreeNearestNeighbors(
		rtreekeyUser, 100, []float64{lat, lng})
	if err != nil {
		glog.Warningf("%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, neighbor := range neighbors {
		if neighbor == r.FormValue("username") {
			continue
		}
		err = conn.Send("PUBLISH", neighbor, b)
		if err != nil {
			glog.Warningf("%v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	conn.Flush()
	for _, neighbor := range neighbors {
		if neighbor == r.FormValue("username") {
			continue
		}
		_, err := conn.Receive()
		if err != nil {
			glog.Warningf("%v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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
	west, err := requiredFloatParam("west", r, w)
	if err != nil {
		return -200, -200, -200, -200, err
	}
	south, err := requiredFloatParam("south", r, w)
	if err != nil {
		return -200, -200, -200, -200, err
	}
	east, err := requiredFloatParam("east", r, w)
	if err != nil {
		return -200, -200, -200, -200, err
	}
	north, err := requiredFloatParam("north", r, w)
	if err != nil {
		return -200, -200, -200, -200, err
	}

	if west >= east {
		err = errors.New(fmt.Sprintf("west %v > east %v", west, east))
		return -200, -200, -200, -200, err
	}
	if south >= north {
		err = errors.New(fmt.Sprintf("south %v > north %v", south, north))
		return -200, -200, -200, -200, err
	}
	return west, south, east, north, nil
}
