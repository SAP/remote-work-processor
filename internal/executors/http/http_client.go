package http

import (
	"net/http"
	"time"

	"github.com/SAP/remote-work-processor/internal/executors/http/tls"
)

const (
	DefaultHttpRequestTimeout = 3 * time.Second
)

func CreateHttpClient(timeoutInS uint64, certAuth *tls.CertificateAuthentication) (*http.Client, error) {
	var tp http.RoundTripper
	if certAuth != nil {
		var err error

		tp, err = tls.NewTLSConfigurationProvider(certAuth).CreateTransport()
		if err != nil {
			return nil, err
		}
	}

	c := &http.Client{
		CheckRedirect: doNotFollowRedirects(),
		Transport:     tp,
	}

	if timeoutInS == 0 {
		c.Timeout = DefaultHttpRequestTimeout
	} else {
		c.Timeout = time.Duration(timeoutInS) * time.Second
	}

	return c, nil
}

func doNotFollowRedirects() func(req *http.Request, via []*http.Request) error {
	return func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
}
