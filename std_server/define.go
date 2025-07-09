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

// ------------------------- PlaceLargeChest -------------------------

type PlaceLargeChestRequest struct {
	BlockName         string `json:"block_name"`
	BlockStatesString string `json:"block_states_string"`

	PairleadChestStructureExist bool   `json:"pairlead_chest_structure_exist"`
	PairleadChestUniqueID       string `json:"pairlead_chest_unique_id"`

	PairedChestStructureExist bool   `json:"paired_chest_structure_exist"`
	PairedChestUniqueID       string `json:"paired_chest_unique_id"`

	PairedChestOffsetX int32 `json:"paired_chest_offset_x"`
	PairedChestOffsetZ int32 `json:"paired_chest_offset_z"`
}

type PlaceLargeChestResponse struct {
	Success   bool   `json:"success"`
	ErrorInfo string `json:"error_info"`

	StructureUniqueID string `json:"structure_unique_id"`
	StructureName     string `json:"structure_name"`
}

// ------------------------- PlaceWaterloggedBlock -------------------------

type PlaceWaterloggedBlockRequest struct {
	NBTBlockName               string `json:"nbt_block_name"`
	NBTBlockStatesString       string `json:"nbt_block_states_string"`
	OffsetNBTBlockStatesString string `json:"offset_nbt_block_states_string"`

	WaterBlockName         string `json:"water_block_name"`
	WaterBlockStatesString string `json:"water_block_states_string"`
	WaterStartOffsetX      int32  `json:"water_start_offset_x"`
	WaterStartOffsetZ      int32  `json:"water_start_offset_z"`
	WaterEndOffsetX        int32  `json:"water_end_offset_x"`
	WaterEndOffsetZ        int32  `json:"water_end_offset_z"`

	NBTStructureExist    bool   `json:"nbt_structure_exist"`
	NBTStructureUniqueID string `json:"nbt_structure_unique_id"`
	NBTStructureOffsetX  int32  `json:"nbt_structure_offset_x"`
	NBTStructureOffsetZ  int32  `json:"nbt_structure_offset_z"`
}

type PlaceWaterloggedBlockResponse struct {
	Success   bool   `json:"success"`
	ErrorInfo string `json:"error_info"`

	StructureUniqueID string `json:"structure_unique_id"`
	StructureName     string `json:"structure_name"`
}

// ------------------------- End -------------------------
