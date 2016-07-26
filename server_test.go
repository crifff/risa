package risa

import (
	"testing"
	"net/http"
	"net/http/httptest"
	//"bytes"
	"strings"
	//"fmt"
)

func post(url string, body string) *http.Response {
	req, _ := http.NewRequest("POST", url, strings.NewReader(body))
	client := &http.Client{}
	resp, _ := client.Do(req)
	return resp
}

func get(url string) *http.Response {
	req, _ := http.NewRequest("GET", url, nil)
	client := &http.Client{}
	resp, _ := client.Do(req)
	return resp
}

func TestServeHTTP(t *testing.T) {
	s := httptest.NewServer(NewJsonRPCServer())
	defer s.Close()
	res1 := get(s.URL)
	if res1.StatusCode != http.StatusMethodNotAllowed {
		t.Error("Status code error")
	}

	var resp *http.Response
	resp = post(s.URL, `{"jsonrpc":"2.0"}`)
	if resp.StatusCode != 200 {
		t.Error()
	}

	resp = post(s.URL, `[{"jsonrpc":"2.0"}, {"jsonrpc":"2.0"}]`)
	if resp.StatusCode != 200 {
		t.Error()
	}

}
