package log

import "github.com/OmineDev/flowers-for-machines/core/minecraft/protocol"

const (
	ReviewStatesUnfinish = iota
	ReviewStatesFinished
)

type LogKey struct {
	LogUniqueID    string `json:"log_unique_id"`
	ReviewStstaes  uint8  `json:"review_states"`
	Source         string `json:"source"`
	UserName       string `json:"user_name"`
	BotName        string `json:"bot_name"`
	CreateUnixTime int64  `json:"create_unix_time"`
	SystemName     string `json:"system_name"`
}

func (l *LogKey) Marshal(io protocol.IO) {
	io.String(&l.LogUniqueID)
	io.Uint8(&l.ReviewStstaes)
	io.String(&l.Source)
	io.String(&l.UserName)
	io.String(&l.BotName)
	io.Int64(&l.CreateUnixTime)
	io.String(&l.SystemName)
}

type LogPayload struct {
	UserRequest string `json:"user_request"`
	ErrorInfo   string `json:"error_info"`
}

func (l *LogPayload) Marshal(io protocol.IO) {
	io.String(&l.UserRequest)
	io.String(&l.ErrorInfo)
}

type FullLogRecord struct {
	LogKey
	LogPayload
}
