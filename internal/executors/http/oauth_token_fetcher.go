package http

import (
	"net/http"

	"github.com/SAP/remote-work-processor/internal/executors/http/tls"
	"github.com/SAP/remote-work-processor/internal/functional"
)

type oAuthTokenFetcher struct {
	HttpExecutor
	tokenUrl           string
	body               string
	certAuthentication *tls.CertificateAuthentication
}

func NewOAuthTokenFetcher(opts ...functional.Option[oAuthTokenFetcher]) TokenFetcher {
	f := &oAuthTokenFetcher{}

	for _, opt := range opts {
		opt(f)
	}

	return f
}

func withExecutor(executor HttpExecutor) functional.Option[oAuthTokenFetcher] {
	return func(f *oAuthTokenFetcher) {
		f.HttpExecutor = executor
	}
}

func withTokenUrl(url string) functional.Option[oAuthTokenFetcher] {
	return func(f *oAuthTokenFetcher) {
		f.tokenUrl = url
	}
}

func withRequestBody(body string) functional.Option[oAuthTokenFetcher] {
	return func(f *oAuthTokenFetcher) {
		f.body = body
	}
}

func withCertificateAuthentication(auth *tls.CertificateAuthentication, p functional.Predicate[*tls.CertificateAuthentication]) functional.Option[oAuthTokenFetcher] {
	return func(f *oAuthTokenFetcher) {
		if p(auth) {
			f.certAuthentication = auth
		}
	}
}

func (f *oAuthTokenFetcher) Fetch() (string, error) {
	p := f.createRequestParameters()

	// TODO: TOTP should be handled here
	r, err := f.HttpExecutor.ExecuteWithParameters(p)
	if err != nil {
		return "", err
	}

	return r.Content, nil
}

func (f *oAuthTokenFetcher) createRequestParameters() *HttpRequestParameters {
	opts := []functional.OptionWithError[HttpRequestParameters]{
		WithUrl(f.tokenUrl),
		WithMethod(http.MethodPost),
		WithHeaders(ContentTypeUrlFormEncoded()),
		WithBody(f.body),
	}

	if f.certAuthentication != nil {
		opts = append(opts, WithCertificateAuthentication(f.certAuthentication))
	}

	return NewHttpRequestParameters(opts...)
}

func ContentTypeUrlFormEncoded() map[string]string {
	return map[string]string{
		CONTENT_TYPE_HEADER: CONTENT_TYPE_URL_ENCODED,
	}
}
