package engine

import (
	"context"
	pb "github.com/SAP/remote-work-processor/build/proto/generated"
)

type ManagerEngine interface {
	SetWatchConfiguration(wc *pb.UpdateConfigRequestMessage)
	WatchResources(ctx context.Context, isEnabled func() bool) error
	IsRunning() bool
	Stop()
}
