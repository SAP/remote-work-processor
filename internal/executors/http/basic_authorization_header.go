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
	str := fmt.Sprintf("%s:%s", h.username, h.password)
	encoded := base64.StdEncoding.EncodeToString([]byte(str))

	return NewAuthorizationHeaderView(fmt.Sprintf("Basic %s", encoded)), nil
}
