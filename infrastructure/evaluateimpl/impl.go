package evaluateimpl

import (
	"github.com/opensourceways/xihe-inference-evaluate/domain"
	"github.com/opensourceways/xihe-inference-evaluate/domain/evaluate"
)

const MetaNameInference = "inference"

func NewEvaluate(cfg *Config) evaluate.Evaluate {
	return &evaluateImpl{}
}

type evaluateImpl struct {
}

func (impl *evaluateImpl) CreateCustom(*domain.CustomEvaluate) error     { return nil }
func (impl *evaluateImpl) CreateStandard(*domain.StandardEvaluate) error { return nil }
