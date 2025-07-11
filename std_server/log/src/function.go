package log

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/OmineDev/flowers-for-machines/std_server/define"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Root(c *gin.Context) {
	c.Writer.WriteString("Hello, World!")
}

func SetAuthKey(c *gin.Context) {
	var request define.SetAuthKeyRequest

	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusOK, define.SetAuthKeyResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("Failed to parse request; err = %v", err),
		})
		return
	}

	if !checkAuth(request.Token) {
		c.JSON(http.StatusOK, define.SetAuthKeyResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("Auth not pass (provided token = %s)", request.Token),
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
			Success:   false,
			ErrorInfo: fmt.Sprintf("Unknown action type %d was found", request.AuthKeyAction),
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusOK, define.SetAuthKeyResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("Set auth key failed; err = %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, define.SetAuthKeyResponse{
		Success: true,
	})
}

func LogRecord(c *gin.Context) {
	var request define.LogRecordRequest

	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusOK, define.LogRecordResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("Failed to parse request; err = %v", err),
		})
		return
	}

	key := LogKey{
		LogUniqueID:    uuid.New().String(),
		ReviewStstaes:  ReviewStatesUnfinish,
		Source:         request.Source,
		UserName:       request.UserName,
		BotName:        request.BotName,
		CreateUnixTime: request.CreateUnixTime,
		SystemName:     request.SystemName,
	}
	payload := LogPayload{
		UserRequest: request.UserRequest,
		ErrorInfo:   request.ErrorInfo,
	}

	err = saveLog(key, payload)
	if err != nil {
		c.JSON(http.StatusOK, define.LogRecordResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("Save log failed; err = %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, define.LogRecordResponse{
		Success:     true,
		LogUniqueID: key.LogUniqueID,
	})
}

func LogReview(c *gin.Context) {
	var request define.LogReviewRequest

	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusOK, define.LogReviewResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("Failed to parse request; err = %v", err),
		})
		return
	}

	if !checkAuth(request.AuthKey) {
		c.JSON(http.StatusOK, define.LogReviewResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("Auth not pass (provided auth key = %s)", request.AuthKey),
		})
		return
	}

	result := filterLogs(request)
	resultString := make([]string, 0)
	for _, value := range result {
		jsonBytes, err := json.Marshal(value)
		if err == nil {
			resultString = append(resultString, string(jsonBytes))
		}
	}

	c.JSON(http.StatusOK, define.LogReviewResponse{
		Success:    true,
		LogRecords: resultString,
	})
}

func LogFinishReview(c *gin.Context) {
	var request define.LogFinishReviewRequest

	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusOK, define.LogFinishReviewResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("Failed to parse request; err = %v", err),
		})
		return
	}

	if !checkAuth(request.AuthKey) {
		c.JSON(http.StatusOK, define.LogFinishReviewResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("Auth not pass (provided auth key = %s)", request.AuthKey),
		})
		return
	}

	if len(request.LogUniqueID) == 0 {
		c.JSON(http.StatusOK, define.LogFinishReviewResponse{Success: true})
		return
	}

	result := filterLogs(
		define.LogReviewRequest{LogUniqueID: request.LogUniqueID},
	)
	for _, value := range result {
		err = updateReviewStates(value.LogKey, value.LogPayload, ReviewStatesFinished)
		if err != nil {
			c.JSON(http.StatusOK, define.LogFinishReviewResponse{
				Success:   false,
				ErrorInfo: fmt.Sprintf("Failed to set review states; err = %v", err),
			})
			return
		}
	}

	c.JSON(http.StatusOK, define.LogFinishReviewResponse{Success: true})
}
