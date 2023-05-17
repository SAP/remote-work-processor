package http

import (
	"net/http"

	"github.com/SAP/remote-work-processor/internal/functional"
	"github.com/SAP/remote-work-processor/internal/utils/array"
	"github.com/SAP/remote-work-processor/internal/utils/maps"
	"github.com/SAP/remote-work-processor/internal/utils/tuple"
)

const CSRF_VERB = "fetch"

var csrfTokenHeaders = [...]string{"X-Csrf-Token", "X-Xsrf-Token"}

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
		headers:          createCsrfHeaders(p.headers, authHeader),
		succeedOnTimeout: p.succeedOnTimeout,
	}
}

func (f *csrfTokenFetcher) Fetch() (string, error) {
	p := f.createRequestParameters()

	r, err := f.HttpExecutor.ExecuteWithParameters(p)
	if err != nil {
		return "", err
	}

	pairs := maps.Pairs(r.Headers)
	filtered := array.Filter(pairs, func(pair tuple.Pair[string, string]) bool {
		// TODO: Optimize
		return array.Contains(csrfTokenHeaders[:], pair.Key)
	})

	// TODO: Error handling

	return filtered[0].Value, nil
}

func createCsrfHeaders(headers HttpHeaders, authHeader AuthorizationHeader) HttpHeaders {
	pairs := array.Map(csrfTokenHeaders[:], func(header string) tuple.Pair[string, string] {
		return tuple.PairOf(header, CSRF_VERB)
	})

	csrfHeaders := map[string]string{}
	for _, p := range pairs {
		csrfHeaders[p.Key] = p.Value
	}

	if authHeader.HasValue() {
		csrfHeaders[authHeader.GetName()] = authHeader.GetValue()
	}

	return csrfHeaders
}

func (f *csrfTokenFetcher) createRequestParameters() *HttpRequestParameters {
	opts := []functional.OptionWithError[HttpRequestParameters]{
		WithUrl(f.csrfUrl),
		WithMethod(http.MethodGet),
		WithHeaders(f.headers),
	}

	return NewHttpRequestParameters(opts...)
}
