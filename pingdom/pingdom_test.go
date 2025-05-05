package pingdom

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	mux    *http.ServeMux
	client *Client
	server *httptest.Server
)

func setup() {
	// test server
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	// test client
	client, _ = NewClientWithConfig(ClientConfig{
		APIToken: "my_api_token",
	})

	url, _ := url.Parse(server.URL)
	client.BaseURL = url
}

func setupWithAPIKey() {
	// test server
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	// test client
	client, _ = NewClientWithConfig(ClientConfig{
		APIKey: "my_api_key",
	})

	url, _ := url.Parse(server.URL)
	client.BaseURL = url
}

func teardown() {
	server.Close()
}

func testMethod(t *testing.T, r *http.Request, want string) {
	assert.Equal(t, want, r.Method)
}

func TestNewClientWithConfig(t *testing.T) {
	// Test with API Token
	c, err := NewClientWithConfig(ClientConfig{
		APIToken: "token",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.DefaultClient, c.client)
	assert.Equal(t, defaultBaseURL, c.BaseURL.String())
	assert.NotNil(t, c.Checks)
	assert.Equal(t, "token", c.APIToken)
	assert.Equal(t, "", c.APIKey)

	// Test with API Key
	c, err = NewClientWithConfig(ClientConfig{
		APIKey: "key",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.DefaultClient, c.client)
	assert.Equal(t, defaultBaseURL, c.BaseURL.String())
	assert.NotNil(t, c.Checks)
	assert.Equal(t, "", c.APIToken)
	assert.Equal(t, "key", c.APIKey)

	// Test with both API Token and API Key
	c, err = NewClientWithConfig(ClientConfig{
		APIToken: "token",
		APIKey:   "key",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.DefaultClient, c.client)
	assert.Equal(t, defaultBaseURL, c.BaseURL.String())
	assert.NotNil(t, c.Checks)
	assert.Equal(t, "token", c.APIToken)
	assert.Equal(t, "key", c.APIKey)

	// Test with no credentials
	c, err = NewClientWithConfig(ClientConfig{})
	assert.Error(t, err)
}

func TestNewClientWithEnvAPITokenDoesNotOverride(t *testing.T) {
	os.Setenv("PINGDOM_API_TOKEN", "envSetToken")
	defer os.Unsetenv("PINGDOM_API_TOKEN")

	c, err := NewClientWithConfig(ClientConfig{
		APIToken: "explicitToken",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.DefaultClient, c.client)
	assert.Equal(t, defaultBaseURL, c.BaseURL.String())
	assert.NotNil(t, c.Checks)
	assert.Equal(t, c.APIToken, "explicitToken")
}

func TestNewClientWithEnvAPIKeyDoesNotOverride(t *testing.T) {
	os.Setenv("PINGDOM_API_KEY", "envSetKey")
	defer os.Unsetenv("PINGDOM_API_KEY")

	c, err := NewClientWithConfig(ClientConfig{
		APIKey: "explicitKey",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.DefaultClient, c.client)
	assert.Equal(t, defaultBaseURL, c.BaseURL.String())
	assert.NotNil(t, c.Checks)
	assert.Equal(t, c.APIKey, "explicitKey")
}

func TestNewClientWithEnvAPITokenWorks(t *testing.T) {
	os.Setenv("PINGDOM_API_TOKEN", "envSetToken")
	defer os.Unsetenv("PINGDOM_API_TOKEN")

	c, err := NewClientWithConfig(ClientConfig{})
	assert.NoError(t, err)
	assert.Equal(t, http.DefaultClient, c.client)
	assert.Equal(t, defaultBaseURL, c.BaseURL.String())
	assert.NotNil(t, c.Checks)
	assert.Equal(t, c.APIToken, "envSetToken")
}

func TestNewClientWithEnvAPIKeyWorks(t *testing.T) {
	os.Setenv("PINGDOM_API_KEY", "envSetKey")
	defer os.Unsetenv("PINGDOM_API_KEY")
	
	// Clear API token in case it's also set
	origToken := os.Getenv("PINGDOM_API_TOKEN")
	os.Unsetenv("PINGDOM_API_TOKEN")
	defer os.Setenv("PINGDOM_API_TOKEN", origToken)

	c, err := NewClientWithConfig(ClientConfig{})
	assert.NoError(t, err)
	assert.Equal(t, http.DefaultClient, c.client)
	assert.Equal(t, defaultBaseURL, c.BaseURL.String())
	assert.NotNil(t, c.Checks)
	assert.Equal(t, c.APIKey, "envSetKey")
}

func TestClientAuthenticationHeaders(t *testing.T) {
	// Test API Token authentication header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	clientWithToken, _ := NewClientWithConfig(ClientConfig{
		APIToken: "token",
	})
	baseURL, _ := url.Parse(server.URL)
	clientWithToken.BaseURL = baseURL
	req, _ := clientWithToken.NewRequest("GET", "/", nil)
	clientWithToken.Do(req, nil)

	// Test API Key authentication header
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer key", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	clientWithKey, _ := NewClientWithConfig(ClientConfig{
		APIKey: "key",
	})
	baseURL, _ = url.Parse(server.URL)
	clientWithKey.BaseURL = baseURL
	req, _ = clientWithKey.NewRequest("GET", "/", nil)
	clientWithKey.Do(req, nil)

	// Test precedence (API Key should be used if both are provided)
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer key", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	clientWithBoth, _ := NewClientWithConfig(ClientConfig{
		APIToken: "token",
		APIKey:   "key",
	})
	baseURL, _ = url.Parse(server.URL)
	clientWithBoth.BaseURL = baseURL
	req, _ = clientWithBoth.NewRequest("GET", "/", nil)
	clientWithBoth.Do(req, nil)
}

func TestNewRequest(t *testing.T) {
	setup()
	defer teardown()

	req, err := client.NewRequest("GET", "/checks", nil)
	assert.NoError(t, err)
	assert.Equal(t, "GET", req.Method)
	assert.Equal(t, client.BaseURL.String()+"/checks", req.URL.String())
}

func TestNewRequestWithAPIKey(t *testing.T) {
	setupWithAPIKey()
	defer teardown()

	req, err := client.NewRequest("GET", "/checks", nil)
	assert.NoError(t, err)
	assert.Equal(t, "GET", req.Method)
	assert.Equal(t, client.BaseURL.String()+"/checks", req.URL.String())
	assert.Equal(t, "Bearer my_api_key", req.Header.Get("Authorization"))
}

func TestDo(t *testing.T) {
	setup()
	defer teardown()

	type foo struct {
		A string
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if m := "GET"; m != r.Method {
			t.Errorf("Request method = %v, want %v", r.Method, m)
		}
		fmt.Fprint(w, `{"A":"a"}`)
	})

	req, _ := client.NewRequest("GET", "/", nil)
	body := new(foo)
	want := &foo{"a"}
	_, err := client.Do(req, body)
	assert.NoError(t, err)
	assert.Equal(t, want, body)
}

func TestValidateResponse(t *testing.T) {
	valid := &http.Response{
		Request:    &http.Request{},
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(strings.NewReader("OK")),
	}
	assert.NoError(t, validateResponse(valid))

	invalid := &http.Response{
		Request:    &http.Request{},
		StatusCode: http.StatusBadRequest,
		Body: ioutil.NopCloser(strings.NewReader(`{
		"error" : {
			"statuscode": 400,
			"statusdesc": "Bad Request",
			"errormessage": "This is an error"
		}
		}`)),
	}
	want := &PingdomError{400, "Bad Request", "This is an error"}
	assert.Equal(t, want, validateResponse(invalid))
}
