package http

import "encoding/json"

type OAuthToken struct {
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
	IdToken     string `json:"id_token,omitempty"`
	ExpiresIn   int64  `json:"expires_in,omitempty"`
	issuedAt    int64  `json:"-"`
}

func NewOAuthToken(token string, issuedAt int64) (*OAuthToken, error) {
	oauth := &OAuthToken{}
	if err := json.Unmarshal([]byte(token), oauth); err != nil {
		return nil, err
	}

	oauth.issuedAt = issuedAt
	return oauth, nil
}

func (t OAuthToken) HasValue() bool {
	return t.AccessToken != ""
}
