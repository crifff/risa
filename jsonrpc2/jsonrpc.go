package jsonrpc2

const Version = "2.0"

type ErrorCode int

const (
	ParseError ErrorCode = -32700
	InvalidRequestError ErrorCode = -32600
	MethodNotFoundError ErrorCode = -32601
	InvalidParamsError ErrorCode = -32602
	InternalError ErrorCode = -32603
)

type Request struct {
	JsonRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  interface{} `json:"params"`
	ID      int `json:"id"`
}

type BatchRequest struct {
	Requests []Request
}

type Error struct {
	Code    ErrorCode    `json:"code"`
	Message string `json:"message"`
	Data    interface{} `json:"data"`
}

func (e Error) Error() string {
	return e.Message
}

type Response struct {
	JsonRPC string  `json:"jsonrpc"`
	Result  string  `json:"result"`
	Error   *Error  `json:"error"`
	ID      string  `json:"id"`
}

func ValidateRequest(r Request) error {
	//jsonrpc MUST be exactly "2.0"
	if version := r.JsonRPC; version != "2.0" {
		return Error{Code:InvalidRequestError, Message:"Invalid Request"}
	}
	return nil
}