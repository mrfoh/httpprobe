package httpclient

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"
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

	requestUrl := c.requestUrl(req.Url, &req.Query)

	if slices.Contains([]string{"POST", "PUT", "PATCH"}, req.Method) {
		request, err = http.NewRequest(req.Method, requestUrl, nil)
	} else {
		request, err = http.NewRequest(req.Method, requestUrl, nil)
	}

	if c.Opts.Headers != nil {
		for k, v := range c.Opts.Headers {
			request.Header.Add(k, fmt.Sprintf("%v", v))
		}
	}

	if req.Headers != nil {
		for k, v := range req.Headers {
			request.Header.Add(k, fmt.Sprintf("%v", v))
		}
	}

	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	resp, err := c.Client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}

	defer resp.Body.Close()

	result = &HttpResponse{
		Status:  resp.StatusCode,
		Headers: resp.Header,
	}

	return result, err
}

// Construct the request URL
func (c *HttpClientImpl) requestUrl(value string, query *map[string]interface{}) string {
	var requestUrl string

	if query != nil {
		q := url.Values{}
		for k, v := range *query {
			q.Add(k, fmt.Sprintf("%v", v))
		}
		requestUrl = q.Encode()
	}

	if c.Opts.BaseUrl != "" {
		requestUrl = fmt.Sprintf("%s/%s", c.Opts.BaseUrl, value)
	} else {
		requestUrl = value
	}

	return requestUrl
}

func (c *HttpClientImpl) Get(requestUrl string, params RequestParams) (*HttpResponse, error) {
	req := HttpRequest{
		Method:  "GET",
		Url:     requestUrl,
		Headers: params.Headers,
		Query:   params.Query,
	}

	return c.makeRequest(req)
}

func (c *HttpClientImpl) Post(requestUrl string, body interface{}, params RequestParams) (*HttpResponse, error) {
	return nil, nil
}

func (c *HttpClientImpl) Put(requestUrl string, body interface{}, params RequestParams) (*HttpResponse, error) {
	return nil, nil
}

func (c *HttpClientImpl) Delete(requestUrl string, params RequestParams) (*HttpResponse, error) {
	return nil, nil
}

func (c *HttpClientImpl) Options(requestUrl string, params RequestParams) (*HttpResponse, error) {
	return nil, nil
}

func (c *HttpClientImpl) Head(requestUrl string, params RequestParams) (*HttpResponse, error) {
	return nil, nil
}

func (c *HttpClientImpl) Patch(requestUrl string, body interface{}, params RequestParams) (*HttpResponse, error) {
	return nil, nil
}
