package define

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
