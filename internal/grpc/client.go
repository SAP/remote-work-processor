package grpc

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	pb "github.com/SAP/remote-work-processor/build/proto/generated"
	meta "github.com/SAP/remote-work-processor/internal/kubernetes/metadata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type RemoteWorkProcessorGrpcClient struct {
	sync.Mutex
	metadata  *ClientMetadata
	stream    pb.RemoteWorkProcessorService_SessionClient
	context   context.Context
	cancelCtx context.CancelFunc
}

func NewClient(metadata meta.RemoteWorkProcessorMetadata, isStandaloneMode bool) *RemoteWorkProcessorGrpcClient {
	return &RemoteWorkProcessorGrpcClient{
		metadata: NewClientMetadata(metadata.AutoPiHost(), metadata.AutoPiPort(), isStandaloneMode).
			WithClientCertificate().
			WithBinaryVersion(metadata.BinaryVersion()).
			BlockWhenDialing(),
	}
}

func (gc *RemoteWorkProcessorGrpcClient) InitSession(baseCtx context.Context, sessionID string) error {
	ctx, cancel := context.WithCancel(baseCtx)
	ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		"X-AutoPilot-SessionId":     sessionID,
		"X-AutoPilot-BinaryVersion": gc.metadata.GetBinaryVersion(),
	}))
	gc.context = ctx
	gc.cancelCtx = cancel

	rpc, err := gc.establishConnection(ctx)
	if err != nil {
		return err
	}
	return gc.startSession(rpc, ctx)
}

func (gc *RemoteWorkProcessorGrpcClient) Send(op *pb.ClientMessage) error {
	select {
	case <-gc.context.Done():
		gc.closeConn()
		return nil
	default:
	}

	gc.Lock()
	defer gc.Unlock()

	if err := gc.stream.Send(op); err != nil {
		gc.closeConn()
		return fmt.Errorf("error occured while sending client message: %v", err)
	}
	return nil
}

func (gc *RemoteWorkProcessorGrpcClient) ReceiveMsg() (*pb.ServerMessage, error) {
	log.Println("Waiting for server message...")
	msg, err := gc.stream.Recv()
	if err == io.EOF {
		log.Println("Server closed the connection.")
		gc.closeConn()
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error occured while receiving message from server: %v", err)
	}
	return msg, nil
}

func (gc *RemoteWorkProcessorGrpcClient) establishConnection(ctx context.Context) (pb.RemoteWorkProcessorServiceClient, error) {
	target := fmt.Sprintf("%s:%s", gc.metadata.GetHost(), gc.metadata.GetPort())
	conn, err := grpc.DialContext(ctx, target, gc.metadata.GetOptions()...)
	if err != nil {
		return nil, fmt.Errorf("could not connect to gRPC server serving at port %s: %v", gc.metadata.GetPort(), err)
	}
	return pb.NewRemoteWorkProcessorServiceClient(conn), nil
}

func (gc *RemoteWorkProcessorGrpcClient) startSession(rpcClient pb.RemoteWorkProcessorServiceClient, ctx context.Context) error {
	stream, err := rpcClient.Session(ctx)
	if err != nil {
		return fmt.Errorf("could not start a session with the server: %v", err)
	}

	gc.stream = stream
	go gc.runHeartbeat()
	return nil
}

func (gc *RemoteWorkProcessorGrpcClient) runHeartbeat() {
	t := time.NewTicker(30 * time.Second)
	defer t.Stop()

Loop:
	for {
		select {
		case <-t.C:
			msg := &pb.ClientMessage{
				Body: &pb.ClientMessage_ProbeSession{
					ProbeSession: &pb.ProbeSessionMessage{},
				},
			}
			if err := gc.Send(msg); err != nil {
				break
			}
		case <-gc.context.Done():
			break Loop
		}
	}
}

func (gc *RemoteWorkProcessorGrpcClient) closeConn() {
	gc.stream.CloseSend()
	gc.cancelCtx()
}
