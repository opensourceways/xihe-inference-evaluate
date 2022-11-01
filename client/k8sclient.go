package client

import (
	"fmt"
	"github.com/opensourceways/xihe-inference-evaluate/domain"
	"github.com/opensourceways/xihe-inference-evaluate/infrastructure/inferenceimpl"
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

const BaseTemplate = `
	{
		"apiVersion": "%s/%s",
    	"kind": "CodeServer",
    	"metadata": {
    		"name": ,
    		"namespace": %s
    	},
    	"spec": {
    		"runtime": "generic",
    		"subdomain": ,
    		"image": "swr.cn-north-4.myhuaweicloud.com/opensourceway/xihe/gradio:51e18ee8a8468a766f3c1958c2ce5274bdd11175",
    		"storageSize": "%sGi",
    		"storageName": "emptyDir",
    		"inactiveAfterSeconds": 0,
    		"recycleAfterSeconds": %d,
    		"restartPolicy": "Never",
    		"resources": {
    			"requests": {
    			"cpu": "0.5",
    			"memory": "512Mi"
    		}
			},
    	"connectProbe": "/",
    	"workspaceLocation": "/workspace",
    	"envs": [
		{
			"name": "GRADIO_SERVER_PORT",
			"value": "8080"
		},
		{
			"name": "GRADIO_SERVER_NAME",
			"value": "0.0.0.0"
		},
		{
			"name": "GITLAB_ENDPOINT",
			"value": "%s"
		},
		{
			"name": "XIHE_USER",
			"value": "%s"
		},
		{
			"name": "XIHE_USER_TOKEN",
			"value": "%s"
		},
		{
			"name": "PROJECT_NAME",
			"value": "%s"
		},
		{
			"name": "LAST_COMMIT",
			"value": "%s"
		},
		{
			"name": "OBS_AK",
			"value": "%s"
		},
		{
			"name": "OBS_SK",
			"value": "%s"
		},
		{
			"name": "OBS_ENDPOINT",
			"value": "%s"
		},
		{
			"name": "OBS_UTIL_PATH",
			"value": "%s"
		},
		{
			"name": "OBS_BUCKET",
			"value": "%s"
		},
		{
			"name": "OBS_LFS_PATH",
			"value": "%s"
		},
    	],
    	"command": [
    	"/bin/bash",
    	"-c",
    	"su mindspore\n python3 obs_folder_download.py --source_dir='xihe-obj/projects/%s/%s/inference/' --source_files='%s' --dest='%s' --obs-ak=%s --obs-sk=%s --obs-bucketname=%s --obs-endpoint=%s\n cd /workspace/content\n pip install --upgrade -i https://pypi.tuna.tsinghua.edu.cn/simple pip\npip install -r requirements.txt -i https://pypi.tuna.tsinghua.edu.cn/simple\n python3 app.py"
    		]
    	}
    }
`

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
		Group:   "cs.opensourceways.com",
		Version: "v1alpha1",
		Kind:    "CodeServer",
	}
	mapping, _ := GetrestMapper().RESTMapping(k.GroupKind(), k.Version)
	return mapping.Resource
}

func GetObj(cfg *inferenceimpl.Config, infer *domain.Inference) (*unstructured.Unstructured, error) {
	var yamldata []byte

	yamldata = []byte(fmt.Sprintf(BaseTemplate,
		"cs.opensourceways.com",
		"v1alpha1",
		"default",
		"10",
		60*60*24,
		cfg.GitlabEndpoint,
		infer.User,
		infer.UserToken,
		infer.ProjectName,
		infer.LastCommit,
		cfg.OBS.AccessKey,
		cfg.OBS.SecretKey,
		cfg.OBS.Endpoint,
		cfg.OBS.OBSUtilPath,
		cfg.OBS.Bucket,
		cfg.OBS.LFSPath,
		infer.Project.Owner.Account(),
		infer.ProjectName,
		"files_string",
		"/workspace/content/",
		cfg.OBS.AccessKey,
		cfg.OBS.SecretKey,
		cfg.OBS.Bucket,
		cfg.OBS.Endpoint))
	obj := &unstructured.Unstructured{}
	_, _, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(yamldata, nil, obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}
