package item_stack_transaction

import (
	"fmt"

	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol"
	"github.com/OmineDev/flowers-for-machines/game_control/game_interface/item_stack_operation"
	"github.com/OmineDev/flowers-for-machines/game_control/resources_control"
)

// itemStackOperationHandler ..
type itemStackOperationHandler struct {
	api                *resources_control.ContainerManager
	constantPacket     *resources_control.ConstantPacket
	virtualInventories *virtualInventories
	responseMapping    *responseMapping
}

// newItemStackOperationHandler ..
func newItemStackOperationHandler(
	api *resources_control.ContainerManager,
	constantPacket *resources_control.ConstantPacket,
	virtualInventories *virtualInventories,
	responseMapping *responseMapping,
) *itemStackOperationHandler {
	return &itemStackOperationHandler{
		api:                api,
		constantPacket:     constantPacket,
		virtualInventories: virtualInventories,
		responseMapping:    responseMapping,
	}
}

// handleMove ..
func (i *itemStackOperationHandler) handleMove(
	op item_stack_operation.Move,
	requestID resources_control.ItemStackRequestID,
) (result []protocol.StackRequestAction, err error) {
	// Basic check
	if op.Source == op.Destination {
		return nil, fmt.Errorf("handleMove: Source is equal to Destination")
	}

	// Get item runtime ID
	srcRID, err := i.virtualInventories.loadAndSetStackNetworkID(op.Source, requestID)
	if err != nil {
		return nil, fmt.Errorf("handleMove: %v", err)
	}
	dstRID, err := i.virtualInventories.loadAndSetStackNetworkID(op.Destination, requestID)
	if err != nil {
		return nil, fmt.Errorf("handleMove: %v", err)
	}

	// Get container ID
	srcCID, found := slotLocationToContainerID(i.api, op.Source)
	if !found {
		return nil, fmt.Errorf("handleMove: Can not find the container ID of given item whose at %#v", op.Source)
	}
	dstCID, found := slotLocationToContainerID(i.api, op.Destination)
	if !found {
		return nil, fmt.Errorf("handleMove: Can not find the container ID of given item whose at %#v", op.Destination)
	}

	// Bind container ID
	i.responseMapping.bind(op.Source.WindowID, srcCID)
	i.responseMapping.bind(op.Destination.WindowID, dstCID)

	// Update item count
	_, err = i.virtualInventories.loadAndAddItemCount(op.Source, -int8(op.Count), false)
	if err != nil {
		return nil, fmt.Errorf("handleMove: %v", err)
	}
	_, err = i.virtualInventories.loadAndAddItemCount(op.Destination, int8(op.Count), false)
	if err != nil {
		return nil, fmt.Errorf("handleMove: %v", err)
	}

	// Get src item stack and dst result count
	srcItemStack, err := i.virtualInventories.loadItemStack(op.Source)
	if err != nil {
		return nil, fmt.Errorf("handleMove: %v", err)
	}
	dstResultCount, err := i.virtualInventories.loadItemCount(op.Destination)
	if err != nil {
		return nil, fmt.Errorf("handleMove: %v", err)
	}

	// Sync item data
	if srcItemStack.Count == 0 {
		err = i.virtualInventories.setAir(op.Source)
		if err != nil {
			return nil, fmt.Errorf("handleMove: %v", err)
		}
	}
	err = i.virtualInventories.setItemStack(op.Destination, srcItemStack)
	if err != nil {
		return nil, fmt.Errorf("handleMove: %v", err)
	}
	err = i.virtualInventories.setItemCount(op.Destination, dstResultCount)
	if err != nil {
		return nil, fmt.Errorf("handleMove: %v", err)
	}

	// Make runtime data
	runtimeData := item_stack_operation.MoveRuntime{
		MoveSrcContainerID:    byte(srcCID),
		MoveSrcStackNetworkID: srcRID,
		MoveDstContainerID:    byte(dstCID),
		MoveDstStackNetworkID: dstRID,
	}
	return op.Make(runtimeData), nil
}

