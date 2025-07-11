package define

type CheckAliveResponse struct {
	Alive     bool   `json:"alive"`
	ErrorInfo string `json:"error_info"`
}
