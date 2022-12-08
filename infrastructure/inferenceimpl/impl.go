package inferenceimpl

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"strings"

	rpcclient "github.com/opensourceways/xihe-grpc-protocol/grpc/client"
	rpcinference "github.com/opensourceways/xihe-grpc-protocol/grpc/inference"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"

	"github.com/opensourceways/xihe-inference-evaluate/domain"
	"github.com/opensourceways/xihe-inference-evaluate/domain/inference"
	"github.com/opensourceways/xihe-inference-evaluate/k8sclient"
)

const (
	keyId             = "id"
	keyUser           = "user"
	keyProjectId      = "project_id"
	keyLastCommit     = "last_commit"
	metaNameInference = "inference"
)

func MetaName() string {
	return metaNameInference
}

func NewInference(cli *k8sclient.Client, cfg *Config, k8sConfig k8sclient.Config) (
	inference.Inference, error,
) {
	txtStr, err := ioutil.ReadFile(cfg.CRD.TemplateFile)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("inference").Parse(string(txtStr))
	if err != nil {
		return nil, err
	}

	rpcCli, err := rpcclient.NewInferenceClient(cfg.RPCEndpoint)
	if err != nil {
		return nil, fmt.Errorf("new inference rpc client error: %s", err.Error())
	}

	return inferenceImpl{
		cfg:         cfg,
		cli:         cli,
		rpcCli:      rpcCli,
		k8sConfig:   k8sConfig,
		crdTemplate: tmpl,
	}, nil
}

type inferenceImpl struct {
	cfg         *Config
	cli         *k8sclient.Client
	rpcCli      *rpcclient.InferenceClient
	k8sConfig   k8sclient.Config
	crdTemplate *template.Template
}

func (impl *inferenceImpl) inferenceIndexString(e *domain.InferenceIndex) string {
	return fmt.Sprintf(
		"%s/%s/%s, meta name:%s",
		e.Project.Owner.Account(), e.Project.Id,
		e.Id, impl.geneMetaName(e),
	)
}

func (impl inferenceImpl) Create(infer *domain.Inference) error {
	s := impl.inferenceIndexString(&infer.InferenceIndex)
	logrus.Debugf("create inference for %s.", s)

	res := new(unstructured.Unstructured)

	if err := impl.getObj(infer, res); err != nil {
		return err
	}

	err := impl.cli.CreateCRD(res)

	logrus.Debugf(
		"create inference:%s in %s, err:%v.",
		s, impl.k8sConfig.Namespace, err,
	)

	return err
}

func (impl inferenceImpl) ExtendSurvivalTime(infer *domain.InferenceIndex, timeToExtend int) error {
	s := impl.inferenceIndexString(infer)
	logrus.Debugf("extend inference for %s to %d.", s, timeToExtend)

	crd, err := impl.cli.GetCRD(impl.geneMetaName(infer))
	if err != nil {
		return err
	}

	if sp, ok := crd.Object["spec"]; ok {
		if spc, ok := sp.(map[string]interface{}); ok {
			spc["increaseRecycleSeconds"] = true
			spc["recycleAfterSeconds"] = timeToExtend
		}
	}

	err = impl.cli.UpdateCRD(crd)

	logrus.Debugf("extend inference for %s to %d, err:%v.", s, timeToExtend, err)

	return err
}

func (impl inferenceImpl) NotifyResult(labels map[string]string, status domain.ContainerDetail) {
	index := rpcinference.InferenceIndex{
		Id:         labels[keyId],
		User:       strings.TrimPrefix(labels[keyUser], keyUser),
		ProjectId:  labels[keyProjectId],
		LastCommit: labels[keyLastCommit],
	}

	info := rpcinference.InferenceInfo{
		Error:     status.ErrorMsg,
		AccessURL: status.AccessUrl,
	}

	if err := impl.rpcCli.SetInferenceInfo(&index, &info); err != nil {
		logrus.Errorf("call inference rpc error:%s", err.Error())
	} else {
		logrus.Debugf(
			"call rpc to set inference(%s/%s/%s/%s) info:(%s/%s)",
			index.User, index.ProjectId, index.LastCommit, index.Id,
			info.Error, info.AccessURL,
		)
	}
}

func (impl inferenceImpl) geneMetaName(index *domain.InferenceIndex) string {
	return fmt.Sprintf("%s-%s", metaNameInference, index.Id)
}

func (impl inferenceImpl) geneLabels(infer *domain.Inference) map[string]string {
	return map[string]string{
		"type":        metaNameInference,
		keyId:         infer.Id,
		keyUser:       keyUser + infer.Project.Owner.Account(),
		keyProjectId:  infer.Project.Id,
		keyLastCommit: infer.LastCommit,
	}
}

func (impl inferenceImpl) getObj(
	infer *domain.Inference, obj *unstructured.Unstructured,
) error {
	crd := &impl.cfg.CRD
	obs := &impl.cfg.OBS
	k8sConfig := &impl.k8sConfig

	data := &crdData{
		Group:          k8sConfig.Group,
		Version:        k8sConfig.Version,
		CodeServer:     k8sConfig.Kind,
		Name:           impl.geneMetaName(&infer.InferenceIndex),
		NameSpace:      k8sConfig.Namespace,
		Image:          crd.CRDImage,
		CPU:            crd.CRDCpuString(),
		Memory:         crd.CRDMemoryString(),
		StorageSize:    20,
		RecycleSeconds: infer.SurvivalTime,
		Labels:         impl.geneLabels(infer),
		ContainerPort:  crd.CRDContainerPortString(),

		GitlabEndPoint: impl.cfg.GitlabEndpoint,
		XiheUser:       infer.Project.Owner.Account(),
		XiheUserToken:  infer.UserToken,
		ProjectName:    infer.ProjectName.ProjectName(),
		LastCommit:     infer.LastCommit,

		ObsAk:       obs.AccessKey,
		ObsSk:       obs.SecretKey,
		ObsEndPoint: obs.Endpoint,
		ObsUtilPath: obs.OBSUtilPath,
		ObsBucket:   obs.Bucket,
		ObsLfsPath:  obs.LFSPath,
	}

	return data.genTemplate(impl.crdTemplate, obj)
}

type crdData struct {
	Group          string
	Version        string
	CodeServer     string
	Name           string
	NameSpace      string
	Image          string
	CPU            string
	Memory         string
	StorageSize    int
	RecycleSeconds int
	Labels         map[string]string
	ContainerPort  string

	GitlabEndPoint string
	XiheUser       string
	XiheUserToken  string
	ProjectName    string
	LastCommit     string

	ObsAk       string
	ObsSk       string
	ObsEndPoint string
	ObsUtilPath string
	ObsBucket   string
	ObsLfsPath  string
}

func (data *crdData) genTemplate(tmpl *template.Template, obj *unstructured.Unstructured) error {
	buf := new(bytes.Buffer)

	if err := tmpl.Execute(buf, data); err != nil {
		return err
	}

	_, _, err := yaml.NewDecodingSerializer(
		unstructured.UnstructuredJSONScheme,
	).Decode(buf.Bytes(), nil, obj)

	return err
}
