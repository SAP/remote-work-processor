package http

import (
	"net/http"

	"github.com/SAP/remote-work-processor/internal/executors/http/tls"
	"github.com/SAP/remote-work-processor/internal/functional"
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
	params, err := f.createRequestParameters()
	if err != nil {
		return "", err
	}

	resp, err := f.HttpExecutor.ExecuteWithParameters(params)
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

func (f *iasTokenFetcher) createRequestParameters() (*HttpRequestParameters, error) {
	opts := []functional.OptionWithError[HttpRequestParameters]{
		WithUrl(f.tokenUrl),
		WithMethod(http.MethodGet),
		WithCertificateAuthentication(
			tls.NewCertAuthentication(
				tls.WithClientCertificate(f.clientCert),
			),
		),
	}
	return NewHttpRequestParameters(opts...)
}
