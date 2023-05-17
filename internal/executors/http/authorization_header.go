package http

import (
	"log"
	"regexp"
	"strconv"

	"github.com/SAP/remote-work-processor/internal/executors"
	"github.com/SAP/remote-work-processor/internal/utils/json"
)

const (
	AUTHORIZATION_HEADER_NAME string = "Authorization"
	IAS_TOKEN_URL_PATTERN     string = "^https:\\/\\/(accounts\\.sap\\.com|[A-Za-z0-9+]+\\.accounts400\\.ondemand\\.com|[A-Za-z0-9+]+\\.accounts\\.ondemand\\.com)"
)

var iasTokenUrlRegex *regexp.Regexp = regexp.MustCompile(IAS_TOKEN_URL_PATTERN)

type AuthorizationHeader interface {
	GetName() string
	GetValue() string
	HasValue() bool
}

type CacheableAuthorizationHeader interface {
	AuthorizationHeader
	GetCachingKey() string
	GetCacheableValue() (string, error)
	ApplyCachedToken(token string) (CacheableAuthorizationHeader, error)
}

type AuthorizationHeaderView struct {
	value string
}

type CacheableAuthorizationHeaderView struct {
	AuthorizationHeaderView
	header *oAuthorizationHeader
}

type CachedToken struct {
	Token     string `json:"token,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

func NewCacheableAuthorizationHeaderView(value string, header *oAuthorizationHeader) CacheableAuthorizationHeaderView {
	return CacheableAuthorizationHeaderView{
		AuthorizationHeaderView: AuthorizationHeaderView{
			value: value,
		},
		header: header,
	}
}

func (h CacheableAuthorizationHeaderView) GetCachingKey() string {
	return h.header.cachingKey
}

func (h CacheableAuthorizationHeaderView) GetCacheableValue() (string, error) {
	token := h.header.token
	if token == nil {
		return "", nil
	}

	t, err := json.ToJson(token)
	if err != nil {
		return "", err
	}

	cached := CachedToken{
		Token:     t,
		Timestamp: strconv.FormatInt(token.issuedAt, 10),
	}

	value, err := json.ToJson(cached)
	if err != nil {
		return "", err
	}

	return string(value), nil
}

func (h CacheableAuthorizationHeaderView) ApplyCachedToken(token string) (CacheableAuthorizationHeader, error) {
	if token == "" {
		return h, nil
	}

	cached := &CachedToken{}
	err := json.FromJson(token, cached)
	if err != nil {
		return nil, err
	}

	if cached.Token == "" || cached.Timestamp == "" {
		return h, nil
	}

	issuedAt, err := strconv.ParseInt(cached.Timestamp, 10, 64)
	if err != nil {
		return nil, err
	}

	h.header.setToken(cached.Token, issuedAt)
	return nil, nil
}

func EmptyAuthorizationHeader() AuthorizationHeaderView {
	return AuthorizationHeaderView{}
}

func NewAuthorizationHeaderView(value string) AuthorizationHeaderView {
	return AuthorizationHeaderView{
		value: value,
	}
}

func (h AuthorizationHeaderView) GetName() string {
	return AUTHORIZATION_HEADER_NAME
}

func (h AuthorizationHeaderView) GetValue() string {
	return h.value
}

func (h AuthorizationHeaderView) HasValue() bool {
	return h.value != ""
}

// Currently Basic authentication and Bearer token authentication is supported, OAuth 2.0 will be added later
func CreateAuthorizationHeader(params *HttpRequestParameters) (AuthorizationHeader, error) {
	extH := params.GetAuthorizationHeader()

	if extH != "" {
		return NewExternalAuthorizationHeader(extH).Generate()
	}

	u := params.GetUser()
	p := params.GetPassword()
	tokenUrl := params.GetTokenUrl()

	if tokenUrl != "" {
		if u != "" && iasTokenUrlRegex.Match([]byte(tokenUrl)) {
			return NewIasAuthorizationHeader(tokenUrl, u, params.GetCertificateAuthentication().GetClientCertificate()).Generate()
		}

		return NewOAuthHeaderGenerator(params).Generate()
	}

	if u != "" {
		return NewBasicAuthorizationHeader(u, p).Generate()
	}

	if noAuthorizationRequired(params) {
		log.Printf("Request does not need any type of authorization header")
		return EmptyAuthorizationHeader(), nil
	}

	return EmptyAuthorizationHeader(), executors.NewNonRetryableError("Input values for the authentication related keys (user, password & authorizationHeader) are not combined properly.")
}

func noAuthorizationRequired(p *HttpRequestParameters) bool {
	switch "" {
	case p.authorizationHeader,
		p.tokenUrl,
		p.clientId,
		p.user,
		p.refreshToken:
		return true
	default:
		return false
	}
}
