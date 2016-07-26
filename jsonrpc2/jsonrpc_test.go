package jsonrpc2

import (
	"testing"
)

func TestValidateRequest(t *testing.T) {
	var err error
	var request *Request

	request = &Request{}
	if err = ValidateRequest(*request); err == nil {
		t.Error()
	}

	request = &Request{JsonRPC:"1.0"}
	if err = ValidateRequest(*request); err == nil {
		t.Error()
	}

	request = &Request{JsonRPC:"2.0", Method:"subtract", Params: []int{1, 2}, ID: 1}
	if err = ValidateRequest(*request); err != nil {
		t.Error()
	}

}
