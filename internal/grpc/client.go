package grpc

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	pb "github.com/SAP/remote-work-processor/build/proto/generated"
	"github.com/SAP/remote-work-processor/internal/grpc/processors"
	meta "github.com/SAP/remote-work-processor/internal/kubernetes/metadata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	HOST string = os.Getenv("AUTOPI_HOSTNAME")
	PORT string = os.Getenv("AUTOPI_PORT")
)

var (
	once   sync.Once
	Client RemoteWorkProcessorGrpcClient
)

type RemoteWorkProcessorGrpcClient struct {
	sync.Mutex
	metadata   *GrpcClientMetadata
	connection *grpc.ClientConn
	context    context.Context
	cancel     context.CancelFunc
	grpcClient pb.RemoteWorkProcessorServiceClient
	stream     pb.RemoteWorkProcessorService_SessionClient
}

func newClient(host string, port string) RemoteWorkProcessorGrpcClient {
	ctx, cf := context.WithCancel(context.Background())
	ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		"X-AutoPilot-SessionId":     meta.Metadata.Id(),
		"X-AutoPilot-BinaryVersion": meta.Metadata.BinaryVersion(),
	}))

	return RemoteWorkProcessorGrpcClient{
		metadata: NewGrpcClientMetadata(host, port).WithClientCertificate().BlockWhenDialing(),
		context:  ctx,
		cancel:   cf,
	}
}

func InitRemoteWorkProcessorGrpcClient() {
	once.Do(func() {
		Client = newClient(HOST, PORT)
		Client.connect()
		Client.openSession()
	})
}

func (gc *RemoteWorkProcessorGrpcClient) Send(op *pb.ClientMessage) {
	gc.Lock()
	defer gc.Unlock()

	if err := gc.stream.Send(op); err != nil {
		log.Fatalf("Error occured while sending client message: %v\n", err)
		gc.stream.CloseSend()
	}
}

func (gc *RemoteWorkProcessorGrpcClient) Receive() <-chan *pb.ServerMessage {
	opChan := make(chan *pb.ServerMessage)
	go func(c chan *pb.ServerMessage) {
		log.Println("Waiting to receive protocol message...")
		for {
			m, recvErr := gc.stream.Recv()
			if recvErr == io.EOF {
				log.Print("Server closed the connection. Bye!")
				gc.stream.CloseSend()
				break
			}

			if recvErr != nil {
				log.Fatalf("Error occured while receiving message from server: %v\n", recvErr)
			}

			c <- m
		}
	}(opChan)

	return opChan
}

func (gc *RemoteWorkProcessorGrpcClient) connect() {
	connection, err := grpc.Dial(fmt.Sprintf("%s:%s", gc.metadata.host, gc.metadata.port), gc.metadata.options...)
	if err != nil {
		log.Fatalf("Couldn't connect to gRPC server serving at port %s: %v\n", PORT, err)
	}

	gc.connection = connection
	gc.grpcClient = pb.NewRemoteWorkProcessorServiceClient(connection)
}

func (gc *RemoteWorkProcessorGrpcClient) openSession() {
	if gc.grpcClient == nil {
		log.Fatalln("Connection to the gRPC server failed and client has no been initialized. Failed to open session")
	}

	stream, err := gc.grpcClient.Session(gc.context)
	if err != nil {
		log.Fatalf("Could not fetch resources watch config from the server: %v\n", err)
	}

	gc.stream = stream

	go func() {
		p := processors.Factory.CreateProbeSessionProcessor()
		for {
			res := <-p.Process()
			if res.Err != nil {
				log.Fatalf("Error occured while sending heartbeat to backend: %v\n", res.Err)
				close(res.Done)
			}

			gc.Send(res.Result)
		}
	}()
}
