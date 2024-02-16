package http

import (
	"fmt"
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
	body := fmt.Sprintf(PASSWORD_GRANT_FORMAT, urlEncoded(p.GetUser()), urlEncoded(p.GetPassword()))

	return NewOAuthorizationHeader(TokenType_ACCESS,
		tokenUrl,
		NewDefaultHttpRequestExecutor(),
		body,
		WithAuthenticationHeader(generateBasicAuthorizationHeader(clientId, clientSecret)))
}

func passwordGrantWithClientCertificateGenerator(p *HttpRequestParameters) AuthorizationHeaderGenerator {
	clientId := p.GetClientId()
	body := fmt.Sprintf(PASSWORD_CREDENTIALS_FORMAT_WITH_CLIENT_ID, urlEncoded(clientId), urlEncoded(p.GetUser()),
		urlEncoded(p.GetPassword()))

	return NewOAuthorizationHeader(TokenType_ACCESS,
		p.GetTokenUrl(),
		NewDefaultHttpRequestExecutor(),
		body,
		UseCertificateAuthentication(p.certAuthentication))
}

func clientCredentialsGenerator(p *HttpRequestParameters, clientId string, clientSecret string) AuthorizationHeaderGenerator {
	tokenUrl := p.GetTokenUrl()
	body := fmt.Sprintf(CLIENT_CREDENTIALS_FORMAT, urlEncoded(clientId), urlEncoded(clientSecret))

	var opt OAuthorizationHeaderOption

	if clientId != "" && p.certAuthentication.GetClientCertificate() == "" {
		opt = WithAuthenticationHeader(generateBasicAuthorizationHeader(clientId, clientSecret))
	} else {
		opt = UseCertificateAuthentication(p.certAuthentication)
	}

	return NewOAuthorizationHeader(TokenType_ACCESS,
		tokenUrl,
		NewDefaultHttpRequestExecutor(),
		body,
		opt)
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
	body := fmt.Sprintf(REFRESH_TOKEN_FORMAT_WITH_CERT, urlEncoded(clientId), urlEncoded(refreshToken))

	return NewOAuthorizationHeader(TokenType_ACCESS,
		tokenUrl,
		NewDefaultHttpRequestExecutor(),
		body,
		UseCertificateAuthentication(certAuthentication))
}

func refreshTokenGrant(tokenUrl, clientId, clientSecret, refreshToken string) AuthorizationHeaderGenerator {
	body := fmt.Sprintf(REFRESH_TOKEN_FORMAT, urlEncoded(refreshToken))

	var opts []OAuthorizationHeaderOption
	if clientId != "" {
		opts = append(opts, WithAuthenticationHeader(generateBasicAuthorizationHeader(clientId, clientSecret)))
	}

	return NewOAuthorizationHeader(TokenType_ACCESS,
		tokenUrl,
		NewDefaultHttpRequestExecutor(),
		body,
		opts...)
}

func generateBasicAuthorizationHeader(clientId string, clientSecret string) AuthorizationHeader {
	header, _ := NewBasicAuthorizationHeader(clientId, clientSecret).Generate()
	return header
}

func urlEncoded(query string) string {
	return url.QueryEscape(query)
}
