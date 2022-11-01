package evaluateimpl

import "github.com/opensourceways/xihe-inference-evaluate/infrastructure/inferenceimpl"

type Config struct {
	Image string                  `json:"image"  required:"true"`
	OBS   inferenceimpl.OBSConfig `json:"-"`
}
