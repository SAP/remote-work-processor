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
		HttpExecutor: DefaultHttpRequestExecutor(),
		tokenUrl:     tokenUrl,
		user:         user,
		clientCert:   clientCert,
	}
}

func (f *iasTokenFetcher) Fetch() (string, error) {
	p := f.createRequestParameters()

	r, err := f.HttpExecutor.ExecuteWithParameters(p)
	if err != nil {
		return "", err
	}

	return r.Content, nil
}

func (f *iasTokenFetcher) createRequestParameters() *HttpRequestParameters {
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
