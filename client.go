package http

import (
	"net/http"
	"time"
)

type Client struct {
	BaseURL   string
	Headers   http.Header
	Timeout   time.Duration
	client    *http.Client
	hooks     []Hooks
	UserAgent string
}

// NewClient creates a new HTTP client with base URL.
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		Headers: http.Header{
			"User-Agent": []string{"go.blk/httpclient"},
		},
		Timeout: 30 * time.Second,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Request starts a new request builder.
func (c *Client) R() *Request {
	return &Request{
		client:  c,
		headers: http.Header{},
		query:   make(map[string]string),
	}
}

// Do is used internally to execute request, called by Request.Do().
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	// Apply default headers
	for k, vals := range c.Headers {
		for _, v := range vals {
			if req.Header.Get(k) == "" {
				req.Header.Add(k, v)
			}
		}
	}

	// Call Pre hooks
	for _, h := range c.hooks {
		if err := h.PreRequest(req); err != nil {
			return nil, err
		}
	}

	// Do request
	resp, err := c.client.Do(req)

	// Call Post hooks
	for _, h := range c.hooks {
		_ = h.PostResponse(resp)
	}

	return resp, err
}

// Use adds middleware hook (logging, retry, etc).
func (c *Client) Use(h Hooks) {
	c.hooks = append(c.hooks, h)
}
