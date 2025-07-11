package define

const (
	ActionSetAuthKey = iota
	ActionRemoveAuthKey
)

type SetAuthKeyRequest struct {
	RequestID     string `json:"request_id"`
	Token         string `json:"token"`
	AuthKeyAction uint8  `json:"auth_key_action"`
	AuthKeyToSet  string `json:"auth_key_to_set"`
}

type SetAuthKeyResponse struct {
	ResponseID string `json:"response_id"`
	Success    bool   `json:"success"`
	ErrorInfo  string `json:"error_info"`
}
