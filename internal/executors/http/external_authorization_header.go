package http

type externalAuthorizationHeader string

func NewExternalAuthorizationHeader(v string) AuthorizationHeaderGenerator {
	return externalAuthorizationHeader(v)
}

func (h externalAuthorizationHeader) Generate() (AuthorizationHeader, error) {
	return NewAuthorizationHeaderView(string(h)), nil
}
