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

func (h *basicAuthorizationHeader) Generate() (string, error) {
	encoded := base64.StdEncoding.EncodeToString(
		fmt.Appendf(nil, "%s:%s", h.username, h.password),
	)
	return fmt.Sprintf("Basic %s", encoded), nil
}
