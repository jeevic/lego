package util

const (
	SUCCESS             = 0
	ERROR               = -1
	AUTHENTICATION_FAIL = 40001
	IllEGAL_PARAMS      = 40002
)

type Response struct {
	Code      int32       `json:"code"`
	Msg       string      `json:"msg"`
	Data      interface{} `json:"data,omitempty"`
	RequestId string      `json:"request_id,omitempty"`
}

func Success(data interface{}, msg string, requestId string) *Response {
	if len(msg) <= 0 {
		msg = Msg(SUCCESS)
	}
	return BuildResponse(SUCCESS, msg, data, requestId)
}

func Error(data interface{}, msg string, requestId string) *Response {
	if len(msg) <= 0 {
		msg = Msg(ERROR)
	}
	return BuildResponse(ERROR, msg, data, requestId)
}

func BuildResponse(code int32, msg string, data interface{}, requestId string) *Response {
	resp := &Response{}
	resp.Code = code
	resp.Data = data
	resp.RequestId = requestId
	resp.Msg = msg
	return resp
}

func Msg(code int32) string {
	var msg = ""
	switch code {
	case SUCCESS:
		msg = "success"
	case ERROR:
		msg = "something error happened in server"
	case AUTHENTICATION_FAIL:
		msg = "authentication fail"
	case IllEGAL_PARAMS:
		msg = "illegal params"
	}
	return msg
}
