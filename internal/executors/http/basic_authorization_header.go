package http

import (
	"encoding/base64"
	"fmt"
	"log"
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
	log.Println("Basic Authorization Header: generating auth header...")
	encoded := base64.StdEncoding.EncodeToString(
		fmt.Appendf(nil, "%s:%s", h.username, h.password),
	)
	return fmt.Sprintf("Basic %s", encoded), nil
}
