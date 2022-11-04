package k8sclient

type Config struct {
	Kind           string `json:"kind"             required:"true"`
	Group          string `json:"group"            required:"true"`
	Version        string `json:"version"          required:"true"`
	KubeConfigFile string `json:"kube_config_file" required:"true"`
}

func (cfg *Config) SetDefault() {
	if cfg.KubeConfigFile == "" {
		cfg.KubeConfigFile = "~/.kube/config"
	}
}
