package http

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"

	"github.com/SAP/remote-work-processor/internal/executors"
	"github.com/SAP/remote-work-processor/internal/executors/http/tls"
)

const (
	CACHING_KEY_FORMAT                         string = "tokenUrl=%s&oAuthUser=%s&oAuthPwd=%s&getTokenBody=%s"
	PASSWORD_GRANT_FORMAT                      string = "grant_type=password&username=%s&password=%s"
	PASSWORD_CREDENTIALS_FORMAT_WITH_CLIENT_ID string = "grant_type=password&client_id=%s&username=%s&password=%s"
	CLIENT_CREDENTIALS_FORMAT                  string = "grant_type=client_credentials&client_id=%s&client_secret=%s"
	REFRESH_TOKEN_FORMAT                       string = "grant_type=refresh_token&refresh_token=%s"
	REFRESH_TOKEN_FORMAT_WITH_CERT             string = "grant_type=refresh_token&client_id=%s&refresh_token=%s"
)

type errorTokenGenerator struct{}

func (errorTokenGenerator) Generate() (string, error) {
	return "", executors.NewNonRetryableError("missing user, client ID or refresh token")
}

func (errorTokenGenerator) GenerateWithCacheAside() (string, error) {
	return "", executors.NewNonRetryableError("missing user, client ID or refresh token")
}

func NewOAuthHeaderGenerator(p *HttpRequestParameters) CacheableAuthorizationHeaderGenerator {
	user := p.GetUser()
	clientId := p.GetClientId()
	refreshToken := p.GetRefreshToken()

	if refreshToken != "" {
		return refreshTokenGenerator(p)
	}

	if user != "" && clientId != "" {
		if p.GetCertificateAuthentication().GetClientCertificate() != "" {
			return passwordGrantWithClientCertificateGenerator(p)
		}

		return passwordGrantGenerator(p)
	}

	if user != "" {
		return clientCredentialsGenerator(p, user, p.GetPassword())
	}

	if clientId != "" {
		return clientCredentialsGenerator(p, clientId, p.GetClientSecret())
	}

	return errorTokenGenerator{}
}

func passwordGrantGenerator(p *HttpRequestParameters) CacheableAuthorizationHeaderGenerator {
	tokenUrl := p.GetTokenUrl()
	clientId := p.GetClientId()
	clientSecret := p.GetClientSecret()
	body := fmt.Sprintf(PASSWORD_GRANT_FORMAT, urlEncoded(p.GetUser()), urlEncoded(p.GetPassword()))

	return NewOAuthorizationHeaderGenerator(TokenType_ACCESS,
		tokenUrl,
		NewDefaultHttpRequestExecutor(),
		body,
		WithAuthenticationHeader(generateBasicAuthorizationHeader(clientId, clientSecret)),
		WithCachingKey(generateCachingKey(tokenUrl, clientId, clientSecret, body)),
		WithCacheStore(p.store))
}

func passwordGrantWithClientCertificateGenerator(p *HttpRequestParameters) CacheableAuthorizationHeaderGenerator {
	tokenUrl := p.GetTokenUrl()
	clientId := p.GetClientId()
	body := fmt.Sprintf(PASSWORD_CREDENTIALS_FORMAT_WITH_CLIENT_ID, urlEncoded(clientId), urlEncoded(p.GetUser()),
		urlEncoded(p.GetPassword()))

	return NewOAuthorizationHeaderGenerator(TokenType_ACCESS,
		p.GetTokenUrl(),
		NewDefaultHttpRequestExecutor(),
		body,
		UseCertificateAuthentication(p.GetCertificateAuthentication()),
		WithCachingKey(generateCachingKey(tokenUrl, clientId, "", body)),
		WithCacheStore(p.store))
}

func clientCredentialsGenerator(p *HttpRequestParameters, clientId string, clientSecret string) CacheableAuthorizationHeaderGenerator {
	tokenUrl := p.GetTokenUrl()
	body := fmt.Sprintf(CLIENT_CREDENTIALS_FORMAT, urlEncoded(clientId), urlEncoded(clientSecret))

	var opt OAuthorizationHeaderOption

	if clientId != "" && p.GetCertificateAuthentication().GetClientCertificate() == "" {
		opt = WithAuthenticationHeader(generateBasicAuthorizationHeader(clientId, clientSecret))
	} else {
		opt = UseCertificateAuthentication(p.GetCertificateAuthentication())
	}

	return NewOAuthorizationHeaderGenerator(TokenType_ACCESS,
		tokenUrl,
		NewDefaultHttpRequestExecutor(),
		body,
		opt,
		WithCachingKey(generateCachingKey(tokenUrl, clientId, clientSecret, body)),
		WithCacheStore(p.store))
}

func refreshTokenGenerator(p *HttpRequestParameters) CacheableAuthorizationHeaderGenerator {
	tokenUrl := p.GetTokenUrl()
	clientId := p.GetClientId()
	clientSecret := p.GetClientSecret()
	refreshToken := p.GetRefreshToken()

	if p.GetCertificateAuthentication().GetClientCertificate() == "" {
		return refreshTokenGrant(tokenUrl, clientId, clientSecret, refreshToken, p.store)
	} else {
		return refreshTokenGrantWithClientCert(tokenUrl, clientId, refreshToken, p.GetCertificateAuthentication(), p.store)
	}
}

func refreshTokenGrantWithClientCert(tokenUrl, clientId, refreshToken string, certAuthentication *tls.CertificateAuthentication,
	store map[string]string) CacheableAuthorizationHeaderGenerator {
	body := fmt.Sprintf(REFRESH_TOKEN_FORMAT_WITH_CERT, urlEncoded(clientId), urlEncoded(refreshToken))

	return NewOAuthorizationHeaderGenerator(TokenType_ACCESS,
		tokenUrl,
		NewDefaultHttpRequestExecutor(),
		body,
		UseCertificateAuthentication(certAuthentication),
		WithCachingKey(generateCachingKey(tokenUrl, clientId, "", body)),
		WithCacheStore(store))
}

func refreshTokenGrant(tokenUrl, clientId, clientSecret, refreshToken string, store map[string]string) CacheableAuthorizationHeaderGenerator {
	body := fmt.Sprintf(REFRESH_TOKEN_FORMAT, urlEncoded(refreshToken))

	var opts []OAuthorizationHeaderOption
	if clientId != "" {
		opts = append(opts, WithAuthenticationHeader(generateBasicAuthorizationHeader(clientId, clientSecret)))
	}
	opts = append(opts, WithCachingKey(generateCachingKey(tokenUrl, clientId, clientSecret, body)),
		WithCacheStore(store))

	return NewOAuthorizationHeaderGenerator(TokenType_ACCESS,
		tokenUrl,
		NewDefaultHttpRequestExecutor(),
		body,
		opts...)
}

func generateBasicAuthorizationHeader(clientId string, clientSecret string) string {
	header, _ := NewBasicAuthorizationHeader(clientId, clientSecret).Generate()
	return header
}

func urlEncoded(query string) string {
	return url.QueryEscape(query)
}

// TODO: TOTP should be considered as part of caching key here as well
func generateCachingKey(tokenUrl string, clientId string, clientSecret string, requestBody string) string {
	h := sha256.New()
	h.Write(fmt.Appendf(nil, CACHING_KEY_FORMAT, tokenUrl, clientId, clientSecret, requestBody))
	return hex.EncodeToString(h.Sum(nil))
}
