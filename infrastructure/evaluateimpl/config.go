package evaluateimpl

import (
	"github.com/opensourceways/xihe-inference-evaluate/infrastructure/config"
	"github.com/opensourceways/xihe-inference-evaluate/infrastructure/inferenceimpl"
)

type Config struct {
	OBS inferenceimpl.OBSConfig `json:"-"`

	CRD config.CRDConfig `json:"crd" required:"true"`
}
