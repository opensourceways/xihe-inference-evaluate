package inferenceimpl

import (
	"github.com/opensourceways/xihe-inference-evaluate/domain"
	"github.com/opensourceways/xihe-inference-evaluate/domain/inference"
)

func NewInference(cfg *Config) inference.Inference {
	return inferenceImpl{}
}

type inferenceImpl struct {
}

func (impl inferenceImpl) Create(*domain.Inference) error {
	return nil
}

func (impl inferenceImpl) ExtendExpiry(*domain.InferenceIndex, int64) error {
	return nil
}
