package watchimpl

type Config struct {
	// RPC Endpoint
	InferenceEndpoint string `json:"inference_endpoint" required:"true"`
	EvaluateEndpoint  string `json:"evaluate_endpoint" required:"true"`
}
