package http

import (
	"fmt"
	"github.com/SAP/remote-work-processor/internal/utils"
	"log"
	"net/http"
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
	log.Println("CSRF token fetcher: fetching new CSRF token from:", f.csrfUrl)
	params, _ := f.createRequestParameters()

	resp, err := f.HttpExecutor.ExecuteWithParameters(params)
	if err != nil {
		log.Println("CSRF token fetcher: failed to fetch CSRF token:", err)
		return "", err
	}

	for key, value := range resp.Headers {
		if utils.Contains(csrfTokenHeaders, key) {
			return value, nil
		}
	}

	log.Println("CSRF token fetcher: CSRF token header not found in response")
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
	return NewHttpRequestParameters(http.MethodGet, f.csrfUrl, WithHeaders(f.headers))
}
