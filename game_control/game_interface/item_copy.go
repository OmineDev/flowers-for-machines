package game_interface

import (
	"fmt"
	"maps"
	"strings"
	"sync"

	"github.com/OmineDev/flowers-for-machines/game_control/resources_control"
)

type (
	// ItemType 指示该物品在单次操作中的物品类型。
	//
	// 物品类型跟实际物品的任何 ID 都是无关的，
	// 它只是为了让内置实现区分不同的物品而设。
	//
	// 因此，在不同的请求中，相同的 ItemType
	// 可以被重新使用。
	//
	// 您可以简单的将不同的物品赋予不同的数字
	ItemType uint8
	// ItemInfo 是物品的信息
	ItemInfo struct {
		Count    uint8    // 物品的数量
		ItemType ItemType // 物品的类型
	}
	// ItemInfoWithSlot 是物品的信息，
	// 同时记载其所在的槽位索引
	ItemInfoWithSlot struct {
		Slot     resources_control.SlotID // 物品所在的槽位索引
		ItemInfo ItemInfo                 // 物品的信息
	}
)

type (
	// ItemGroupElement 是 ItemGroup 中的单个元素
	ItemGroupElement struct {
		Slot  resources_control.SlotID // 该物品所在的槽位索引
		Count uint8                    // 该物品的数量
	}
	// ItemGroup 描述 ItemType 一致的物品，
	// 并且记载这些物品在目标容器中的分布情况。
	// 我们把这一系列物品称为一个 Group
	ItemGroup struct {
		// Parent 是这个 Group 的父结点。
		// 父节点最初来源于已有的基物品，
		// 可以保证它一定是存在的
		Parent ItemGroupElement
		// Child 是这个 Group 的所有子结点
		Child []ItemGroupElement
		// ChildHaveGrown 指示每个子结点
		// 是否已经存在至少 1 个物品。
		//
		// 在每个生成轮次中，应当优先保证
		// 所有子结点都已被诞生，即每个子
		// 结点都存在至少 1 个物品。
		//
		// 如果存在没有诞生的子节点，则该 Group
		// 应当重复子结点生成操作，直到全部诞生
		ChildHaveGrown []bool
	}
)

// ItemCopy 是互斥的复杂物品拷贝实现
type ItemCopy struct {
	mu        *sync.Mutex
	api       *ContainerOpenAndClose
	itemStack *ItemStackOperation
	commands  *Commands
	structure *StructureBackup

	inventory [27]ItemInfo
	container [27]ItemInfo

	containerInfo     UseItemOnBlocks
	containerIsOpened bool

	ItemGroups map[ItemType]ItemGroup
}

// NewItemCopy 基于 api、commands、itemStack 和 structure 创建并返回一个新的 ItemCopy
func NewItemCopy(
	api *ContainerOpenAndClose,
	commands *Commands,
	itemStack *ItemStackOperation,
	structure *StructureBackup,
) *ItemCopy {
	return &ItemCopy{
		mu:                new(sync.Mutex),
		api:               api,
		itemStack:         itemStack,
		commands:          commands,
		structure:         structure,
		inventory:         [27]ItemInfo{},
		container:         [27]ItemInfo{},
		containerInfo:     UseItemOnBlocks{},
		containerIsOpened: false,
		ItemGroups:        make(map[ItemType]ItemGroup),
	}
}

