package http

import (
	"fmt"
	"net/http"

	"github.com/SAP/remote-work-processor/internal/functional"
	"github.com/SAP/remote-work-processor/internal/utils/array"
)

const CSRF_VERB = "fetch"

var csrfTokenHeaders = []string{"X-Csrf-Token", "X-Xsrf-Token"}

type csrfTokenFetcher struct {
	HttpExecutor
	csrfUrl          string
	headers          map[string]string
	succeedOnTimeout bool
}

func NewCsrfTokenFetcher(p *HttpRequestParameters, authHeader AuthorizationHeader) TokenFetcher {
	return &csrfTokenFetcher{
		HttpExecutor:     NewHttpRequestExecutor(authHeader),
		csrfUrl:          p.csrfUrl,
		headers:          createCsrfHeaders(authHeader),
		succeedOnTimeout: p.succeedOnTimeout,
	}
}

func (f *csrfTokenFetcher) Fetch() (string, error) {
	params, err := f.createRequestParameters()
	if err != nil {
		return "", err
	}

	resp, err := f.HttpExecutor.ExecuteWithParameters(params)
	if err != nil {
		return "", err
	}

	for key, value := range resp.Headers {
		if array.Contains(csrfTokenHeaders, key) {
			return value, nil
		}
	}
	return "", fmt.Errorf("no csrf header present")
}

func createCsrfHeaders(authHeader AuthorizationHeader) HttpHeaders {
	csrfHeaders := make(map[string]string)
	for _, headerKey := range csrfTokenHeaders {
		csrfHeaders[headerKey] = CSRF_VERB
	}

	if authHeader.HasValue() {
		csrfHeaders[authHeader.GetName()] = authHeader.GetValue()
	}
	return csrfHeaders
}

func (f *csrfTokenFetcher) createRequestParameters() (*HttpRequestParameters, error) {
	opts := []functional.OptionWithError[HttpRequestParameters]{
		WithUrl(f.csrfUrl),
		WithMethod(http.MethodGet),
		WithHeaders(f.headers),
	}
	return NewHttpRequestParameters(opts...)
}
