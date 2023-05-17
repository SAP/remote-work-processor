package http

import (
	"encoding/base64"
	"fmt"
)

type basicAuthorizationHeader struct {
	username string
	password string
}

func NewBasicAuthorizationHeader(u string, p string) AuthorizationHeaderGenerator {
	return &basicAuthorizationHeader{
		username: u,
		password: p,
	}
}

func (h *basicAuthorizationHeader) Generate() (AuthorizationHeader, error) {
	c := fmt.Sprintf("%s:%s", h.username, h.password)

	return NewAuthorizationHeaderView(fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(c)))), nil
}
