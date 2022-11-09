package inferenceimpl

import (
	"errors"
	"path/filepath"

	"github.com/opensourceways/xihe-inference-evaluate/infrastructure/config"
)

type configValidate interface {
	Validate() error
}

type configSetDefault interface {
	SetDefault()
}

type Config struct {
	OBS            OBSConfig        `json:"obs"              required:"true"`
	CRD            config.CRDConfig `json:"crd"              required:"true"`
	GitlabEndpoint string           `json:"gitlab_endpiont"  required:"true"`
}

func (cfg *Config) configItems() []interface{} {
	return []interface{}{
		&cfg.OBS,
		&cfg.CRD,
	}
}

func (cfg *Config) SetDefault() {
	items := cfg.configItems()

	for _, i := range items {
		if f, ok := i.(configSetDefault); ok {
			f.SetDefault()
		}
	}
}

func (cfg *Config) Validate() error {
	items := cfg.configItems()

	for _, i := range items {
		if f, ok := i.(configValidate); ok {
			if err := f.Validate(); err != nil {
				return err
			}
		}
	}

	return nil
}

type OBSConfig struct {
	OBSUtilPath string `json:"obsutil_path"  required:"true"`
	AccessKey   string `json:"access_key"    required:"true"`
	SecretKey   string `json:"secret_key"    required:"true"`
	Endpoint    string `json:"endpoint"      required:"true"`
	LFSPath     string `json:"lfs_path"      required:"true"`
	Bucket      string `json:"bucket"        required:"true"`
}

func (c *OBSConfig) Validate() error {
	if !filepath.IsAbs(c.OBSUtilPath) {
		return errors.New("obsutil_path must be an absolute path")
	}

	if filepath.IsAbs(c.LFSPath) {
		return errors.New("lfs_path can't start with /")
	}

	return nil
}
