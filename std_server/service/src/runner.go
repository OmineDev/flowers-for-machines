package service

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/OmineDev/flowers-for-machines/client"
	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol"
	"github.com/OmineDev/flowers-for-machines/game_control/game_interface"
	"github.com/OmineDev/flowers-for-machines/game_control/resources_control"
	"github.com/OmineDev/flowers-for-machines/nbt_assigner"
	"github.com/OmineDev/flowers-for-machines/nbt_assigner/nbt_cache"
	"github.com/OmineDev/flowers-for-machines/nbt_assigner/nbt_console"

	"github.com/pterm/pterm"
)

var userName string

var (
	mu            *sync.Mutex
	mcClient      *client.Client
	resources     *resources_control.Resources
	gameInterface *game_interface.GameInterface
	console       *nbt_console.Console
	cache         *nbt_cache.NBTCacheSystem
	wrapper       *nbt_assigner.NBTAssigner
)

func init() {
	mu = new(sync.Mutex)
}

func RunServer(
	rentalServerCode string,
	rentalServerPasscode string,
	authServerAddress string,
	authServerToken string,
	standardServerPort int,
	consoleDimensionID int,
	consoleCenterX int,
	consoleCenterY int,
	consoleCenterZ int,
) {
	var err error
	cfg := client.Config{
		AuthServerAddress:    authServerAddress,
		AuthServerToken:      authServerToken,
		RentalServerCode:     rentalServerCode,
		RentalServerPasscode: rentalServerPasscode,
	}

	for {
		c, err := client.LoginRentalServer(cfg)
		if err != nil {
			if strings.Contains(fmt.Sprintf("%v", err), "netease.report.kick.hint") {
				continue
			}
			panic(err)
		}
		mcClient = c
		break
	}

	resources = resources_control.NewResourcesControl(mcClient)
	gameInterface = game_interface.NewGameInterface(resources)
	requestPermission()

	console, err = nbt_console.NewConsole(
		gameInterface,
		uint8(consoleDimensionID),
		protocol.BlockPos{
			int32(consoleCenterX),
			int32(consoleCenterY),
			int32(consoleCenterZ),
		},
	)
	if err != nil {
		panic(err)
	}
	cache = nbt_cache.NewNBTCacheSystem(console)
	wrapper = nbt_assigner.NewNBTAssigner(console, cache)

	runHttpServer(standardServerPort)
}

func requestPermission() {
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	for {
		resp, err := gameInterface.Commands().SendWSCommandWithResp("querytarget @s")
		if err != nil {
			panic(err)
		}

		if resp.SuccessCount == 0 {
			pterm.Warning.Printfln("缺少管理员权限，请给予 %s 管理员权限", gameInterface.GetBotInfo().BotName)
			<-ticker.C
			continue
		}

		break
	}
}