// handleSwap ..
func (i *itemStackOperationHandler) handleSwap(
	op item_stack_operation.Swap,
	requestID resources_control.ItemStackRequestID,
) (result []protocol.StackRequestAction, err error) {
	// Basic check
	if op.Source == op.Destination {
		return nil, fmt.Errorf("handleSwap: Source is equal to Destination")
	}

	// Get item runtime ID
	srcRID, err := i.virtualInventories.loadAndSetStackNetworkID(op.Source, requestID)
	if err != nil {
		return nil, fmt.Errorf("handleSwap: %v", err)
	}
	dstRID, err := i.virtualInventories.loadAndSetStackNetworkID(op.Destination, requestID)
	if err != nil {
		return nil, fmt.Errorf("handleSwap: %v", err)
	}

	// Get container ID
	srcCID, found := slotLocationToContainerID(i.api, op.Source)
	if !found {
		return nil, fmt.Errorf("handleSwap: Can not find the container ID of given item whose at %#v", op.Source)
	}
	dstCID, found := slotLocationToContainerID(i.api, op.Destination)
	if !found {
		return nil, fmt.Errorf("handleSwap: Can not find the container ID of given item whose at %#v", op.Destination)
	}

	// Bind container ID
	i.responseMapping.bind(op.Source.WindowID, srcCID)
	i.responseMapping.bind(op.Destination.WindowID, dstCID)

	// Get item origin data
	srcItemStack, err := i.virtualInventories.loadItemStack(op.Source)
	if err != nil {
		return nil, fmt.Errorf("handleSwap: %v", err)
	}
	dstItemStack, err := i.virtualInventories.loadItemStack(op.Destination)
	if err != nil {
		return nil, fmt.Errorf("handleSwap: %v", err)
	}

	// Sync item data
	err = i.virtualInventories.setItemStack(op.Source, dstItemStack)
	if err != nil {
		return nil, fmt.Errorf("handleSwap: %v", err)
	}
	err = i.virtualInventories.setItemStack(op.Destination, srcItemStack)
	if err != nil {
		return nil, fmt.Errorf("handleSwap: %v", err)
	}

	// Make runtime data
	runtimeData := item_stack_operation.SwapRuntime{
		SwapSrcContainerID:    byte(srcCID),
		SwapSrcStackNetworkID: srcRID,
		SwapDstContainerID:    byte(dstCID),
		SwapDstStackNetworkID: dstRID,
	}
	return op.Make(runtimeData), nil
}

// handleDrop ..
func (i *itemStackOperationHandler) handleDrop(
	op item_stack_operation.Drop,
	requestID resources_control.ItemStackRequestID,
) (result []protocol.StackRequestAction, err error) {
	// Get item runtime ID
	srcRID, err := i.virtualInventories.loadAndSetStackNetworkID(op.Path, requestID)
	if err != nil {
		return nil, fmt.Errorf("handleDrop: %v", err)
	}

	// Get container ID
	srcCID, found := slotLocationToContainerID(i.api, op.Path)
	if !found {
		return nil, fmt.Errorf("handleDrop: Can not find the container ID of given item whose at %#v", op.Path)
	}

	// Bind container ID
	i.responseMapping.bind(op.Path.WindowID, srcCID)

	// Update item data
	_, err = i.virtualInventories.loadAndAddItemCount(op.Path, -int8(op.Count), false)
	if err != nil {
		return nil, fmt.Errorf("handleDrop: %v", err)
	}
	srcResultCount, err := i.virtualInventories.loadItemCount(op.Path)
	if err != nil {
		return nil, fmt.Errorf("handleDrop: %v", err)
	}
	if srcResultCount == 0 {
		err = i.virtualInventories.setAir(op.Path)
		if err != nil {
			return nil, fmt.Errorf("handleDrop: %v", err)
		}
	}

	// Make runtime data
	runtimeData := item_stack_operation.DropRuntime{
		DropSrcContainerID:    byte(srcCID),
		DropSrcStackNetworkID: srcRID,
		Randomly:              false,
	}
	return op.Make(runtimeData), nil
}

