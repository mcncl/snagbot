package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcncl/snagbot/internal/api"
	"github.com/mcncl/snagbot/internal/config"
	"github.com/stretchr/testify/assert"
)

// TestHealthCheckEndpoint tests the health check endpoint
func TestHealthCheckEndpoint(t *testing.T) {
	// Create a test config
	cfg := config.New()

	// Create handler with the simple router
	handler := api.SetupSimpleRouter(cfg)

	// Create a test server
	server := httptest.NewServer(handler)
	defer server.Close()

	// Make a request to the health check endpoint
	resp, err := http.Get(server.URL + "/health")
	assert.NoError(t, err)

	// Check response status code
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Parse response body
	var response api.Response
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)

	// Check response content
	assert.Equal(t, "Service is healthy", response.Message)
	assert.Equal(t, "OK", response.Status)
}

// TestHelloWorldEndpoint tests the hello world endpoint
func TestHelloWorldEndpoint(t *testing.T) {
	// Create a test config
	cfg := config.New()

	// Create handler with the simple router
	handler := api.SetupSimpleRouter(cfg)

	// Create a test server
	server := httptest.NewServer(handler)
	defer server.Close()

	// Make a request to the hello world endpoint
	resp, err := http.Get(server.URL + "/hello")
	assert.NoError(t, err)

	// Check response status code
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Parse response body
	var response api.Response
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)

	// Check response content
	assert.Equal(t, "Hello, world! SnagBot is running.", response.Message)
	assert.Equal(t, "OK", response.Status)
}

// TestNonExistentEndpoint tests accessing a non-existent endpoint
func TestNonExistentEndpoint(t *testing.T) {
	// Create a test config
	cfg := config.New()

	// Create handler with the simple router
	handler := api.SetupSimpleRouter(cfg)

	// Create a test server
	server := httptest.NewServer(handler)
	defer server.Close()

	// Make a request to a non-existent endpoint
	resp, err := http.Get(server.URL + "/non-existent")
	assert.NoError(t, err)

	// Check response status code - should be 404 Not Found
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestMethodNotAllowed tests using an incorrect HTTP method
func TestMethodNotAllowed(t *testing.T) {
	// Create a test config
	cfg := config.New()

	// Create handler with the simple router
	handler := api.SetupSimpleRouter(cfg)

	// Create a test server
	server := httptest.NewServer(handler)
	defer server.Close()

	// Make a POST request to the health endpoint, which only accepts GET
	resp, err := http.Post(server.URL+"/health", "application/json", nil)
	assert.NoError(t, err)

	// Check response status code - should be 405 Method Not Allowed
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}
