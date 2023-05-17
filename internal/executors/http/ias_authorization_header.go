package http

import (
	"fmt"

	"github.com/SAP/remote-work-processor/internal/utils/json"
)

const PASSCODE string = "passcode"

type iasAuthorizationHeader struct {
	user    string
	fetcher TokenFetcher
}

func NewIasAuthorizationHeader(tokenUrl, user, clientCert string) AuthorizationHeaderGenerator {
	return &iasAuthorizationHeader{
		user:    user,
		fetcher: NewIasTokenFetcher(tokenUrl, user, clientCert),
	}
}

func (h *iasAuthorizationHeader) Generate() (AuthorizationHeader, error) {
	raw, err := h.fetcher.Fetch()
	if err != nil {
		return nil, err
	}

	parsed := map[string]any{}
	if err := json.FromJson(raw, &parsed); err != nil {
		return nil, err
	}

	pass, prs := parsed[PASSCODE]
	if !prs {
		return nil, fmt.Errorf("passcode does not exist in the http response")
	}

	return NewBasicAuthorizationHeader(h.user, pass.(string)).Generate()
}
