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
	authHeader         string
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

func withAuthHeader(header string) functional.Option[oAuthTokenFetcher] {
	return func(f *oAuthTokenFetcher) {
		f.authHeader = header
	}
}

func withCertificateAuthentication(auth *tls.CertificateAuthentication) functional.Option[oAuthTokenFetcher] {
	return func(f *oAuthTokenFetcher) {
		f.certAuthentication = auth
	}
}

func (f *oAuthTokenFetcher) Fetch() (string, error) {
	params, _ := f.createRequestParameters()

	// TODO: TOTP should be handled here
	req, err := f.HttpExecutor.ExecuteWithParameters(params)
	if err != nil {
		return "", err
	}

	return req.Content, nil
}

func (f *oAuthTokenFetcher) createRequestParameters() (*HttpRequestParameters, error) {
	opts := []functional.OptionWithError[HttpRequestParameters]{
		WithHeaders(ContentTypeUrlFormEncoded()),
		WithBody(f.body),
		WithAuthorizationHeader(f.authHeader),
	}

	if f.certAuthentication != nil {
		opts = append(opts, WithCertificateAuthentication(f.certAuthentication))
	}

	return NewHttpRequestParameters(http.MethodPost, f.tokenUrl, opts...)
}

func ContentTypeUrlFormEncoded() map[string]string {
	return map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}
}
