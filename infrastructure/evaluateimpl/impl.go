package evaluateimpl

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io/ioutil"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"

	"github.com/opensourceways/xihe-inference-evaluate/domain"
	"github.com/opensourceways/xihe-inference-evaluate/domain/evaluate"
	"github.com/opensourceways/xihe-inference-evaluate/k8sclient"
)

const metaNameEvaluate = "evaluate"

func MetaName() string {
	return metaNameEvaluate
}

func NewEvaluate(cfg *Config, k8sConfig k8sclient.Config) (evaluate.Evaluate, error) {
	txtStr, err := ioutil.ReadFile(cfg.CRD.TemplateFile)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("evaluate").Parse(string(txtStr))
	if err != nil {
		return nil, err
	}

	return &evaluateImpl{
		cfg:         cfg,
		k8sConfig:   k8sConfig,
		crdTemplate: tmpl,
	}, nil
}

type evaluateImpl struct {
	cfg         *Config
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

	ns := k8sclient.GetNamespace(impl.cfg.CRD.CRDNamespace)
	_, err := ns.Create(context.TODO(), res, metav1.CreateOptions{})

	logrus.Debugf(
		"gen crd for custom evaluate:%s in %s, err:%v.",
		s, impl.cfg.CRD.CRDNamespace, err,
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

	ns := k8sclient.GetNamespace(impl.cfg.CRD.CRDNamespace)
	_, err := ns.Create(context.TODO(), res, metav1.CreateOptions{})

	logrus.Debugf(
		"gen crd for standard evaluate:%s in %s, err:%v.",
		s, impl.cfg.CRD.CRDNamespace, err,
	)

	return err
}

func (impl *evaluateImpl) geneMetaName(eva *domain.EvaluateIndex) string {
	return fmt.Sprintf("%s-%s", metaNameEvaluate, eva.Id)
}

func (impl *evaluateImpl) geneLabels(eva *domain.EvaluateIndex) map[string]string {
	m := make(map[string]string)
	m["id"] = eva.Id
	m["user"] = eva.Project.Owner.Account()
	m["project_id"] = eva.Project.Id
	m["training_id"] = eva.TrainingId
	m["type"] = metaNameEvaluate
	return m
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
		NameSpace:      crd.CRDNamespace,
		Image:          crd.CRDImage,
		CPU:            crd.CRDCpuString(),
		Memory:         crd.CRDMemoryString(),
		StorageSize:    10,
		RecycleSeconds: survivalTime,
		Labels:         impl.geneLabels(index),

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
