package http

import "fmt"

const (
	INVALID_OAUTH_TOKEN_ERROR_MESSAGE = "Invalid oAuth 2.0 token response.\nURL: %s\nMethod: %s\nResponse code: %s"
)

type IllegalTokenTypeError struct {
	tokenType TokenType
}

func NewIllegalTokenTypeError(tokenType TokenType) *IllegalTokenTypeError {
	return &IllegalTokenTypeError{
		tokenType: tokenType,
	}
}

func (e *IllegalTokenTypeError) Error() string {
	return fmt.Sprintf("Invalid value for token type '%s'", e.tokenType)
}

type OAuthTokenParseError struct {
	url          string
	method       string
	responseCode string
}

func NewOAuthTokenParseError(url string, method string, responseCode string) *OAuthTokenParseError {
	return &OAuthTokenParseError{
		url:          url,
		method:       method,
		responseCode: responseCode,
	}
}

func (e *OAuthTokenParseError) Error() string {
	return fmt.Sprintf(INVALID_OAUTH_TOKEN_ERROR_MESSAGE, e.url, e.method, e.responseCode)
}
