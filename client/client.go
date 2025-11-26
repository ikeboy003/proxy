package client

import (
	"fmt"
	"net/url"
	"time"

	"resty.dev/v3"
)

// Client represents a generic HTTP client using Resty for GET requests.
type Client struct {
	restyClient *resty.Client
	BaseURL     string
}

// NewClient creates a new base HTTP client with Resty.
func NewClient(baseURL string) *Client {
	client := resty.New().
		SetTimeout(30 * time.Second).
		SetRetryCount(0)

	return &Client{
		restyClient: client,
		BaseURL:     baseURL,
	}
}

// SetBaseURL updates the base URL for API requests.
func (c *Client) SetBaseURL(baseURL string) {
	c.BaseURL = baseURL
}

// prepareURL constructs a full URL by appending query parameters.
func prepareURL(baseURL string, params map[string]string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	// Add query parameters
	q := u.Query()
	for key, value := range params {
		q.Set(key, value)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// GetRequest makes a GET request with the given endpoint, params, and headers.
// Returns the response body as []byte and any error.
func (c *Client) GetRequest(endpoint string, params map[string]string, headers map[string]string) ([]byte, error) {
	// Construct full URL
	fullURL := c.BaseURL + endpoint
	fullURL, err := prepareURL(fullURL, params)
	if err != nil {
		return nil, fmt.Errorf("error constructing request URL: %w", err)
	}
	fmt.Println(fullURL)
	// Create Resty request
	req := c.restyClient.R()

	// Set headers
	for key, value := range headers {
		req.SetHeader(key, value)
	}

	// Execute GET request
	resp, err := req.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Return response body as bytes
	return []byte(resp.String()), nil
}
