package http

import (
	"fmt"
	"github.com/SAP/remote-work-processor/internal/utils"
	"net/http"

	"github.com/SAP/remote-work-processor/internal/functional"
)

const CsrfVerb = "fetch"

var csrfTokenHeaders = []string{"X-Csrf-Token", "X-Xsrf-Token"}

type csrfTokenFetcher struct {
	HttpExecutor
	csrfUrl          string
	headers          map[string]string
	succeedOnTimeout bool
}

func NewCsrfTokenFetcher(p *HttpRequestParameters, authHeader string) TokenFetcher {
	return &csrfTokenFetcher{
		HttpExecutor:     NewDefaultHttpRequestExecutor(),
		csrfUrl:          p.csrfUrl,
		headers:          createCsrfHeaders(authHeader),
		succeedOnTimeout: p.succeedOnTimeout,
	}
}

func (f *csrfTokenFetcher) Fetch() (string, error) {
	params, _ := f.createRequestParameters()

	resp, err := f.HttpExecutor.ExecuteWithParameters(params)
	if err != nil {
		return "", err
	}

	for key, value := range resp.Headers {
		if utils.Contains(csrfTokenHeaders, key) {
			return value, nil
		}
	}
	return "", fmt.Errorf("no csrf header present in response from %s", f.csrfUrl)
}

func createCsrfHeaders(authHeader string) HttpHeaders {
	csrfHeaders := make(map[string]string)
	for _, headerKey := range csrfTokenHeaders {
		csrfHeaders[headerKey] = CsrfVerb
	}

	if authHeader != "" {
		csrfHeaders[AuthorizationHeaderName] = authHeader
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
