package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/OmineDev/flowers-for-machines/std_server/define"
	"github.com/google/uuid"
)

const AuthKey = "..."

func main() {
	setAuth()
	reviewLogs()
}

func setAuth() {
	var response define.SetAuthKeyResponse

	request := define.SetAuthKeyRequest{
		RequestID:     uuid.New().String(),
		Token:         AuthKey,
		AuthKeyAction: define.ActionRemoveAuthKey,
		AuthKeyToSet:  "",
	}

	jsonBytes, err := json.Marshal(request)
	if err != nil {
		panic(err)
	}

	resp, err := http.Post(
		"https://log-record.eulogist-api.icu/set_auth_key",
		"application/json",
		bytes.NewBuffer(jsonBytes),
	)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != 200 {
		panic("reviewLogs: resp.StatusCode is not equal to 200")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		panic(err)
	}

	jsonBytes, err = json.MarshalIndent(response, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonBytes))
}

func reviewLogs() {
	var response define.LogReviewResponse

	request := define.LogReviewRequest{
		ReviewRequestID: uuid.New().String(),
		AuthKey:         AuthKey,
		Source:          []string{},
		RequestID:       []string{},
		UserName:        []string{},
		BotName:         []string{},
		StartUnixTime:   0,
		EndUnixTime:     0,
		SystemName:      []string{},
	}

	jsonBytes, err := json.Marshal(request)
	if err != nil {
		panic(err)
	}

	resp, err := http.Post(
		"https://log-record.eulogist-api.icu/log_review",
		"application/json",
		bytes.NewBuffer(jsonBytes),
	)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != 200 {
		panic("reviewLogs: resp.StatusCode is not equal to 200")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		panic(err)
	}

	jsonBytes, err = json.MarshalIndent(response, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonBytes))
}
