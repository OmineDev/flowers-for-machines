package resources_control

import (
	"fmt"
	"time"

	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol"
	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol/packet"
	"github.com/OmineDev/flowers-for-machines/core/py_rpc"

	"github.com/pterm/pterm"
)

// Block interaction actions were require bot to send
// packet.PlayerAuthInput since nemc 1.20.50.
//
// However, for command teleport, we should handle
// this by send packet.PlayerAuthInput with flag
// packet.InputFlagHandledTeleport, or the server will
// ignore the packet.PlayerAuthInput after bot was
// teleported.
func (r *Resources) handleMovePlayer(p *packet.MovePlayer) {
	if p.EntityRuntimeID == r.BotInfo().EntityRuntimeID && p.Mode == packet.MoveModeTeleport {
		_ = r.WritePacket(&packet.PlayerAuthInput{
			InputData: packet.InputFlagHandledTeleport,
			Position:  p.Position,
		})
	}
}

// respawn process
func (r *Resources) handleRespawn(p *packet.Respawn) {
	if p.State == packet.RespawnStateSearchingForSpawn {
		entityRuntimeID := r.BotInfo().EntityRuntimeID
		_ = r.WritePacket(&packet.Respawn{
			State:           packet.RespawnStateClientReadyToSpawn,
			EntityRuntimeID: entityRuntimeID,
		})
		_ = r.WritePacket(&packet.PlayerAction{
			EntityRuntimeID: entityRuntimeID,
			ActionType:      protocol.PlayerActionRespawn,
			BlockFace:       -1,
		})
		for range 5 {
			_ = r.WritePacket(&packet.PlayerAuthInput{
				InputData: packet.InputFlagStartFlying,
				Position:  p.Position,
			})
			time.Sleep(time.Second / 20 * 3)
		}
	}
}

// command request callback
func (r *Resources) handleCommandOutput(p *packet.CommandOutput) {
	r.commands.onCommandOutput(p)
}

// heart beat response (netease pyrpc)
func (r *Resources) handlePyRpc(p *packet.PyRpc) {
	// prepare
	if p.Value == nil {
		return
	}
	// unmarshal
	content, err := py_rpc.Unmarshal(p.Value)
	if err != nil {
		pterm.Warning.Sprintf("handlePyRpc: %v", err)
		return
	}
	// unmarshal
	switch c := content.(type) {
	case *py_rpc.HeartBeat:
		// heart beat to test the device is still alive?
		// it seems that we just need to return it back to the server is OK
		c.Type = py_rpc.ClientToServerHeartBeat
		r.client.Conn().WritePacket(&packet.PyRpc{
			Value:         py_rpc.Marshal(c),
			OperationType: packet.PyRpcOperationTypeSend,
		})
	}
}

// inventory contents(basic)
func (r *Resources) handleInventoryContent(p *packet.InventoryContent) {
	windowID := WindowID(p.WindowID)
	for key, value := range p.Content {
		slotID := SlotID(key)
		r.inventory.setItemStack(windowID, slotID, &value)
	}
}

// inventory contents(for enchant command...)
func (r *Resources) handleInventoryTransaction(p *packet.InventoryTransaction) {
	for _, value := range p.Actions {
		if value.SourceType == protocol.InventoryActionSourceCreative {
			continue
		}
		windowID, slotID := WindowID(value.WindowID), SlotID(value.InventorySlot)
		r.inventory.setItemStack(windowID, slotID, &value.NewItem)
	}
}

// inventory contents(for chest...) [NOT TEST]
func (r *Resources) handleInventorySlot(p *packet.InventorySlot) {
	windowID, slotID := WindowID(p.WindowID), SlotID(p.Slot)
	r.inventory.setItemStack(windowID, slotID, &p.NewItem)
}

