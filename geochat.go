package geochat

import (
  "encoding/json"
  "fmt"
  "html/template"
  "net/http"
  "regexp"
  "time"
  "github.com/golang/glog"
)

func init() {
  initConfig()
  http.HandleFunc("/stream", stream)
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

func stream(w http.ResponseWriter, r *http.Request) {
  west, err := requiredFloatParam("west", r, w); if err != nil { return }
  south, err := requiredFloatParam("south", r, w); if err != nil { return }
  east, err := requiredFloatParam("east", r, w); if err != nil { return }
  north, err := requiredFolatParam("north", r, w); if err != nil { return }
  username := randByteSlice()

  // Subscribe to username in Redis
  c, err := NewRedisConn()
  if err != nil {
    glog.Warningf("%v", err)
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  defer c.Close()
  subscriber := NewRedisSubscriber(c, string(username))

  // Record client's map border
  err = rtreeClient.RtreeInsert(rtreekeyUser, username,
                                []float64{west, south}, []float64{east, north})
  if err != nil {
    glog.Warningf("%v", err)
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  defer func() {
    err = rtreeClient.RtreeDelete(rtreekeyUser, username)
    if err != nil { glog.Errorf("%v", err) }
  }()

  // Inform client about her/his username
  sse := NewServerSideEventWriter(w)
  err = sse.EventWrite("username", username) if err != nil { return }

  L:
  for {
    select {
    case msg <- subscriber:
      switch v := msg.(type) {
      case redis.Message:
        sse.Write(v.Data)
      case error:
        break L
      }
    case <- sse.ConnClosed:
      glog.Infof("Closing %v", username)
      break L
    }
  }
}

func say(w http.ResponseWriter, r *http.Request) {
  msg, err := requiredStringParam("msg", r, w); if err != nil { return }
  lat, err := requiredFloatParam("latitude", r, w); if err != nil { return }
  lng, err := requiredFloatParam("longitude", r, w); if err != nil { return }
  data := map[string]interface{}{
    "msg": msg,
    "created_at": time.Now().Unix(),
    "latitude": lat,
    "longitude": lng,
  }
  conn := redisPool.Get()
  defer conn.Close()

  // Store the message into the chatlogs.
  err = maptileStore(rediskeyTileChatlog, data, conn)
  if err != nil {
    glog.Warningf("%v", err)
  }

  // Broadcast message to others
  b, _ := json.Marshal(data)
  neighbors, err := rtreeClient.RtreeNearestNeighbors(
                      rtreekeyUser, 100, []float64{lat, lng})
  if err != nil {
    glog.Warningf("%v", err)
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  for _, neighbor := range neighbors {
    err = conn.Send("PUBLISH", neighbor, b)
    if err != nil {
      glog.Warningf("%v", err)
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
  }
  conn.Flush()
  for _, _ = range neighbors {
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
  x := matches[1]
  y := matches[2]
  z := matches[3]
  fmt.Println(x, y, z)
  fmt.Fprintf(w, "hello")
}
