package quester

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"time"
)

type Request struct {
	client            *Client
	method            string
	path              string
	headers           http.Header
	query             map[string]string
	body              any
	ctx               context.Context
	basicAuthUsername string
	basicAuthPassword string
	bearerToken       string
	enableTrace       bool
}

// EnableTrace enables HTTP trace/debug.
func (r *Request) EnableTrace() *Request {
	r.enableTrace = true
	return r
}

// SetMethod sets the HTTP method.
func (r *Request) SetMethod(method string) *Request {
	r.method = strings.ToUpper(method)
	return r
}

func (r *Request) SetBasicAuth(username, password string) *Request {
	r.basicAuthUsername = username
	r.basicAuthPassword = password
	return r
}

// SetBearerToken sets bearer token in Authorization header.
func (r *Request) SetBearerToken(token string) *Request {
	r.bearerToken = token
	return r
}

// SetPath sets the request path (relative to base URL).
func (r *Request) SetPath(path string) *Request {
	r.path = path
	return r
}

// SetHeader sets a custom header.
func (r *Request) SetHeader(key, value string) *Request {
	r.headers.Set(key, value)
	return r
}

// SetQuery adds a query parameter.
func (r *Request) SetQuery(key, value string) *Request {
	r.query[key] = value
	return r
}

// SetQueries adds a query parameter.
func (r *Request) SetQueries(queries map[string]string) *Request {
	for k, q := range queries {
		r.query[k] = q
	}
	return r
}

// SetBody sets the request body.
func (r *Request) SetBody(body any) *Request {
	r.body = body
	return r
}

// SetContext sets a custom context.
func (r *Request) SetContext(ctx context.Context) *Request {
	r.ctx = ctx
	return r
}

// SetTimeout sets a timeout for the request context.
func (r *Request) SetTimeout(d time.Duration) *Request {
	if r.ctx != nil {
		var cancel context.CancelFunc
		r.ctx, cancel = context.WithTimeout(r.ctx, d)
		go func() {
			<-r.ctx.Done()
			cancel()
		}()
	} else {
		var cancel context.CancelFunc
		r.ctx, cancel = context.WithTimeout(context.Background(), d)
		go func() {
			<-r.ctx.Done()
			cancel()
		}()
	}
	return r
}

// Do sends the request and decodes the response into result.
func (r *Request) Do(result any) (*Response, error) {
	fullURL := r.client.BaseURL + r.path

	// Build query
	if len(r.query) > 0 {
		q := url.Values{}
		for k, v := range r.query {
			q.Set(k, v)
		}
		fullURL += "?" + q.Encode()
	}

	var bodyReader io.Reader
	switch b := r.body.(type) {
	case nil:
	case io.Reader:
		bodyReader = b
	default:
		buf := &bytes.Buffer{}
		if err := json.NewEncoder(buf).Encode(b); err != nil {
			return nil, err
		}
		bodyReader = buf
		if r.headers.Get("Content-Type") == "" {
			r.headers.Set("Content-Type", "application/json")
		}
	}

	var trace *httptrace.ClientTrace
	if r.enableTrace {
		trace = &httptrace.ClientTrace{
			DNSStart: func(info httptrace.DNSStartInfo) {
				println("[TRACE] DNS Start:", info.Host)
			},
			DNSDone: func(info httptrace.DNSDoneInfo) {
				println("[TRACE] DNS Done:", info.Addrs)
			},
			ConnectStart: func(network, addr string) {
				println("[TRACE] Connect Start:", network, addr)
			},
			ConnectDone: func(network, addr string, err error) {
				println("[TRACE] Connect Done:", network, addr, err)
			},
			GotFirstResponseByte: func() {
				println("[TRACE] Got First Byte:", time.Now().Format(time.RFC3339Nano))
			},
		}
	}

	ctx := r.ctxOrDefault()
	if trace != nil {
		ctx = httptrace.WithClientTrace(ctx, trace)
	}

	// Build request
	req, err := http.NewRequestWithContext(ctx, r.method, fullURL, bodyReader)
	if err != nil {
		return nil, err
	}

	// Set Basic Auth if present
	if r.basicAuthUsername != "" || r.basicAuthPassword != "" {
		req.SetBasicAuth(r.basicAuthUsername, r.basicAuthPassword)
	}

	// Set Bearer Token if present
	if r.bearerToken != "" && req.Header.Get("Authorization") == "" {
		req.Header.Set("Authorization", "Bearer "+r.bearerToken)
	}

	// Add per-request headers
	for k, vals := range r.headers {
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}

	// Send
	res, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Read body
	resp := &Response{
		Status:     res.StatusCode,
		Headers:    res.Header,
		Body:       res,
		StatusText: res.Status,
	}

	// Decode response if provided
	if result != nil {
		contentType := res.Header.Get("Content-Type")
		switch {
		case strings.Contains(contentType, "application/json"):
			err = json.NewDecoder(res.Body).Decode(result)
		case strings.Contains(contentType, "application/xml"), strings.Contains(contentType, "text/xml"):
			err = xml.NewDecoder(res.Body).Decode(result)
		default:
			resp.Body, _ = io.ReadAll(res.Body)
		}
	}

	return resp, err
}

func (r *Request) ctxOrDefault() context.Context {
	if r.ctx != nil {
		return r.ctx
	}
	return context.Background()
}
