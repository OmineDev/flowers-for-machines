package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol/packet"
	"github.com/OmineDev/flowers-for-machines/utils"

	"github.com/google/uuid"
	"github.com/pterm/pterm"
)

func SystemTestingUUIDSafeString() {
	tA := time.Now()

	// Chat
	for range 100 {
		ud := uuid.New()
		chatContent := utils.MakeUUIDSafeString(ud)
		channel := make(chan struct{})

		equalUUID, _ := utils.FromUUIDSafeString(chatContent)
		if equalUUID != ud {
			panic("SystemTestingUUIDSafeString: UUID Safe String not equal")
		}

		doOnce := new(sync.Once)
		uniqueID, err := api.PacketListener().ListenPacket(
			[]uint32{packet.IDText},
			func(p packet.Packet, connCloseErr error) {
				if connCloseErr != nil {
					panic(fmt.Sprintf("SystemTestingUUIDSafeString: Failed due to %v", connCloseErr))
				}
				if p.(*packet.Text).Message == chatContent {
					doOnce.Do(func() { close(channel) })
				}
			},
		)
		if err != nil {
			panic(fmt.Sprintf("SystemTestingUUIDSafeString: Failed due to %v", err))
		}
		api.Commands().SendChat(chatContent)

		timer := time.NewTimer(time.Second * 5)
		defer timer.Stop()
		select {
		case <-timer.C:
			panic("SystemTestingUUIDSafeString: Time out")
		case <-channel:
			api.PacketListener().DestroyListener(uniqueID)
		}
	}

	pterm.Success.Printfln("SystemTestingUUIDSafeString: PASS (Time used = %v)", time.Since(tA))
}
