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
	reply.Message = s.Text
	return nil
}

func (s *HelloService) Set(t *http.Request, args *HelloArgs, reply *HelloReply) error {
	s.Text = "Hello " + args.Name + "!"
	return nil
}

func TestNewJsonRPCServer(t *testing.T) {
	rpcHandler := NewJsonRPCServer()
	errRegister := rpcHandler.Register(new(HelloService))
	if errRegister != nil {
		t.Error(errRegister.Error())
	}
	s := httptest.NewServer(rpcHandler)
	defer s.Close()
	payload := `{"jsonrpc":"2.0","method":"HelloService.Get","params":[{"Name":"John"}], "id":1}`
	req1, _ := http.NewRequest("POST", s.URL, strings.NewReader(payload))
	client := &http.Client{}
	resp, errDo := client.Do(req1)
	//fmt.Println(resp.Body)
	if errDo != nil {
		t.Error(errDo.Error())
	}
	if resp.StatusCode != 200 {
		t.Error()
	}
	body, _ := ioutil.ReadAll(resp.Body);
	if `{"jsonrpc":"2.0","result":"{\"Message\":\"\"}","error":null,"id":""}` != string(body) {
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
	payload := `[{"jsonrpc":"2.0","method":"HelloService.Get","params":[{"Name":"John"}], "id":1},{"jsonrpc":"2.0","method":"HelloService.Get","params":[{"Name":"John"}], "id":2}]`
	req1, _ := http.NewRequest("POST", s.URL, strings.NewReader(payload))
	client := &http.Client{}
	resp, errDo := client.Do(req1)
	//fmt.Println(resp.Body)
	if errDo != nil {
		t.Error(errDo.Error())
	}
	if resp.StatusCode != 200 {
		t.Error()
	}
	body, _ := ioutil.ReadAll(resp.Body);
	if `[{"jsonrpc":"2.0","result":"{\"Message\":\"\"}","error":null,"id":""},{"jsonrpc":"2.0","result":"{\"Message\":\"\"}","error":null,"id":""}]` != string(body) {
		t.Errorf("Data Error. %s", string(body))
	}
}
