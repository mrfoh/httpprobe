package easyreq

import (
	"time"

	"github.com/mrfoh/httpprobe/internal/logging"
)

type HttpClientOptions struct {
	Logger logging.Logger
	// BaseUrl is the base URL to use for all requests
	BaseUrl string
	// Timeout is the timeout in milliseconds for the request
	Timeout int
	// Headers is a map of headers to include in every request. This will be merged with any other headers passed in
	Headers map[string]interface{}
}

func NewOptions() *HttpClientOptions {
	// Default options
	return &HttpClientOptions{
		Timeout: 10000,
		Headers: make(map[string]interface{}),
	}
}

func (o *HttpClientOptions) WithLogger(logger logging.Logger) *HttpClientOptions {
	o.Logger = logger
	return o
}

func (o *HttpClientOptions) WithTimeout(timeout int) *HttpClientOptions {
	o.Timeout = timeout
	return o
}

func (o *HttpClientOptions) WithBaseUrl(baseUrl string) *HttpClientOptions {
	o.BaseUrl = baseUrl
	return o
}

func (o *HttpClientOptions) WithHeaders(headers map[string]interface{}) *HttpClientOptions {
	o.Headers = headers
	return o
}

func (o *HttpClientOptions) GetTimeout() time.Duration {
	return time.Duration(o.Timeout) * time.Millisecond
}
