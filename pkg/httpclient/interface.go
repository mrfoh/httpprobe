package httpclient

type HttpClient interface {
	Get(requestUrl string, params RequestParams) (*HttpResponse, error)
	Post(requestUrl string, body interface{}, params RequestParams) (*HttpResponse, error)
	Put(requestUrl string, body interface{}, params RequestParams) (*HttpResponse, error)
	Delete(requestUrl string, params RequestParams) (*HttpResponse, error)
	Options(requestUrl string, params RequestParams) (*HttpResponse, error)
	Head(requestUrl string, params RequestParams) (*HttpResponse, error)
	Patch(requestUrl string, body interface{}, params RequestParams) (*HttpResponse, error)
}

type RequestParams struct {
	Headers map[string]interface{}
	Query   map[string]interface{}
}

type HttpResponse struct {
	Status  int
	Body    []byte
	Headers map[string][]string
}
