package http

import (
	"github.com/SAP/remote-work-processor/internal/executors"
	"regexp"
)

const (
	AuthorizationHeaderName = "Authorization"
	IasTokenUrlPattern      = "^https:\\/\\/(accounts\\.sap\\.com|[A-Za-z0-9+]+\\.accounts400\\.ondemand\\.com|[A-Za-z0-9+]+\\.accounts\\.ondemand\\.com)"
)

var iasTokenUrlRegex = regexp.MustCompile(IasTokenUrlPattern)

// Currently only Basic and Bearer token authentication is supported.
// OAuth 2.0 will be added later

func CreateAuthorizationHeader(params *HttpRequestParameters) (string, error) {
	authHeader := params.GetAuthorizationHeader()

	if authHeader != "" {
		return authHeader, nil
	}

	user := params.GetUser()
	pass := params.GetPassword()
	tokenUrl := params.GetTokenUrl()

	if tokenUrl != "" {
		if user != "" && iasTokenUrlRegex.Match([]byte(tokenUrl)) {
			return NewIasAuthorizationHeader(tokenUrl, user, params.GetCertificateAuthentication().GetClientCertificate()).Generate()
		}
		return NewOAuthHeaderGenerator(params).GenerateWithCacheAside()
	}

	if user != "" {
		return NewBasicAuthorizationHeader(user, pass).Generate()
	}

	if noAuthorizationRequired(params) {
		return "", nil
	}

	return "", executors.NewNonRetryableError("Input values for the authentication-related keys " +
		"(user, password & authorizationHeader) are not combined properly.")
}

func noAuthorizationRequired(p *HttpRequestParameters) bool {
	isEmpty := func(s string) bool { return len(s) == 0 }
	isAnyEmpty := func(strings ...string) bool {
		for _, s := range strings {
			if isEmpty(s) {
				return true
			}
		}
		return false
	}
	return isAnyEmpty(p.authorizationHeader, p.tokenUrl, p.clientId, p.user, p.refreshToken)
}
