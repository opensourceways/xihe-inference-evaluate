package k8sclient

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	k8sConfig *rest.Config
	k8sClient *kubernetes.Clientset
	resource  dynamic.NamespaceableResourceInterface
)

func Init(cfg *Config) (err error) {
	k8sConfig, err = clientcmd.BuildConfigFromFlags("", cfg.KubeConfigFile)
	if err != nil {
		return
	}

	k8sClient, err = kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return
	}

	dyna, err := dynamic.NewForConfig(k8sConfig)
	if err != nil {
		return
	}

	dis, err := discovery.NewDiscoveryClientForConfig(k8sConfig)
	if err != nil {
		return
	}

	restm := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dis))

	k := schema.GroupVersionKind{
		Group:   cfg.Group,
		Version: cfg.Version,
		Kind:    cfg.Kind,
	}

	mapping, err := restm.RESTMapping(k.GroupKind(), k.Version)
	if err != nil {
		return
	}

	resource = dyna.Resource(mapping.Resource)

	return
}

func GetClient() *kubernetes.Clientset {
	return k8sClient
}

func GetResource() dynamic.NamespaceableResourceInterface {
	return resource
}

func GetNamespace(ns string) dynamic.ResourceInterface {
	return resource.Namespace(ns)
}
