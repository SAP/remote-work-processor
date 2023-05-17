package http

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/SAP/remote-work-processor/internal/executors/http/tls"
)

const (
	CONTENT_TYPE_HEADER              string  = "Content-Type"
	CONTENT_TYPE_URL_ENCODED         string  = "application/x-www-form-urlencoded"
	TOKEN_EXPIRATION_TIME_PERCENTAGE float32 = 0.95
)

type oAuthorizationHeaderOption func(*oAuthorizationHeader)

type oAuthorizationHeader struct {
	tokenType          TokenType
	grantType          GrantType
	token              *OAuthToken
	tokenUrl           string
	executor           HttpExecutor
	requestBody        string
	certAuthentication *tls.CertificateAuthentication
	cachingKey         string
	fetcher            TokenFetcher
	m                  *sync.Mutex
}

func NewOAuthorizationHeader(tokenType TokenType, grantType GrantType, tokenUrl string, executor HttpExecutor, requestBody string, cachingKey string, opts ...oAuthorizationHeaderOption) AuthorizationHeaderGenerator {
	h := &oAuthorizationHeader{
		tokenType:   tokenType,
		grantType:   grantType,
		token:       &OAuthToken{},
		tokenUrl:    tokenUrl,
		executor:    executor,
		requestBody: requestBody,
		cachingKey:  cachingKey,
		m:           &sync.Mutex{},
	}

	for _, opt := range opts {
		opt(h)
	}

	h.fetcher = NewOAuthTokenFetcher(
		withExecutor(executor),
		withTokenUrl(tokenUrl),
		withRequestBody(requestBody),
		withCertificateAuthentication(h.certAuthentication, func(auth *tls.CertificateAuthentication) bool { return auth != nil }),
	)

	return h
}

func UseCertificateAuthentication(certAuthentication *tls.CertificateAuthentication) oAuthorizationHeaderOption {
	return func(h *oAuthorizationHeader) {
		h.certAuthentication = certAuthentication
	}
}

func (h *oAuthorizationHeader) Generate() (AuthorizationHeader, error) {
	h.m.Lock()
	defer h.m.Unlock()

	if !h.token.HasValue() || h.tokenAboutToExpire() {
		if err := h.fetchToken(); err != nil {
			return nil, err
		}
	}

	var token string
	switch h.tokenType {
	case TokenType_ACCESS:
		token = h.token.AccessToken
	case TokenType_ID:
		token = h.token.IdToken
	default:
		return nil, NewIllegalTokenTypeError(h.tokenType)
	}

	return NewCacheableAuthorizationHeaderView(bearerToken(token), h), nil
}

func bearerToken(token string) string {
	return fmt.Sprintf("Bearer %s", token)
}

func (h *oAuthorizationHeader) tokenAboutToExpire() bool {
	issuedAt := h.token.issuedAt
	if issuedAt <= 0.0 {
		log.Fatalf("OAuth token is not initialized properly.")
	}

	return float32(issuedAt+h.token.ExpiresIn) >= TOKEN_EXPIRATION_TIME_PERCENTAGE*float32(time.Now().UnixMilli())
}

func (h *oAuthorizationHeader) setToken(token string, issuedAt int64) error {
	t, err := NewOAuthToken(token, issuedAt)
	if err != nil {
		return err
	}

	h.token = t
	return nil
}

func (h *oAuthorizationHeader) fetchToken() error {
	token, err := h.fetcher.Fetch()
	if err != nil {
		return err
	}

	issuedAt := time.Now().UnixMilli()

	err = h.setToken(token, issuedAt)
	if err != nil {
		return err
	}
	return nil
}
