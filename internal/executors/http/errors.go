package http

import "fmt"

type IllegalTokenTypeError struct {
	tokenType TokenType
}

func NewIllegalTokenTypeError(tokenType TokenType) *IllegalTokenTypeError {
	return &IllegalTokenTypeError{
		tokenType: tokenType,
	}
}

func (e *IllegalTokenTypeError) Error() string {
	return fmt.Sprintf("invalid value for token type %q", e.tokenType)
}
