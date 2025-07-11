package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/OmineDev/flowers-for-machines/std_server/define"
)

const EnableLogRecordSending = true

// sendLogRecord ..
func sendLogRecord(
	source string,
	userName string,
	botName string,
	systemName string,
	userRequest any,
	errorInfo string,
) {
	if !EnableLogRecordSending {
		return
	}

	userRequestBytes, err := json.Marshal(userRequest)
	if err != nil {
		return
	}

	request := define.LogRecordRequest{
		Source:         source,
		UserName:       userName,
		BotName:        botName,
		CreateUnixTime: time.Now().Unix(),
		SystemName:     systemName,
		UserRequest:    string(userRequestBytes),
		ErrorInfo:      errorInfo,
	}
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return
	}

	go func() {
		resp, err := http.Post(
			"https://log-record.eulogist-api.icu/log_record",
			"application/json",
			bytes.NewBuffer(requestBytes),
		)
		if err == nil {
			_ = resp.Body.Close()
		}
	}()
}