// handleCreativeItem ..
func (i *itemStackOperationHandler) handleCreativeItem(
	op item_stack_operation.CreativeItem,
	requestID resources_control.ItemStackRequestID,
) (result []protocol.StackRequestAction, err error) {
	// Get item runtime ID
	rid, err := i.virtualInventories.loadAndSetStackNetworkID(op.Path, requestID)
	if err != nil {
		return nil, fmt.Errorf("handleCreativeItem: %v", err)
	}

	// Get container ID
	cid, found := slotLocationToContainerID(i.api, op.Path)
	if !found {
		return nil, fmt.Errorf("handleCreativeItem: Can not find the container ID of given item whose at %#v", op.Path)
	}

	// Bind container ID
	i.responseMapping.bind(op.Path.WindowID, cid)

	// Update item count
	_, err = i.virtualInventories.loadAndAddItemCount(op.Path, int8(op.Count), false)
	if err != nil {
		return nil, fmt.Errorf("handleCreativeItem: %v", err)
	}

	// Sync item data
	resultCount, err := i.virtualInventories.loadItemCount(op.Path)
	if err != nil {
		return nil, fmt.Errorf("handleCreativeItem: %v", err)
	}
	err = i.virtualInventories.setItemStack(
		op.Path,
		i.constantPacket.CreativeItemByCNI(op.CINI).Item,
	)
	if err != nil {
		return nil, fmt.Errorf("handleCreativeItem: %v", err)
	}
	err = i.virtualInventories.setItemCount(op.Path, resultCount)
	if err != nil {
		return nil, fmt.Errorf("handleCreativeItem: %v", err)
	}

	// Make runtime data
	return op.Make(
		item_stack_operation.CreativeItemRuntime{
			RequestID:             int32(requestID),
			DstContainerID:        byte(cid),
			DstItemStackID:        rid,
			CreativeItemNetworkID: op.CINI,
		},
	), nil
}

