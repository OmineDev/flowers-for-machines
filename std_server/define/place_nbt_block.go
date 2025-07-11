package define

const (
	ResponseErrorTypeParseError = iota
	ResponseErrorTypeRuntimeError
)

type PlaceNBTBlockRequest struct {
	BlockName            string `json:"block_name"`
	BlockStatesString    string `json:"block_states_string"`
	BlockNBTBase64String string `json:"block_nbt_base64_string"`
}

type PlaceNBTBlockResponse struct {
	Success   bool   `json:"success"`
	ErrorType int    `json:"error_type"`
	ErrorInfo string `json:"error_info"`

	CanFast           bool   `json:"can_fast"`
	StructureUniqueID string `json:"structure_unique_id"`
	StructureName     string `json:"structure_name"`

	OffsetX int32 `json:"offset_x"`
	OffsetY int32 `json:"offset_y"`
	OffsetZ int32 `json:"offset_z"`
}
