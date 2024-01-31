package engine

import (
	"context"
	pb "github.com/SAP/remote-work-processor/build/proto/generated"
)

type ManagerEngine interface {
	StartManager(ctx context.Context, isEnabled func() bool) error
	StopManager()
	IsStarted() bool
	SetWatchConfiguration(wc *pb.UpdateConfigRequestMessage)
}
