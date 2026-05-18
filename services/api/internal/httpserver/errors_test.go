package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteErrorUsesConsistentShape(t *testing.T) {
	response := httptest.NewRecorder()

	writeError(response, NewAPIError(http.StatusBadRequest, "invalid_request", "title is required"))

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}

	var body errorResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.Error.Code != "invalid_request" {
		t.Fatalf("expected error code, got %q", body.Error.Code)
	}
	if body.Error.Message != "title is required" {
		t.Fatalf("expected error message, got %q", body.Error.Message)
	}
}
