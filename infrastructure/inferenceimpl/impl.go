package inferenceimpl

import (
	"context"
	"fmt"

	"github.com/opensourceways/xihe-inference-evaluate/domain"
	"github.com/opensourceways/xihe-inference-evaluate/domain/inference"
	"github.com/opensourceways/xihe-inference-evaluate/k8sclient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const MetaNameInference = "inference"

func NewInference(cfg *Config) inference.Inference {
	return inferenceImpl{
		cfg: cfg,
	}
}

type inferenceImpl struct {
	cfg *Config
}

func (impl inferenceImpl) Create(infer *domain.Inference) error {
	cli := k8sclient.GetDyna()
	resource := k8sclient.GetResource()

	res, err := impl.GetObj(infer)

	dr := cli.Resource(resource).Namespace(impl.cfg.CRD.CRDNamespace)
	_, err = dr.Create(context.TODO(), res, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}
func (impl inferenceImpl) ExtendSurvivalTime(infer *domain.InferenceIndex, timeToExtend int) error {
	cli := k8sclient.GetDyna()
	resource := k8sclient.GetResource()

	get, err := cli.Resource(resource).Namespace(impl.cfg.CRD.CRDNamespace).Get(
		context.TODO(), impl.geneMetaName(infer), metav1.GetOptions{},
	)
	if err != nil {
		return err
	}

	if sp, ok := get.Object["spec"]; ok {
		if spc, ok := sp.(map[string]interface{}); ok {
			spc["increaseRecycleSeconds"] = true
			spc["recycleAfterSeconds"] = timeToExtend
		}
	}
	_, err = cli.Resource(resource).Namespace(impl.cfg.CRD.CRDNamespace).Update(
		context.TODO(), get, metav1.UpdateOptions{},
	)
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

func (impl inferenceImpl) GetObj(infer *domain.Inference) (*unstructured.Unstructured, error) {
	name := impl.geneMetaName(&infer.InferenceIndex)
	labels := impl.GeneLabels(infer)
	crd := &impl.cfg.CRD

	var data = &k8sclient.CrdData{
		Group:          k8sclient.Cfg.Group,
		Version:        k8sclient.Cfg.Version,
		CodeServer:     k8sclient.Cfg.Kind,
		Name:           name,
		NameSpace:      crd.CRDNamespace,
		Image:          crd.CRDImage,
		GitlabEndPoint: impl.cfg.GitlabEndpoint,
		XiheUser:       infer.Project.Owner.Account(),
		XiheUserToken:  infer.UserToken,
		ProjectName:    infer.ProjectName.ProjectName(),
		LastCommit:     infer.LastCommit,
		ObsAk:          impl.cfg.OBS.AccessKey,
		ObsSk:          impl.cfg.OBS.SecretKey,
		ObsEndPoint:    impl.cfg.OBS.Endpoint,
		ObsUtilPath:    impl.cfg.OBS.OBSUtilPath,
		ObsBucket:      impl.cfg.OBS.Bucket,
		ObsLfsPath:     impl.cfg.OBS.LFSPath,
		StorageSize:    10,
		RecycleSeconds: infer.SurvivalTime,
		CPU:            crd.CRDCpuString(),
		Memory:         crd.CRDMemoryString(),
		Labels:         labels,
	}
	return k8sclient.GetObj(data)
}
