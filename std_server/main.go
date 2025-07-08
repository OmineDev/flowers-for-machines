package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
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

var (
	mcClient      *client.Client
	resources     *resources_control.Resources
	gameInterface *game_interface.GameInterface
	console       *nbt_console.Console
	cache         *nbt_cache.NBTCacheSystem
	wrapper       *nbt_assigner.NBTAssigner
)

var (
	rentalServerCode     *string
	rentalServerPasscode *string
	authServerAddress    *string
	authServerToken      *string
	standardServerPort   *int
	consoleDimensionID   *int
	consoleCenterX       *int
	consoleCenterY       *int
	consoleCenterZ       *int
)

func init() {
	rentalServerCode = flag.String("rsn", "", "The rental server number.")
	rentalServerPasscode = flag.String("rsp", "", "The pass code of the rental server.")
	authServerAddress = flag.String("asa", "", "The auth server address.")
	authServerToken = flag.String("ast", "", "The auth server token.")
	standardServerPort = flag.Int("ssp", 0, "The server port to running.")
	consoleDimensionID = flag.Int("cdi", 0, "The dimension ID of the console. (e.g. overworld = 0, nether = 1, end = 2, dmT = T, etc.)")
	consoleCenterX = flag.Int("ccx", 0, "The X position of the center of the console.")
	consoleCenterY = flag.Int("ccy", 0, "The Y position of the center of the console.")
	consoleCenterZ = flag.Int("ccz", 0, "The Z position of the center of the console.")

	flag.Parse()
	if len(*rentalServerCode) == 0 {
		log.Fatalln("Please provide your rental server number.\n\te.g. -rsn=\"123456\"")
	}
	if len(*authServerAddress) == 0 {
		log.Fatalln("Please provide your auth server address.\n\te.g. -asa=\"http://127.0.0.1\"")
	}
	if *standardServerPort == 0 {
		log.Fatalln("Please provide the server port to running.\n\te.g. -ssp=0")
	}
}

func main() {
	var err error
	cfg := client.Config{
		AuthServerAddress:    *authServerAddress,
		AuthServerToken:      *authServerToken,
		RentalServerCode:     *rentalServerCode,
		RentalServerPasscode: *rentalServerPasscode,
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
		uint8(*consoleDimensionID),
		protocol.BlockPos{
			int32(*consoleCenterX),
			int32(*consoleCenterY),
			int32(*consoleCenterZ),
		},
	)
	if err != nil {
		panic(err)
	}
	cache = nbt_cache.NewNBTCacheSystem(console)
	wrapper = nbt_assigner.NewNBTAssigner(console, cache)

	RunServer()
}

func requestPermission() {
	api := gameInterface.Commands()

	_, err := api.SendWSCommandWithResp("deop @s")
	if err != nil {
		panic(err)
	}

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
