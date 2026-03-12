package response

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

const (
	CodeSuccess      = 0
	CodeBadRequest   = 400
	CodeUnauthorized = 401
	CodeForbidden    = 403
	CodeNotFound     = 404
	CodeServerError  = 500
)

func Success(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

func SuccessMsg(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Code:    CodeSuccess,
		Message: message,
	})
}

func Error(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(getHTTPStatus(code))
	json.NewEncoder(w).Encode(ErrorResponse{
		Code:    code,
		Message: message,
	})
}

func ErrorMsg(w http.ResponseWriter, message string) {
	Error(w, CodeServerError, message)
}

func BadRequest(w http.ResponseWriter, message string) {
	Error(w, CodeBadRequest, message)
}

func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, CodeUnauthorized, message)
}

func NotFound(w http.ResponseWriter, message string) {
	Error(w, CodeNotFound, message)
}

func ServerError(w http.ResponseWriter, message string) {
	Error(w, CodeServerError, message)
}

func getHTTPStatus(code int) int {
	switch code {
	case CodeBadRequest:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
