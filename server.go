package risa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hoshina85/risa/jsonrpc2"
	"github.com/hoshina85/risa/rpc"
	"gopkg.in/guregu/null.v3"
	"io"
	"io/ioutil"
	"net/http"
	"github.com/pkg/errors"
	"os"
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
			response := responseError(jsonrpc2.ParseError, "", err)
			writeResponse(w, response)
			return
		}
		if len(requests) == 0 {
			response := responseError(jsonrpc2.InvalidRequestError, "", errors.New("request is blank"))
			writeResponse(w, response)
			return
		}
		cnt := len(requests)

		var responses []jsonrpc2.Response = make([]jsonrpc2.Response, cnt)
		for i, req := range requests {
			response, errExecute := s.execute(req, r)
			if errExecute != nil {
				response = responseError(jsonrpc2.InternalError, req.ID, errExecute)
			}
			responses[i] = response
		}
		writeResponse(w, responses)
	} else {
		req, err := request(rdr2)
		if err != nil {
			response := responseError(jsonrpc2.ParseError, req.ID, err)
			writeResponse(w, response)
			return
		}

		response, errExecute := s.execute(req, r)
		if errExecute != nil {
			response := responseError(jsonrpc2.InternalError, req.ID, err)
			writeResponse(w, response)
			return
		}
		writeResponse(w, response)
	}
}

func (s *JsonRPCServer) execute(req jsonrpc2.Request, r *http.Request) (jsonrpc2.Response, error) {

	if err := jsonrpc2.ValidateRequest(req); err != nil {
		return jsonrpc2.Response{}, errors.Wrap(err, "Request Invalid")
	}
	if s.ServiceMap.HasMethod(req.Method) == false {
		response := responseError(jsonrpc2.MethodNotFoundError, req.ID, errors.New("Request not found"))
		return response, nil
	}

	reply, errCall := s.ServiceMap.Call(req, r)
	if errCall != nil {
		return jsonrpc2.Response{}, errors.Wrap(errCall, "Failed Execute")
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

func responseError(code jsonrpc2.ErrorCode, id string, e error) jsonrpc2.Response {
	rpcError := &jsonrpc2.Error{
		Code:    code,
		Message: jsonrpc2.ErrorMessage[code],
	}
	if e != nil && os.Getenv("RISA_RETURN_STACKTRACE") != "" {
		rpcError.Data = map[string]string{
			"stacktrace":fmt.Sprintf("%#+v", e),
		}
	}
	response := jsonrpc2.Response{
		JsonRPC: jsonrpc2.Version,
		Error: rpcError,
		ID: null.NewString(id, id != ""),
	}
	return response
}

func writeResponse(w http.ResponseWriter, response interface{}) error {
	jsonStr, err := json.Marshal(response)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	_, errPrint := fmt.Fprint(w, string(jsonStr))
	return errPrint
}
