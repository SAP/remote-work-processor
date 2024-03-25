package http

import (
	"github.com/SAP/remote-work-processor/internal/executors"
	"github.com/SAP/remote-work-processor/internal/executors/http/tls"
	"github.com/SAP/remote-work-processor/internal/functional"
)

const (
	METHOD                    string = "method"
	URL                       string = "url"
	TOKEN_URL                 string = "tokenUrl"
	CSRF_URL                  string = "csrfUrl"
	CLIENT_ID                 string = "clientId"
	CLIENT_SECRET             string = "clientSecret"
	REFRESH_TOKEN             string = "refreshToken"
	RESPONSE_BODY_TRANSFORMER string = "responseBodyTransformer"
	HEADERS                   string = "headers"
	BODY                      string = "body"
	USER                      string = "user"
	PASSWORD                  string = "password"
	TIMEOUT                   string = "timeout"
	SUCCESS_RESPONSE_CODES    string = "successResponseCodes"
	SUCCEED_ON_TIMEOUT        string = "succeedOnTimeout"
	TRUSTED_CERTS             string = "trustedCerts"
	CLIENT_CERT               string = "clientCert"
	TRUST_ANY_CERT            string = "trustAnyCert"
	AUTHORIZATION_HEADER      string = "authorizationHeader"
)

var defaultSuccessResponseCodes = []string{"2xx"}

type HttpRequestParameters struct {
	method                  string
	url                     string
	tokenUrl                string
	csrfUrl                 string
	clientId                string
	clientSecret            string
	refreshToken            string
	responseBodyTransformer string
	headers                 map[string]string
	body                    string
	user                    string
	password                string
	timeout                 uint64
	successResponseCodes    []string
	succeedOnTimeout        bool
	certAuthentication      *tls.CertificateAuthentication
	authorizationHeader     string

	store map[string]string
}

func NewHttpRequestParametersFromContext(ctx executors.Context) (*HttpRequestParameters, error) {
	method, err := ctx.GetRequiredString(METHOD)
	if err != nil {
		return nil, nonRetryableError(err)
	}

	url, err := ctx.GetRequiredString(URL)
	if err != nil {
		return nil, nonRetryableError(err)
	}

	opts := []functional.OptionWithError[HttpRequestParameters]{
		withTokenUrlFromContext(ctx),
		withCsrfUrlFromContext(ctx),
		withClientIdFromContext(ctx),
		withClientSecretFromContext(ctx),
		withRefreshTokenFromContext(ctx),
		withResponseBodyTransformerFromContext(ctx),
		withHeadersFromContext(ctx),
		withBodyFromContext(ctx),
		withUserFromContext(ctx),
		withPasswordFromContext(ctx),
		withTimeoutFromContext(ctx),
		withSuccessResponseCodesFromContext(ctx),
		withSucceedOnTimeoutFromContext(ctx),
		withCertAuthenticationFromContext(ctx),
		withAuthorizationHeaderFromContext(ctx),
		withStoreFromContext(ctx),
	}
	return NewHttpRequestParameters(method, url, opts...)
}

func NewHttpRequestParameters(method, url string, opts ...functional.OptionWithError[HttpRequestParameters]) (*HttpRequestParameters, error) {
	p := &HttpRequestParameters{
		method: method,
		url:    url,
	}

	for _, opt := range opts {
		if err := opt(p); err != nil {
			return nil, err
		}
	}
	return p, nil
}

func (p HttpRequestParameters) GetTokenUrl() string {
	return p.tokenUrl
}

func (p HttpRequestParameters) GetCsrfUrl() string {
	return p.csrfUrl
}

func (p HttpRequestParameters) GetClientId() string {
	return p.clientId
}

func (p HttpRequestParameters) GetClientSecret() string {
	return p.clientSecret
}

func (p HttpRequestParameters) GetRefreshToken() string {
	return p.refreshToken
}

func (p HttpRequestParameters) GetUser() string {
	return p.user
}

func (p HttpRequestParameters) GetPassword() string {
	return p.password
}

func (p HttpRequestParameters) GetAuthorizationHeader() string {
	return p.authorizationHeader
}