// item stack request
func (r *Resources) handleItemStackResponse(p *packet.ItemStackResponse) {
	r.itemStack.mu.Lock()
	defer r.itemStack.mu.Unlock()

	select {
	case <-r.itemStack.ctx.Done():
		return
	default:
	}

	for _, response := range p.Responses {
		requestID := ItemStackRequestID(response.RequestID)
		itemRepeatChecker := make(map[SlotLocation]bool)

		callback, ok := r.itemStack.itemStackCallback[requestID]
		if !ok {
			panic(fmt.Sprintf("handleItemStackResponse: Item stack request with id %d set no callback", response.RequestID))
		}
		delete(r.itemStack.itemStackCallback, requestID)

		containerIDToWindowID, ok := r.itemStack.itemStackMapping[requestID]
		if !ok {
			panic(fmt.Sprintf("handleItemStackResponse: Item stack request with id %d set no container ID to Window ID mapping", response.RequestID))
		}
		delete(r.itemStack.itemStackMapping, requestID)

		itemUpdater := r.itemStack.itemStackUpdater[requestID]
		delete(r.itemStack.itemStackUpdater, requestID)

		if response.Status != protocol.ItemStackResponseStatusOK {
			resp := response
			go callback(&resp, nil)
			continue
		}

		for _, containerInfo := range response.ContainerInfo {
			windowID, existed := containerIDToWindowID[ContainerID(containerInfo.ContainerID)]
			if !existed {
				panic(
					fmt.Sprintf(
						"handleItemStackResponse: ContainerID %d not existed in underlying container ID to window ID mapping %#v (request id = %d)",
						containerInfo.ContainerID, containerIDToWindowID, response.RequestID,
					),
				)
			}

			for _, slotInfo := range containerInfo.SlotInfo {
				slotID := SlotID(slotInfo.Slot)

				slotLocation := SlotLocation{WindowID: windowID, SlotID: slotID}
				if _, ok := itemRepeatChecker[slotLocation]; ok {
					panic(fmt.Sprintf("handleItemStackResponse: The item at %#v was found duplicates (Should never happened)", slotLocation))
				}
				itemRepeatChecker[slotLocation] = true

				item, inventoryExisted := r.inventory.GetItemStack(windowID, slotID)
				if !inventoryExisted {
					panic(
						fmt.Sprintf("handleItemStackResponse: Inventory whose window ID is %d is not existed (request id = %d)",
							windowID, response.RequestID,
						),
					)
				}

				UpdateNetworkItem(
					item,
					SlotLocation{WindowID: windowID, SlotID: slotID},
					slotInfo, itemUpdater,
				)
				r.inventory.setItemStack(windowID, slotID, item)
			}
		}

		resp := response
		go callback(&resp, nil)
	}
}

// when a container is opened
func (r *Resources) handleContainerOpen(p *packet.ContainerOpen) {
	r.inventory.createInventory(WindowID(p.WindowID))
	r.container.onContainerOpen(p)
}

// when a container has been closed
func (r *Resources) handleContainerClose(p *packet.ContainerClose) {
	switch p.WindowID {
	case protocol.WindowIDInventory, protocol.WindowIDOffHand:
	case protocol.WindowIDArmour, protocol.WindowIDUI:
	default:
		r.inventory.deleteInventory(WindowID(p.WindowID))
	}
	r.container.onContainerClose(p)
}

// 根据收到的数据包更新客户端的资源数据
func (r *Resources) handlePacket(pk packet.Packet) {
	// internal
	switch p := pk.(type) {
	case *packet.MovePlayer:
		r.handleMovePlayer(p)
	case *packet.Respawn:
		r.handleRespawn(p)
	case *packet.CommandOutput:
		r.handleCommandOutput(p)
	case *packet.PyRpc:
		r.handlePyRpc(p)
	case *packet.InventoryContent:
		r.handleInventoryContent(p)
	case *packet.InventoryTransaction:
		r.handleInventoryTransaction(p)
	case *packet.InventorySlot:
		r.handleInventorySlot(p)
	case *packet.ItemStackResponse:
		r.handleItemStackResponse(p)
	case *packet.ContainerOpen:
		r.handleContainerOpen(p)
	case *packet.ContainerClose:
		r.handleContainerClose(p)
	case *packet.CreativeContent:
		r.constant.onCreativeContent(p)
	case *packet.AvailableCommands:
		r.constant.onAvailableCommands(p)
	case *packet.CraftingData:
		r.constant.onCraftingData(p)
	}
	// for other implements
	r.listener.onPacket(pk)
}

// handleConnClose ..
func (r *Resources) handleConnClose(err error) {
	r.commands.handleConnClose(err)
	r.itemStack.handleConnClose(err)
	r.container.handleConnClose(err)
	r.listener.handleConnClose(err)
}
