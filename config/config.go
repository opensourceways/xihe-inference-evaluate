package config

import (
	"github.com/opensourceways/community-robot-lib/utils"

	"github.com/opensourceways/xihe-inference-evaluate/infrastructure/evaluateimpl"
	"github.com/opensourceways/xihe-inference-evaluate/infrastructure/inferenceimpl"
	"github.com/opensourceways/xihe-inference-evaluate/infrastructure/watchimpl"
	"github.com/opensourceways/xihe-inference-evaluate/k8sclient"
)

func LoadConfig(path string, cfg interface{}) error {
	if err := utils.LoadFromYaml(path, cfg); err != nil {
		return err
	}

	if f, ok := cfg.(ConfigSetDefault); ok {
		f.SetDefault()
	}

	if f, ok := cfg.(ConfigValidate); ok {
		if err := f.Validate(); err != nil {
			return err
		}
	}

	return nil
}

type ConfigValidate interface {
	Validate() error
}

type ConfigSetDefault interface {
	SetDefault()
}

type Config struct {
	Inference inferenceimpl.Config `json:"inference"  required:"true"`
	Evaluate  evaluateimpl.Config  `json:"evaluate"   required:"true"`
	Watch     watchimpl.Config     `json:"watch"      required:"true"`
	K8sClient k8sclient.Config     `json:"k8SClient"  required:"true"`
}

func (cfg *Config) configItems() []interface{} {
	return []interface{}{
		&cfg.Inference,
		&cfg.Evaluate,
		&cfg.Watch,
		&cfg.K8sClient,
	}
}

func (cfg *Config) SetDefault() {
	items := cfg.configItems()
	for _, i := range items {
		if f, ok := i.(ConfigSetDefault); ok {
			f.SetDefault()
		}
	}

	cfg.Evaluate.OBS = cfg.Inference.OBS
}

func (cfg *Config) Validate() error {
	if _, err := utils.BuildRequestBody(cfg, ""); err != nil {
		return err
	}

	items := cfg.configItems()
	for _, i := range items {
		if f, ok := i.(ConfigValidate); ok {
			if err := f.Validate(); err != nil {
				return err
			}
		}
	}

	return nil
}
