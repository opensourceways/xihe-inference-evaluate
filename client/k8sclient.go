package client

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"log"
	"os/user"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	CrdGroup     = "cs.opensourceways.com"
	CrdVersion   = "v1alpha1"
	CrdKind      = "CodeServer"
	CrdNameSpace = "default"
)

type CrdData struct {
	Group          string
	Version        string
	Name           string
	NameSpace      string
	Image          string
	GitlabEndPoint string
	XiheUser       string
	XiheUserToken  string
	ProjectName    string
	LastCommit     string
	ObsAk          string
	ObsSk          string
	ObsEndPoint    string
	ObsUtilPath    string
	ObsBucket      string
	ObsLfsPath     string
	StorageSize    int
	RecycleSeconds int
	Labels         map[string]string
	OBSPath        string
	EvaluateType   string
	LearningScope  string
	BatchScope     string
	MomentumScope  string
}

var (
	k8sConfig *rest.Config
	k8sClient *kubernetes.Clientset
	dyna      dynamic.Interface
	restm     *restmapper.DeferredDiscoveryRESTMapper
)

func getHome() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}

	return u.HomeDir
}

func Init() (err error) {
	k8sConfig, err = clientcmd.BuildConfigFromFlags("", getHome()+"/.kube/config")
	if err != nil {
		log.Println(err)
		return
	}

	k8sClient, err = kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		log.Println(err)
		return
	}
	dyna, err = dynamic.NewForConfig(k8sConfig)
	if err != nil {
		log.Println(err)
		return
	}

	dis, err := discovery.NewDiscoveryClientForConfig(k8sConfig)
	if err != nil {
		log.Println("NewDiscoveryClientForConfig err", err)
		return
	}

	restm = restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dis))
	return nil
}

func GetClient() *kubernetes.Clientset {
	return k8sClient
}

func GetDyna() dynamic.Interface {
	return dyna
}

func GetrestMapper() *restmapper.DeferredDiscoveryRESTMapper {
	return restm
}

func GetK8sConfig() *rest.Config {
	return k8sConfig
}

func GetResource2() schema.GroupVersionResource {
	k := schema.GroupVersionKind{
		Group:   CrdGroup,
		Version: CrdVersion,
		Kind:    CrdKind,
	}
	mapping, _ := GetrestMapper().RESTMapping(k.GroupKind(), k.Version)
	return mapping.Resource
}

func GetObj(data *CrdData) (*unstructured.Unstructured, error) {
	txtStr, err := ioutil.ReadFile("./template/crd-resource.yaml")
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("crd").Parse(string(txtStr))
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if err = tmpl.Execute(buf, data); err != nil {
		return nil, err
	}

	obj := &unstructured.Unstructured{}
	_, _, err = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(buf.Bytes(), nil, obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}
