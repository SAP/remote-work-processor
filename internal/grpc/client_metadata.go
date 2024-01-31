package grpc

import (
	"crypto/tls"
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
	host          string
	port          string
	binaryVersion string
	options       []grpc.DialOption
}

func NewGrpcClientMetadata(host string, port string) *ClientMetadata {
	return &ClientMetadata{
		host: host,
		port: port,
	}
}

func (cm *ClientMetadata) WithClientCertificate() *ClientMetadata {
	//TODO: these will be passed differently in standalone mode
	// either:
	// - passed to stdin
	// - path to key and cert passed as env vars
	// - path to key and cert passed as cmd flags
	clientCert, err := tls.LoadX509KeyPair(CERTIFICATE_MOUTH_PATH+CERTIFICATE_KEY, CERTIFICATE_MOUTH_PATH+PRIVATE_KEY)
	if err != nil {
		log.Fatalf("Could not load client cert: %v\n", err)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
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
