package mocks

import (
	"net/http"
	"net/http/httptest"
)

// NewMockServer creates a test HTTP server that returns the given response
func NewMockServer(response string, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write([]byte(response))
	}))
}

// NewMockServerWithHandler creates a test HTTP server with custom handler
func NewMockServerWithHandler(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}