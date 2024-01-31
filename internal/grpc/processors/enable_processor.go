package processors

import (
	"context"
	"fmt"

	pb "github.com/SAP/remote-work-processor/build/proto/generated"
)

type EnableProcessor struct {
	enableFunc func()
}

func NewEnableProcessor(enableFunc func()) EnableProcessor {
	return EnableProcessor{
		enableFunc: enableFunc,
	}
}

func (p EnableProcessor) Process(_ context.Context) (*pb.ClientMessage, error) {
	fmt.Println("Enabling work processor...")

	p.enableFunc()

	return &pb.ClientMessage{
		Body: &pb.ClientMessage_ConfirmEnabled{
			ConfirmEnabled: &pb.ConfirmEnabledMessage{},
		},
	}, nil
}
