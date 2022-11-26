package k8sclient

import (
	"bytes"
	"context"
	"io"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
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

	cli.podCli = cli.k8sClient.CoreV1().Pods(cfg.Namespace)
	cli.resource = dyna.Resource(mapping.Resource)
	cli.resourceCli = cli.resource.Namespace(cfg.Namespace)

	return
}

type Client struct {
	k8sClient   *kubernetes.Clientset
	resource    dynamic.NamespaceableResourceInterface
	resourceCli dynamic.ResourceInterface
	podCli      typedcorev1.PodInterface
}

func (cli *Client) GetResource() dynamic.NamespaceableResourceInterface {
	return cli.resource
}

func (cli *Client) CreateCRD(res *unstructured.Unstructured) error {
	_, err := cli.resourceCli.Create(context.TODO(), res, metav1.CreateOptions{})

	return err
}

func (cli *Client) UpdateCRD(res *unstructured.Unstructured) error {
	_, err := cli.resourceCli.Update(context.TODO(), res, metav1.UpdateOptions{})

	return err
}

func (cli *Client) GetCRD(name string) (*unstructured.Unstructured, error) {
	return cli.resourceCli.Get(context.TODO(), name, metav1.GetOptions{})
}

func (cli *Client) DeleteCRD(name string) error {
	return cli.resourceCli.Delete(context.TODO(), name, metav1.DeleteOptions{})
}

func (cli *Client) ListPods() ([]corev1.Pod, error) {
	v, err := cli.podCli.List(
		context.TODO(), metav1.ListOptions{LabelSelector: "app=codeserver"},
	)
	if err != nil {
		return nil, err
	}

	return v.Items, nil
}

func (cli *Client) FailedPodLog(pod *corev1.Pod) (string, error) {
	v, err := cli.getFailedPodLog(pod, false)

	if v1, err1 := cli.getFailedPodLog(pod, true); err1 == nil && v1 != "" {
		if v == "" {
			v = v1
		} else {
			v += ". more detail: " + v1
		}
	}

	return v, err
}

func (cli *Client) getFailedPodLog(pod *corev1.Pod, previous bool) (string, error) {
	var opt *corev1.PodLogOptions
	if previous {
		opt = &corev1.PodLogOptions{Previous: true}
	}

	v := cli.podCli.GetLogs(pod.GetName(), opt)

	s, err := v.Stream(context.TODO())
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)

	if _, err = io.Copy(buf, s); err == nil {
		return buf.String(), nil
	}

	return "", err
}

func (cli *Client) IsPodFailed(pod *corev1.Pod) bool {
	v := pod.Status.ContainerStatuses
	for i := range v {
		if v[i].RestartCount > 0 {
			return true
		}
	}

	return false
}
