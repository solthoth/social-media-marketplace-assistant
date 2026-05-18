package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ServerSuite struct {
	suite.Suite
}

func TestServerSuite(t *testing.T) {
	suite.Run(t, new(ServerSuite))
}

func (s *ServerSuite) TestHealthz() {
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()

	NewRouter().ServeHTTP(response, request)

	s.Equal(http.StatusOK, response.Code)

	var body healthResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		s.T().Fatalf("decode response: %v", err)
	}

	s.Equal("ok", body.Status)
	s.NotEmpty(body.Service)
}
