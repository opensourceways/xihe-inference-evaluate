package config

import (
	"fmt"
	"strconv"
)

type CRDConfig struct {
	CRDImage     string `json:"crd_image"        required:"true"`
	TemplateFile string `json:"crd_template"     required:"true"`
	CRDNamespace string `json:"crd_namespace"    required:"true"`

	// Specifies the terminal container port for connection
	ContainerPort int `json:"container_port"   required:"true"`

	// CrdCpu specifies the number of cpu
	CRDCpu float32 `json:"crd_cpu"               required:"true"`

	// CrdMemory specifies the memory in megabyte.
	CRDMemory int `json:"crd_memory"             required:"true"`
}

func (cfg *CRDConfig) CRDCpuString() string {
	return fmt.Sprintf("%v", cfg.CRDCpu)
}

func (cfg *CRDConfig) CRDMemoryString() string {
	return fmt.Sprintf("%vMi", cfg.CRDMemory)
}

func (cfg *CRDConfig) CRDContainerPortString() string {
	return strconv.Itoa(cfg.ContainerPort)
}

func (cfg *CRDConfig) SetDefault() {
	if cfg.ContainerPort <= 0 {
		cfg.ContainerPort = 8080
	}
}
