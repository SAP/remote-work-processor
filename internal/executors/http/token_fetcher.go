package http

type TokenFetcher interface {
	Fetch() (string, error)
	createRequestParameters() *HttpRequestParameters
}
