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

	"github.com/SAP/remote-work-processor/internal/executors"
)

const (
	PEM_CERTIFICATE_BLOCK_TYPE = "CERTIFICATE"
)

type ConfigurationProvider struct {
	*CertificateAuthentication
	certPool *x509.CertPool
}

func NewTLSConfigurationProvider(certAuth *CertificateAuthentication) *ConfigurationProvider {
	return &ConfigurationProvider{
		CertificateAuthentication: certAuth,
		certPool:                  ensureCertificatePool(),
	}
}

func (p *ConfigurationProvider) CreateTransport() (http.RoundTripper, error) {
	t := &http.Transport{
		TLSClientConfig: &tls.Config{},
	}

	t.TLSClientConfig.InsecureSkipVerify = p.TrustAnyCertificate()

	if p.UseTrustedCertificates() {
		if err := p.trustCertificate(t, p.trustedCerts, "Failed to register the trusted certificate"); err != nil {
			return nil, err
		}
	}

	if p.UseClientCertificate() {
		if err := p.registerClientCertificate(t, p.clientCert); err != nil {
			return nil, err
		}
	}

	return t, nil
}

func (p *ConfigurationProvider) registerClientCertificate(tr *http.Transport, certs string) error {
	certs = decodeIfBase64Certificate(certs)

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
		b, rest := pem.Decode(certs)
		if b == nil && len(rest) == 0 {
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

func (p *ConfigurationProvider) trustCertificate(tr *http.Transport, certs string, errMessage string) error {
	certs = decodeIfBase64Certificate(certs)
	ok := p.certPool.AppendCertsFromPEM([]byte(certs))
	if !ok {
		return executors.NewNonRetryableError(errMessage)
	}

	tr.TLSClientConfig.RootCAs = p.certPool
	return nil
}

func decodeIfBase64Certificate(certs string) string {
	decoded, err := base64.StdEncoding.DecodeString(certs)
	if err != nil {
		return certs
	}
	return string(decoded)
}

func ensureCertificatePool() *x509.CertPool {
	pool, err := x509.SystemCertPool()
	if err != nil {
		log.Println("Failed to get system certificate pool, a new one will be created.")
		pool = x509.NewCertPool()
	}
	return pool
}
