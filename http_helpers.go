package geochat

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"
)

type paramParser struct {
	R   *http.Request
	W   http.ResponseWriter
	Err error
}

const errInt int = math.MinInt32

func (parser *paramParser) RequiredIntParam(key string) int {
	if parser.Err != nil {
		return errInt
	}
	v := parser.R.FormValue(key)
	if v == "" {
		errMsg := fmt.Sprintf("missing param: %v", key)
		http.Error(parser.W, errMsg, http.StatusBadRequest)
		parser.Err = errors.New(errMsg)
		return errInt
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		errMsg := fmt.Sprintf("Wrong integer format: %v", v)
		http.Error(parser.W, errMsg, http.StatusBadRequest)
		parser.Err = errors.New(errMsg)
		return errInt
	}
	return i
}

func (parser *paramParser) OptionalIntParam(key string, defaultVal int) int {
	if parser.Err != nil {
		return errInt
	}
	v := parser.R.FormValue(key)
	if v == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		errMsg := fmt.Sprintf("Wrong integer format: %v", v)
		http.Error(parser.W, errMsg, http.StatusBadRequest)
		parser.Err = errors.New(errMsg)
		return errInt
	}
	return i
}

const errFloat float64 = math.SmallestNonzeroFloat64

func (parser *paramParser) RequiredFloatParam(key string) float64 {
	if parser.Err != nil {
		return errFloat
	}
	v := parser.R.FormValue(key)
	if v == "" {
		errMsg := fmt.Sprintf("missing param: %v", key)
		http.Error(parser.W, errMsg, http.StatusBadRequest)
		parser.Err = errors.New(errMsg)
		return errFloat
	}
	f, err := strconv.ParseFloat(v, 32)
	if err != nil {
		errMsg := fmt.Sprintf("Wrong integer format: %v", v)
		http.Error(parser.W, errMsg, http.StatusBadRequest)
		parser.Err = errors.New(errMsg)
		return errFloat
	}
	return f
}

func (parser *paramParser) RequiredStringParam(key string) string {
	if parser.Err != nil {
		return ""
	}
	v := parser.R.FormValue(key)
	if v == "" {
		errMsg := fmt.Sprintf("missing param: %v", key)
		http.Error(parser.W, errMsg, http.StatusBadRequest)
		parser.Err = errors.New(errMsg)
		return ""
	}
	return v
}

func jsonResp(w http.ResponseWriter, o map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	b, err := json.Marshal(o)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(b)
}

type byteWriter struct {
	respWriter http.ResponseWriter
	err        error
}

func (w *byteWriter) Write(b []byte) {
	if w.err != nil {
		return
	}
	_, w.err = w.respWriter.Write(b)
}

const SSEHeatbeat = "heartbeat"

type Sse struct {
	w          http.ResponseWriter
	ticker     *time.Ticker
	ConnClosed chan bool
}

func NewServerSideEventWriter(w http.ResponseWriter) Sse {
	headers := w.Header()
	headers.Set("Content-Type", "text/event-stream")
	headers.Set("Cache-Control", "no-cache")
	headers.Set("Connection", "keep-alive")

	// Openshift proxy's keep-alive has a timeout of 15, we need to be shorter.
	sse := Sse{w, time.NewTicker(10 * time.Second), make(chan bool, 1)}
	go func() {
		for _ = range sse.ticker.C {
			err := sse.EventWrite(SSEHeatbeat, make([]byte, 0))
			if err != nil {
				sse.ConnClosed <- true
				return
			}
		}
	}()
	return sse
}

func (sse Sse) Write(b []byte) error {
	bw := &byteWriter{respWriter: sse.w}
	bw.Write([]byte("data: "))
	bw.Write(b)
	bw.Write([]byte("\n\n"))
	if bw.err != nil {
		return bw.err
	}
	if f, ok := sse.w.(http.Flusher); ok {
		f.Flush()
	}
	return nil
}

func (sse Sse) EventWrite(event string, b []byte) error {
	bw := &byteWriter{respWriter: sse.w}
	bw.Write([]byte("event: "))
	bw.Write([]byte(event))
	bw.Write([]byte("\n"))
	if bw.err != nil {
		return bw.err
	}
	return sse.Write(b)
}