func (p HttpRequestParameters) GetCertificateAuthentication() *tls.CertificateAuthentication {
	return p.certAuthentication
}

func WithTokenUrl(u string) functional.OptionWithError[HttpRequestParameters] {
	return func(params *HttpRequestParameters) error {
		params.tokenUrl = u

		return nil
	}
}

func WithCsrfUrl(u string) functional.OptionWithError[HttpRequestParameters] {
	return func(params *HttpRequestParameters) error {
		params.csrfUrl = u

		return nil
	}
}

func WithClientId(id string) functional.OptionWithError[HttpRequestParameters] {
	return func(params *HttpRequestParameters) error {
		params.clientId = id

		return nil
	}
}

func WithClientSecret(s string) functional.OptionWithError[HttpRequestParameters] {
	return func(params *HttpRequestParameters) error {
		params.clientSecret = s

		return nil
	}
}

func WithRefreshToken(rt string) functional.OptionWithError[HttpRequestParameters] {
	return func(params *HttpRequestParameters) error {
		params.refreshToken = rt

		return nil
	}
}

func WithHeaders(h map[string]string) functional.OptionWithError[HttpRequestParameters] {
	return func(params *HttpRequestParameters) error {
		params.headers = h

		return nil
	}
}

func WithBody(b string) functional.OptionWithError[HttpRequestParameters] {
	return func(params *HttpRequestParameters) error {
		params.body = b

		return nil
	}
}

func WithUser(u string) functional.OptionWithError[HttpRequestParameters] {
	return func(params *HttpRequestParameters) error {
		params.user = u

		return nil
	}
}

func WithPassword(p string) functional.OptionWithError[HttpRequestParameters] {
	return func(params *HttpRequestParameters) error {
		params.password = p

		return nil
	}
}

func WithTimeout(t uint64) functional.OptionWithError[HttpRequestParameters] {
	return func(params *HttpRequestParameters) error {
		params.timeout = t

		return nil
	}
}

func WithSuccessResponseCodes(src []string) functional.OptionWithError[HttpRequestParameters] {
	return func(params *HttpRequestParameters) error {
		params.successResponseCodes = src

		return nil
	}
}

func WithSucceedOnTimeout(s bool) functional.OptionWithError[HttpRequestParameters] {
	return func(params *HttpRequestParameters) error {
		params.succeedOnTimeout = s

		return nil
	}
}

func WithCertificateAuthentication(cauth *tls.CertificateAuthentication) functional.OptionWithError[HttpRequestParameters] {
	return func(params *HttpRequestParameters) error {
		params.certAuthentication = cauth

		return nil
	}
}

func WithAuthorizationHeader(h string) functional.OptionWithError[HttpRequestParameters] {
	return func(params *HttpRequestParameters) error {
		params.authorizationHeader = h

		return nil
	}
}

func withTokenUrlFromContext(ctx executors.Context) functional.OptionWithError[HttpRequestParameters] {
	return func(params *HttpRequestParameters) error {
		u := ctx.GetString(TOKEN_URL)

		params.tokenUrl = u
		return nil
	}
}

func withCsrfUrlFromContext(ctx executors.Context) functional.OptionWithError[HttpRequestParameters] {
	return func(params *HttpRequestParameters) error {
		u := ctx.GetString(CSRF_URL)

		params.csrfUrl = u
		return nil
	}
}

func withClientIdFromContext(ctx executors.Context) functional.OptionWithError[HttpRequestParameters] {
	return func(params *HttpRequestParameters) error {
		id := ctx.GetString(CLIENT_ID)

		params.clientId = id
		return nil
	}
}

func withClientSecretFromContext(ctx executors.Context) functional.OptionWithError[HttpRequestParameters] {
	return func(params *HttpRequestParameters) error {
		s := ctx.GetString(CLIENT_SECRET)

		params.clientSecret = s
		return nil
	}
}

func withRefreshTokenFromContext(ctx executors.Context) functional.OptionWithError[HttpRequestParameters] {
	return func(params *HttpRequestParameters) error {
		rt := ctx.GetString(REFRESH_TOKEN)

		params.refreshToken = rt
		return nil
	}
}

