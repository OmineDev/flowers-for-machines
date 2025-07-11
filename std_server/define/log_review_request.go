package define

type LogReviewRequest struct {
	AuthKey         string   `json:"auth_key"`
	IncludeFinished bool     `json:"include_finished"`
	Source          []string `json:"source"`
	LogUniqueID     []string `json:"log_unique_id"`
	UserName        []string `json:"user_name"`
	BotName         []string `json:"bot_name"`
	StartUnixTime   int64    `json:"start_unix_time"`
	EndUnixTime     int64    `json:"end_unix_time"`
	SystemName      []string `json:"system_name"`
}

type LogReviewResponse struct {
	Success    bool     `json:"success"`
	ErrorInfo  string   `json:"error_info"`
	LogRecords []string `json:"log_records"`
}