// CopyItem 根据给定的基物品 baseItems 和蓝图 targetItems 向相应的容器载入物品。
// 应确保调用 CopyItem 前没有打开任何容器。
//
// containerInfo 提供的信息指示机器人应当如何打开目标容器。由于此函数不会自动切换物品栏，
// 因此您需要确保在调用前已经切换物品栏到 containerInfo.HotbarSlotID 的值。
//
// targetItems 指示是最终容器的物品状态，即机器人将按照背包中已有的 baseItems 物品，
// 通过多次的物品拷贝操作，使得容器中物品的状态为 targetItems。
//
// 因此，您应当确保 targetItems 的所有物品都可以在 baseItems 中找到。
// 如果存在至少 1 个物品不能被找到，则返回错误。
//
// 应当说明的是，targetItems 的长度应该与目标容器的格子数量相等，并且空气物品应当
// 置为 nil；除此外，CopyItem 在完成后，蓝图 baseItems 本身将被消耗，并且背包将被
// 清空；如果返回错误，不保证蓝图 baseItems 仍然按预期存在。
//
// 当然，targetItems 的长度不应超过 27，这意味着目标容器的格子数量被限制在 27 最大；
// 同时，您不应操作一个连体的大箱子，即便它具有 54 个格子。
//
// 另外，CopyItem 是阻塞的，这意味着如果存在多个 go 惯例调用 CopyItem，则每个调用
// 都将会阻塞，直到上一个调用完成
func (i *ItemCopy) CopyItem(
	containerInfo UseItemOnBlocks,
	baseItems []ItemInfoWithSlot,
	targetItems []*ItemInfo,
) error {
	if len(baseItems) == 0 || len(targetItems) == 0 {
		return nil
	}

	if len(baseItems) > 27 {
		return fmt.Errorf("CopyItem: Given baseItems have %d elements, but a maximum of 27 are allowed", len(baseItems))
	}
	if len(targetItems) > 27 {
		return fmt.Errorf("CopyItem: Given targetItems have %d elements, but a maximum of 27 are allowed", len(baseItems))
	}

	baseItemTypeSet := make(map[ItemType]bool)
	for _, item := range baseItems {
		if baseItemTypeSet[item.ItemInfo.ItemType] {
			return fmt.Errorf("CopyItem: Given baseItems found same item type; item = %#v; baseItems = %#v", item, baseItems)
		}
		baseItemTypeSet[item.ItemInfo.ItemType] = true
	}
	for _, item := range targetItems {
		if item == nil {
			continue
		}
		if !baseItemTypeSet[item.ItemType] {
			return fmt.Errorf("CopyItem: Item %#v is not existed in given baseItems; baseItems = %#v", item, baseItems)
		}
	}

	i.mu.Lock()
	defer func() {
		if i.containerIsOpened {
			_ = i.api.CloseContainer()
			i.containerIsOpened = false
		}
		i.mu.Unlock()
	}()

	err := i.copyItem(containerInfo, baseItems, targetItems)
	if err != nil {
		return fmt.Errorf("CopyItem: %v", err)
	}
	return nil
}

// copyItem 是内部实现细节，
// 不应该被其他人所使用
func (i *ItemCopy) copyItem(
	containerInfo UseItemOnBlocks,
	baseItems []ItemInfoWithSlot,
	targetItems []*ItemInfo,
) error {
	i.inventory = [27]ItemInfo{}
	i.container = [27]ItemInfo{}
	i.containerInfo = containerInfo
	i.containerIsOpened = false
	i.ItemGroups = make(map[ItemType]ItemGroup)

	// Step 1: Convert target items to item groups
	for index, item := range targetItems {
		if item == nil {
			continue
		}

		group, ok := i.ItemGroups[item.ItemType]
		if !ok {
			i.ItemGroups[item.ItemType] = ItemGroup{
				Parent: ItemGroupElement{
					Slot:  resources_control.SlotID(index),
					Count: item.Count,
				},
				Child:          nil,
				ChildHaveGrown: nil,
			}
			continue
		}

		group.Child = append(group.Child, ItemGroupElement{
			Slot:  resources_control.SlotID(index),
			Count: item.Count,
		})
		group.ChildHaveGrown = append(group.ChildHaveGrown, false)
		i.ItemGroups[item.ItemType] = group
	}

	// Step 2.1: Open the container
	success, err := i.api.OpenContainer(i.containerInfo, false)
	if err != nil {
		return fmt.Errorf("copyItem: %v", err)
	}
	if !success {
		return fmt.Errorf("copyItem: Failed to open the container in the every beginning")
	}
	i.containerIsOpened = true

	// Step 2.2: Get base item mapping
	baseItemMapping := make(map[ItemType]ItemInfoWithSlot)
	for _, item := range baseItems {
		baseItemMapping[item.ItemInfo.ItemType] = item
	}

	// Step 2.3: Move base item to the container
	transaction := i.itemStack.OpenTransaction()
	for itemType, group := range i.ItemGroups {
		agent := baseItemMapping[itemType]

		_ = transaction.MoveToContainer(
			agent.Slot,
			group.Parent.Slot,
			agent.ItemInfo.Count,
		)

		for index, child := range group.Child {
			if agent.ItemInfo.Count == 1 {
				break
			}
			agent.ItemInfo.Count--

			_ = transaction.MoveBetweenContainer(
				group.Parent.Slot,
				child.Slot,
				1,
			)

			i.container[child.Slot] = ItemInfo{
				Count:    1,
				ItemType: itemType,
			}
			i.ItemGroups[itemType].ChildHaveGrown[index] = true
		}

		i.container[group.Parent.Slot] = agent.ItemInfo
	}
	success, _, _, err = transaction.Commit()
	if err != nil {
		return fmt.Errorf("copyItem: %v", err)
	}
	if !success {
		return fmt.Errorf("copyItem: Failed to move baseItems to the container")
	}

	// Step 3: Clean inventory
	_, err = i.commands.SendWSCommandWithResp("clear")
	if err != nil {
		return fmt.Errorf("copyItem: %v", err)
	}

	// Step 3: Copy item
	for {
		err = i.stepGetAllThingBack()
		if err != nil {
			return fmt.Errorf("copyItem: %v", err)
		}
		canStop, err := i.stepMergeToContainer()
		if err != nil {
			return fmt.Errorf("copyItem: %v", err)
		}
		if canStop {
			break
		}
	}

	// Step 4: Return
	return nil
}

