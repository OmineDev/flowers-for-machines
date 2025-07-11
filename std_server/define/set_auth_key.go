package define

const (
	ActionSetAuthKey = iota
	ActionRemoveAuthKey
)

type SetAuthKeyRequest struct {
	Token         string `json:"token"`
	AuthKeyAction uint8  `json:"auth_key_action"`
	AuthKeyToSet  string `json:"auth_key_to_set"`
}

type SetAuthKeyResponse struct {
	Success   bool   `json:"success"`
	ErrorInfo string `json:"error_info"`
}
