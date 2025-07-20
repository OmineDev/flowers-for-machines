package main

import (
	"time"

	"github.com/OmineDev/flowers-for-machines/client"
	"github.com/OmineDev/flowers-for-machines/game_control/game_interface"
	"github.com/OmineDev/flowers-for-machines/game_control/resources_control"

	"github.com/pterm/pterm"
)

func SystemTestingLogin() {
	var err error
	tA := time.Now()

	cfg := client.Config{
		AuthServerAddress:    "...",
		AuthServerToken:      "...",
		RentalServerCode:     "48285363",
		RentalServerPasscode: "",
	}

	c, err = client.LoginRentalServer(cfg)
	if err != nil {
		panic(err)
	}
	resources = resources_control.NewResourcesControl(c)
	api = game_interface.NewGameInterface(resources)

	pterm.Success.Printfln("SystemTestingLogin: PASS (Time used = %v)", time.Since(tA))
}
