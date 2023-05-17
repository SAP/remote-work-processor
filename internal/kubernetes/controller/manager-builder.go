package controller

import (
	"log"
	"os"
	"time"

	pb "github.com/SAP/remote-work-processor/build/proto/generated"
	"github.com/SAP/remote-work-processor/internal/cache"
	"github.com/SAP/remote-work-processor/internal/kubernetes/dynamic"
	"github.com/SAP/remote-work-processor/internal/kubernetes/selector"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	setupLog = ctrl.Log.WithName("setup")
)

type ControllerManagerBuilder struct {
	ControllerManager
}

type ManagerBuilder interface {
	WithConfig(config *rest.Config) *ControllerManagerBuilder
	WithOptions(scheme *runtime.Scheme, enableLeaderElection bool) *ControllerManagerBuilder
	WithoutLeaderElection() *ControllerManagerBuilder
	WithWatchConfiguration(wc *pb.UpdateConfigRequestMessage) *ControllerManagerBuilder
	Build() ControllerManager
}

func CreateManagerBuilder() *ControllerManagerBuilder {
	return &ControllerManagerBuilder{
		ControllerManager: ControllerManager{},
	}
}

func (cm *ControllerManagerBuilder) WithConfig(config *rest.Config) *ControllerManagerBuilder {
	cm.config = config
	return cm
}

func (cm *ControllerManagerBuilder) WithOptions(scheme *runtime.Scheme) *ControllerManagerBuilder {
	t := 0 * time.Second
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

func (cm *ControllerManagerBuilder) WithWatchConfiguration(wc *pb.UpdateConfigRequestMessage) *ControllerManagerBuilder {
	cm.watchConfig = wc
	cm.initSelectors(wc.Resources)
	return cm
}

func (cm *ControllerManagerBuilder) initSelectors(rs map[string]*pb.Resource) {
	cm.selectorCache = cache.NewInMemoryCache[string, selector.Selector]()
	for k, r := range rs {
		cm.selectorCache.Write(k, selector.NewSelector(r.GetLabelSelectors(), r.GetFieldSelectors()))
	}
}

func (cm *ControllerManagerBuilder) Build() ControllerManager {
	cm.dynamicClient = buildDynamicClient(cm.config)
	cm.manager = buildInternalManager(cm.config, cm.options)
	return cm.ControllerManager
}

func buildDynamicClient(config *rest.Config) *dynamic.DynamicClient {
	dc, err := dynamic.NewDynamicClient(config)
	if err != nil {
		log.Fatalf("unable to create dynamic client: %v\n", err)
	}

	return dc
}

func buildInternalManager(config *rest.Config, options manager.Options) (mgr manager.Manager) {
	mgr, err := ctrl.NewManager(config, options)

	if err != nil {
		log.Fatalf("unable to start manager: %v\n", err)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	return
}
