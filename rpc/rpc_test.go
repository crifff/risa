package rpc

import (
	"testing"
	"net/http"
	"errors"
	"reflect"
)

type noExportType struct{}

type ExportArgs struct{ Name string }
type noExportArgs struct{}
type ExportReply struct{}
type noExportReply struct{}
type ExportType1 struct{}

func (e *ExportType1) Get(r *http.Request, arg *ExportArgs, reply *ExportReply) error {
	return errors.New("return by ExportType1.Get from " + arg.Name)
}

type ExportType2 struct{}

func (e *ExportType2) Get(http.Request, ExportArgs, ExportReply) error {
	return nil
}

type ExportType3 struct{}

func (e *ExportType3) Get(*http.Request, ExportArgs, ExportReply) error {
	return nil
}

type ExportType4 struct{}

func (e *ExportType4) Get(*http.Request, *ExportArgs, ExportReply) error {
	return nil
}

type ExportType5 struct{}

func (e *ExportType5) Get(*http.Request, *noExportArgs, *ExportReply) error {
	return nil
}

type ExportType6 struct{}

func (e *ExportType6) Get(*http.Request, *ExportArgs, *noExportReply) error {
	return nil
}

func TestRegister(t *testing.T) {
	s := ServiceMap{}
	var err error

	if err = s.Register(nil); err == nil {
		t.Error(err)
	}

	// no export Type id ignored
	if err = s.Register(new(noExportType)); err != nil {
		t.Error(err)
	}

	if err = s.Register(new(ExportType2)); err == nil {
		t.Error(err)
	}
	if err = s.Register(new(ExportType3)); err == nil {
		t.Error(err)
	}
	if err = s.Register(new(ExportType4)); err == nil {
		t.Error(err)
	}
	if err = s.Register(new(ExportType5)); err == nil {
		t.Error(err)
	}
	if err = s.Register(new(ExportType6)); err == nil {
		t.Error(err)
	}

	if err = s.Register(new(ExportType1)); err != nil {
		t.Error(err)
	}
	if s.HasMethod("ExportType1.Get") == false {
		t.Error()
	}
}

func TestGet(t *testing.T) {

	s := ServiceMap{}
	if err := s.Register(new(ExportType1)); err != nil {
		t.Error(err)
	}

	r, _ := http.NewRequest("POST", "", nil)
	var errValue []reflect.Value
	args := reflect.ValueOf(&ExportArgs{Name:"John"})
	reply := reflect.ValueOf(&ExportReply{})
	service, methodSpec, _ := s.Get("ExportType1.Get")
	errValue = methodSpec.method.Func.Call([]reflect.Value{
		service.rcvr,
		reflect.ValueOf(r),
		args,
		reply,
	})
	err := errValue[0].Interface().(error)
	if err.Error() != "return by ExportType1.Get from John" {
		t.Error(err)
	}
}