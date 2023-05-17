package engine

import (
	pb "github.com/SAP/remote-work-processor/build/proto/generated"
)

type ManagerEngine interface {
	StartManager() error
	StopManager()
	ManagerStartedAtLeastOnce() bool
	WithWatchConfiguration(wc *pb.UpdateConfigRequestMessage)
	WithContext()
}
