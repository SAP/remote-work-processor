package http

type AuthorizationHeaderGenerator interface {
	Generate() (AuthorizationHeader, error)
}
