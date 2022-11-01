package inferenceimpl

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/opensourceways/xihe-inference-evaluate/client"
	"github.com/opensourceways/xihe-inference-evaluate/domain"
	"github.com/opensourceways/xihe-inference-evaluate/domain/inference"
)

const MetaNameInference = "inference"

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

func NewInference(cfg *Config) inference.Inference {
	return inferenceImpl{
		cfg: cfg,
	}
}

type inferenceImpl struct {
	cfg *Config
}

func (impl inferenceImpl) Create(infer *domain.Inference) error {
	cli := client.GetDyna()
	resource := client.GetResource2()
	res, err := impl.GetObj(impl.cfg, infer)

	res.Object["metadata"] = map[string]interface{}{
		"name":   impl.geneMetaName(&infer.InferenceIndex),
		"labels": impl.GeneLabels(infer),
	}

	dr := cli.Resource(resource).Namespace("default")
	_, err = dr.Create(context.TODO(), res, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (impl inferenceImpl) ExtendExpiry(infer *domain.InferenceIndex, expiry int64) error {
	cli := client.GetDyna()
	resource := client.GetResource2()

	get, err := cli.Resource(resource).Namespace("default").Get(context.TODO(), impl.geneMetaName(infer), metav1.GetOptions{})
	if err != nil {
		return err
	}

	if sp, ok := get.Object["spec"]; ok {
		if spc, ok := sp.(map[string]interface{}); ok {
			spc["add"] = true
			spc["recycleAfterSeconds"] = expiry
		}
	}
	_, err = cli.Resource(resource).Namespace("default").Update(context.TODO(), get, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (impl inferenceImpl) geneMetaName(index *domain.InferenceIndex) string {
	return fmt.Sprintf("%s-%s", MetaNameInference, index.Id)
}

func (impl inferenceImpl) GeneLabels(infer *domain.Inference) map[string]string {
	m := make(map[string]string)
	m["id"] = infer.Id
	m["user"] = infer.Project.Owner.Account()
	m["project_id"] = infer.Project.Id
	m["last_commit"] = infer.LastCommit
	m["type"] = MetaNameInference
	return m
}

func (impl inferenceImpl) GetObj(cfg *Config, infer *domain.Inference) (*unstructured.Unstructured, error) {
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
