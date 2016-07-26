package risa

import (
	"net/http"
	"encoding/json"
	"hoshina85/risa/jsonrpc2"
	"io/ioutil"
	"bytes"
	"io"
)

type JsonRPCServer struct {
}

func NewJsonRPCServer() *JsonRPCServer {
	return &JsonRPCServer{}
}

func (s *JsonRPCServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}
	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
	rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))

	firstChar := make([]byte, 1)
	if rdr1.Read(firstChar); string(firstChar) == "[" {
		_, err := batchRequest(rdr2)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
		}
	} else {
		_, err := request(rdr2)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
		}
	}
}

func request(reader io.ReadCloser) (jsonrpc2.Request, error) {
	var req jsonrpc2.Request

	decoder := json.NewDecoder(reader)
	decoder.UseNumber()
	err := decoder.Decode(&req)
	if err != nil {
		return req, err
	}
	return req, nil
}

func batchRequest(reader io.ReadCloser) ([]jsonrpc2.Request, error) {
	var req []jsonrpc2.Request

	decoder := json.NewDecoder(reader)
	decoder.UseNumber()
	err := decoder.Decode(&req)
	if err != nil {
		return req, err
	}
	return req, nil
}