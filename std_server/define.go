package main

const (
	ResponseErrorTypeParseError = iota
	ResponseErrorTypeRuntimeError
)

// ------------------------- CheckAlive -------------------------

type CheckAliveResponse struct {
	Alive     bool   `json:"alive"`
	ErrorInfo string `json:"error_info"`
}

// ------------------------- ChangeConsolePosition -------------------------

type ChangeConsolePosRequest struct {
	DimensionID uint8 `json:"dimension_id"`
	CenterX     int32 `json:"center_x"`
	CenterY     int32 `json:"center_y"`
	CenterZ     int32 `json:"center_z"`
}

type ChangeConsolePosResponse struct {
	Success   bool   `json:"success"`
	ErrorInfo string `json:"error_info"`
}

// ------------------------- PlaceNBTBlock -------------------------

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

// ------------------------- End -------------------------
