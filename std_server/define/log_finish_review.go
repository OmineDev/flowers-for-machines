package define

type LogFinishReviewRequest struct {
	AuthKey     string   `json:"auth_key"`
	LogUniqueID []string `json:"log_unique_id"`
}

type LogFinishReviewResponse struct {
	Success   bool   `json:"success"`
	ErrorInfo string `json:"error_info"`
}
