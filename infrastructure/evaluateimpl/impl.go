package evaluateimpl

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"strings"

	rpcclient "github.com/opensourceways/xihe-grpc-protocol/grpc/client"
	rpcevaluate "github.com/opensourceways/xihe-grpc-protocol/grpc/evaluate"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"

	"github.com/opensourceways/xihe-inference-evaluate/domain"
	"github.com/opensourceways/xihe-inference-evaluate/domain/evaluate"
	"github.com/opensourceways/xihe-inference-evaluate/k8sclient"
)

const (
	keyId            = "id"
	keyUser          = "user"
	keyProjectId     = "project_id"
	keyTrainingId    = "training_id"
	metaNameEvaluate = "evaluate"
)

func MetaName() string {
	return metaNameEvaluate
}

func NewEvaluate(cli *k8sclient.Client, cfg *Config, k8sConfig k8sclient.Config) (evaluate.Evaluate, error) {
	txtStr, err := ioutil.ReadFile(cfg.CRD.TemplateFile)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("evaluate").Parse(string(txtStr))
	if err != nil {
		return nil, err
	}

	rpcCli, err := rpcclient.NewEvaluateClient(cfg.RPCEndpoint)
	if err != nil {
		return nil, fmt.Errorf("new evaluate rpc client error: %s", err.Error())
	}

	return &evaluateImpl{
		cfg:         cfg,
		cli:         cli,
		rpcCli:      rpcCli,
		k8sConfig:   k8sConfig,
		crdTemplate: tmpl,
	}, nil
}

type evaluateImpl struct {
	cfg         *Config
	cli         *k8sclient.Client
	rpcCli      *rpcclient.EvaluateClient
	k8sConfig   k8sclient.Config
	crdTemplate *template.Template
}

func (impl *evaluateImpl) evaluateIndexString(e *domain.EvaluateIndex) string {
	return fmt.Sprintf(
		"%s/%s/%s/%s, meta name:%s",
		e.Project.Owner.Account(), e.Project.Id,
		e.TrainingId, e.Id, impl.geneMetaName(e),
	)
}

func (impl *evaluateImpl) CreateCustom(ce *domain.CustomEvaluate) error {
	s := impl.evaluateIndexString(&ce.EvaluateIndex)
	logrus.Debugf("receive custom evaluate for %s.", s)

	res := new(unstructured.Unstructured)

	if err := impl.getCustomObj(ce, res); err != nil {
		return err
	}

	err := impl.cli.CreateCRD(res)

	logrus.Debugf(
		"gen crd for custom evaluate:%s in %s, err:%v.",
		s, impl.k8sConfig.Namespace, err,
	)

	return err
}

func (impl *evaluateImpl) CreateStandard(se *domain.StandardEvaluate) error {
	s := impl.evaluateIndexString(&se.EvaluateIndex)
	logrus.Debugf("receive standard evaluate for %s.", s)

	res := new(unstructured.Unstructured)

	if err := impl.getStandardObj(se, res); err != nil {
		return err
	}

	err := impl.cli.CreateCRD(res)

	logrus.Debugf(
		"gen crd for standard evaluate:%s in %s, err:%v.",
		s, impl.k8sConfig.Namespace, err,
	)

	return err
}

func (impl *evaluateImpl) NotifyResult(labels map[string]string, status domain.ContainerDetail) {
	index := rpcevaluate.EvaluateIndex{
		Id:         labels[keyId],
		User:       strings.TrimPrefix(labels[keyUser], keyUser),
		ProjectId:  labels[keyProjectId],
		TrainingID: labels[keyTrainingId],
	}
	info := rpcevaluate.EvaluateInfo{
		Error:     status.ErrorMsg,
		AccessURL: status.AccessUrl,
	}

	if err := impl.rpcCli.SetEvaluateInfo(&index, &info); err != nil {
		logrus.Errorf("call evaluate rpc error: %s", err.Error())
	} else {
		logrus.Debugf(
			"call rpc to set evaluate(%s/%s/%s/%s) info:(%s/%s)",
			index.User, index.ProjectId, index.TrainingID, index.Id,
			info.Error, info.AccessURL,
		)
	}
}

func (impl *evaluateImpl) geneMetaName(eva *domain.EvaluateIndex) string {
	return fmt.Sprintf("%s-%s", metaNameEvaluate, eva.Id)
}

func (impl *evaluateImpl) geneLabels(eva *domain.EvaluateIndex) map[string]string {
	return map[string]string{
		"type":        metaNameEvaluate,
		keyId:         eva.Id,
		keyUser:       keyUser + eva.Project.Owner.Account(),
		keyProjectId:  eva.Project.Id,
		keyTrainingId: eva.TrainingId,
	}
}

func (impl *evaluateImpl) getCustomObj(
	ce *domain.CustomEvaluate,
	obj *unstructured.Unstructured,
) error {
	data := new(crdData)
	impl.genCrdData(data, &ce.EvaluateIndex, ce.SurvivalTime)

	data.OBSPath = ce.AimPath
	data.EvaluateType = ce.Type()

	return data.genTemplate(impl.crdTemplate, obj)
}

func (impl *evaluateImpl) getStandardObj(
	se *domain.StandardEvaluate,
	obj *unstructured.Unstructured,
) error {
	data := new(crdData)
	impl.genCrdData(data, &se.EvaluateIndex, se.SurvivalTime)

	data.OBSPath = se.LogPath
	data.EvaluateType = se.Type()
	data.LearningScope = se.LearningRateScope.String()
	data.BatchScope = se.BatchSizeScope.String()
	data.MomentumScope = se.MomentumScope.String()

	return data.genTemplate(impl.crdTemplate, obj)
}

func (impl *evaluateImpl) genCrdData(
	data *crdData,
	index *domain.EvaluateIndex, survivalTime int,
) {
	crd := &impl.cfg.CRD
	obs := &impl.cfg.OBS
	k8sConfig := &impl.k8sConfig

	*data = crdData{
		Group:          k8sConfig.Group,
		Version:        k8sConfig.Version,
		CodeServer:     k8sConfig.Kind,
		Name:           impl.geneMetaName(index),
		NameSpace:      k8sConfig.Namespace,
		Image:          crd.CRDImage,
		CPU:            crd.CRDCpuString(),
		Memory:         crd.CRDMemoryString(),
		StorageSize:    10,
		RecycleSeconds: survivalTime,
		Labels:         impl.geneLabels(index),
		ContainerPort:  crd.CRDContainerPortString(),

		ObsAk:       obs.AccessKey,
		ObsSk:       obs.SecretKey,
		ObsBucket:   obs.Bucket,
		ObsEndPoint: obs.Endpoint,
		ObsUtilPath: obs.OBSUtilPath,
	}
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

	ObsAk       string
	ObsSk       string
	ObsBucket   string
	ObsEndPoint string
	ObsUtilPath string

	OBSPath       string
	EvaluateType  string
	LearningScope string
	BatchScope    string
	MomentumScope string
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
