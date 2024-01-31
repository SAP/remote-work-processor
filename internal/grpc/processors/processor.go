package processors

import (
	"context"

	pb "github.com/SAP/remote-work-processor/build/proto/generated"
)

type Processor interface {
	Process(ctx context.Context) (*pb.ClientMessage, error)
}
