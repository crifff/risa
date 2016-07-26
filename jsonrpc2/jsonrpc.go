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

type Params map[string]interface{}

type Request struct {
	JsonRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  Params `json:"params"`
	ID      string `json:"id"`
}

type BatchRequest struct {
	Requests []Request
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    interface{} `json:"data"`
}

type Response struct {
	JsonRPC string  `json:"jsonrpc"`
	Result  string  `json:"result"`
	Error   *Error  `json:"error"`
	ID      string  `json:"id"`
}