// handleRenaming ..
func (i *itemStackOperationHandler) handleRenaming(
	op item_stack_operation.Renaming,
	requestID resources_control.ItemStackRequestID,
) (result []protocol.StackRequestAction, err error) {
	// Get opening container data
	containerData, _, existed := i.api.ContainerData()
	if !existed {
		return nil, fmt.Errorf("handleRenaming: Anvil is not opened")
	}

	// Get item runtime ID
	srcRID, err := i.virtualInventories.loadAndSetStackNetworkID(op.Path, requestID)
	if err != nil {
		return nil, fmt.Errorf("handleRenaming: %v", err)
	}
	anvilRID, err := i.virtualInventories.loadAndSetStackNetworkID(
		resources_control.SlotLocation{
			WindowID: resources_control.WindowID(containerData.WindowID),
			SlotID:   1,
		},
		requestID,
	)
	if err != nil {
		return nil, fmt.Errorf("handleRenaming: %v", err)
	}

	// Get container ID
	srcCID, found := slotLocationToContainerID(i.api, op.Path)
	if !found {
		return nil, fmt.Errorf("handleRenaming: Can not find the container ID of given item whose at %#v", op.Path)
	}

	// Bind container ID
	i.responseMapping.bind(resources_control.WindowID(containerData.WindowID), protocol.ContainerAnvilInput)
	i.responseMapping.bind(op.Path.WindowID, srcCID)

	// Get item stack
	srcItemStack, err := i.virtualInventories.loadItemStack(op.Path)
	if err != nil {
		return nil, fmt.Errorf("handleRenaming: %v", err)
	}

	// Update item data
	err = i.virtualInventories.updateFromUpdater(
		op.Path,
		resources_control.ExpectedNewItem{
			NBT: resources_control.ItemNewNBTData{
				UseNBTData:       false,
				ChangeRepairCost: true,
				RepairCostDelta:  0,
				UseOriginDamage:  false,
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("handleRenaming: %v", err)
	}

	// Make runtime data
	runtimeData := item_stack_operation.RenamingRuntime{
		RequestID:               int32(requestID),
		ItemCount:               uint8(srcItemStack.Count),
		SrcContainerID:          byte(srcCID),
		SrcStackNetworkID:       srcRID,
		AnvilSlotStackNetworkID: anvilRID,
	}
	return op.Make(runtimeData), nil
}

// handleLooming ..
func (i *itemStackOperationHandler) handleLooming(
	op item_stack_operation.Looming,
	requestID resources_control.ItemStackRequestID,
) (result []protocol.StackRequestAction, err error) {
	// Prepare
	runtimeData := item_stack_operation.LoomingRuntime{
		RequestID: int32(requestID),
	}

	// Basic check
	if op.BannerPath == op.DyePath {
		return nil, fmt.Errorf("handleLooming: BannerPath is equal to DyePath")
	}
	if op.UsePattern {
		if op.PatternPath == op.BannerPath {
			return nil, fmt.Errorf("handleLooming: PatternPath is equal to BannerPath")
		}
		if op.PatternPath == op.DyePath {
			return nil, fmt.Errorf("handleLooming: PatternPath is equal to DyePath")
		}
	}

	// Get opening container data
	containerData, _, existed := i.api.ContainerData()
	if !existed {
		return nil, fmt.Errorf("handleLooming: Loom is not opened")
	}

	// Process pattern
	if op.UsePattern {
		// Prepare
		loomSlot := resources_control.SlotLocation{
			WindowID: resources_control.WindowID(containerData.WindowID),
			SlotID:   11,
		}

		// Get item runtime ID
		rid, err := i.virtualInventories.loadAndSetStackNetworkID(op.PatternPath, requestID)
		if err != nil {
			return nil, fmt.Errorf("handleLooming: %v", err)
		}
		loomRID, err := i.virtualInventories.loadAndSetStackNetworkID(loomSlot, requestID)
		if err != nil {
			return nil, fmt.Errorf("handleLooming: %v", err)
		}

		// Get container ID
		cid, found := slotLocationToContainerID(i.api, op.PatternPath)
		if !found {
			return nil, fmt.Errorf("handleLooming: Can not find the container ID of given item whose at %#v", op.PatternPath)
		}

		// Bind container ID
		i.responseMapping.bind(op.PatternPath.WindowID, cid)
		i.responseMapping.bind(resources_control.WindowID(containerData.WindowID), protocol.ContainerLoomMaterial)

		// Set runtime data (Pattern related)
		runtimeData.LoomPatternStackNetworkID = loomRID
		runtimeData.MovePatternSrcContainerID = byte(cid)
		runtimeData.MovePatternSrcStackNetworkID = rid
	}

	// Process banner
	{
		// Prepare
		loomSlot := resources_control.SlotLocation{
			WindowID: resources_control.WindowID(containerData.WindowID),
			SlotID:   9,
		}

		// Get item runtime ID
		rid, err := i.virtualInventories.loadAndSetStackNetworkID(op.BannerPath, requestID)
		if err != nil {
			return nil, fmt.Errorf("handleLooming: %v", err)
		}
		loomRID, err := i.virtualInventories.loadAndSetStackNetworkID(loomSlot, requestID)
		if err != nil {
			return nil, fmt.Errorf("handleLooming: %v", err)
		}

		// Get container ID
		cid, found := slotLocationToContainerID(i.api, op.BannerPath)
		if !found {
			return nil, fmt.Errorf("handleLooming: Can not find the container ID of given item whose at %#v", op.BannerPath)
		}

		// Bind container ID
		i.responseMapping.bind(op.BannerPath.WindowID, cid)
		i.responseMapping.bind(resources_control.WindowID(containerData.WindowID), protocol.ContainerLoomInput)

		// Update banner data
		err = i.virtualInventories.updateFromUpdater(op.BannerPath, op.ResultItem)
		if err != nil {
			return nil, fmt.Errorf("handleLooming: %v", err)
		}

		// Set runtime data (Banner related)
		runtimeData.LoomBannerStackNetworkID = loomRID
		runtimeData.MoveBannerSrcContainerID = byte(cid)
		runtimeData.MoveBannerSrcStackNetworkID = rid
	}

	// Process dye
	{
		// Prepare
		loomSlot := resources_control.SlotLocation{
			WindowID: resources_control.WindowID(containerData.WindowID),
			SlotID:   10,
		}

		// Get item runtime ID
		rid, err := i.virtualInventories.loadAndSetStackNetworkID(op.DyePath, requestID)
		if err != nil {
			return nil, fmt.Errorf("handleLooming: %v", err)
		}
		loomRID, err := i.virtualInventories.loadAndSetStackNetworkID(loomSlot, requestID)
		if err != nil {
			return nil, fmt.Errorf("handleLooming: %v", err)
		}

		// Get container ID
		cid, found := slotLocationToContainerID(i.api, op.DyePath)
		if !found {
			return nil, fmt.Errorf("handleLooming: Can not find the container ID of given item whose at %#v", op.DyePath)
		}

		// Bind container ID
		i.responseMapping.bind(op.DyePath.WindowID, cid)
		i.responseMapping.bind(resources_control.WindowID(containerData.WindowID), protocol.ContainerLoomDye)

		// Update item count
		_, err = i.virtualInventories.loadAndAddItemCount(op.DyePath, -1, false)
		if err != nil {
			return nil, fmt.Errorf("handleLooming: %v", err)
		}

		// Sync item data
		resultCount, err := i.virtualInventories.loadItemCount(op.DyePath)
		if err != nil {
			return nil, fmt.Errorf("handleLooming: %v", err)
		}
		if resultCount == 0 {
			err = i.virtualInventories.setAir(op.DyePath)
			if err != nil {
				return nil, fmt.Errorf("handleLooming: %v", err)
			}
		}

		// Set runtime data (Dye related)
		runtimeData.LoomDyeStackNetworkID = loomRID
		runtimeData.MoveDyeSrcContainerID = byte(cid)
		runtimeData.MoveDyeSrcStackNetworkID = rid
	}

	// Make runtime data
	return op.Make(runtimeData), nil
}

// handleCrafting ..
func (i *itemStackOperationHandler) handleCrafting(
	op item_stack_operation.Crafting,
	requestID resources_control.ItemStackRequestID,
) (result []protocol.StackRequestAction, err error) {
	// Prepare
	runtimeData := item_stack_operation.CraftingRuntime{
		RequestID: int32(requestID),
	}

	// Search items that can be consumed
	for slotID, item := range i.virtualInventories.allItemInstances(protocol.WindowIDCrafting) {
		// Check
		if item.Stack.Count == 0 {
			continue
		}

		// Prepare
		location := resources_control.SlotLocation{
			WindowID: protocol.WindowIDCrafting,
			SlotID:   slotID,
		}

		// Get item runtime ID
		rid, err := i.virtualInventories.loadAndSetStackNetworkID(location, requestID)
		if err != nil {
			return nil, fmt.Errorf("handleCrafting: %v", err)
		}

		// Set to air
		err = i.virtualInventories.setAir(location)
		if err != nil {
			return nil, fmt.Errorf("handleCrafting: %v", err)
		}

		// Append
		runtimeData.Consumes = append(runtimeData.Consumes, item_stack_operation.CraftingConsume{
			Slot:           location.SlotID,
			StackNetworkID: rid,
			Count:          uint8(item.Stack.Count),
		})
	}

	// Process result item
	{
		// Prepare
		resultPath := resources_control.SlotLocation{
			WindowID: protocol.WindowIDInventory,
			SlotID:   op.ResultSlotID,
		}

		// Get item runtime ID
		rid, err := i.virtualInventories.loadAndSetStackNetworkID(resultPath, requestID)
		if err != nil {
			return nil, fmt.Errorf("handleCrafting: %v", err)
		}

		// Update item count
		_, err = i.virtualInventories.loadAndAddItemCount(resultPath, int8(op.ResultCount), false)
		if err != nil {
			return nil, fmt.Errorf("handleCrafting: %v", err)
		}

		// Sync item data
		err = i.virtualInventories.updateFromUpdater(resultPath, op.ResultItem)
		if err != nil {
			return nil, fmt.Errorf("handleCrafting: %v", err)
		}

		// Set runtime data
		runtimeData.ResultStackNetworkID = rid
	}

	// Bind container ID
	i.responseMapping.bind(protocol.WindowIDInventory, protocol.ContainerCombinedHotBarAndInventory)
	i.responseMapping.bind(protocol.WindowIDCrafting, protocol.ContainerCraftingInput)

	// Make runtime data
	return op.Make(runtimeData), nil
}

// handleTrimming ..
func (i *itemStackOperationHandler) handleTrimming(
	op item_stack_operation.Trimming,
	requestID resources_control.ItemStackRequestID,
) (result []protocol.StackRequestAction, err error) {
	// Prepare
	runtimeData := item_stack_operation.TrimmingRuntime{
		RequestID:       int32(requestID),
		RecipeNetworkID: i.constantPacket.TrimRecipeNetworkID(),
	}

	// Basic check
	if op.TrimItem == op.Material {
		return nil, fmt.Errorf("handleTrimming: TrimItem (path) is equal to Marterial (path)")
	}
	if op.TrimItem == op.Template {
		return nil, fmt.Errorf("handleTrimming: TrimItem (path) is equal to Template (path)")
	}
	if op.Material == op.Template {
		return nil, fmt.Errorf("handleTrimming: Material (path) is equal to Template (path)")
	}

	// Get opening container data
	containerData, _, existed := i.api.ContainerData()
	if !existed {
		return nil, fmt.Errorf("handleTrimming: Smithing table is not opened")
	}

	// Trim item
	{
		// Prepare
		loomSlot := resources_control.SlotLocation{
			WindowID: resources_control.WindowID(containerData.WindowID),
			SlotID:   0x33,
		}

		// Get item runtime ID
		rid, err := i.virtualInventories.loadAndSetStackNetworkID(op.TrimItem, requestID)
		if err != nil {
			return nil, fmt.Errorf("handleTrimming: %v", err)
		}
		loomRID, err := i.virtualInventories.loadAndSetStackNetworkID(loomSlot, requestID)
		if err != nil {
			return nil, fmt.Errorf("handleTrimming: %v", err)
		}

		// Get container ID
		cid, found := slotLocationToContainerID(i.api, op.TrimItem)
		if !found {
			return nil, fmt.Errorf("handleTrimming: Can not find the container ID of given item whose at %#v", op.TrimItem)
		}

		// Bind container ID
		i.responseMapping.bind(op.TrimItem.WindowID, cid)
		i.responseMapping.bind(resources_control.WindowID(containerData.WindowID), protocol.ContainerSmithingTableInput)

		// Update trim item data
		err = i.virtualInventories.updateFromUpdater(op.TrimItem, op.ResultItem)
		if err != nil {
			return nil, fmt.Errorf("handleTrimming: %v", err)
		}

		// Set runtime data (TrimItem related)
		runtimeData.TrimItemStackNetworkID = loomRID
		runtimeData.MoveTrimItemSrcContainerID = byte(cid)
		runtimeData.MoveTrimItemSrcStackNetworkID = rid
	}

	// Material
	{
		// Prepare
		loomSlot := resources_control.SlotLocation{
			WindowID: resources_control.WindowID(containerData.WindowID),
			SlotID:   0x34,
		}

		// Get item runtime ID
		rid, err := i.virtualInventories.loadAndSetStackNetworkID(op.Material, requestID)
		if err != nil {
			return nil, fmt.Errorf("handleTrimming: %v", err)
		}
		loomRID, err := i.virtualInventories.loadAndSetStackNetworkID(loomSlot, requestID)
		if err != nil {
			return nil, fmt.Errorf("handleTrimming: %v", err)
		}

		// Get container ID
		cid, found := slotLocationToContainerID(i.api, op.Material)
		if !found {
			return nil, fmt.Errorf("handleTrimming: Can not find the container ID of given item whose at %#v", op.Material)
		}

		// Bind container ID
		i.responseMapping.bind(op.Material.WindowID, cid)
		i.responseMapping.bind(resources_control.WindowID(containerData.WindowID), protocol.ContainerSmithingTableMaterial)

		// Update item count
		_, err = i.virtualInventories.loadAndAddItemCount(op.Material, -1, false)
		if err != nil {
			return nil, fmt.Errorf("handleTrimming: %v", err)
		}

		// Sync item data
		resultCount, err := i.virtualInventories.loadItemCount(op.Material)
		if err != nil {
			return nil, fmt.Errorf("handleTrimming: %v", err)
		}
		if resultCount == 0 {
			err = i.virtualInventories.setAir(op.Material)
			if err != nil {
				return nil, fmt.Errorf("handleTrimming: %v", err)
			}
		}

		// Set runtime data (Material related)
		runtimeData.MaterialStackNetworkID = loomRID
		runtimeData.MoveMaterialSrcContainerID = byte(cid)
		runtimeData.MoveMaterialSrcStackNetworkID = rid
	}

	// Template
	{
		// Prepare
		loomSlot := resources_control.SlotLocation{
			WindowID: resources_control.WindowID(containerData.WindowID),
			SlotID:   0x35,
		}

		// Get item runtime ID
		rid, err := i.virtualInventories.loadAndSetStackNetworkID(op.Template, requestID)
		if err != nil {
			return nil, fmt.Errorf("handleTrimming: %v", err)
		}
		loomRID, err := i.virtualInventories.loadAndSetStackNetworkID(loomSlot, requestID)
		if err != nil {
			return nil, fmt.Errorf("handleTrimming: %v", err)
		}

		// Get container ID
		cid, found := slotLocationToContainerID(i.api, op.Template)
		if !found {
			return nil, fmt.Errorf("handleTrimming: Can not find the container ID of given item whose at %#v", op.Template)
		}

		// Bind container ID
		i.responseMapping.bind(op.Template.WindowID, cid)
		i.responseMapping.bind(resources_control.WindowID(containerData.WindowID), protocol.ContainerSmithingTableTemplate)

		// Update item count
		_, err = i.virtualInventories.loadAndAddItemCount(op.Template, -1, false)
		if err != nil {
			return nil, fmt.Errorf("handleTrimming: %v", err)
		}

		// Sync item data
		resultCount, err := i.virtualInventories.loadItemCount(op.Template)
		if err != nil {
			return nil, fmt.Errorf("handleTrimming: %v", err)
		}
		if resultCount == 0 {
			err = i.virtualInventories.setAir(op.Template)
			if err != nil {
				return nil, fmt.Errorf("handleTrimming: %v", err)
			}
		}

		// Set runtime data (Template related)
		runtimeData.TemplateStackNetworkID = loomRID
		runtimeData.MoveTemplateSrcContainerID = byte(cid)
		runtimeData.MoveTemplateSrcStackNetworkID = rid
	}

	// Make runtime data
	return op.Make(runtimeData), nil
}
