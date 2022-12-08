package evaluate

import (
	"github.com/opensourceways/xihe-inference-evaluate/domain"
)

type Evaluate interface {
	CreateCustom(*domain.CustomEvaluate) error
	CreateStandard(*domain.StandardEvaluate) error
	NotifyResult(labels map[string]string, status domain.ContainerDetail)
}
