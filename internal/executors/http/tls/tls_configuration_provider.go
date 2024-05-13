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

	log.Println("TLS transport: trust any certificate:", p.TrustAnyCertificate())
	t.TLSClientConfig.InsecureSkipVerify = p.TrustAnyCertificate()

	if p.UseTrustedCertificates() {
		log.Println("TLS transport: adding trusted certificates...")
		if err := p.trustCertificate(t, p.trustedCerts); err != nil {
			log.Println("TLS transport: failed to add trusted certificate:", err)
			return nil, err
		}
	}

	if p.UseClientCertificate() {
		log.Println("TLS transport: adding client certificate...")
		if err := p.registerClientCertificate(t, p.clientCert); err != nil {
			log.Println("TLS transport: failed to add client certificate:", err)
			return nil, err
		}
	}

	return t, nil
}

func (p *ConfigurationProvider) registerClientCertificate(tr *http.Transport, certs string) error {
	certs = decodeIfBase64(certs)

	cert, err := parseCertificate([]byte(certs))
	if err != nil {
		return err
	}

	tr.TLSClientConfig.Certificates = append(tr.TLSClientConfig.Certificates, cert)
	return nil
}

func parseCertificate(certWithKey []byte) (tls.Certificate, error) {
	log.Println("TLS transport: parsing certificate...")
	cert := tls.Certificate{}
	var err error

	for {
		block, rest := pem.Decode(certWithKey)
		if block == nil {
			break
		}

		if block.Type == "CERTIFICATE" {
			cert.Certificate = append(cert.Certificate, block.Bytes)
		} else {
			if cert.PrivateKey, err = parsePK(block.Bytes); err != nil {
				return tls.Certificate{}, err
			}
		}

		certWithKey = rest
	}

	return cert, nil
}

func parsePK(block []byte) (crypto.PrivateKey, error) {
	log.Println("TLS transport: parsing private key...")
	if pk, err := x509.ParsePKCS8PrivateKey(block); err == nil {
		switch pk.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey, *ed25519.PrivateKey:
			return pk, nil
		default:
			log.Println("TLS transport: failed to parse private key: unrecognized private key format")
			return nil, executors.NewNonRetryableError("Unrecognized private key format")
		}
	}

	if pk, err := x509.ParsePKCS1PrivateKey(block); err == nil {
		return pk, nil
	}

	log.Println("TLS transport: failed to parse private key: unsupported algorithm")
	return nil, executors.NewNonRetryableError("Failed to parse client private key")
}

func (p *ConfigurationProvider) trustCertificate(tr *http.Transport, certs string) error {
	certs = decodeIfBase64(certs)
	ok := p.certPool.AppendCertsFromPEM([]byte(certs))
	if !ok {
		log.Println("TLS transport: failed to register certificate to certificate pool")
		return executors.NewNonRetryableError("Failed to register the trusted certificate")
	}

	tr.TLSClientConfig.RootCAs = p.certPool
	return nil
}

func decodeIfBase64(certs string) string {
	log.Println("TLS transport: decoding certificates...")
	decoded, err := base64.StdEncoding.DecodeString(certs)
	if err != nil {
		log.Println("TLS transport: certificates not in base64 format. Using raw certificate data")
		return certs
	}
	return string(decoded)
}

func ensureCertificatePool() *x509.CertPool {
	pool, err := x509.SystemCertPool()
	if err != nil {
		log.Println("TLS transport: failed to get system certificate pool, a new one will be created")
		pool = x509.NewCertPool()
	}
	return pool
}
