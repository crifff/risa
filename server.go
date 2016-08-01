package risa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-openapi/errors"
	"github.com/hoshina85/risa/jsonrpc2"
	"github.com/hoshina85/risa/rpc"
	"gopkg.in/guregu/null.v3"
	"io"
	"io/ioutil"
	"net/http"
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
			response := responseError(jsonrpc2.ParseError, "")
			writeResponse(w, response)
			return
		}
		if len(requests) == 0 {
			response := responseError(jsonrpc2.InvalidRequestError, "")
			writeResponse(w, response)
			return
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
		writeResponse(w, responses)
	} else {
		req, err := request(rdr2)
		if err != nil {
			response := responseError(jsonrpc2.ParseError, req.ID)
			writeResponse(w, response)
			return
		}

		response, errExecute := s.execute(req, r)
		if errExecute != nil {
			http.Error(w, "Bad Request", http.StatusInternalServerError)
		}
		writeResponse(w, response)
	}
}

func (s *JsonRPCServer) execute(req jsonrpc2.Request, r *http.Request) (jsonrpc2.Response, error) {

	if err := jsonrpc2.ValidateRequest(req); err != nil {
		return jsonrpc2.Response{}, errors.New(http.StatusBadRequest, "Bad Request")
	}
	if s.ServiceMap.HasMethod(req.Method) == false {
		response := responseError(jsonrpc2.MethodNotFoundError, req.ID)
		return response, nil
	}

	reply, errCall := s.ServiceMap.Call(req, r)
	if errCall != nil {
		return jsonrpc2.Response{}, errors.New(http.StatusBadRequest, "Bad Request")
	}

	response := jsonrpc2.Response{
		JsonRPC: jsonrpc2.Version,
		Result:  reply.Elem().Interface(),
		Error:   nil,
		ID:      null.NewString(req.ID, req.ID != ""),
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

func responseError(code jsonrpc2.ErrorCode, id string) jsonrpc2.Response {
	response := jsonrpc2.Response{
		JsonRPC: jsonrpc2.Version,
		Error: &jsonrpc2.Error{
			Code:    code,
			Message: jsonrpc2.ErrorMessage[code],
		},
		ID: null.NewString(id, id != ""),
	}
	return response
}

func writeResponse(w http.ResponseWriter, response interface{}) error {
	json, err := json.Marshal(response)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	_, errPrint := fmt.Fprint(w, string(json))
	return errPrint
}
