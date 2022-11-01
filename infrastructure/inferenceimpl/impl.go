package inferenceimpl

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/opensourceways/xihe-inference-evaluate/client"
	"github.com/opensourceways/xihe-inference-evaluate/domain"
	"github.com/opensourceways/xihe-inference-evaluate/domain/inference"
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
	cli := client.GetDyna()
	resource := client.GetResource2()
	res, err := client.GetObj(impl.cfg, infer)

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
