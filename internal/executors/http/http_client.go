package http

import (
	"log"
	"net/http"
	"time"

	"github.com/SAP/remote-work-processor/internal/executors/http/tls"
)

const (
	DefaultHttpRequestTimeout = 3 * time.Second
)

func CreateHttpClient(timeoutInS uint64, certAuth *tls.CertificateAuthentication) (*http.Client, error) {
	log.Println("Creating HTTP Client...")
	var tp http.RoundTripper
	if certAuth != nil {
		var err error

		log.Println("Creating TLS transport...")
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
	log.Println("HTTP Client: using timeout:", c.Timeout.String())

	return c, nil
}

func doNotFollowRedirects() func(req *http.Request, via []*http.Request) error {
	return func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
}
