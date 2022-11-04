package k8sclient

import (
	"bytes"
	"html/template"
	"io/ioutil"

	"github.com/sirupsen/logrus"
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

type CrdData struct {
	Group          string
	Version        string
	CodeServer     string
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
	CPU            string
	Memory         string
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
	resource  schema.GroupVersionResource
	Cfg       *Config
)

func Init(cfg *Config) (err error) {
	Cfg = cfg
	k8sConfig, err = clientcmd.BuildConfigFromFlags("", cfg.KubeConfigFile)
	if err != nil {
		logrus.Error(err)
		return
	}

	k8sClient, err = kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		logrus.Error(err)
		return
	}
	dyna, err = dynamic.NewForConfig(k8sConfig)
	if err != nil {
		logrus.Error(err)
		return
	}

	dis, err := discovery.NewDiscoveryClientForConfig(k8sConfig)
	if err != nil {
		logrus.Errorf("NewDiscoveryClientForConfig err:%s", err)
		return
	}

	restm = restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dis))

	k := schema.GroupVersionKind{
		Group:   cfg.Group,
		Version: cfg.Version,
		Kind:    cfg.Kind,
	}
	mapping, err := GetrestMapper().RESTMapping(k.GroupKind(), k.Version)
	if err != nil {
		logrus.Errorf("init resource err:%s", err)
	}
	resource = mapping.Resource
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

func GetResource() schema.GroupVersionResource {
	return resource
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
