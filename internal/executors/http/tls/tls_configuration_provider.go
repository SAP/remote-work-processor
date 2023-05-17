package tls

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"log"
	"net/http"
	"regexp"

	"github.com/SAP/remote-work-processor/internal/executors"
)

const (
	BASE64_ENCODING_PATTERN    = "^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{3}=|[A-Za-z0-9+/]{2}==)?$"
	PEM_CERTIFICATE_BLOCK_TYPE = "CERTIFICATE"
)

type TLSConfigurationProvider struct {
	*CertificateAuthentication
	certPool *x509.CertPool
}

func NewTLSConfigurationProvider(certAuth *CertificateAuthentication) *TLSConfigurationProvider {
	provider := &TLSConfigurationProvider{
		CertificateAuthentication: certAuth,
		certPool:                  ensureCertificatePool(),
	}

	return provider
}

func (p *TLSConfigurationProvider) CreateTransport() (http.RoundTripper, error) {
	t := &http.Transport{
		TLSClientConfig: &tls.Config{},
	}

	if p.TrustAnyCertificate() {
		t.TLSClientConfig.InsecureSkipVerify = p.TrustAnyCertificate()
	} else if p.UseTrustedCertificates() {
		p.trustCertificate(t, p.trustedCerts, "Failed to register the trusted certificate")
	}

	if p.UseClientCertificate() {
		p.registerClientCertificate(t, p.clientCert, "Failed to register the client certificate and its' private key")
	}

	return t, nil
}

func (p *TLSConfigurationProvider) registerClientCertificate(tr *http.Transport, certs string, errMessage string) error {
	certs, err := decodeIfBase64Certificate(certs, errMessage)
	if err != nil {
		return err
	}

	cert, err := parseCertificate([]byte(certs))
	if err != nil {
		return err
	}

	tr.TLSClientConfig.Certificates = append(tr.TLSClientConfig.Certificates, cert)
	return nil
}

func parseCertificate(certs []byte) (tls.Certificate, error) {
	cert := tls.Certificate{}
	var err error

	for {
		b, rest := pem.Decode([]byte(certs))
		if b == nil && len(rest) == 0 {
			log.Println("All PEM blocks have been read")
			break
		}

		if b.Type == PEM_CERTIFICATE_BLOCK_TYPE {
			cert.Certificate = append(cert.Certificate, b.Bytes)
		} else {
			if cert.PrivateKey, err = parsePK(b.Bytes); err != nil {
				return tls.Certificate{}, err
			}
		}

		certs = rest
	}

	return cert, nil
}

func parsePK(block []byte) (crypto.PrivateKey, error) {
	if pk, err := x509.ParsePKCS8PrivateKey(block); err == nil {
		switch pk.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey, *ed25519.PrivateKey:
			return pk, nil
		default:
			return nil, executors.NewNonRetryableError("Unrecognized private key format")
		}
	}

	if pk, err := x509.ParsePKCS1PrivateKey(block); err == nil {
		return pk, nil
	}

	return nil, executors.NewNonRetryableError("Failed to parse client private key")
}

func (p *TLSConfigurationProvider) trustCertificate(tr *http.Transport, certs string, errMessage string) error {
	certs, err := decodeIfBase64Certificate(certs, errMessage)
	if err != nil {
		return err
	}

	ok := p.certPool.AppendCertsFromPEM([]byte(certs))
	if !ok {
		return executors.NewNonRetryableError(errMessage)
	}

	tr.TLSClientConfig.RootCAs = p.certPool
	return nil
}

func decodeIfBase64Certificate(certs string, errMessage string) (string, error) {
	if !isBase64(certs) {
		return certs, nil
	}

	d, err := base64.StdEncoding.DecodeString(certs)
	if err != nil {
		return "", executors.NewNonRetryableError(errMessage)
	}

	return string(d), nil
}

func isBase64(s string) bool {
	return regexp.MustCompile(BASE64_ENCODING_PATTERN).MatchString(s)
}

func ensureCertificatePool() *x509.CertPool {
	pool, err := x509.SystemCertPool()
	if err != nil {
		log.Printf("Failed to get system certificate pool, a new one will be created.")
		pool = x509.NewCertPool()
	}

	return pool
}
