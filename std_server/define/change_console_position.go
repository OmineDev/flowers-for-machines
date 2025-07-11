package define

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
