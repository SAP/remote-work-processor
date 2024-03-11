package http

import (
	"encoding/base64"
	"fmt"
)

type basicAuthorizationHeader struct {
	username string
	password []byte
}

func NewBasicAuthorizationHeader(u string, p string) AuthorizationHeaderGenerator {
	return &basicAuthorizationHeader{
		username: u,
		password: []byte(p),
	}
}

func (h *basicAuthorizationHeader) Generate() (string, error) {
	str := fmt.Sprintf("%s:%s", h.username, string(h.password))
	encoded := base64.StdEncoding.EncodeToString([]byte(str))

	return fmt.Sprintf("Basic %s", encoded), nil
}
