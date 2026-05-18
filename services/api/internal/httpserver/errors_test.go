package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type ErrorSuite struct {
	suite.Suite
}

func TestErrorSuite(t *testing.T) {
	suite.Run(t, new(ErrorSuite))
}

func (s *ErrorSuite) TestWriteErrorUsesConsistentShape() {
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)

	writeError(ctx, NewAPIError(http.StatusBadRequest, "invalid_request", "title is required"))

	s.Equal(http.StatusBadRequest, response.Code)

	var body errorResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		s.T().Fatalf("decode response: %v", err)
	}

	s.Equal("invalid_request", body.Error.Code)
	s.Equal("title is required", body.Error.Message)
}
