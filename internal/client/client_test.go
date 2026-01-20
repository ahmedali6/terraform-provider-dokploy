package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewDokployClient(t *testing.T) {
	baseURL := "https://test.example.com"
	apiKey := "test-api-key"

	client := NewDokployClient(baseURL, apiKey)

	if client.BaseURL != baseURL {
		t.Errorf("Expected BaseURL %s, got %s", baseURL, client.BaseURL)
	}

	if client.APIKey != apiKey {
		t.Errorf("Expected APIKey %s, got %s", apiKey, client.APIKey)
	}

	if client.HTTPClient == nil {
		t.Error("Expected HTTPClient to be initialized")
	}

	if client.HTTPClient.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", client.HTTPClient.Timeout)
	}
}

func TestDoRequestSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("Expected x-api-key test-key, got %s", r.Header.Get("x-api-key"))
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id": "123", "name": "test"}`))
	}))
	defer server.Close()

	client := NewDokployClient(server.URL, "test-key")
	resp, err := client.doRequest("POST", "test/endpoint", map[string]string{"key": "value"})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if string(resp) != `{"id": "123", "name": "test"}` {
		t.Errorf("Expected response body, got %s", string(resp))
	}
}

func TestDoRequestGET(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": "test"}`))
	}))
	defer server.Close()

	client := NewDokployClient(server.URL, "test-key")
	resp, err := client.doRequest("GET", "test/path", nil)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if string(resp) != `{"data": "test"}` {
		t.Errorf("Expected response body, got %s", string(resp))
	}
}

func TestDoRequest404Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "not found"}`))
	}))
	defer server.Close()

	client := NewDokployClient(server.URL, "test-key")
	_, err := client.doRequest("GET", "test/path", nil)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if err == nil || err.Error() == "" {
		t.Fatal("Expected error, got nil or empty error")
	}
}

func TestDoRequest400Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "bad request"}`))
	}))
	defer server.Close()

	client := NewDokployClient(server.URL, "test-key")
	_, err := client.doRequest("POST", "test/path", nil)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expectedError := "API error: 400 Bad Request"
	if err.Error()[:len(expectedError)] != expectedError {
		t.Errorf("Expected error to start with %s, got %v", expectedError, err)
	}
}

func TestDoRequest500Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	client := NewDokployClient(server.URL, "test-key")
	_, err := client.doRequest("GET", "test/path", nil)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestDoRequestInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewDokployClient(server.URL, "test-key")
	resp, err := client.doRequest("GET", "test/path", nil)

	if err != nil {
		t.Fatalf("Expected no error for invalid JSON response, got %v", err)
	}

	if string(resp) != "invalid json" {
		t.Errorf("Expected raw response, got %s", string(resp))
	}
}

func TestDoRequestNetworkError(t *testing.T) {
	client := NewDokployClient("http://invalid-url-that-does-not-exist.local", "test-key")
	_, err := client.doRequest("GET", "test/path", nil)

	if err == nil {
		t.Fatal("Expected network error, got nil")
	}
}

type MockRoundTripper struct {
	Response *http.Response
	Err      error
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.Response, m.Err
}

func TestDoRequestWithCustomClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	client := NewDokployClient(server.URL, "test-key")
	customClient := &http.Client{Timeout: 5 * time.Second}
	client.HTTPClient = customClient

	resp, err := client.doRequest("GET", "test", nil)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if string(resp) != `{"success": true}` {
		t.Errorf("Unexpected response: %s", string(resp))
	}
}

func TestDoRequestWithNilBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if len(body) > 0 {
			t.Errorf("Expected empty body, got %s", string(body))
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	client := NewDokployClient(server.URL, "test-key")
	resp, err := client.doRequest("POST", "test/path", nil)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if string(resp) != `{"status": "ok"}` {
		t.Errorf("Expected response, got %s", string(resp))
	}
}

func TestDoRequestBodyMarshalError(t *testing.T) {
	type InvalidType struct {
		Func func() string
	}

	client := NewDokployClient("http://test.local", "test-key")
	_, err := client.doRequest("POST", "test", InvalidType{Func: func() string { return "test" }})

	if err == nil {
		t.Fatal("Expected error for unmarshalable type, got nil")
	}
}

