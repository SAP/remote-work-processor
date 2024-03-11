package http

type AuthorizationHeaderGenerator interface {
	Generate() (string, error)
}

type CacheableAuthorizationHeaderGenerator interface {
	AuthorizationHeaderGenerator
	GenerateWithCacheAside() (string, error)
}
