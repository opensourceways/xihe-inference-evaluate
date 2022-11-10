package k8sclient

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

func Init(cfg *Config) (cli Client, err error) {
	k8sConfig, err := clientcmd.BuildConfigFromFlags("", cfg.KubeConfigFile)
	if err != nil {
		return
	}

	cli.k8sClient, err = kubernetes.NewForConfig(k8sConfig)
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

	cli.resource = dyna.Resource(mapping.Resource)
	cli.namespace = cfg.Namespace

	return
}

type Client struct {
	k8sClient *kubernetes.Clientset
	resource  dynamic.NamespaceableResourceInterface
	namespace string
}

func (cli *Client) GetResource() dynamic.NamespaceableResourceInterface {
	return cli.resource
}

func (cli *Client) getNamespace() dynamic.ResourceInterface {
	return cli.resource.Namespace(cli.namespace)
}

func (cli *Client) GetPodClient() corev1.PodInterface {
	return cli.k8sClient.CoreV1().Pods(cli.namespace)
}

func (cli *Client) CreateCRD(res *unstructured.Unstructured) error {
	ns := cli.getNamespace()

	_, err := ns.Create(context.TODO(), res, metav1.CreateOptions{})

	return err
}

func (cli *Client) UpdateCRD(res *unstructured.Unstructured) error {
	ns := cli.getNamespace()

	_, err := ns.Update(context.TODO(), res, metav1.UpdateOptions{})

	return err
}

func (cli *Client) GetCRD(name string) (*unstructured.Unstructured, error) {
	ns := cli.getNamespace()

	return ns.Get(context.TODO(), name, metav1.GetOptions{})
}

func (cli *Client) DeleteCRD(name string) error {
	ns := cli.getNamespace()

	return ns.Delete(context.TODO(), name, metav1.DeleteOptions{})
}
