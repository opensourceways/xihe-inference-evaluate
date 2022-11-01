package controller

const (
	errorSystemError     = "system_error"
	errorBadRequestBody  = "bad_request_body"
	errorBadRequestParam = "bad_request_param"
)

var (
	respBadRequestBody = newResponseCodeMsg(
		errorBadRequestBody, "can't fetch request body",
	)
)

// responseData is the response data to client
type responseData struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func newResponseError(err error) responseData {
	code := errorSystemError

	return responseData{
		Code: code,
		Msg:  err.Error(),
	}
}

func newResponseData(data interface{}) responseData {
	return responseData{
		Data: data,
	}
}

func newResponseCodeError(code string, err error) responseData {
	return responseData{
		Code: code,
		Msg:  err.Error(),
	}
}

func newResponseCodeMsg(code, msg string) responseData {
	return responseData{
		Code: code,
		Msg:  msg,
	}
}
