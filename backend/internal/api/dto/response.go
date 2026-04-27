package dto

// Response 统一响应格式
type Response struct {
	BaseResponse `json:",inline"`
	Data         interface{} `json:"data,omitempty"`
}

// BaseResponse 基础响应格式
type BaseResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse 错误响应格式
type ErrorResponse struct {
	BaseResponse `json:",inline"`
	Details      string `json:"details,omitempty"`
}

// Success 成功响应
func Success(data interface{}) *Response {
	return &Response{
		BaseResponse: BaseResponse{
			Code:    CodeSuccess,
			Message: "success",
		},
		Data: data,
	}
}

// Error 错误响应
func Error(code int, message string) *ErrorResponse {
	return &ErrorResponse{
		BaseResponse: BaseResponse{
			Code:    code,
			Message: message,
		},
	}
}

// ErrorWithDetails 带详情的错误响应
func ErrorWithDetails(code int, message string, details string) *ErrorResponse {
	return &ErrorResponse{
		BaseResponse: BaseResponse{
			Code:    code,
			Message: message,
		},
		Details: details,
	}
}
