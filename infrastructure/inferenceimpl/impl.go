package inferenceimpl

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
	"github.com/opensourceways/xihe-inference-evaluate/domain/inference"
	"github.com/opensourceways/xihe-inference-evaluate/k8sclient"
)

const metaNameInference = "inference"

func MetaName() string {
	return metaNameInference
}

func NewInference(cfg *Config, k8sConfig k8sclient.Config) (inference.Inference, error) {
	txtStr, err := ioutil.ReadFile(cfg.CRD.TemplateFile)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("inference").Parse(string(txtStr))
	if err != nil {
		return nil, err
	}

	return inferenceImpl{
		cfg:         cfg,
		k8sConfig:   k8sConfig,
		crdTemplate: tmpl,
	}, nil
}

type inferenceImpl struct {
	cfg         *Config
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

	ns := k8sclient.GetNamespace(impl.cfg.CRD.CRDNamespace)
	_, err := ns.Create(context.TODO(), res, metav1.CreateOptions{})

	logrus.Debugf(
		"create inference:%s in %s, err:%v.",
		s, impl.cfg.CRD.CRDNamespace, err,
	)

	return err
}

func (impl inferenceImpl) ExtendSurvivalTime(infer *domain.InferenceIndex, timeToExtend int) error {
	s := impl.inferenceIndexString(infer)
	logrus.Debugf("extend inference for %s to %d.", s, timeToExtend)

	ns := k8sclient.GetNamespace(impl.cfg.CRD.CRDNamespace)

	get, err := ns.Get(context.TODO(), impl.geneMetaName(infer), metav1.GetOptions{})
	if err != nil {
		return err
	}

	if sp, ok := get.Object["spec"]; ok {
		if spc, ok := sp.(map[string]interface{}); ok {
			spc["increaseRecycleSeconds"] = true
			spc["recycleAfterSeconds"] = timeToExtend
		}
	}

	_, err = ns.Update(context.TODO(), get, metav1.UpdateOptions{})

	logrus.Debugf("extend inference for %s to %d, err:%v.", s, timeToExtend, err)

	return err
}

func (impl inferenceImpl) geneMetaName(index *domain.InferenceIndex) string {
	return fmt.Sprintf("%s-%s", metaNameInference, index.Id)
}

func (impl inferenceImpl) geneLabels(infer *domain.Inference) map[string]string {
	return map[string]string{
		"id":          infer.Id,
		"type":        metaNameInference,
		"user":        infer.Project.Owner.Account(),
		"project_id":  infer.Project.Id,
		"last_commit": infer.LastCommit,
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
		NameSpace:      crd.CRDNamespace,
		Image:          crd.CRDImage,
		CPU:            crd.CRDCpuString(),
		Memory:         crd.CRDMemoryString(),
		StorageSize:    10,
		RecycleSeconds: infer.SurvivalTime,
		Labels:         impl.geneLabels(infer),

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
