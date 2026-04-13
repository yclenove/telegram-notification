package model

type NotifyRequest struct {
	Title     string            `json:"title"`
	Message   string            `json:"message"`
	Level     string            `json:"level"`
	Source    string            `json:"source"`
	EventTime string            `json:"event_time"`
	EventID   string            `json:"event_id"`
	Labels    map[string]string `json:"labels"`
}

// NotifyResponse 为 /notify 的标准响应结构。
type NotifyResponse struct {
	RequestID string `json:"request_id"`
	Status    string `json:"status"`
}
