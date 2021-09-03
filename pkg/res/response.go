package res

type JsonResult struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func JsonSuccess() *JsonResult {
	return &JsonResult{
		Success: true,
		Code:    0,
		Message: "success",
		Data:    nil,
	}
}

func JsonData(data interface{}) *JsonResult {
	return &JsonResult{
		Success: true,
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

func JsonError(code int) *JsonResult {
	return &JsonResult{
		Code:    code,
		Message: "error",
		Data:    "",
		Success: false,
	}
}

func JsonErrorMessage(code int, message string) *JsonResult {
	return &JsonResult{
		Code:    code,
		Message: message,
		Data:    "",
		Success: false,
	}
}
