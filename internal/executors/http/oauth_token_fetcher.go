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

func withCertificateAuthentication(auth *tls.CertificateAuthentication, p func(*tls.CertificateAuthentication) bool) functional.Option[oAuthTokenFetcher] {
	return func(f *oAuthTokenFetcher) {
		if p(auth) {
			f.certAuthentication = auth
		}
	}
}

func (f *oAuthTokenFetcher) Fetch() (string, error) {
	params, err := f.createRequestParameters()
	if err != nil {
		return "", err
	}

	// TODO: TOTP should be handled here
	req, err := f.HttpExecutor.ExecuteWithParameters(params)
	if err != nil {
		return "", err
	}

	return req.Content, nil
}

func (f *oAuthTokenFetcher) createRequestParameters() (*HttpRequestParameters, error) {
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
		"Content-Type": "application/x-www-form-urlencoded",
	}
}
