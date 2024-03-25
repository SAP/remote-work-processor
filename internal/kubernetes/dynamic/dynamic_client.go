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

type Client struct {
	mapper meta.RESTMapper
	client dynamic.Interface
}

func NewDynamicClient(config *rest.Config) (*Client, error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}
	dc := &Client{}
	dc.mapper = restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(discoveryClient))

	if dc.client, err = dynamic.NewForConfig(config); err != nil {
		return nil, err
	}
	return dc, nil
}

func (dc *Client) GetResourceInterface(resource schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
	return dc.client.Resource(resource)
}

func (dc *Client) GetGVR(gvk *schema.GroupVersionKind) (*meta.RESTMapping, error) {
	return dc.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
}
