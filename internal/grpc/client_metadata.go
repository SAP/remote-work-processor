package grpc

import (
	"crypto/tls"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	CERTIFICATE_MOUTH_PATH = "/etc/"
	CERTIFICATE_KEY        = "crt"
	PRIVATE_KEY            = "pk"
)

type GrpcClientMetadata struct {
	host    string
	port    string
	options []grpc.DialOption
}

func NewGrpcClientMetadata(host string, port string) *GrpcClientMetadata {
	return &GrpcClientMetadata{
		host:    host,
		port:    port,
		options: make([]grpc.DialOption, 0),
	}
}

func (gm *GrpcClientMetadata) WithClientCertificate() *GrpcClientMetadata {
	clientCert, err := tls.LoadX509KeyPair(CERTIFICATE_MOUTH_PATH+CERTIFICATE_KEY, CERTIFICATE_MOUTH_PATH+PRIVATE_KEY)
	if err != nil {
		log.Fatalf("could not load client cert: %v", err)
	}

	config := &tls.Config{
		Certificates:       []tls.Certificate{clientCert},
		InsecureSkipVerify: false,
	}

	gm.options = append(gm.options, grpc.WithTransportCredentials(credentials.NewTLS(config)))
	return gm
}

func (gm *GrpcClientMetadata) BlockWhenDialing() *GrpcClientMetadata {
	gm.options = append(gm.options, grpc.WithBlock())
	return gm
}
