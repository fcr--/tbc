package terabox

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

// MockRoundTripper implements http.RoundTripper
type MockRoundTripper struct {
	RoundTripFunc func(req *http.Request) *http.Response
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.RoundTripFunc(req), nil
}

func TestNewClient(t *testing.T) {
	mockClient := &http.Client{
		Transport: &MockRoundTripper{
			RoundTripFunc: func(req *http.Request) *http.Response {
				if req.URL.Path == "/main" {
					return &http.Response{
						StatusCode: 200,
						Body: io.NopCloser(bytes.NewBufferString(
							`"bdstoken":"mock_bds_token" window.jsToken%20%3D%20a%7D%3Bfn%28%22mock_js_token`,
						)),
						Header: make(http.Header),
					}
				}
				return &http.Response{StatusCode: 404, Body: io.NopCloser(bytes.NewBufferString(""))}
			},
		},
	}
	client, err := NewClient("test_cookie", mockClient)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	if client == nil {
		t.Fatal("Client is nil")
	}
}

func TestGetHomeInfo(t *testing.T) {
	// Mock response body
	mockResponse := `{"errno":0,"errmsg":"","data":{"username":"testuser","uk":"12345","sign1":"s1","timestamp":1234567890,"sign3":"s3"}}`

	// Create a mock HTTP client
	mockClient := &http.Client{
		Transport: &MockRoundTripper{
			RoundTripFunc: func(req *http.Request) *http.Response {
				if req.URL.Path == "/main" {
					return &http.Response{
						StatusCode: 200,
						Body: io.NopCloser(bytes.NewBufferString(
							`"bdstoken":"mock_bds_token" window.jsToken%20%3D%20a%7D%3Bfn%28%22mock_js_token`,
						)),
						Header: make(http.Header),
					}
				}
				// Verify request URL
				if req.URL.Path != "/api/home/info" {
					t.Errorf("Expected path /api/home/info, got %s", req.URL.Path)
				}
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(mockResponse)),
					Header:     make(http.Header),
				}
			},
		},
	}

	client, err := NewClient("test_cookie", mockClient)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	// Inject tokens manually since we are not testing login flow here
	client.jsToken = "test_js_token"

	info, err := client.getHomeInfo()
	if err != nil {
		t.Fatalf("getHomeInfo failed: %v", err)
	}

	if info.Errno != 0 {
		t.Errorf("Expected errno 0, got %d", info.Errno)
	}
	if info.Data.Username != "testuser" {
		t.Errorf("Expected username testuser, got %s", info.Data.Username)
	}
}
