package middleware

// ResponseData represents the standard API response structure
type ResponseData struct {
	Status    int         `json:"status"`
	Message   string      `json:"message"`
	Result    interface{} `json:"result"`
	Timestamp string      `json:"timestamp"`
	Path      string      `json:"path"`
}
