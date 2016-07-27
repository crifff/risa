package jsonrpc2

import (
	"testing"
	"gopkg.in/guregu/null.v3"
	"encoding/json"
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

	payload := `{"jsonrpc": "2.0", "method": "subtract", "params": [42,23], "id": "2"}`
	request = &Request{}
	json.Unmarshal([]byte(payload), request)
	if err = ValidateRequest(*request); err != nil {
		t.Error()
	}

}

func TestJsonUnMarshal(t *testing.T) {

	r := Response{
		ID: null.NewString("", false),
	}

	json, _ := json.Marshal(r)

	if `{"jsonrpc":"","id":null}` != string(json) {
		t.Error()
	}
}
