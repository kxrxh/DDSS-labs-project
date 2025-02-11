package server

type Request struct {
	Action string         `json:"action"`
	Data   map[string]any `json:"data"`
}

type Response struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Data    any    `json:"data"`
}