func withResponseBodyTransformerFromContext(ctx executors.Context) functional.OptionWithError[HttpRequestParameters] {
	return func(params *HttpRequestParameters) error {
		t := ctx.GetString(RESPONSE_BODY_TRANSFORMER)

		params.responseBodyTransformer = t
		return nil
	}
}

func withHeadersFromContext(ctx executors.Context) functional.OptionWithError[HttpRequestParameters] {
	return func(params *HttpRequestParameters) error {
		h, err := ctx.GetMap(HEADERS)
		if err != nil {
			return nonRetryableError(err)
		}

		params.headers = h
		return nil
	}
}

func withBodyFromContext(ctx executors.Context) functional.OptionWithError[HttpRequestParameters] {
	return func(params *HttpRequestParameters) error {
		b := ctx.GetString(BODY)

		params.body = b
		return nil
	}
}

func withUserFromContext(ctx executors.Context) functional.OptionWithError[HttpRequestParameters] {
	return func(params *HttpRequestParameters) error {
		u := ctx.GetString(USER)

		params.user = u
		return nil
	}
}

func withPasswordFromContext(ctx executors.Context) functional.OptionWithError[HttpRequestParameters] {
	return func(params *HttpRequestParameters) error {
		p := ctx.GetString(PASSWORD)

		params.password = p
		return nil
	}
}

func withTimeoutFromContext(ctx executors.Context) functional.OptionWithError[HttpRequestParameters] {
	return func(params *HttpRequestParameters) error {
		timeout, err := ctx.GetNumber(TIMEOUT)
		if err != nil {
			return nonRetryableError(err)
		}

		params.timeout = timeout
		return nil
	}
}

func withSuccessResponseCodesFromContext(ctx executors.Context) functional.OptionWithError[HttpRequestParameters] {
	return func(params *HttpRequestParameters) error {
		src, err := ctx.GetList(SUCCESS_RESPONSE_CODES)
		if err != nil {
			return nonRetryableError(err)
		}

		if len(src) == 0 {
			params.successResponseCodes = defaultSuccessResponseCodes
		} else {
			params.successResponseCodes = src
		}
		return nil
	}
}

func withSucceedOnTimeoutFromContext(ctx executors.Context) functional.OptionWithError[HttpRequestParameters] {
	return func(params *HttpRequestParameters) error {
		s, err := ctx.GetBoolean(SUCCEED_ON_TIMEOUT)
		if err != nil {
			return nonRetryableError(err)
		}

		params.succeedOnTimeout = s
		return nil
	}
}

func withCertAuthenticationFromContext(ctx executors.Context) functional.OptionWithError[HttpRequestParameters] {
	return func(params *HttpRequestParameters) error {
		var opts []tls.CertificateAuthenticationOption

		tCerts := ctx.GetString(TRUSTED_CERTS)
		if len(tCerts) > 0 {
			opts = append(opts, tls.TrustCertificates(tCerts))
		}

		cCert := ctx.GetString(CLIENT_CERT)
		if len(cCert) > 0 {
			opts = append(opts, tls.WithClientCertificate(cCert))
		}

		trustAnyCert, err := ctx.GetBoolean(TRUST_ANY_CERT)
		if err != nil {
			return nonRetryableError(err)
		}
		opts = append(opts, tls.TrustAnyCertificate(trustAnyCert))

		// TODO: Validation can be done before creating CertificateAuthentication object
		params.certAuthentication = tls.NewCertAuthentication(opts...)
		return nil
	}
}

func withAuthorizationHeaderFromContext(ctx executors.Context) functional.OptionWithError[HttpRequestParameters] {
	return func(params *HttpRequestParameters) error {
		h := ctx.GetString(AUTHORIZATION_HEADER)

		params.authorizationHeader = h
		return nil
	}
}

func withStoreFromContext(ctx executors.Context) functional.OptionWithError[HttpRequestParameters] {
	return func(params *HttpRequestParameters) error {
		params.store = ctx.GetStore()
		return nil
	}
}

func nonRetryableError(cause error) error {
	return executors.NewNonRetryableError(cause.Error()).WithCause(cause)
}
