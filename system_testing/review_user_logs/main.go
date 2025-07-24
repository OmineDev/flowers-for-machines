package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/OmineDev/flowers-for-machines/client"
	"github.com/OmineDev/flowers-for-machines/core/minecraft/nbt"
	"github.com/OmineDev/flowers-for-machines/game_control/game_interface"
	"github.com/OmineDev/flowers-for-machines/game_control/resources_control"
	"github.com/OmineDev/flowers-for-machines/nbt_assigner"
	"github.com/OmineDev/flowers-for-machines/nbt_assigner/nbt_cache"
	"github.com/OmineDev/flowers-for-machines/nbt_assigner/nbt_console"
	"github.com/OmineDev/flowers-for-machines/std_server/define"
	log "github.com/OmineDev/flowers-for-machines/std_server/log/src"
	"github.com/OmineDev/flowers-for-machines/utils"
	"github.com/pterm/pterm"
)

var (
	c         *client.Client
	resources *resources_control.Resources
	api       *game_interface.GameInterface
)

var console *nbt_console.Console

func main() {
	tA := time.Now()

	SystemTestingLogin()
	defer func() {
		c.Conn().Close()
		time.Sleep(time.Second)
	}()

	cache := nbt_cache.NewNBTCacheSystem(console)
	assigner := nbt_assigner.NewNBTAssigner(console, cache)

	var nbts []string
	var records []log.FullLogRecord
	fileBytes, err := os.ReadFile("nbts.log")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(fileBytes, &nbts)
	if err != nil {
		panic(err)
	}
	for _, value := range nbts {
		var temp log.FullLogRecord
		err = json.Unmarshal([]byte(value), &temp)
		if err != nil {
			panic(err)
		}
		records = append(records, temp)
	}

	var okOne []string
	var okOneMapping map[string]bool = make(map[string]bool)
	fileBytes, err = os.ReadFile("okOne.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(fileBytes, &okOne)
	if err != nil {
		panic(err)
	}
	for _, value := range okOne {
		okOneMapping[value] = true
	}

	for index, record := range records {
		var request define.PlaceNBTBlockRequest
		var blockNBT map[string]any

		if record.SystemName != define.SystemNamePlaceNBTBlock {
			continue
		}
		if okOneMapping[record.LogUniqueID] {
			continue
		}
		err = json.Unmarshal([]byte(record.UserRequest), &request)
		if err != nil {
			panic(err)
		}

		blockNBTBytes, err := base64.StdEncoding.DecodeString(request.BlockNBTBase64String)
		if err != nil {
			panic(err)
		}
		err = nbt.UnmarshalEncoding(blockNBTBytes, &blockNBT, nbt.LittleEndian)
		if err != nil {
			panic(err)
		}

		_, _, _, err = assigner.PlaceNBTBlock(
			request.BlockName,
			utils.ParseBlockStatesString(request.BlockStatesString),
			blockNBT,
		)
		if err == nil {
			okOne = append(okOne, record.LogUniqueID)
			okOneMapping[record.LogUniqueID] = true
			jsonBytes, err := json.Marshal(okOne)
			if err != nil {
				panic(err)
			}
			err = os.WriteFile("okOne.json", jsonBytes, 0600)
			if err != nil {
				panic(err)
			}
		} else {
			pterm.Warning.Println(record.LogUniqueID)
			pterm.Warning.Printf("%#v\n", err)
		}

		fmt.Println(index+1, len(records))
	}

	pterm.Success.Printfln("System Testing: ALL PASS (Time used = %v)", time.Since(tA))
}

type PlaceNBTBlockRequest struct {
	BlockName            string `json:"block_name"`
	BlockStatesString    string `json:"block_states_string"`
	BlockNBTBase64String string `json:"block_nbt_base64_string"`
}
