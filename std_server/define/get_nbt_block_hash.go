package define

const (
	RequestTypeFullHash = iota
	RequestTypeNBTHash
	RequestTypeContainerSetHash
)

type GetNBTBlockHashRequest struct {
	RequestType          uint8  `json:"request_type"`
	BlockName            string `json:"block_name"`
	BlockStatesString    string `json:"block_states_string"`
	BlockNBTBase64String string `json:"block_nbt_base64_string"`
}

type GetNBTBlockHashResponse struct {
	Success   bool   `json:"success"`
	ErrorInfo string `json:"error_info"`
	Hash      uint64 `json:"hash"`
}
