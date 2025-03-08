package easyreq

import (
	"github.com/stretchr/testify/mock"
	"net/http"
)

// HttpClientMock is a mock implementation of the easyreq.HttpClient interface for testing
type HttpClientMock struct {
	mock.Mock
	
	// Track method calls
	GetCalls     []string
	PostCalls    []string
	PutCalls     []string
	DeleteCalls  []string
	PatchCalls   []string
	HeadCalls    []string
	OptionsCalls []string
	
	// Default response to return if mock.Called is not used
	MockResponse *HttpResponse
	MockError    error
	
	// Custom implementations for HTTP methods (override mock.Called behavior)
	CustomGet     func(url string, params RequestParams) (*HttpResponse, error)
	CustomPost    func(url string, body interface{}, params RequestParams) (*HttpResponse, error)
	CustomPut     func(url string, body interface{}, params RequestParams) (*HttpResponse, error)
	CustomDelete  func(url string, params RequestParams) (*HttpResponse, error)
	CustomPatch   func(url string, body interface{}, params RequestParams) (*HttpResponse, error)
	CustomHead    func(url string, params RequestParams) (*HttpResponse, error)
	CustomOptions func(url string, params RequestParams) (*HttpResponse, error)
}

// NewHttpClientMock creates a new instance of HttpClientMock
func NewHttpClientMock() *HttpClientMock {
	// Default mock response
	defaultResponse := &HttpResponse{
		Status:  200,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(`{"result":"success"}`),
	}
	
	return &HttpClientMock{
		GetCalls:     make([]string, 0),
		PostCalls:    make([]string, 0),
		PutCalls:     make([]string, 0),
		DeleteCalls:  make([]string, 0),
		PatchCalls:   make([]string, 0),
		HeadCalls:    make([]string, 0),
		OptionsCalls: make([]string, 0),
		MockResponse: defaultResponse,
	}
}

// Custom HTTP methods with flexible implementation patterns
func (m *HttpClientMock) Get(url string, params RequestParams) (*HttpResponse, error) {
	// Track the call
	m.GetCalls = append(m.GetCalls, url)
	
	// Use custom implementation if provided
	if m.CustomGet != nil {
		return m.CustomGet(url, params)
	}
	
	// Use testify/mock if args are defined
	if len(m.Mock.ExpectedCalls) > 0 {
		args := m.Called(url, params)
		return args.Get(0).(*HttpResponse), args.Error(1)
	}
	
	// Default behavior
	return m.MockResponse, m.MockError
}

func (m *HttpClientMock) Post(url string, body interface{}, params RequestParams) (*HttpResponse, error) {
	// Track the call
	m.PostCalls = append(m.PostCalls, url)
	
	// Use custom implementation if provided
	if m.CustomPost != nil {
		return m.CustomPost(url, body, params)
	}
	
	// Use testify/mock if args are defined
	if len(m.Mock.ExpectedCalls) > 0 {
		args := m.Called(url, body, params)
		return args.Get(0).(*HttpResponse), args.Error(1)
	}
	
	// Default behavior
	return m.MockResponse, m.MockError
}

func (m *HttpClientMock) Put(url string, body interface{}, params RequestParams) (*HttpResponse, error) {
	// Track the call
	m.PutCalls = append(m.PutCalls, url)
	
	// Use custom implementation if provided
	if m.CustomPut != nil {
		return m.CustomPut(url, body, params)
	}
	
	// Use testify/mock if args are defined
	if len(m.Mock.ExpectedCalls) > 0 {
		args := m.Called(url, body, params)
		return args.Get(0).(*HttpResponse), args.Error(1)
	}
	
	// Default behavior
	return m.MockResponse, m.MockError
}

func (m *HttpClientMock) Delete(url string, params RequestParams) (*HttpResponse, error) {
	// Track the call
	m.DeleteCalls = append(m.DeleteCalls, url)
	
	// Use custom implementation if provided
	if m.CustomDelete != nil {
		return m.CustomDelete(url, params)
	}
	
	// Use testify/mock if args are defined
	if len(m.Mock.ExpectedCalls) > 0 {
		args := m.Called(url, params)
		return args.Get(0).(*HttpResponse), args.Error(1)
	}
	
	// Default behavior
	return m.MockResponse, m.MockError
}

func (m *HttpClientMock) Options(url string, params RequestParams) (*HttpResponse, error) {
	// Track the call
	m.OptionsCalls = append(m.OptionsCalls, url)
	
	// Use custom implementation if provided
	if m.CustomOptions != nil {
		return m.CustomOptions(url, params)
	}
	
	// Use testify/mock if args are defined
	if len(m.Mock.ExpectedCalls) > 0 {
		args := m.Called(url, params)
		return args.Get(0).(*HttpResponse), args.Error(1)
	}
	
	// Default behavior
	return m.MockResponse, m.MockError
}

func (m *HttpClientMock) Head(url string, params RequestParams) (*HttpResponse, error) {
	// Track the call
	m.HeadCalls = append(m.HeadCalls, url)
	
	// Use custom implementation if provided
	if m.CustomHead != nil {
		return m.CustomHead(url, params)
	}
	
	// Use testify/mock if args are defined
	if len(m.Mock.ExpectedCalls) > 0 {
		args := m.Called(url, params)
		return args.Get(0).(*HttpResponse), args.Error(1)
	}
	
	// Default behavior
	return m.MockResponse, m.MockError
}

func (m *HttpClientMock) Patch(url string, body interface{}, params RequestParams) (*HttpResponse, error) {
	// Track the call
	m.PatchCalls = append(m.PatchCalls, url)
	
	// Use custom implementation if provided
	if m.CustomPatch != nil {
		return m.CustomPatch(url, body, params)
	}
	
	// Use testify/mock if args are defined
	if len(m.Mock.ExpectedCalls) > 0 {
		args := m.Called(url, body, params)
		return args.Get(0).(*HttpResponse), args.Error(1)
	}
	
	// Default behavior
	return m.MockResponse, m.MockError
}
