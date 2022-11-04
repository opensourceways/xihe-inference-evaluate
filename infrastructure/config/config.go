package config

import "fmt"

type CRDConfig struct {
	CRDImage     string `json:"crd_image"        required:"true"`
	CRDNamespace string `json:"crd_namespace"    required:"true"`

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
