package rpc

import (
	"reflect"
	"sync"
	"fmt"
	"unicode/utf8"
	"unicode"
	"net/http"
	"strings"
	"errors"
)

var (
	typeOfError = reflect.TypeOf((*error)(nil)).Elem()
	typeOfRequest = reflect.TypeOf((*http.Request)(nil)).Elem()
)

type service struct {
	name    string
	rcvr    reflect.Value
	typ     reflect.Type
	methods map[string]*methodType
	passReq bool
}

type methodType struct {
	method    reflect.Method
	ArgsType  reflect.Type
	ReplyType reflect.Type
}

type serviceMap struct {
	mu       sync.RWMutex
	services map[string]*service
}

func (serviceMap *serviceMap) register(rcvr interface{}) error {
	serviceMap.mu.Lock()
	defer serviceMap.mu.Unlock()

	if rcvr == nil {
		return errors.New("rcvr is nil")
	}

	name := reflect.TypeOf(rcvr).Elem().Name()
	if name == "" {
		return fmt.Errorf("rpc: no service name for type %q", name)
	}

	s := &service{
		name:     name,
		rcvr:     reflect.ValueOf(rcvr),
		typ: reflect.TypeOf(rcvr),
		methods:  make(map[string]*methodType),
	}

	for i := 0; i < s.typ.NumMethod(); i++ {
		method := s.typ.Method(i)
		mtype := method.Type

		// Method must be exported.
		if method.PkgPath != "" {
			continue
		}

		offset := 1

		// Method needs four ins: receiver, *http.Request, *args, *reply
		if mtype.NumIn() != 3 + offset {
			return errors.New("need four ins")
		}

		// First argument must be  *http.Request
		reqType := mtype.In(1)
		if reqType.Kind() != reflect.Ptr || reqType.Elem() != typeOfRequest {
			return errors.New("First argument must be http.Request")
		}
		// Second argument must be *args and must be exported
		args := mtype.In(1 + offset)
		if args.Kind() != reflect.Ptr || !isExportedOrBuiltin(args) {
			return errors.New("Second argument must be Pointer")
		}
		// Third argument must be *reply and must be exported
		reply := mtype.In(2 + offset)
		if reply.Kind() != reflect.Ptr || !isExportedOrBuiltin(reply) {
			return errors.New("Third argument must be pointer")
		}

		// Method return value must be a error
		if mtype.NumOut() != 1 {
			return errors.New("Method return value must be a error")
		}
		if returnType := mtype.Out(0); returnType != typeOfError {
			return errors.New("error type")
		}
		s.methods[method.Name] = &methodType{
			method:    method,
			ArgsType:  args.Elem(),
			ReplyType: reply.Elem(),
		}
	}

	if serviceMap.services == nil {
		serviceMap.services = make(map[string]*service)
	} else if _, ok := serviceMap.services[s.name]; ok {
		return fmt.Errorf("rpc: service already defined: %q", s.name)
	}
	serviceMap.services[s.name] = s

	return nil
}

func (serviceMap *serviceMap) Get(method string) (*service, *methodType, error) {
	parts := strings.Split(method, ".")
	if len(parts) != 2 {
		return nil, nil, fmt.Errorf("invalid request: %q", method)
	}
	serviceMap.mu.Lock()
	service := serviceMap.services[parts[0]]
	serviceMap.mu.Unlock()
	if service == nil {
		err := fmt.Errorf("service not found: %q", method)
		return nil, nil, err
	}
	serviceMethod := service.methods[parts[1]]
	if serviceMethod == nil {
		err := fmt.Errorf("method not found: %q", method)
		return nil, nil, err
	}
	return service, serviceMethod, nil
}

func (m *serviceMap) HasMethod(method string) bool {
	if _, _, err := m.Get(method); err == nil {
		return true
	}
	return false
}

func isExported(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r)
}

func isExportedOrBuiltin(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return isExported(t.Name()) || t.PkgPath() == ""
}