package grpc

import (
	"crypto/tls"
	"encoding/base64"
	"github.com/SAP/remote-work-processor/internal/utils"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	CERTIFICATE_MOUTH_PATH = "/etc/auth/"
	CERTIFICATE_KEY        = "crt"
	PRIVATE_KEY            = "pk"
)

type ClientMetadata struct {
	host           string
	port           string
	binaryVersion  string
	options        []grpc.DialOption
	standaloneMode bool
}

func NewClientMetadata(host string, port string, isStandaloneMode bool) *ClientMetadata {
	return &ClientMetadata{
		host:           host,
		port:           port,
		standaloneMode: isStandaloneMode,
	}
}

func (cm *ClientMetadata) WithClientCertificate() *ClientMetadata {
	cert := cm.getClientCert()
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	cm.options = append(cm.options, grpc.WithTransportCredentials(credentials.NewTLS(config)))
	return cm
}

func (cm *ClientMetadata) WithBinaryVersion(version string) *ClientMetadata {
	cm.binaryVersion = version
	return cm
}

func (cm *ClientMetadata) BlockWhenDialing() *ClientMetadata {
	cm.options = append(cm.options, grpc.WithBlock())
	return cm
}

func (cm *ClientMetadata) GetHost() string {
	return cm.host
}

func (cm *ClientMetadata) GetPort() string {
	return cm.port
}

func (cm *ClientMetadata) GetBinaryVersion() string {
	return cm.binaryVersion
}

func (cm *ClientMetadata) GetOptions() []grpc.DialOption {
	return cm.options
}

func (cm *ClientMetadata) getClientCert() tls.Certificate {
	if cm.standaloneMode {
		certChain := utils.GetRequiredEnv("CERT_CHAIN")
		privateKey := utils.GetRequiredEnv("PRIVATE_KEY")

		certChainBytes, err := base64.StdEncoding.DecodeString(certChain)
		if err != nil {
			log.Fatalln("Could not decode certificate chain from environment:", err)
		}

		privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKey)
		if err != nil {
			log.Fatalln("Could not decode private key from environment:", err)
		}

		cert, err := tls.X509KeyPair(certChainBytes, privateKeyBytes)
		if err != nil {
			log.Fatalln("Could not load client certificate from environment:", err)
		}
		return cert
	} else {
		cert, err := tls.LoadX509KeyPair(CERTIFICATE_MOUTH_PATH+CERTIFICATE_KEY, CERTIFICATE_MOUTH_PATH+PRIVATE_KEY)
		if err != nil {
			log.Fatalln("Could not load client certificate from files:", err)
		}
		return cert
	}
}
