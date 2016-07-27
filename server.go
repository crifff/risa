package risa

import (
	"net/http"
	"encoding/json"
	"hoshina85/risa/jsonrpc2"
	"io/ioutil"
	"bytes"
	"io"
	"hoshina85/risa/rpc"
	"fmt"
	"github.com/go-openapi/errors"
)

type JsonRPCServer struct {
	ServiceMap *rpc.ServiceMap
}

func NewJsonRPCServer() *JsonRPCServer {
	return &JsonRPCServer{
		ServiceMap: rpc.NewServiceMap(),
	}
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
		requests, err := batchRequest(rdr2)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
		}
		cnt := len(requests)

		var responses []jsonrpc2.Response = make([]jsonrpc2.Response, cnt)
		for i, req := range requests {
			response, errExecute := s.execute(req, r)
			if errExecute != nil {
				http.Error(w, "Bad Request", http.StatusInternalServerError)
			}
			responses[i] = response
		}
		json, _ := json.Marshal(responses)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(json))
	} else {

		req, err := request(rdr2)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
		}

		response, errExecute := s.execute(req, r)

		if errExecute != nil {
			http.Error(w, "Bad Request", http.StatusInternalServerError)

		}
		json, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(json))

	}
}
func (s *JsonRPCServer) execute(req jsonrpc2.Request, r *http.Request) (jsonrpc2.Response, error) {

	if err := jsonrpc2.ValidateRequest(req); err != nil {
		//http.Error(w, "Bad Request", http.StatusBadRequest)
		return jsonrpc2.Response{}, errors.New(http.StatusBadRequest, "Bad Request")
	}

	reply, errCall := s.ServiceMap.Call(req.Method, r)
	if errCall != nil {
		return jsonrpc2.Response{}, errors.New(http.StatusBadRequest, "Bad Request")
	}
	jsonReply, _ := json.Marshal(reply.Elem().Interface())

	response := jsonrpc2.Response{
		JsonRPC:jsonrpc2.Version,
		Result: string(jsonReply),
		Error: nil,
		ID: "",
	}
	return response, nil
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

func (s *JsonRPCServer) Register(service interface{}) error {
	return s.ServiceMap.Register(service)
}