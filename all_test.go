package risa

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"strings"
	//"fmt"
	"io/ioutil"
	//"fmt"
	//"encoding/json"
	//"hoshina85/risa/jsonrpc2"
	"github.com/pkg/errors"
)

type HelloArgs struct {
	Name string
}
type HelloReply struct {
	Message string
}
type HelloService struct {
	Text string
}

func (s *HelloService) Get(t *http.Request, args *HelloArgs, reply *HelloReply) error {
	reply.Message = "Hello " + args.Name + "!"
	return nil
}

func (s *HelloService) Error(t *http.Request, args *HelloArgs, reply *HelloReply) error {
	return errors.New("return error")
}

func TestNewJsonRPCServer(t *testing.T) {
	rpcHandler := NewJsonRPCServer()
	errRegister := rpcHandler.Register(new(HelloService))
	if errRegister != nil {
		t.Error(errRegister.Error())
	}
	s := httptest.NewServer(rpcHandler)
	defer s.Close()

	payload := `[
	{"jsonrpc":"2.0","method":"HelloService.Get","params":[{"Name":"John"}], "id":"1"},
	{"jsonrpc":"2.0","method":"HelloService.Get","params":[{"Name":"John"}], "id":"2"}
	]`
	resp, _ := http.Post(s.URL, "application/json", strings.NewReader(payload))
	if resp.StatusCode != 200 {
		t.Error()
	}
	body, _ := ioutil.ReadAll(resp.Body);
	if `[{"jsonrpc":"2.0","result":{"Message":"Hello John!"},"id":"1"},{"jsonrpc":"2.0","result":{"Message":"Hello John!"},"id":"2"}]` != string(body) {
		t.Errorf("Data Error. %s", string(body))
	}

}

func TestBatchRequest(t *testing.T) {
	rpcHandler := NewJsonRPCServer()
	errRegister := rpcHandler.Register(new(HelloService))
	if errRegister != nil {
		t.Error(errRegister.Error())
	}
	s := httptest.NewServer(rpcHandler)
	defer s.Close()
	payload := `[{"jsonrpc":"2.0","method":"HelloService.Get","params":[{"Name":"John"}], "id":"1"},{"jsonrpc":"2.0","method":"HelloService.Get","params":[{"Name":"Sam"}], "id":"2"}]`
	req1, _ := http.NewRequest("POST", s.URL, strings.NewReader(payload))
	client := &http.Client{}
	resp, errDo := client.Do(req1)
	if errDo != nil {
		t.Error(errDo.Error())
	}
	if resp.StatusCode != 200 {
		t.Error()
	}
	body, _ := ioutil.ReadAll(resp.Body);
	if `[{"jsonrpc":"2.0","result":{"Message":"Hello John!"},"id":"1"},{"jsonrpc":"2.0","result":{"Message":"Hello Sam!"},"id":"2"}]` != string(body) {
		t.Errorf("Data Error. %s", string(body))
	}
}

func TestFailRequest(t *testing.T) {
	rpcHandler := NewJsonRPCServer()
	errRegister := rpcHandler.Register(new(HelloService))
	if errRegister != nil {
		t.Error(errRegister.Error())
	}
	s := httptest.NewServer(rpcHandler)
	defer s.Close()

	payload := `{"jsonrpc": "2.0", "method": "foobar", "id": "1"}`
	resp, _ := http.Post(s.URL, "application/json", strings.NewReader(payload))
	body, _ := ioutil.ReadAll(resp.Body);
	if `{"jsonrpc":"2.0","error":{"code":-32601,"message":"Method not found"},"id":"1"}` != string(body) {
		t.Errorf("Data Error. %s", string(body))
	}

	payload2 := `{"jsonrpc": "2.0", "method": "foobar, "params": "bar", "baz]`
	resp2, _ := http.Post(s.URL, "application/json", strings.NewReader(payload2))
	body2, _ := ioutil.ReadAll(resp2.Body);
	if `{"jsonrpc":"2.0","error":{"code":-32700,"message":"Parse error"},"id":null}` != string(body2) {
		t.Errorf("Data Error. %s", string(body2))
	}
}

func TestFailBatchRequest(t *testing.T) {
	rpcHandler := NewJsonRPCServer()
	errRegister := rpcHandler.Register(new(HelloService))
	if errRegister != nil {
		t.Error(errRegister.Error())
	}
	s := httptest.NewServer(rpcHandler)
	defer s.Close()

	payload := `[
  {"jsonrpc": "2.0", "method": "sum", "params": [1,2,4], "id": "1"},
  {"jsonrpc": "2.0", "method"
]`
	resp, _ := http.Post(s.URL, "application/json", strings.NewReader(payload))
	body, _ := ioutil.ReadAll(resp.Body);
	if `{"jsonrpc":"2.0","error":{"code":-32700,"message":"Parse error"},"id":null}` != string(body) {
		t.Errorf("Data Error. %s", string(body))
	}

	payload2 := `[]`
	resp2, _ := http.Post(s.URL, "application/json", strings.NewReader(payload2))
	body2, _ := ioutil.ReadAll(resp2.Body);
	if `{"jsonrpc":"2.0","error":{"code":-32600,"message":"Invalid request"},"id":null}` != string(body2) {
		t.Errorf("Data Error. %s", string(body2))
	}

}

func TestServerError(t *testing.T) {
	rpcHandler := NewJsonRPCServer()
	errRegister := rpcHandler.Register(new(HelloService))
	if errRegister != nil {
		t.Error(errRegister.Error())
	}
	s := httptest.NewServer(rpcHandler)
	defer s.Close()
	payload := `{"jsonrpc":"2.0","method":"HelloService.Error","params":[{"Name":"John"}], "id":"1"}`
	req1, _ := http.NewRequest("POST", s.URL, strings.NewReader(payload))
	client := &http.Client{}
	resp, errDo := client.Do(req1)
	if errDo != nil {
		t.Error(errDo.Error())
	}
	if resp.StatusCode != 200 {
		t.Error()
	}
	body, _ := ioutil.ReadAll(resp.Body);
	if `{"jsonrpc":"2.0","error":{"code":-32603,"message":"Internal error"},"id":"1"}` != string(body) {
		t.Errorf("Data Error. %s", string(body))
	}
}