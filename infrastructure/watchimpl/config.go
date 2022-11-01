package watchimpl

type Config struct {
	// RPC Endpoint
	Endpoint string `json:"endpoint" required:"true"`
}
