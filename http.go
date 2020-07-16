package easyHttpCrud

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

func SendM(w http.ResponseWriter, s int, m string) {
	var o struct {
		Text string `json:"message"`
	}
	o.Text = m
	SendObject(w, s, o)
}

func SendObject(w http.ResponseWriter, s int, o interface{}) {
	js, _ := json.Marshal(o)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(s)
	w.Write(js)
}

func SendRes(w http.ResponseWriter, r *http.Response) error {
	var body interface{}
	_, err := ParseResBody(r, &body)
	if err != nil {
		return err
	}
	SendObject(w, r.StatusCode, body)
	return nil
}

func ParseBody(r *http.Request, v interface{}) ([]byte, error) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1024*512))
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, v)
	if err != nil {
		return nil, err
	}
	return body, nil
}
func ParseResBody(r *http.Response, v interface{}) ([]byte, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body.Close()
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	err = json.Unmarshal(body, v)
	if err != nil {
		if string(body) == "Not Found" {
			return body, nil
		}
		return nil, err
	}
	return body, nil
}
