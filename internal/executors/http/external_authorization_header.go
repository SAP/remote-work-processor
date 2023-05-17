package http

type externalAuthorizationHeader struct {
	value string
}

func NewExternalAuthorizationHeader(v string) AuthorizationHeaderGenerator {
	return &externalAuthorizationHeader{
		value: v,
	}
}

func (h *externalAuthorizationHeader) Generate() (AuthorizationHeader, error) {
	return NewAuthorizationHeaderView(h.value), nil
}
