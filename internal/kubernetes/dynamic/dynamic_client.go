package dynamic

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

type DynamicClient struct {
	DiscoveryClient *discovery.DiscoveryClient
	Client          dynamic.Interface
}

func NewDynamicClient(config *rest.Config) (*DynamicClient, error) {
	dc := &DynamicClient{}

	if err := dc.createDiscoveryClient(config); err != nil {
		return nil, err
	}

	if err := dc.createDynamicClient(config); err != nil {
		return nil, err
	}

	return dc, nil
}

func (dc *DynamicClient) createDynamicClient(config *rest.Config) error {
	d, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	dc.Client = d
	return nil
}

func (dc *DynamicClient) createDiscoveryClient(config *rest.Config) error {
	c, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return err
	}

	dc.DiscoveryClient = c
	return nil
}

func (dc *DynamicClient) GetGVR(gvk *schema.GroupVersionKind) (*meta.RESTMapping, error) {
	m := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc.DiscoveryClient)) // TODO: Check cache lifecycle and invalidation mechanisms
	return m.RESTMapping(gvk.GroupKind(), gvk.Version)
}
