package nbaclient

import (
	"forwardproxy/client"
)

// NBAClient wraps the base client with NBA-specific headers.
type NBAClient struct {
	*client.Client
	defaultHeaders map[string]string
}

// NewNBAClient creates a new NBA client with NBA-specific headers.
func NewNBAClient() *NBAClient {
	return &NBAClient{
		Client: client.NewClient("https://stats.nba.com/stats/"),
		defaultHeaders: map[string]string{
			"User-Agent":               "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:72.0) Gecko/20100101 Firefox/72.0",
			"Accept":                   "application/json, text/plain, */*",
			"Accept-Language":          "en-US,en;q=0.5",
			"Accept-Encoding":          "gzip, deflate, br",
			"x-endpoints-stats-origin": "stats",
			"x-endpoints-stats-token":  "true",
			"Connection":               "keep-alive",
			"Referer":                  "https://stats.nba.com/",
			"Pragma":                   "no-cache",
			"Cache-Control":            "no-cache",
		},
	}
}

// GetNBAData makes a GET request to the NBA API with the specified endpoint and params.
// NBA-specific headers are automatically included.
func (n *NBAClient) GetNBAData(endpoint string, params map[string]string, customHeaders map[string]string) ([]byte, error) {
	// Merge default headers with custom headers (custom headers take precedence)
	headers := make(map[string]string)
	for k, v := range n.defaultHeaders {
		headers[k] = v
	}
	for k, v := range customHeaders {
		headers[k] = v
	}

	// Use base client's GetRequest method
	return n.Client.GetRequest(endpoint, params, headers)
}
