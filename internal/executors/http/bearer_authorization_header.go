package http

import (
	"fmt"
)

type bearerAuthorizationHeader struct {
	token string
}

func NewBearerAuthorizationHeader(t string) AuthorizationHeaderGenerator {
	return &bearerAuthorizationHeader{
		token: t,
	}
}

func (h *bearerAuthorizationHeader) Generate() (AuthorizationHeader, error) {
	return NewAuthorizationHeaderView(fmt.Sprintf("Bearer %s", h.token)), nil
}
