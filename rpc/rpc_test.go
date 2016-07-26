package rpc

import (
	"testing"
	"net/http"
)

type noExportType struct{}

type ExportArgs struct{}
type ExportReply struct{}
type ExportType struct{}

func (e *ExportType) Get(r *http.Request, arg *ExportArgs, reply *ExportReply) error {
	return nil
}

func TestRegister(t *testing.T) {
	s := serviceMap{}
	var err error

	if err = s.register(nil, "", true); err == nil {
		t.Error(err)
	}

	if err = s.register(new(noExportType), "", true); err == nil {
		t.Error(err)
	}

	if err = s.register(new(noExportType), "Name", true); err == nil {
		t.Error(err)
	}

	if err = s.register(new(ExportType), "", true); err != nil {
		t.Error(err)
	}
}