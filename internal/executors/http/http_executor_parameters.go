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

var (
	defaultSuccessResponseCodes [1]string = [...]string{"2xx"}
)

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
}

func NewHttpRequestParametersFromContext(ctx executors.ExecutorContext) *HttpRequestParameters {
	opts := []functional.OptionWithError[HttpRequestParameters]{
		withMethodFromContext(&ctx),
		withUrlFromContext(&ctx),
		withTokenUrlFromContext(&ctx),
		withCsrfUrlFromContext(&ctx),
		withClientIdFromContext(&ctx),
		withClientSecretFromContext(&ctx),
		withRefreshTokenFromContext(&ctx),
		withResponseBodyTransformerFromContext(&ctx),
		withHeadersFromContext(&ctx),
		withBodyFromContext(&ctx),
		withUserFromContext(&ctx),
		withPasswordFromContext(&ctx),
		withTimeoutFromContext(&ctx),
		withSuccessResponseCodesFromContext(&ctx),
		withSucceedOnTimeoutFromContext(&ctx),
		withCertAuthenticationFromContext(&ctx),
		withAuthorizationHeaderFromContext(&ctx),
	}

	return applyBuildOptions(&HttpRequestParameters{}, opts...)
}

func NewHttpRequestParameters(opts ...functional.OptionWithError[HttpRequestParameters]) *HttpRequestParameters {
	return applyBuildOptions(&HttpRequestParameters{}, opts...)
}

func applyBuildOptions(p *HttpRequestParameters, opts ...functional.OptionWithError[HttpRequestParameters]) *HttpRequestParameters {
	for _, opt := range opts {
		opt(p)
	}

	return p
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

func WithMethod(m string) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		hrp.method = m

		return nil
	}
}

func WithUrl(u string) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		hrp.url = u

		return nil
	}
}

func WithTokenUrl(u string) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		hrp.tokenUrl = u

		return nil
	}
}

func WithCsrfUrl(u string) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		hrp.csrfUrl = u

		return nil
	}
}

func WithClientId(id string) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		hrp.clientId = id

		return nil
	}
}

func WithClientSecret(s string) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		hrp.clientSecret = s

		return nil
	}
}

func WithRefreshToken(rt string) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		hrp.refreshToken = rt

		return nil
	}
}

func WithHeaders(h map[string]string) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		hrp.headers = h

		return nil
	}
}

func WithBody(b string) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		hrp.body = b

		return nil
	}
}

func WithUser(u string) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		hrp.user = u

		return nil
	}
}

func WithPassword(p string) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		hrp.password = p

		return nil
	}
}

func WithTimeout(t uint64) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		hrp.timeout = t

		return nil
	}
}

func WithSuccessResponseCodes(src []string) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		hrp.successResponseCodes = src

		return nil
	}
}

func WithSucceedOnTimeout(s bool) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		hrp.succeedOnTimeout = s

		return nil
	}
}

func WithCertificateAuthentication(cauth *tls.CertificateAuthentication) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		hrp.certAuthentication = cauth

		return nil
	}
}

func WithAuthorizationHeader(h string) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		hrp.authorizationHeader = h

		return nil
	}
}

func withMethodFromContext(ctx *executors.ExecutorContext) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		m, err := ctx.GetRequiredString(METHOD)
		if err != nil {
			return nonRetryableError(err)
		}

		hrp.method = m
		return nil
	}
}

func withUrlFromContext(ctx *executors.ExecutorContext) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		u, err := ctx.GetRequiredString(URL)
		if err != nil {
			nonRetryableError(err)
		}

		hrp.url = u
		return nil
	}
}

func withTokenUrlFromContext(ctx *executors.ExecutorContext) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		u := ctx.GetString(TOKEN_URL)

		hrp.tokenUrl = u
		return nil
	}
}

func withCsrfUrlFromContext(ctx *executors.ExecutorContext) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		u := ctx.GetString(CSRF_URL)

		hrp.csrfUrl = u
		return nil
	}
}

func withClientIdFromContext(ctx *executors.ExecutorContext) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		id := ctx.GetString(CLIENT_ID)

		hrp.clientId = id
		return nil
	}
}

func withClientSecretFromContext(ctx *executors.ExecutorContext) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		s := ctx.GetString(CLIENT_SECRET)

		hrp.clientSecret = s
		return nil
	}
}

func withRefreshTokenFromContext(ctx *executors.ExecutorContext) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		rt := ctx.GetString(REFRESH_TOKEN)

		hrp.refreshToken = rt
		return nil
	}
}

func withResponseBodyTransformerFromContext(ctx *executors.ExecutorContext) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		t := ctx.GetString(RESPONSE_BODY_TRANSFORMER)

		hrp.responseBodyTransformer = t
		return nil
	}
}

func withHeadersFromContext(ctx *executors.ExecutorContext) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		h, err := ctx.GetMap(HEADERS)
		if err != nil {
			nonRetryableError(err)
		}

		hrp.headers = h
		return nil
	}
}

func withBodyFromContext(ctx *executors.ExecutorContext) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		b := ctx.GetString(BODY)

		hrp.body = b
		return nil
	}
}

func withUserFromContext(ctx *executors.ExecutorContext) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		u := ctx.GetString(USER)

		hrp.user = u
		return nil
	}
}

func withPasswordFromContext(ctx *executors.ExecutorContext) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		p := ctx.GetString(PASSWORD)

		hrp.password = p
		return nil
	}
}

func withTimeoutFromContext(ctx *executors.ExecutorContext) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		t, err := ctx.GetNumber(TIMEOUT)
		if err != nil {
			return nonRetryableError(err)
		}

		hrp.timeout = t
		return nil
	}
}

func withSuccessResponseCodesFromContext(ctx *executors.ExecutorContext) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		src, err := ctx.GetList(SUCCESS_RESPONSE_CODES)
		if err != nil {
			return nonRetryableError(err)
		}

		if len(src) == 0 {
			hrp.successResponseCodes = defaultSuccessResponseCodes[:]
		} else {
			hrp.successResponseCodes = src
		}
		return nil
	}
}

func withSucceedOnTimeoutFromContext(ctx *executors.ExecutorContext) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		s, err := ctx.GetBoolean(SUCCEED_ON_TIMEOUT)
		if err != nil {
			return nonRetryableError(err)
		}

		hrp.succeedOnTimeout = s
		return nil
	}
}

func withCertAuthenticationFromContext(ctx *executors.ExecutorContext) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		opts := []tls.CertificateAuthenticationOption{}

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
		hrp.certAuthentication = tls.NewCertAuthentication(opts...)
		return nil
	}
}

func withAuthorizationHeaderFromContext(ctx *executors.ExecutorContext) functional.OptionWithError[HttpRequestParameters] {
	return func(hrp *HttpRequestParameters) error {
		h := ctx.GetString(AUTHORIZATION_HEADER)

		hrp.authorizationHeader = h
		return nil
	}
}

func nonRetryableError(cause error) error {
	return executors.NewNonRetryableError(cause.Error()).WithCause(cause)
}
