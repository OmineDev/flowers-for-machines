package define

type LogReviewRequest struct {
	ReviewRequestID string   `json:"review_request_id"`
	AuthKey         string   `json:"auth_key"`
	Source          []string `json:"source"`
	RequestID       []string `json:"request_id"`
	UserName        []string `json:"user_name"`
	BotName         []string `json:"bot_name"`
	StartUnixTime   int64    `json:"start_unix_time"`
	EndUnixTime     int64    `json:"end_unix_time"`
	SystemName      []string `json:"system_name"`
}

type LogReviewResponse struct {
	ResponseID string   `json:"response_id"`
	Success    bool     `json:"success"`
	ErrorInfo  string   `json:"error_info"`
	LogRecords []string `json:"log_records"`
}