// stepGetAllThingBack 将容器中的全部物品放入背包
func (i *ItemCopy) stepGetAllThingBack() error {
	// Step 1: Backup structure
	uniqueID, err := i.structure.BackupStructure(i.containerInfo.BlockPos)
	if err != nil {
		return fmt.Errorf("stepGetAllThingBack: %v", err)
	}
	defer i.structure.DeleteStructure(uniqueID)

	// Step 2: Move all items back to inventory
	transaction := i.itemStack.OpenTransaction()
	for itemType, group := range i.ItemGroups {
		for _, child := range append([]ItemGroupElement{group.Parent}, group.Child...) {
			haveCount := i.container[child.Slot].Count
			extraCount := i.inventory[child.Slot].Count

			if haveCount == 0 {
				continue
			}

			childAllGrown := true
			for _, grow := range group.ChildHaveGrown {
				if !grow {
					childAllGrown = false
					break
				}
			}
			if haveCount >= child.Count && childAllGrown {
				continue
			}

			if extraCount != 0 {
				// This should never happened, or there happened some underlying internal problems
				panic("stepMergeToContainer: Should never happened")
			}

			_ = transaction.MoveToInventory(
				child.Slot,
				child.Slot,
				haveCount,
			)
			i.inventory[child.Slot] = ItemInfo{
				Count:    haveCount,
				ItemType: itemType,
			}
		}
	}
	success, _, _, err := transaction.Commit()
	if err != nil {
		return fmt.Errorf("stepGetAllThingBack: %v", err)
	}
	if !success {
		return fmt.Errorf("stepGetAllThingBack: Move item back unsuccessful")
	}

	// Step 3: Close container
	err = i.api.CloseContainer()
	if err != nil {
		return fmt.Errorf("stepGetAllThingBack: %v", err)
	}
	i.containerIsOpened = false

	// Step 4: Revert structure
	err = i.structure.RevertStructure(uniqueID, i.containerInfo.BlockPos)
	if err != nil {
		return fmt.Errorf("stepGetAllThingBack: %v", err)
	}

	// Step 5: Special process for barrel
	if strings.Contains(i.containerInfo.BlockName, "barrel") {
		openBit, _ := i.containerInfo.BlockStates["open_bit"].(byte)
		if openBit == 0 {
			newBlockStates := make(map[string]any)
			maps.Copy(newBlockStates, i.containerInfo.BlockStates)
			newBlockStates["open_bit"] = byte(1)
			i.containerInfo.BlockStates = newBlockStates
		}
	}

	return nil
}

