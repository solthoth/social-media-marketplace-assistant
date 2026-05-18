package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

type SwaggerSuite struct {
	suite.Suite
	router http.Handler
}

func TestSwaggerSuite(t *testing.T) {
	suite.Run(t, new(SwaggerSuite))
}

func (s *SwaggerSuite) SetupTest() {
	s.router = NewRouter()
}

func (s *SwaggerSuite) TestOpenAPIDocumentIncludesCoreRoutes() {
	response := s.request(http.MethodGet, "/swagger/doc.json")

	s.Equal(http.StatusOK, response.Code)

	var body openAPIDocument
	s.Require().NoError(json.NewDecoder(response.Body).Decode(&body))
	s.Equal("3.0.3", body.OpenAPI)
	s.Equal("Social Media Marketplace Assistant API", body.Info.Title)
	s.Contains(body.Paths, "/healthz")
	s.Contains(body.Paths, "/items")
	s.Contains(body.Paths, "/items/{id}")
	itemResponse := body.Components.Schemas["ItemResponse"].(map[string]any)
	properties := itemResponse["properties"].(map[string]any)
	s.Contains(properties, "original_purchase_price_cents")
	s.Contains(properties, "selling_price_cents")
}

func (s *SwaggerSuite) TestSwaggerUIIsServed() {
	response := s.request(http.MethodGet, "/swagger/index.html")

	s.Equal(http.StatusOK, response.Code)
	s.Contains(response.Body.String(), "/swagger/doc.json")
}

func (s *SwaggerSuite) TestSwaggerRouteRedirectsToIndex() {
	response := s.request(http.MethodGet, "/swagger")

	s.Equal(http.StatusMovedPermanently, response.Code)
	s.Equal("/swagger/index.html", response.Header().Get("Location"))
}

func (s *SwaggerSuite) request(method string, target string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, target, nil)
	response := httptest.NewRecorder()
	s.router.ServeHTTP(response, request)
	return response
}
