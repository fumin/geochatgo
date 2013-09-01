package geochat

import (
  "errors"
  "fmt"
  "encoding/json"
  "net/http"
  "strconv"
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
