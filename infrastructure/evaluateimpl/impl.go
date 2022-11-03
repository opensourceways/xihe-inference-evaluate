package evaluateimpl

import (
	"context"
	"fmt"

	"github.com/opensourceways/xihe-inference-evaluate/client"
	"github.com/opensourceways/xihe-inference-evaluate/domain"
	"github.com/opensourceways/xihe-inference-evaluate/domain/evaluate"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const MetaNameEvaluate = "evaluate"

func NewEvaluate(cfg *Config) evaluate.Evaluate {
	return &evaluateImpl{
		cfg: cfg,
	}
}

type evaluateImpl struct {
	cfg *Config
}

func (impl *evaluateImpl) CreateCustom(ceva *domain.CustomEvaluate) error {
	cli := client.GetDyna()
	resource := client.GetResource2()

	res, err := impl.GetCustomObj(ceva)
	dr := cli.Resource(resource).Namespace(client.CrdNameSpace)

	_, err = dr.Create(context.TODO(), res, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}
func (impl *evaluateImpl) CreateStandard(seva *domain.StandardEvaluate) error {
	cli := client.GetDyna()
	resource := client.GetResource2()

	res, err := impl.GetStandardObj(seva)
	dr := cli.Resource(resource).Namespace(client.CrdNameSpace)

	_, err = dr.Create(context.TODO(), res, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (impl *evaluateImpl) geneMetaName(eva *domain.EvaluateIndex) string {
	return fmt.Sprintf("%s-%s", MetaNameEvaluate, eva.Id)
}

func (impl *evaluateImpl) GeneLabels(eva *domain.EvaluateIndex) map[string]string {
	m := make(map[string]string)
	m["id"] = eva.Id
	m["user"] = eva.Project.Owner.Account()
	m["project_id"] = eva.Project.Id
	m["training_id"] = eva.TrainingId
	m["type"] = MetaNameEvaluate
	return m
}

func (impl *evaluateImpl) GetCustomObj(ceva *domain.CustomEvaluate) (*unstructured.Unstructured, error) {
	name := impl.geneMetaName(&ceva.EvaluateIndex)
	labels := impl.GeneLabels(&ceva.EvaluateIndex)

	var data = &client.CrdData{
		Group:          client.CrdGroup,
		Version:        client.CrdVersion,
		Name:           name,
		NameSpace:      client.CrdNameSpace,
		Image:          impl.cfg.Image,
		ObsAk:          impl.cfg.OBS.AccessKey,
		ObsSk:          impl.cfg.OBS.SecretKey,
		ObsEndPoint:    impl.cfg.OBS.Endpoint,
		ObsUtilPath:    impl.cfg.OBS.OBSUtilPath,
		ObsBucket:      impl.cfg.OBS.Bucket,
		ObsLfsPath:     impl.cfg.OBS.LFSPath,
		StorageSize:    10,
		RecycleSeconds: ceva.SurvivalTime,
		Labels:         labels,
		OBSPath:        ceva.AimPath,
		EvaluateType:   ceva.Type(),
	}
	return client.GetObj(data)
}

func (impl *evaluateImpl) GetStandardObj(seva *domain.StandardEvaluate) (*unstructured.Unstructured, error) {
	name := impl.geneMetaName(&seva.EvaluateIndex)
	labels := impl.GeneLabels(&seva.EvaluateIndex)

	var data = &client.CrdData{
		Group:          client.CrdGroup,
		Version:        client.CrdVersion,
		Name:           name,
		NameSpace:      client.CrdNameSpace,
		Image:          impl.cfg.Image,
		ObsAk:          impl.cfg.OBS.AccessKey,
		ObsSk:          impl.cfg.OBS.SecretKey,
		ObsEndPoint:    impl.cfg.OBS.Endpoint,
		ObsUtilPath:    impl.cfg.OBS.OBSUtilPath,
		ObsBucket:      impl.cfg.OBS.Bucket,
		ObsLfsPath:     impl.cfg.OBS.LFSPath,
		StorageSize:    10,
		RecycleSeconds: seva.SurvivalTime,
		Labels:         labels,
		OBSPath:        seva.LogPath,
		EvaluateType:   seva.Type(),
		LearningScope:  seva.LearningRateScope.String(),
		BatchScope:     seva.BatchSizeScope.String(),
		MomentumScope:  seva.MomentumScope.String(),
	}
	return client.GetObj(data)
}
