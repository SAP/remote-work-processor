package controller

import (
	"github.com/SAP/remote-work-processor/internal/grpc"
	"log"
	"time"

	pb "github.com/SAP/remote-work-processor/build/proto/generated"
	"github.com/SAP/remote-work-processor/internal/kubernetes/dynamic"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type ControllerManagerBuilder struct {
	ControllerManager //TODO: do not embed
}

type ManagerBuilder interface {
	WithConfig(config *rest.Config) *ControllerManagerBuilder
	WithOptions(scheme *runtime.Scheme) *ControllerManagerBuilder
	WithoutLeaderElection() *ControllerManagerBuilder
	SetWatchConfiguration(wc *pb.UpdateConfigRequestMessage)
	Build() ControllerManager
}

func CreateManagerBuilder(client *grpc.RemoteWorkProcessorGrpcClient) *ControllerManagerBuilder {
	return &ControllerManagerBuilder{
		ControllerManager: ControllerManager{
			grpcClient: client,
		},
	}
}

func (cm *ControllerManagerBuilder) WithConfig(config *rest.Config) *ControllerManagerBuilder {
	cm.config = config
	return cm
}

func (cm *ControllerManagerBuilder) WithOptions(scheme *runtime.Scheme) *ControllerManagerBuilder {
	t := time.Duration(0)
	cm.options = manager.Options{
		Scheme:                  scheme,
		GracefulShutdownTimeout: &t,
		WebhookServer:           nil,
		HealthProbeBindAddress:  "localhost:8811",
		MetricsBindAddress:      "0",
	}
	return cm
}

func (cm *ControllerManagerBuilder) WithoutLeaderElection() *ControllerManagerBuilder {
	cm.options.LeaderElection = false
	return cm
}

func (cm *ControllerManagerBuilder) SetWatchConfiguration(wc *pb.UpdateConfigRequestMessage) {
	cm.watchConfig = wc
	cm.InitSelectors(wc.Resources)
}

func (cm *ControllerManagerBuilder) Build() ControllerManager {
	cm.dynamicClient = buildDynamicClient(cm.config)
	cm.manager = buildInternalManager(cm.config, cm.options)
	return cm.ControllerManager
}

func buildDynamicClient(config *rest.Config) *dynamic.Client {
	dc, err := dynamic.NewDynamicClient(config)
	if err != nil {
		log.Fatalln("unable to create dynamic client:", err)
	}
	return dc
}

func buildInternalManager(config *rest.Config, options manager.Options) manager.Manager {
	mgr, err := ctrl.NewManager(config, options)
	if err != nil {
		log.Fatalln("unable to create manager:", err)
		return nil
	}

	// these can only fail if the manager has been started prior to calling them
	mgr.AddHealthzCheck("healthz", healthz.Ping)
	mgr.AddReadyzCheck("readyz", healthz.Ping)
	return mgr
}
