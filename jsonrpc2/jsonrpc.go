package jsonrpc2

import (
	null "gopkg.in/guregu/null.v3"
	"encoding/json"
)

const Version = "2.0"

type ErrorCode int

const (
	ParseError ErrorCode = -32700
	InvalidRequestError ErrorCode = -32600
	MethodNotFoundError ErrorCode = -32601
	InvalidParamsError ErrorCode = -32602
	InternalError ErrorCode = -32603
)

var ErrorMessage map[ErrorCode]string = map[ErrorCode]string{
	ParseError:"Parse error",
	InvalidRequestError:"Invalid request",
	MethodNotFoundError:"Method not found",
	InvalidParamsError:"Invalid params",
	InternalError:"Internal error",
}

type Request struct {
	JsonRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  *json.RawMessage `json:"params"`
	ID      string `json:"id"`
}

type BatchRequest struct {
	Requests []Request
}

type Error struct {
	Code    ErrorCode    `json:"code"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e Error) Error() string {
	return e.Message
}

type Response struct {
	JsonRPC string  `json:"jsonrpc"`
	Result  interface{}  `json:"result,omitempty"`
	Error   *Error  `json:"error,omitempty"`
	ID      null.String `json:"id"`
}

func ValidateRequest(r Request) error {
	//jsonrpc MUST be exactly "2.0"
	if version := r.JsonRPC; version != "2.0" {
		return Error{Code:InvalidRequestError, Message:"Invalid Request"}
	}
	return nil
}