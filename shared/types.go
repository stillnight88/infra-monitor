package shared

type MetricsPayload struct {
	AgentID   string  `json:"agent_id"`
	Hostname  string  `json:"hostname"`
	CPU       float64 `json:"cpu"`
	RAM       float64 `json:"ram"`
	Disk      float64 `json:"disk"`
	Timestamp int64   `json:"timestamp"`
}
