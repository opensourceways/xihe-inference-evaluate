package inference

import (
	"github.com/opensourceways/xihe-inference-evaluate/domain"
)

type Inference interface {
	Create(*domain.Inference) error
	ExtendExpiry(*domain.InferenceIndex, int64) error
}
