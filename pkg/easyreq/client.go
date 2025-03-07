package easyreq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	
	"go.uber.org/zap"
)

type HttpClientImpl struct {
	Client *http.Client
	Opts   *HttpClientOptions
}

type HttpRequest struct {
	Method  string
	Url     string
	Headers map[string]interface{}
	Query   map[string]interface{}
	Body    interface{}
}

func New(opts *HttpClientOptions) HttpClient {
	return &HttpClientImpl{
		Opts: opts,
		Client: &http.Client{
			Timeout: opts.GetTimeout(),
		},
	}
}

// makeRequest is a private method that makes the actual request to the server
func (c *HttpClientImpl) makeRequest(req HttpRequest) (*HttpResponse, error) {
	var request *http.Request
	var result *HttpResponse
	var err error

	requestUrl := c.requestUrl(req.Url, req.Query)

	if slices.Contains([]string{"POST", "PUT", "PATCH"}, req.Method) && req.Body != nil {
		var body []byte
		body, err = json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %v", err)
		}
		request, err = http.NewRequest(req.Method, requestUrl, bytes.NewBuffer(body))
		if err == nil {
			request.Header.Set("Content-Type", "application/json")
		}
	} else {
		request, err = http.NewRequest(req.Method, requestUrl, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Add client-level headers
	if c.Opts.Headers != nil {
		for k, v := range c.Opts.Headers {
			request.Header.Add(k, fmt.Sprintf("%v", v))
		}
	}

	// Add request-specific headers (overriding client-level headers)
	if req.Headers != nil {
		for k, v := range req.Headers {
			request.Header.Set(k, fmt.Sprintf("%v", v))
		}
	}

	// Execute the request
	resp, err := c.Client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	result = &HttpResponse{
		Status:  resp.StatusCode,
		Headers: resp.Header,
		Body:    bodyBytes,
	}

	return result, nil
}

// requestUrl constructs the full request URL
func (c *HttpClientImpl) requestUrl(value string, query map[string]interface{}) string {
	var baseUrl string

	// Determine base URL
	if c.Opts.BaseUrl != "" {
		baseUrl = fmt.Sprintf("%s/%s", c.Opts.BaseUrl, value)
	} else {
		baseUrl = value
	}

	// Add query parameters if present
	if len(query) > 0 {
		parsedUrl, err := url.Parse(baseUrl)
		if err != nil {
			if c.Opts.Logger != nil {
				c.Opts.Logger.Warn(fmt.Sprintf("Error parsing URL: %s", baseUrl), zap.Error(err))
			}
			return baseUrl
		}

		q := parsedUrl.Query()
		for k, v := range query {
			q.Add(k, fmt.Sprintf("%v", v))
		}
		parsedUrl.RawQuery = q.Encode()
		return parsedUrl.String()
	}

	return baseUrl
}

func (c *HttpClientImpl) Get(requestUrl string, params RequestParams) (*HttpResponse, error) {
	req := HttpRequest{
		Method:  http.MethodGet,
		Url:     requestUrl,
		Headers: params.Headers,
		Query:   params.Query,
	}

	return c.makeRequest(req)
}

func (c *HttpClientImpl) Post(requestUrl string, body interface{}, params RequestParams) (*HttpResponse, error) {
	req := HttpRequest{
		Method:  http.MethodPost,
		Url:     requestUrl,
		Headers: params.Headers,
		Query:   params.Query,
		Body:    body,
	}

	return c.makeRequest(req)
}

func (c *HttpClientImpl) Put(requestUrl string, body interface{}, params RequestParams) (*HttpResponse, error) {
	req := HttpRequest{
		Method:  http.MethodPut,
		Url:     requestUrl,
		Headers: params.Headers,
		Query:   params.Query,
		Body:    body,
	}

	return c.makeRequest(req)
}

func (c *HttpClientImpl) Delete(requestUrl string, params RequestParams) (*HttpResponse, error) {
	req := HttpRequest{
		Method:  http.MethodDelete,
		Url:     requestUrl,
		Headers: params.Headers,
		Query:   params.Query,
	}

	return c.makeRequest(req)
}

func (c *HttpClientImpl) Options(requestUrl string, params RequestParams) (*HttpResponse, error) {
	req := HttpRequest{
		Method:  http.MethodOptions,
		Url:     requestUrl,
		Headers: params.Headers,
		Query:   params.Query,
	}

	return c.makeRequest(req)
}

func (c *HttpClientImpl) Head(requestUrl string, params RequestParams) (*HttpResponse, error) {
	req := HttpRequest{
		Method:  http.MethodHead,
		Url:     requestUrl,
		Headers: params.Headers,
		Query:   params.Query,
	}

	return c.makeRequest(req)
}

func (c *HttpClientImpl) Patch(requestUrl string, body interface{}, params RequestParams) (*HttpResponse, error) {
	req := HttpRequest{
		Method:  http.MethodPatch,
		Url:     requestUrl,
		Headers: params.Headers,
		Query:   params.Query,
		Body:    body,
	}

	return c.makeRequest(req)
}