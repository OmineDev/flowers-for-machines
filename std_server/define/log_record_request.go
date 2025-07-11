package define

const (
	SourceDefault     = SourceToolDelta
	SystemNameDefault = SystemNamePlaceNBTBlock
)

const (
	SourceOmegaBuilder = "OmegaBuilder"
	SourceToolDelta    = "ToolDelta"
	SourceFunOnBuilder = "FunOnBuilder"
	SourceYsCloud      = "YsCloud"
)

const (
	SystemNameChangeConsolePosition = "ChangeConsolePosition"
	SystemNamePlaceNBTBlock         = "PlaceNBTBlock"
	SystemNamePlaceLargeChest       = "PlaceLargeChest"
	SystemNameGetNBTBlockHash       = "GetNBTBlockHash"
)

type LogRecordRequest struct {
	Source         string `json:"source"`
	UserName       string `json:"user_name"`
	BotName        string `json:"bot_name"`
	CreateUnixTime int64  `json:"create_unix_time"`
	SystemName     string `json:"system_name"`
	UserRequest    string `json:"user_request"`
	ErrorInfo      string `json:"error_info"`
}

type LogRecordResponse struct {
	Success     bool   `json:"success"`
	ErrorInfo   string `json:"error_info"`
	LogUniqueID string `json:"log_unique_id"`
}
