package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol/packet"

	"github.com/pterm/pterm"
)

func SystemTestingCommands() {
	tA := time.Now()

	// Chat
	{
		channel := make(chan struct{})

		doOnce := new(sync.Once)
		uniqueID, err := api.PacketListener().ListenPacket(
			[]uint32{packet.IDText},
			func(p packet.Packet, connCloseErr error) {
				if connCloseErr != nil {
					panic(fmt.Sprintf("SystemTestingCommands: `SendChat` failed due to %v", connCloseErr))
				}
				if p.(*packet.Text).Message == "System Testing" {
					doOnce.Do(func() { close(channel) })
				}
			},
		)
		if err != nil {
			panic(fmt.Sprintf("SystemTestingCommands: `SendChat` failed due to %v", err))
		}
		api.Commands().SendChat("System Testing")

		timer := time.NewTimer(time.Second * 5)
		defer timer.Stop()
		select {
		case <-timer.C:
			panic("SystemTestingCommands: `SendChat` time out")
		case <-channel:
			api.PacketListener().DestroyListener(uniqueID)
		}
	}

	// AwaitChangesGeneral
	{
		channel := make(chan struct{})

		go func() {
			api.Commands().AwaitChangesGeneral()
			close(channel)
		}()

		timer := time.NewTimer(time.Second * 5)
		defer timer.Stop()
		select {
		case <-timer.C:
			panic("SystemTestingCommands: `AwaitChangesGeneral` time out")
		case <-channel:
		}
	}

	// SendSettingsCommand
	{
		channel := make(chan struct{})

		doOnce := new(sync.Once)
		uniqueID, err := api.PacketListener().ListenPacket(
			[]uint32{packet.IDGameRulesChanged},
			func(p packet.Packet, connCloseErr error) {
				if connCloseErr != nil {
					panic(fmt.Sprintf("SystemTestingCommands: `SendSettingsCommand` failed due to %v", connCloseErr))
				}
				doOnce.Do(func() { close(channel) })
			},
		)
		if err != nil {
			panic(fmt.Sprintf("SystemTestingCommands: `SendSettingsCommand` failed due to %v", err))
		}

		api.Commands().SendSettingsCommand("gamemode 1", false)
		api.Commands().SendSettingsCommand("gamerule sendcommandfeedback false", false)
		time.Sleep(time.Second)
		api.Commands().SendSettingsCommand("gamerule sendcommandfeedback true", false)

		timer := time.NewTimer(time.Second * 5)
		defer timer.Stop()
		select {
		case <-timer.C:
			panic("SystemTestingCommands: `SendSettingsCommand` time out")
		case <-channel:
			api.PacketListener().DestroyListener(uniqueID)
		}
	}

	// SendPlayerCommand
	{
		channel := make(chan struct{})

		doOnce := new(sync.Once)
		uniqueID, err := api.PacketListener().ListenPacket(
			[]uint32{packet.IDText},
			func(p packet.Packet, connCloseErr error) {
				if connCloseErr != nil {
					panic(fmt.Sprintf("SystemTestingCommands: `SendPlayerCommand` failed due to %v", connCloseErr))
				}
				if p.(*packet.Text).Message == "System Testing" {
					doOnce.Do(func() { close(channel) })
				}
			},
		)
		if err != nil {
			panic(fmt.Sprintf("SystemTestingCommands: `SendPlayerCommand` failed due to %v", err))
		}

		api.Commands().SendPlayerCommand(fmt.Sprintf("msg @s %s", "System Testing"))

		timer := time.NewTimer(time.Second * 5)
		defer timer.Stop()
		select {
		case <-timer.C:
			panic("SystemTestingCommands: `SendPlayerCommand` time out")
		case <-channel:
			api.PacketListener().DestroyListener(uniqueID)
		}
	}

	// SendWSCommand
	{
		channel := make(chan struct{})

		doOnce := new(sync.Once)
		uniqueID, err := api.PacketListener().ListenPacket(
			[]uint32{packet.IDText},
			func(p packet.Packet, connCloseErr error) {
				if connCloseErr != nil {
					panic(fmt.Sprintf("SystemTestingCommands: `SendWSCommand` failed due to %v", connCloseErr))
				}
				if p.(*packet.Text).Message == "System Testing" {
					doOnce.Do(func() { close(channel) })
				}
			},
		)
		if err != nil {
			panic(fmt.Sprintf("SystemTestingCommands: `SendWSCommand` failed due to %v", err))
		}
		api.Commands().SendWSCommand(fmt.Sprintf("msg @s %s", "System Testing"))

		timer := time.NewTimer(time.Second * 5)
		defer timer.Stop()
		select {
		case <-timer.C:
			panic("SystemTestingCommands: `SendWSCommand` time out")
		case <-channel:
			api.PacketListener().DestroyListener(uniqueID)
		}
	}

	// SendPlayerCommandWithResp,
	// SendPlayerCommandWithTimeout,
	// SendWSCommand,
	// SendWSCommandWithTimeout
	{
		api.Commands().SendPlayerCommandWithResp("System Testing")

		_, isTimeout, _ := api.Commands().SendPlayerCommandWithTimeout("say System Testing", 0)
		if !isTimeout {
			panic("SystemTestingCommands: `SendPlayerCommandWithTimeout` failed")
		}

		api.Commands().SendWSCommand("System Testing")

		_, isTimeout, _ = api.Commands().SendWSCommandWithTimeout("say System Testing", 0)
		if isTimeout {
			panic("SystemTestingCommands: `SendWSCommandWithTimeout` failed")
		}
	}

	// Title
	{
		channel := make(chan struct{})

		doOnce := new(sync.Once)
		uniqueID, err := api.PacketListener().ListenPacket(
			[]uint32{packet.IDSetTitle},
			func(p packet.Packet, connCloseErr error) {
				if connCloseErr != nil {
					panic(fmt.Sprintf("SystemTestingCommands: `Title` failed due to %v", connCloseErr))
				}
				if strings.Contains(p.(*packet.SetTitle).Text, "System Testing") {
					doOnce.Do(func() { close(channel) })
				}
			},
		)
		if err != nil {
			panic(fmt.Sprintf("SystemTestingCommands: `Title` failed due to %v", err))
		}
		api.Commands().Title("System Testing")

		timer := time.NewTimer(time.Second * 5)
		defer timer.Stop()
		select {
		case <-timer.C:
			panic("SystemTestingCommands: `Title` time out")
		case <-channel:
			api.PacketListener().DestroyListener(uniqueID)
		}
	}

	// Long command
	{
		_, isTimeout, err := api.Commands().SendWSCommandWithTimeout("system testing"+strings.Repeat("-", 256), 0)
		if isTimeout {
			panic("SystemTestingCommands: `Long command` time out")
		}
		if err != nil {
			panic(fmt.Sprintf("SystemTestingCommands: `Long command` failed due to %v", err))
		}
	}

	pterm.Success.Printfln("SystemTestingCommands: PASS (Time used = %v)", time.Since(tA))
}