// stepMergeToContainer 将背包的物品合并到目标容器中。
// canStop 指示目标容器是否已经完成构造
func (i *ItemCopy) stepMergeToContainer() (canStop bool, err error) {
	// Step 1: Open container
	success, err := i.api.OpenContainer(i.containerInfo, false)
	if err != nil {
		return false, fmt.Errorf("stepMergeToContainer: %v", err)
	}
	if !success {
		return false, fmt.Errorf("stepMergeToContainer: Failed to open the container")
	}
	i.containerIsOpened = true

	// Step 2: Open transaction
	transaction := i.itemStack.OpenTransaction()

	// Step 3: Grow child and merge
	for itemType, group := range i.ItemGroups {
		// Try to grow child from their parent
		for index, child := range group.Child {
			if i.container[group.Parent.Slot].Count <= group.Parent.Count {
				break
			}

			if group.ChildHaveGrown[index] {
				continue
			}

			_ = transaction.MoveBetweenContainer(
				group.Parent.Slot,
				child.Slot,
				1,
			)

			i.container[group.Parent.Slot].Count--
			i.container[child.Slot] = ItemInfo{
				Count:    1,
				ItemType: itemType,
			}
			i.ItemGroups[itemType].ChildHaveGrown[index] = true
		}

		// Try to grow child from inventory
		for index, child := range group.Child {
			if group.ChildHaveGrown[index] {
				continue
			}

			slot, _, found := i.searchItemFromInventory(itemType)
			if !found {
				break
			}
			_ = transaction.MoveToContainer(slot, child.Slot, 1)

			i.inventory[slot].Count--
			if i.inventory[slot].Count == 0 {
				i.inventory[slot].ItemType = 0
			}

			i.container[child.Slot] = ItemInfo{
				Count:    1,
				ItemType: itemType,
			}
			i.ItemGroups[itemType].ChildHaveGrown[index] = true
		}

		// If their still have child not grown,
		// then that means we don't have enough
		// items to grown more child, and also
		// have no items to merge
		stillHaveChildNotGrown := false
		for _, grown := range group.ChildHaveGrown {
			if !grown {
				stillHaveChildNotGrown = true
				break
			}
		}
		if stillHaveChildNotGrown {
			continue
		}

		allElements := append([]ItemGroupElement{group.Parent}, group.Child...)
		unusedContainerCount := uint8(0)
		unusedInventoryMapping := make(map[resources_control.SlotID]uint8)

		// (Merge stage 1)
		// Here we merge each child in a simple way, which is
		// from inventory[child.Slot] to container[child.Slot]
		for index, child := range allElements {
			haveCount := i.container[child.Slot].Count
			moveCount := i.inventory[child.Slot].Count
			unusedCount := uint8(0)

			// This only can be happened for a parent.
			// By this way, record the unused items and continue.
			if child.Count < haveCount {
				// It is impossible for a child to meet child.Count < haveCount,
				// so we panic for those we are child.
				if index != 0 {
					panic("stepMergeToContainer: Should never happened")
				}
				unusedContainerCount = haveCount - child.Count
				continue
			}

			// It's possibly for the item have zero count
			if moveCount == 0 {
				continue
			}
			// Compute the actual move count and the item we will not used
			if moveCount+haveCount > child.Count {
				unusedCount = moveCount - (child.Count - haveCount)
				moveCount -= unusedCount
			}
			// We have some item that not used, and here we mark them.
			if unusedCount > 0 {
				unusedInventoryMapping[child.Slot] = unusedCount
			}
			// Current item is finished, and can continue here
			if moveCount == 0 {
				continue
			}

			// Do move operation.
			_ = transaction.MoveToContainer(child.Slot, child.Slot, moveCount)

			// Update underlying virtual inventories information.
			i.container[child.Slot].Count += moveCount
			i.inventory[child.Slot].Count -= moveCount
			if i.inventory[child.Slot].Count == 0 {
				i.inventory[child.Slot].ItemType = 0
			}
		}

		// (Merge stage 2)
		// Here we try to use these unused item from merge stage 1
		for _, child := range allElements {
			// We here to check the finish states of current child
			haveCount := i.container[child.Slot].Count
			moveCount := child.Count - haveCount
			if child.Count <= haveCount {
				continue
			}

			// unusedContainerCount > 0 means there exist extra item
			// from parent. And now we can use them.
			if unusedContainerCount > 0 {
				moveCount = min(moveCount, unusedContainerCount)
				unusedContainerCount -= moveCount

				_ = transaction.MoveBetweenContainer(
					allElements[0].Slot,
					child.Slot,
					moveCount,
				)

				i.container[child.Slot].Count += moveCount
				i.container[allElements[0].Slot].Count -= moveCount
				if i.container[allElements[0].Slot].Count == 0 {
					i.container[allElements[0].Slot].ItemType = 0
				}
			}

			// Check again because we maybe use item from parent
			haveCount = i.container[child.Slot].Count
			moveCount = child.Count - haveCount
			if child.Count < haveCount {
				// This should never happened, or there happened some underlying internal problems
				panic("stepMergeToContainer: Should never happened")
			}
			if moveCount == 0 {
				continue
			}

			// Here we use these unused item from inventory.
			for slotID, canUseCount := range unusedInventoryMapping {
				if canUseCount == 0 {
					continue
				}

				moveCount = min(moveCount, canUseCount)
				unusedInventoryMapping[slotID] -= moveCount

				_ = transaction.MoveToContainer(slotID, child.Slot, moveCount)

				i.container[child.Slot].Count += moveCount
				i.inventory[slotID].Count -= moveCount
				if i.inventory[slotID].Count == 0 {
					i.inventory[slotID].ItemType = 0
				}

				haveCount = i.container[child.Slot].Count
				moveCount = child.Count - haveCount
				if child.Count < haveCount {
					// This should never happened, or there happened some underlying internal problems
					panic("stepMergeToContainer: Should never happened")
				}
				if moveCount == 0 {
					break
				}
			}
		}
	}

	// Step 4: Commit changes
	success, _, _, err = transaction.Commit()
	if err != nil {
		return false, fmt.Errorf("stepMergeToContainer: %v", err)
	}
	if !success {
		return false, fmt.Errorf("stepMergeToContainer: Complex item stack operation failed")
	}

	parentExistMoreThanSituation := false

	// Step 5: Check can stop
	for _, group := range i.ItemGroups {
		count := i.container[group.Parent.Slot].Count
		if group.Parent.Count < count {
			parentExistMoreThanSituation = true
		}
		if group.Parent.Count > count {
			return false, nil
		}
		for _, child := range group.Child {
			if child.Count < i.container[child.Slot].Count {
				// This should never happened, or there happened some underlying internal problems
				panic("stepMergeToContainer: Should never happened")
			}
			if child.Count > i.container[child.Slot].Count {
				return false, nil
			}
		}
	}

	// Step 6: Clean inventory
	_, err = i.commands.SendWSCommandWithResp("clear")
	if err != nil {
		return false, fmt.Errorf("stepMergeToContainer: %v", err)
	}

	// Step 7: Return
	if !parentExistMoreThanSituation {
		return true, nil
	}

	// Step 8.1: Clean extra item from each parent
	for itemType, group := range i.ItemGroups {
		count := i.container[group.Parent.Slot].Count
		if group.Parent.Count > count {
			// This should never happened, or there happened some underlying internal problems
			panic("stepMergeToContainer: Should never happened")
		}
		if group.Parent.Count == count {
			continue
		}

		cleanCount := count - group.Parent.Count
		airSlot, found := i.searchAirFromInventory()
		if !found {
			// This should never happened, or there happened some underlying internal problems
			panic("stepMergeToContainer: Should never happened")
		}
		_ = transaction.MoveToInventory(group.Parent.Slot, airSlot, cleanCount)

		i.container[group.Parent.Slot].Count -= cleanCount
		i.inventory[group.Parent.Slot] = ItemInfo{
			Count:    cleanCount,
			ItemType: itemType,
		}
	}

	// Step 8.2: Commit changes
	success, _, _, err = transaction.Commit()
	if err != nil {
		return false, fmt.Errorf("stepMergeToContainer: %v", err)
	}
	if !success {
		return false, fmt.Errorf("stepMergeToContainer: Commit clean failed")
	}

	// Step 8.3: Clean inventory
	_, err = i.commands.SendWSCommandWithResp("clear")
	if err != nil {
		return false, fmt.Errorf("stepMergeToContainer: %v", err)
	}

	// Step 9: Return
	return true, nil
}

// searchItemFromInventory 从背包查找物品类型为 itemType 的一个物品
func (i *ItemCopy) searchItemFromInventory(itemType ItemType) (
	slotID resources_control.SlotID,
	item ItemInfo,
	found bool,
) {
	for index, item := range i.inventory {
		if item.Count == 0 {
			continue
		}
		if item.ItemType == itemType {
			return resources_control.SlotID(index), item, true
		}
	}
	return 0, ItemInfo{}, false
}

// searchAirFromInventory 从背包查找一个空气物品
func (i *ItemCopy) searchAirFromInventory() (
	slotID resources_control.SlotID,
	found bool,
) {
	for index, item := range i.inventory {
		if item.Count == 0 {
			return resources_control.SlotID(index), true
		}
	}
	return 0, false
}
