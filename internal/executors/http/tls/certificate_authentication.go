package tls

type CertificateAuthentication struct {
	trustedCerts string
	clientCert   string
	trustAnyCert bool
}

type CertificateAuthenticationOption func(*CertificateAuthentication)

func NewCertAuthentication(opts ...CertificateAuthenticationOption) *CertificateAuthentication {
	auth := &CertificateAuthentication{}

	for _, opt := range opts {
		opt(auth)
	}

	return auth
}

func (ca *CertificateAuthentication) GetTrustedCertificates() string {
	return ca.trustedCerts
}

func (ca *CertificateAuthentication) GetClientCertificate() string {
	return ca.clientCert
}

func (ca *CertificateAuthentication) UseTrustedCertificates() bool {
	return ca.trustedCerts != ""
}

func (ca *CertificateAuthentication) UseClientCertificate() bool {
	return ca.clientCert != ""
}

func (ca *CertificateAuthentication) TrustAnyCertificate() bool {
	return ca.trustAnyCert
}

func TrustCertificates(trustedCerts string) CertificateAuthenticationOption {
	return func(ca *CertificateAuthentication) {
		ca.trustedCerts = trustedCerts
	}
}

func WithClientCertificate(clientCert string) CertificateAuthenticationOption {
	return func(ca *CertificateAuthentication) {
		ca.clientCert = clientCert
	}
}

func TrustAnyCertificate(trustAnyCert bool) CertificateAuthenticationOption {
	return func(ca *CertificateAuthentication) {
		ca.trustAnyCert = trustAnyCert
	}
}
