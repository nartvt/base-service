package response

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp int64             `json:"timestamp"`
	Services  map[string]string `json:"services,omitempty"`
	Version   string            `json:"version,omitempty"`
}
