package http

import (
	"fmt"
	"github.com/SAP/remote-work-processor/internal/utils"
	"log"
	"time"

	"github.com/SAP/remote-work-processor/internal/executors/http/tls"
)

type OAuthorizationHeaderOption func(*oAuthorizationHeaderGenerator)

type oAuthorizationHeaderGenerator struct {
	tokenType          TokenType
	certAuthentication *tls.CertificateAuthentication
	authHeader         string
	cachingKey         string
	requestStore       map[string]string
	fetcher            TokenFetcher
}

type cachedToken struct {
	*OAuthToken
	IssuedAt int64 `json:"timestamp,omitempty"`
}

func NewOAuthorizationHeaderGenerator(tokenType TokenType, tokenUrl string, executor HttpExecutor, requestBody string,
	opts ...OAuthorizationHeaderOption) CacheableAuthorizationHeaderGenerator {
	h := &oAuthorizationHeaderGenerator{
		tokenType: tokenType,
	}

	for _, opt := range opts {
		opt(h)
	}

	h.fetcher = NewOAuthTokenFetcher(
		withExecutor(executor),
		withTokenUrl(tokenUrl),
		withRequestBody(requestBody),
		withCertificateAuthentication(h.certAuthentication),
		withAuthHeader(h.authHeader),
	)

	return h
}

func UseCertificateAuthentication(certAuthentication *tls.CertificateAuthentication) OAuthorizationHeaderOption {
	return func(h *oAuthorizationHeaderGenerator) {
		h.certAuthentication = certAuthentication
	}
}

func WithAuthenticationHeader(header string) OAuthorizationHeaderOption {
	return func(h *oAuthorizationHeaderGenerator) {
		h.authHeader = header
	}
}

func WithCachingKey(cacheKey string) OAuthorizationHeaderOption {
	return func(h *oAuthorizationHeaderGenerator) {
		h.cachingKey = cacheKey
	}
}

func (h *oAuthorizationHeaderGenerator) Generate() (string, error) {
	oAuthToken, err := h.fetchToken()
	if err != nil {
		return "", err
	}

	return h.formatToken(oAuthToken)
}

func (h *oAuthorizationHeaderGenerator) GenerateWithCacheAside() (string, error) {
	log.Println("OAuth token header: checking for token in cache...")
	var cached cachedToken
	if cachedValue, inCache := h.requestStore[h.cachingKey]; inCache {
		log.Println("OAuth token header: token found in cache")
		if err := utils.FromJson(cachedValue, &cached); err != nil {
			log.Println("OAuth token header: error decoding cached token:", err)
			return "", fmt.Errorf("failed to deserialize cached OAuth token: %v", err)
		}
	}

	if h.tokenAboutToExpire(cached) {
		log.Println("OAuth token header: token is close to expiry. regenerating...")
		newToken, err := h.fetchToken()
		if err != nil {
			return "", err
		}

		cached = cachedToken{
			OAuthToken: newToken,
			IssuedAt:   time.Now().UnixMilli(),
		}

		newCachedToken, err := utils.ToJson(cached)
		if err != nil {
			log.Println("OAuth token header: failed to serialize token:", err)
			return "", fmt.Errorf("failed to serialize OAuth token: %v", err)
		}

		log.Println("OAuth token header: setting new token in cache")
		h.requestStore[h.cachingKey] = newCachedToken
	}

	return h.formatToken(cached.OAuthToken)
}

func (h *oAuthorizationHeaderGenerator) tokenAboutToExpire(token cachedToken) bool {
	// copied from OAuth2BearerAuthorizationHeader.java::isTokenAboutToExpire
	return time.Now().Add(30 * time.Second).After(time.UnixMilli(token.IssuedAt + token.ExpiresIn))
}

func (h *oAuthorizationHeaderGenerator) fetchToken() (*OAuthToken, error) {
	rawToken, err := h.fetcher.Fetch()
	if err != nil {
		log.Println("OAuth token header: failed to fetch token:", err)
		return nil, fmt.Errorf("failed to fetch OAuth token: %v", err)
	}
	return NewOAuthToken(rawToken)
}

func (h *oAuthorizationHeaderGenerator) formatToken(oAuthToken *OAuthToken) (string, error) {
	log.Println("OAuth token header: formatting token...")
	var token string
	switch h.tokenType {
	case TokenType_ACCESS:
		token = oAuthToken.AccessToken
	case TokenType_ID:
		token = oAuthToken.IdToken
	default:
		log.Println("OAuth token header: invalid token type:", h.tokenType)
		return "", NewIllegalTokenTypeError(h.tokenType)
	}

	return fmt.Sprintf("Bearer %s", token), nil
}
