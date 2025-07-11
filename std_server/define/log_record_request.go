package define

import "github.com/OmineDev/flowers-for-machines/core/minecraft/protocol"

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
	RequestID      string `json:"request_id"`
	Source         string `json:"source"`
	UserName       string `json:"user_name"`
	BotName        string `json:"bot_name"`
	CreateUnixTime int64  `json:"create_unix_time"`
	SystemName     string `json:"system_name"`
	UserRequest    string `json:"user_request"`
	ErrorInfo      string `json:"error_info"`
}

type LogRecordResponse struct {
	ResponseID string `json:"response_id"`
	Success    bool   `json:"success"`
	ErrorInfo  string `json:"error_info"`
}

func (l *LogRecordRequest) MarshalKey(io protocol.IO) {
	io.String(&l.RequestID)
	io.String(&l.Source)
	io.String(&l.UserName)
	io.String(&l.BotName)
	io.Int64(&l.CreateUnixTime)
	io.String(&l.SystemName)
}

func (l *LogRecordRequest) MarshalPayload(io protocol.IO) {
	io.String(&l.UserRequest)
	io.String(&l.ErrorInfo)
}
