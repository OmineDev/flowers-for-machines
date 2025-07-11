package log

import (
	"bytes"
	"cmp"
	"fmt"
	"slices"

	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol"
	"github.com/OmineDev/flowers-for-machines/std_server/define"
	"go.etcd.io/bbolt"
)

const (
	DatabaseFile          = "log_record.db"
	DatabaseAuthKeyBucket = "auth_key"
	DatabseLogBucket      = "logs"
)

var database *bbolt.DB

func init() {
	var err error

	database, err = bbolt.Open(DatabaseFile, 0600, &bbolt.Options{
		FreelistType: bbolt.FreelistMapType,
	})
	if err != nil {
		panic(err)
	}

	err = database.Update(func(tx *bbolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte(DatabaseAuthKeyBucket))
		return err
	})
	if err != nil {
		panic(err)
	}

	err = database.Update(func(tx *bbolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte(DatabseLogBucket))
		return err
	})
	if err != nil {
		panic(err)
	}
}

func checkAuth(key string) (result bool) {
	_ = database.View(func(tx *bbolt.Tx) error {
		payload := tx.Bucket([]byte(DatabaseAuthKeyBucket)).Get([]byte(key))
		if len(payload) != 1 || payload[0] != 1 {
			return nil
		}
		result = true
		return nil
	})
	return
}

func setAuth(key string) error {
	err := database.Update(func(tx *bbolt.Tx) error {
		return tx.
			Bucket([]byte(DatabaseAuthKeyBucket)).
			Put([]byte(key), []byte{1})
	})
	if err != nil {
		return fmt.Errorf("setAuth: %v", err)
	}
	return nil
}

func removeAuth(key string) error {
	err := database.Update(func(tx *bbolt.Tx) error {
		return tx.
			Bucket([]byte(DatabaseAuthKeyBucket)).
			Delete([]byte(key))
	})
	if err != nil {
		return fmt.Errorf("removeAuth: %v", err)
	}
	return nil
}

func saveLog(key LogKey, payload LogPayload) error {
	keyBuf := bytes.NewBuffer(nil)
	key.Marshal(protocol.NewWriter(keyBuf, 0))

	payloadBuf := bytes.NewBuffer(nil)
	payload.Marshal(protocol.NewWriter(payloadBuf, 0))

	err := database.Update(func(tx *bbolt.Tx) error {
		return tx.
			Bucket([]byte(DatabseLogBucket)).
			Put(keyBuf.Bytes(), payloadBuf.Bytes())
	})
	if err != nil {
		return fmt.Errorf("saveLog: %v", err)
	}

	return nil
}

func deleteLog(key LogKey) error {
	buf := bytes.NewBuffer(nil)
	key.Marshal(protocol.NewWriter(buf, 0))

	err := database.Update(func(tx *bbolt.Tx) error {
		return tx.
			Bucket([]byte(DatabseLogBucket)).
			Delete(buf.Bytes())
	})
	if err != nil {
		return fmt.Errorf("deleteLog: %v", err)
	}

	return nil
}

func updateReviewStates(key LogKey, payload LogPayload, newStates uint8) error {
	err := deleteLog(key)
	if err != nil {
		return fmt.Errorf("updateReviewStates: %v", err)
	}

	key.ReviewStstaes = newStates
	err = saveLog(key, payload)
	if err != nil {
		return fmt.Errorf("updateReviewStates: %v", err)
	}

	return nil
}

func filterLogs(request define.LogReviewRequest) []FullLogRecord {
	result := make([]FullLogRecord, 0)

	sourceMapping := make(map[string]bool)
	logUniqueIDMapping := make(map[string]bool)
	userNameMapping := make(map[string]bool)
	botNameMapping := make(map[string]bool)
	systemNameMapping := make(map[string]bool)

	for _, value := range request.Source {
		sourceMapping[value] = true
	}
	for _, value := range request.LogUniqueID {
		logUniqueIDMapping[value] = true
	}
	for _, value := range request.UserName {
		userNameMapping[value] = true
	}
	for _, value := range request.BotName {
		botNameMapping[value] = true
	}
	for _, value := range request.SystemName {
		systemNameMapping[value] = true
	}

	_ = database.View(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte(DatabseLogBucket)).ForEach(func(k, v []byte) error {
			var key LogKey
			var payload LogPayload
			key.Marshal(protocol.NewReader(bytes.NewBuffer(k), 0, false))

			if !request.IncludeFinished && key.ReviewStstaes != ReviewStatesUnfinish {
				return nil
			}
			if len(sourceMapping) > 0 && !sourceMapping[key.Source] {
				return nil
			}
			if len(logUniqueIDMapping) > 0 && !logUniqueIDMapping[key.LogUniqueID] {
				return nil
			}
			if len(userNameMapping) > 0 && !userNameMapping[key.UserName] {
				return nil
			}
			if len(botNameMapping) > 0 && !botNameMapping[key.BotName] {
				return nil
			}
			if request.StartUnixTime != 0 && request.EndUnixTime != 0 {
				if key.CreateUnixTime > request.EndUnixTime || key.CreateUnixTime < request.StartUnixTime {
					return nil
				}
			}
			if len(systemNameMapping) > 0 && !systemNameMapping[key.SystemName] {
				return nil
			}

			payload.Marshal(protocol.NewReader(bytes.NewBuffer(v), 0, false))
			result = append(result, FullLogRecord{
				LogKey:     key,
				LogPayload: payload,
			})

			return nil
		})
	})

	slices.SortStableFunc(result, func(a FullLogRecord, b FullLogRecord) int {
		return cmp.Compare(a.CreateUnixTime, b.CreateUnixTime)
	})
	return result
}
