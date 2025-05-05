package pingdom

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	defaultBaseURL = "https://api.pingdom.com/api/3.1"
)

// Client represents a client to the Pingdom API.
type Client struct {
	APIToken     string
	APIKey       string // Added field for API Key authentication
	BaseURL      *url.URL
	client       *http.Client
	Checks       *CheckService
	Contacts     *ContactService
	Maintenances *MaintenanceService
	Occurrences  *OccurrenceService
	Probes       *ProbeService
	Teams        *TeamService
	TMSCheck     *TMSCheckService
}

// ClientConfig represents a configuration for a pingdom client.
type ClientConfig struct {
	APIToken   string
	APIKey     string // Added API Key option
	BaseURL    string
	HTTPClient *http.Client
}

// NewClientWithConfig returns a Pingdom client.
func NewClientWithConfig(config ClientConfig) (*Client, error) {
	var baseURL *url.URL
	var err error
	if config.BaseURL != "" {
		baseURL, err = url.Parse(config.BaseURL)
	} else {
		baseURL, err = url.Parse(defaultBaseURL)
	}
	if err != nil {
		return nil, err
	}

	c := &Client{
		BaseURL: baseURL,
	}

	// Handle API Token configuration
	if config.APIToken == "" {
		if envAPIToken, set := os.LookupEnv("PINGDOM_API_TOKEN"); set {
			c.APIToken = envAPIToken
		}
	} else {
		c.APIToken = config.APIToken
	}

	// Handle API Key configuration
	if config.APIKey == "" {
		if envAPIKey, set := os.LookupEnv("PINGDOM_API_KEY"); set {
			c.APIKey = envAPIKey
		}
	} else {
		c.APIKey = config.APIKey
	}

	// Ensure at least one authentication method is provided
	if c.APIToken == "" && c.APIKey == "" {
		return nil, fmt.Errorf("either API Token or API Key must be provided")
	}

	if config.HTTPClient != nil {
		c.client = config.HTTPClient
	} else {
		c.client = http.DefaultClient
	}

	c.Checks = &CheckService{client: c}
	c.Contacts = &ContactService{client: c}
	c.Maintenances = &MaintenanceService{client: c}
	c.Occurrences = &OccurrenceService{client: c}
	c.Probes = &ProbeService{client: c}
	c.Teams = &TeamService{client: c}
	c.TMSCheck = &TMSCheckService{client: c}
	return c, nil
}

// addAuthHeaders adds the appropriate authentication headers to the request
func (pc *Client) addAuthHeaders(req *http.Request) {
	if pc.APIKey != "" {
		// Use API Key authentication if available
		req.Header.Add("Authorization", "Bearer "+pc.APIKey)
	} else if pc.APIToken != "" {
		// Fall back to Bearer token authentication
		req.Header.Add("Authorization", "Bearer "+pc.APIToken)
	}
}

// NewRequest makes a new HTTP Request.  The method param should be an HTTP method in
// all caps such as GET, POST, PUT, DELETE.  The rsc param should correspond with
// a restful resource.  Params can be passed in as a map of strings
func (pc *Client) NewRequest(method string, rsc string, params map[string]string) (*http.Request, error) {
	baseURL, err := url.Parse(pc.BaseURL.String() + rsc)
	if err != nil {
		return nil, err
	}

	if params != nil {
		ps := url.Values{}
		for k, v := range params {
			ps.Set(k, v)
		}
		baseURL.RawQuery = ps.Encode()
	}

	req, err := http.NewRequest(method, baseURL.String(), nil)
	if err != nil {
		return nil, err
	}
	
	// Add authentication headers
	pc.addAuthHeaders(req)
	
	return req, nil
}

func (pc *Client) NewRequestMultiParamValue(method string, rsc string, params map[string][]string) (*http.Request, error) {
	baseURL, err := url.Parse(pc.BaseURL.String() + rsc)
	if err != nil {
		return nil, err
	}

	if params != nil {
		ps := url.Values{}
		for k, mv := range params {
			for _, v := range mv {
				ps.Add(k, v)
			}
		}
		baseURL.RawQuery = ps.Encode()
	}

	req, err := http.NewRequest(method, baseURL.String(), nil)
	if err != nil {
		return nil, err
	}
	
	// Add authentication headers
	pc.addAuthHeaders(req)
	
	return req, nil
}

// NewJSONRequest makes a new HTTP Request.  The method param should be an HTTP method in
// all caps such as GET, POST, PUT, DELETE.  The rsc param should correspond with
// a restful resource.  Params should be a json formatted string.
func (pc *Client) NewJSONRequest(method string, rsc string, params string) (*http.Request, error) {
	baseURL, err := url.Parse(pc.BaseURL.String() + rsc)
	if err != nil {
		return nil, err
	}

	reqBody := strings.NewReader(params)

	req, err := http.NewRequest(method, baseURL.String(), reqBody)
	if err != nil {
		return nil, err
	}
	
	// Add authentication headers
	pc.addAuthHeaders(req)
	req.Header.Add("Content-Type", "application/json")
	
	return req, nil
}

// Do makes an HTTP request and will unmarshal the JSON response in to the
// passed in interface.  If the HTTP response is outside of the 2xx range the
// response will be returned along with the error.
func (pc *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := pc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := validateResponse(resp); err != nil {
		return resp, err
	}

	err = decodeResponse(resp, v)
	return resp, err
}

func decodeResponse(r *http.Response, v interface{}) error {
	if v == nil {
		return fmt.Errorf("nil interface provided to decodeResponse")
	}

	bodyBytes, _ := ioutil.ReadAll(r.Body)
	bodyString := string(bodyBytes)
	err := json.Unmarshal([]byte(bodyString), &v)
	return err
}

// Takes an HTTP response and determines whether it was successful.
// Returns nil if the HTTP status code is within the 2xx range.  Returns
// an error otherwise.
func validateResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}

	bodyBytes, _ := ioutil.ReadAll(r.Body)
	bodyString := string(bodyBytes)
	m := &errorJSONResponse{}
	err := json.Unmarshal([]byte(bodyString), &m)
	if err != nil {
		return err
	}

	return m.Error
}
