package log

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/OmineDev/flowers-for-machines/std_server/define"
	"github.com/gin-gonic/gin"
)

func Root(c *gin.Context) {
	c.Writer.WriteString("Hello, World!")
}

func LogRecord(c *gin.Context) {
	var request define.LogRecordRequest

	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusOK, define.LogRecordResponse{
			ResponseID: request.RequestID,
			Success:    false,
			ErrorInfo:  fmt.Sprintf("Failed to parse request; err = %v", err),
		})
		return
	}

	err = writeLog(request)
	if err != nil {
		c.JSON(http.StatusOK, define.LogRecordResponse{
			ResponseID: request.RequestID,
			Success:    false,
			ErrorInfo:  fmt.Sprintf("Write log failed; err = %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, define.LogRecordResponse{
		ResponseID: request.RequestID,
		Success:    true,
	})
}

func LogReview(c *gin.Context) {
	var request define.LogReviewRequest

	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusOK, define.LogReviewResponse{
			ResponseID: request.ReviewRequestID,
			Success:    false,
			ErrorInfo:  fmt.Sprintf("Failed to parse request; err = %v", err),
		})
		return
	}

	if !checkAuth(request.AuthKey) {
		c.JSON(http.StatusOK, define.LogReviewResponse{
			ResponseID: request.ReviewRequestID,
			Success:    false,
			ErrorInfo:  fmt.Sprintf("Auth not pass (provided auth key = %s)", request.AuthKey),
		})
		return
	}

	result := reviewLogs(request)
	resultString := make([]string, 0)
	for _, value := range result {
		jsonBytes, err := json.Marshal(value)
		if err == nil {
			resultString = append(resultString, string(jsonBytes))
		}
	}

	c.JSON(http.StatusOK, define.LogReviewResponse{
		ResponseID: request.ReviewRequestID,
		Success:    true,
		LogRecords: resultString,
	})
}

func SetAuthKey(c *gin.Context) {
	var request define.SetAuthKeyRequest

	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusOK, define.SetAuthKeyResponse{
			ResponseID: request.RequestID,
			Success:    false,
			ErrorInfo:  fmt.Sprintf("Failed to parse request; err = %v", err),
		})
		return
	}

	if !checkAuth(request.Token) {
		c.JSON(http.StatusOK, define.SetAuthKeyResponse{
			ResponseID: request.RequestID,
			Success:    false,
			ErrorInfo:  fmt.Sprintf("Auth not pass (provided token = %s)", request.Token),
		})
		return
	}

	switch request.AuthKeyAction {
	case define.ActionSetAuthKey:
		err = setAuth(request.AuthKeyToSet)
	case define.ActionRemoveAuthKey:
		err = removeAuth(request.AuthKeyToSet)
	default:
		c.JSON(http.StatusOK, define.SetAuthKeyResponse{
			ResponseID: request.RequestID,
			Success:    false,
			ErrorInfo:  fmt.Sprintf("Unknown action type %d was found", request.AuthKeyAction),
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusOK, define.SetAuthKeyResponse{
			ResponseID: request.RequestID,
			Success:    false,
			ErrorInfo:  fmt.Sprintf("Set auth key failed; err = %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, define.SetAuthKeyResponse{
		ResponseID: request.RequestID,
		Success:    true,
	})
}
