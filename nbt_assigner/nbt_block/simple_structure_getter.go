package nbt_block

import (
	"fmt"
	"sync"

	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol"
	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol/packet"
	"github.com/OmineDev/flowers-for-machines/game_control/game_interface"
	"github.com/OmineDev/flowers-for-machines/nbt_assigner/nbt_console"
)

// TODO: Merge this to game interface
func simpleStructureGetter(console *nbt_console.Console) (nbtMap map[string]any, err error) {
	var (
		api         *game_interface.GameInterface = console.API()
		resp        *packet.StructureTemplateDataResponse
		terminalErr error
	)
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("simpleStructureGetter: %v", r)
		}
	}()

	doOnce := new(sync.Once)
	channel := make(chan struct{})
	uniqueID, err := api.PacketListener().ListenPacket(
		[]uint32{packet.IDStructureTemplateDataResponse},
		func(p packet.Packet, connCloseErr error) {
			doOnce.Do(func() {
				if connCloseErr != nil {
					terminalErr = connCloseErr
				} else {
					resp = p.(*packet.StructureTemplateDataResponse)
				}
				close(channel)
			})
		},
	)
	if err != nil {
		return nil, fmt.Errorf("simpleStructureGetter: %v", err)
	}
	defer api.PacketListener().DestroyListener(uniqueID)

	err = api.Resources().WritePacket(
		&packet.StructureTemplateDataRequest{
			StructureName: "mystructure:simpleStructureGetter",
			Position:      console.Center(),
			Settings: protocol.StructureSettings{
				PaletteName:               "default",
				IgnoreEntities:            true,
				IgnoreBlocks:              false,
				Size:                      protocol.BlockPos{1, 1, 1},
				Offset:                    protocol.BlockPos{0, 0, 0},
				LastEditingPlayerUniqueID: api.GetBotInfo().EntityUniqueID,
				Rotation:                  0,
				Mirror:                    0,
				Integrity:                 100,
				Seed:                      0,
				AllowNonTickingChunks:     false,
			},
			RequestType: packet.StructureTemplateRequestExportFromSave,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("simpleStructureGetter: %v", err)
	}

	<-channel
	if terminalErr != nil {
		return nil, fmt.Errorf("simpleStructureGetter: %v", terminalErr)
	}

	m := resp.StructureTemplate
	m = m["structure"].(map[string]any)["palette"].(map[string]any)["default"].(map[string]any)
	m = m["block_position_data"].(map[string]any)["0"].(map[string]any)["block_entity_data"].(map[string]any)

	return m, nil
}
