package inferenceimpl

import (
	"errors"
	"path/filepath"

	"github.com/opensourceways/xihe-inference-evaluate/infrastructure/config"
)

type Config struct {
	OBS            OBSConfig        `json:"obs"              required:"true"`
	CRD            config.CRDConfig `json:"crd"              required:"true"`
	GitlabEndpoint string           `json:"gitlab_endpiont"  required:"true"`
}

type OBSConfig struct {
	OBSUtilPath string `json:"obsutil_path"  required:"true"`
	AccessKey   string `json:"access_key"    required:"true"`
	SecretKey   string `json:"secret_key"    required:"true"`
	Endpoint    string `json:"endpoint"      required:"true"`
	LFSPath     string `json:"lfs_path"      required:"true"`
	Bucket      string `json:"bucket"        required:"true"`
}

func (c *Config) Validate() error {
	if !filepath.IsAbs(c.OBS.OBSUtilPath) {
		return errors.New("obsutil_path must be an absolute path")
	}

	if filepath.IsAbs(c.OBS.LFSPath) {
		return errors.New("lfs_path can't start with /")
	}

	return nil
}