func TestDoRequestWithComplexBody(t *testing.T) {
	complexBody := map[string]interface{}{
		"string":  "value",
		"number":  123,
		"boolean": true,
		"nested": map[string]string{
			"key": "value",
		},
		"array": []int{1, 2, 3},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var received map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&received)

		if received["string"] != "value" {
			t.Errorf("Expected string value, got %v", received["string"])
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"created": true}`))
	}))
	defer server.Close()

	client := NewDokployClient(server.URL, "test-key")
	resp, err := client.doRequest("POST", "test/complex", complexBody)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if string(resp) != `{"created": true}` {
		t.Errorf("Expected response, got %s", string(resp))
	}
}

func TestDoRequestEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
		_, _ = w.Write([]byte{})
	}))
	defer server.Close()

	client := NewDokployClient(server.URL, "test-key")
	resp, err := client.doRequest("DELETE", "test/delete", nil)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(resp) != 0 {
		t.Errorf("Expected empty response, got %s", string(resp))
	}
}

func TestParseEnv(t *testing.T) {
	tests := []struct {
		name     string
		env      string
		expected map[string]string
	}{
		{
			name:     "empty string",
			env:      "",
			expected: map[string]string{},
		},
		{
			name: "single variable",
			env:  "KEY=value",
			expected: map[string]string{
				"KEY": "value",
			},
		},
		{
			name: "multiple variables",
			env:  "KEY1=value1\nKEY2=value2\nKEY3=value3",
			expected: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
				"KEY3": "value3",
			},
		},
		{
			name: "with comments",
			env:  "# This is a comment\nKEY=value\n# Another comment\nKEY2=value2",
			expected: map[string]string{
				"KEY":  "value",
				"KEY2": "value2",
			},
		},
		{
			name: "with empty lines",
			env:  "KEY1=value1\n\n\nKEY2=value2",
			expected: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
		},
		{
			name: "with spaces around lines",
			env:  "  KEY1=value1  \n  KEY2=value2  ",
			expected: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
		},
		{
			name: "value with equals sign",
			env:  "KEY=value=with=equals",
			expected: map[string]string{
				"KEY": "value=with=equals",
			},
		},
		{
			name: "complex values",
			env:  "DB_HOST=localhost\nDB_PORT=5432\nDB_NAME=mydb\nDB_USER=admin\nDB_PASS=secret",
			expected: map[string]string{
				"DB_HOST": "localhost",
				"DB_PORT": "5432",
				"DB_NAME": "mydb",
				"DB_USER": "admin",
				"DB_PASS": "secret",
			},
		},
		{
			name: "invalid lines without equals",
			env:  "KEY1=value1\nINVALID_LINE\nKEY2=value2",
			expected: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseEnv(tt.env)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d entries, got %d", len(tt.expected), len(result))
			}
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("Expected %s=%s, got %s=%s", k, v, k, result[k])
				}
			}
		})
	}
}

func TestFormatEnv(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]string
		expected string
	}{
		{
			name:     "empty map",
			env:      map[string]string{},
			expected: "",
		},
		{
			name: "single variable",
			env: map[string]string{
				"KEY": "value",
			},
			expected: "KEY=value",
		},
		{
			name: "multiple variables",
			env: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
				"KEY3": "value3",
			},
			expected: "KEY1=value1\nKEY2=value2\nKEY3=value3",
		},
		{
			name: "complex values",
			env: map[string]string{
				"DB_HOST": "localhost",
				"DB_PORT": "5432",
			},
			expected: "DB_HOST=localhost\nDB_PORT=5432",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatEnv(tt.env)
			// Parse both to compare maps since order may vary
			resultMap := ParseEnv(result)
			expectedMap := ParseEnv(tt.expected)

			if len(resultMap) != len(expectedMap) {
				t.Errorf("Expected %d entries, got %d", len(expectedMap), len(resultMap))
			}

			for k, v := range expectedMap {
				if resultMap[k] != v {
					t.Errorf("Expected %s=%s, got %s=%s", k, v, k, resultMap[k])
				}
			}
		})
	}
}

func TestParseEnvFormatEnvRoundtrip(t *testing.T) {
	original := `KEY1=value1
KEY2=value2
# This is a comment
KEY3=value3`

	parsed := ParseEnv(original)
	formatted := formatEnv(parsed)

	// After parsing and formatting, comments should be removed
	// and the order might be different, so we just check that
	// the content is the same when parsed again
	reparsed := ParseEnv(formatted)

	if len(reparsed) != 3 {
		t.Errorf("Expected 3 entries after roundtrip, got %d", len(reparsed))
	}

	expectedValues := map[string]string{
		"KEY1": "value1",
		"KEY2": "value2",
		"KEY3": "value3",
	}

	for k, v := range expectedValues {
		if reparsed[k] != v {
			t.Errorf("Expected %s=%s, got %s=%s", k, v, k, reparsed[k])
		}
	}
}
