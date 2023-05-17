package http

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/url"

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

func NewOAuthHeaderGenerator(p *HttpRequestParameters) AuthorizationHeaderGenerator {
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

	return nil
}

func passwordGrantGenerator(p *HttpRequestParameters) AuthorizationHeaderGenerator {
	tokenUrl := p.GetTokenUrl()
	clientId := p.GetClientId()
	clientSecret := p.GetClientSecret()
	b := fmt.Sprintf(PASSWORD_GRANT_FORMAT, urlEncoded(p.GetUser()), urlEncoded(p.GetPassword()))

	return NewOAuthorizationHeader(
		TokenType_ACCESS,
		GrantType_PASSWORD,
		tokenUrl,
		NewHttpRequestExecutor(generateBasicAuthorizationHeader(clientId, clientSecret)),
		b,
		generateCachingKey(tokenUrl, clientId, clientSecret, b),
	)
}

func passwordGrantWithClientCertificateGenerator(p *HttpRequestParameters) AuthorizationHeaderGenerator {
	tokenUrl := p.GetTokenUrl()
	clientId := p.GetClientId()
	b := fmt.Sprintf(PASSWORD_CREDENTIALS_FORMAT_WITH_CLIENT_ID, urlEncoded(clientId), urlEncoded(p.GetUser()), urlEncoded(p.GetPassword()))

	return NewOAuthorizationHeader(
		TokenType_ACCESS,
		GrantType_PASSWORD,
		p.GetTokenUrl(),
		DefaultHttpRequestExecutor(),
		b,
		generateCachingKey(tokenUrl, clientId, "", b),
		UseCertificateAuthentication(p.certAuthentication),
	)
}

func clientCredentialsGenerator(p *HttpRequestParameters, clientId string, clientSecret string) AuthorizationHeaderGenerator {
	tokenUrl := p.GetTokenUrl()
	b := fmt.Sprintf(CLIENT_CREDENTIALS_FORMAT, urlEncoded(clientId), urlEncoded(clientSecret))

	var h AuthorizationHeader

	if clientId != "" && p.certAuthentication.GetClientCertificate() == "" {
		h = generateBasicAuthorizationHeader(clientId, clientSecret)
	}

	return NewOAuthorizationHeader(
		TokenType_ACCESS,
		GrantType_CLIENT_CREDENTIALS,
		tokenUrl,
		resolveHttpExecutor(h),
		b,
		generateCachingKey(tokenUrl, clientId, clientSecret, b),
		UseCertificateAuthentication(p.certAuthentication),
	)
}

func refreshTokenGenerator(p *HttpRequestParameters) AuthorizationHeaderGenerator {
	tokenUrl := p.GetTokenUrl()
	clientId := p.GetClientId()
	clientSecret := p.GetClientSecret()
	refreshToken := p.GetRefreshToken()

	if p.certAuthentication.GetClientCertificate() == "" {
		return refreshTokenGrant(tokenUrl, clientId, clientSecret, refreshToken)
	} else {
		return refreshTokenGrantWithClientCert(tokenUrl, clientId, refreshToken, p.certAuthentication)
	}
}

func refreshTokenGrantWithClientCert(tokenUrl, clientId, refreshToken string, certAuthentication *tls.CertificateAuthentication) AuthorizationHeaderGenerator {
	b := fmt.Sprintf(REFRESH_TOKEN_FORMAT_WITH_CERT, urlEncoded(clientId), urlEncoded(refreshToken))
	emptyClientSecret := ""

	return NewOAuthorizationHeader(
		TokenType_ACCESS,
		GrantType_REFRESH_TOKEN,
		tokenUrl,
		DefaultHttpRequestExecutor(),
		b,
		generateCachingKey(tokenUrl, clientId, emptyClientSecret, b),
		UseCertificateAuthentication(certAuthentication),
	)
}

func refreshTokenGrant(tokenUrl, clientId, clientSecret, refreshToken string) AuthorizationHeaderGenerator {
	b := fmt.Sprintf(REFRESH_TOKEN_FORMAT, urlEncoded(refreshToken))

	var h AuthorizationHeader

	if clientId != "" {
		h = generateBasicAuthorizationHeader(clientId, clientSecret)
	}

	return NewOAuthorizationHeader(
		TokenType_ACCESS,
		GrantType_REFRESH_TOKEN,
		tokenUrl,
		resolveHttpExecutor(h),
		b,
		generateCachingKey(tokenUrl, clientId, clientSecret, b),
	)
}

func generateBasicAuthorizationHeader(clientId string, clientSecret string) AuthorizationHeader {
	h, err := NewBasicAuthorizationHeader(clientId, clientSecret).Generate()

	if err != nil {
		log.Fatalf("Error occurred while trying to get refresh token: %v\n", err)
	}

	return h
}

func resolveHttpExecutor(h AuthorizationHeader) HttpExecutor {
	if h != nil {
		return NewHttpRequestExecutor(h)
	} else {
		return DefaultHttpRequestExecutor()
	}
}

func urlEncoded(query string) string {
	return url.QueryEscape(query)
}

// TODO: TOTP should be considered as part of caching key here as well
func generateCachingKey(tokenUrl string, clientId string, clientSecret string, requestBody string) string {
	h := sha256.New()
	v := fmt.Sprintf(CACHING_KEY_FORMAT, tokenUrl, clientId, clientSecret, requestBody)

	h.Write([]byte(v))
	return hex.EncodeToString(h.Sum(nil))
}
