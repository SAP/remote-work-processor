package http

import (
	"net/http"

	"github.com/SAP/remote-work-processor/internal/executors/http/tls"
)

type iasTokenFetcher struct {
	HttpExecutor
	tokenUrl   string
	user       string
	clientCert string
}

func NewIasTokenFetcher(tokenUrl, user, clientCert string) TokenFetcher {
	return &iasTokenFetcher{
		HttpExecutor: NewDefaultHttpRequestExecutor(),
		tokenUrl:     tokenUrl,
		user:         user,
		clientCert:   clientCert,
	}
}

func (f *iasTokenFetcher) Fetch() (string, error) {
	params, _ := f.createRequestParameters()

	resp, err := f.HttpExecutor.ExecuteWithParameters(params)
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

func (f *iasTokenFetcher) createRequestParameters() (*HttpRequestParameters, error) {
	return NewHttpRequestParameters(http.MethodGet, f.tokenUrl, WithCertificateAuthentication(
		tls.NewCertAuthentication(
			tls.WithClientCertificate(f.clientCert),
		),
	))
}
