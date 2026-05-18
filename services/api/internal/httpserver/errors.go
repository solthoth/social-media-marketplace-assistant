package httpserver

import "github.com/gin-gonic/gin"

type APIError struct {
	Status  int
	Code    string
	Message string
}

type errorResponse struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewAPIError(status int, code string, message string) APIError {
	return APIError{
		Status:  status,
		Code:    code,
		Message: message,
	}
}

func writeError(c *gin.Context, err APIError) {
	c.JSON(err.Status, errorResponse{
		Error: errorDetail{
			Code:    err.Code,
			Message: err.Message,
		},
	})
}
