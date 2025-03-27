package multiplexer

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchHandler_ValidRequest(t *testing.T) {
	m := NewMultiplexer(Options{})

	reqBody, _ := json.Marshal(fetchRequest{
		URLs: []string{"https://example.com"},
	})

	req, err := http.NewRequest(http.MethodPost, "/fetch", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(m.FetchHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

}

func TestFetchHandler_TooManyURLs(t *testing.T) {
	m := NewMultiplexer(Options{})

	urls := make([]string, 21)
	for i := range urls {
		urls[i] = "https://example.com"
	}
	reqBody, _ := json.Marshal(fetchRequest{URLs: urls})

	req, _ := http.NewRequest(http.MethodPost, "/fetch", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(m.FetchHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

}
