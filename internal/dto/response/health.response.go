package response

type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp int64             `json:"timestamp"`
	Services  map[string]string `json:"services"`
	Version   string            `json:"version,omitempty"`
}
