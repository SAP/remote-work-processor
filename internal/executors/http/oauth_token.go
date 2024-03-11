package http

import (
	"encoding/json"
	"fmt"
)

type OAuthToken struct {
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
	IdToken     string `json:"id_token,omitempty"`
	ExpiresIn   int64  `json:"expires_in,omitempty"`
}

func NewOAuthToken(token string) (*OAuthToken, error) {
	oauth := &OAuthToken{}
	if err := json.Unmarshal([]byte(token), oauth); err != nil {
		return nil, fmt.Errorf("failed to parse OAuth token: %v", err)
	}
	return oauth, nil
}
