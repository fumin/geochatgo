package geochat

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func requiredIntParam(key string, r *http.Request, w http.ResponseWriter) (int, error) {
	v := r.FormValue(key)
	if v == "" {
		errMsg := fmt.Sprintf("missing param: %v", key)
		http.Error(w, errMsg, http.StatusBadRequest)
		return 0, errors.New(errMsg)
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		errMsg := fmt.Sprintf("Wrong integer format: %v", v)
		http.Error(w, errMsg, http.StatusBadRequest)
		return 0, errors.New(errMsg)
	}
	return i, nil
}

func optionalIntParam(key string, defaultVal int, r *http.Request, w http.ResponseWriter) (int, error) {
	v := r.FormValue(key)
	if v == "" {
		return defaultVal, nil
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		errMsg := fmt.Sprintf("Wrong integer format: %v", v)
		http.Error(w, errMsg, http.StatusBadRequest)
		return 0, errors.New(errMsg)
	}
	return i, nil
}

func requiredFloatParam(key string, r *http.Request, w http.ResponseWriter) (float64, error) {
	v := r.FormValue(key)
	if v == "" {
		errMsg := fmt.Sprintf("missing param: %v", key)
		http.Error(w, errMsg, http.StatusBadRequest)
		return 0, errors.New(errMsg)
	}
	f, err := strconv.ParseFloat(v, 32)
	if err != nil {
		errMsg := fmt.Sprintf("Wrong integer format: %v", v)
		http.Error(w, errMsg, http.StatusBadRequest)
		return 0, errors.New(errMsg)
	}
	return f, nil
}

func requiredStringParam(key string, r *http.Request, w http.ResponseWriter) (string, error) {
	v := r.FormValue(key)
	if v == "" {
		errMsg := fmt.Sprintf("missing param: %v", key)
		http.Error(w, errMsg, http.StatusBadRequest)
		return "", errors.New(errMsg)
	}
	return v, nil
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

	sse := Sse{w, time.NewTicker(60 * time.Second), make(chan bool, 1)}
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
