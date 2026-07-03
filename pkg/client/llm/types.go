package llm

type LLMEndpoint struct {
	EndpointID       string `json:"endpoint_id"`
	EndpointName     string `json:"endpoint_name"`
	Model            string `json:"model"`
	Project          string `json:"project"`
	Status           string `json:"status"`
	CreatedTimestamp string `json:"created_timestamp"`
	EndpointAddress  string `json:"endpoint_address"`
}

var AllowedModelList = []string{"GLM-4.7", "GLM-4.6", "GPT OSS 120b"}

const (
	StatusPending = "pending"
	StatusRunning = "running"
)

func IsAllowedModel(model string) bool {
	for _, m := range AllowedModelList {
		if m == model {
			return true
		}
	}
	return false
}
