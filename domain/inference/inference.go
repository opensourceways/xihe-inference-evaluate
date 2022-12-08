package inference

import (
	"github.com/opensourceways/xihe-inference-evaluate/domain"
)

type Inference interface {
	Create(*domain.Inference) error
	ExtendSurvivalTime(index *domain.InferenceIndex, time int) error
	NotifyResult(labels map[string]string, status domain.ContainerDetail)
}
