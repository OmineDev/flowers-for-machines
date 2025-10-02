package nbt_block

import (
	"fmt"

	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol"
	"github.com/OmineDev/flowers-for-machines/game_control/game_interface"
	"github.com/OmineDev/flowers-for-machines/game_control/resources_control"
	"github.com/OmineDev/flowers-for-machines/nbt_assigner/block_helper"
	nbt_assigner_interface "github.com/OmineDev/flowers-for-machines/nbt_assigner/interface"
	"github.com/OmineDev/flowers-for-machines/nbt_assigner/nbt_cache"
	"github.com/OmineDev/flowers-for-machines/nbt_assigner/nbt_console"
	nbt_assigner_utils "github.com/OmineDev/flowers-for-machines/nbt_assigner/utils"
	nbt_parser_block "github.com/OmineDev/flowers-for-machines/nbt_parser/block"
	nbt_hash "github.com/OmineDev/flowers-for-machines/nbt_parser/hash"
	nbt_parser_interface "github.com/OmineDev/flowers-for-machines/nbt_parser/interface"
	nbt_parser_item "github.com/OmineDev/flowers-for-machines/nbt_parser/item"
	"github.com/OmineDev/flowers-for-machines/utils"
)

// 容器
type Container struct {
	console *nbt_console.Console
	cache   *nbt_cache.NBTCacheSystem
	data    nbt_parser_block.Container
}

func (Container) Offset() protocol.BlockPos {
	return protocol.BlockPos{0, 0, 0}
}

func (c *Container) Make() error {
	err := c.makeNormal()

	_, _ = c.console.API().Commands().SendWSCommandWithResp("clear")
	c.console.CleanInventory()
	_ = c.console.ChangeAndUpdateHotbarSlotID(nbt_console.DefaultHotbarSlot)

	if err != nil {
		return fmt.Errorf("Make: %v", err)
	}

	c.console.UseHelperBlock(
		nbt_console.RequesterUser,
		nbt_console.ConsoleIndexCenterBlock,
		block_helper.ContainerBlockHelper{
			OpenInfo: block_helper.ContainerBlockOpenInfo{
				Name:                  c.data.BlockName(),
				States:                c.data.BlockStates(),
				ConsiderOpenDirection: c.data.ConsiderOpenDirection(),
				ShulkerFacing:         c.data.NBT.ShulkerFacing,
			},
		},
	)
	return nil
}

func (c *Container) makeNormal() error {
	api := c.console.API()

	// Step 1: 检查该容器是否命中集合校验和
	{
		// 尝试从底层缓存命中系统加载
		structure, hit, isSetHashHit, err := c.cache.NBTBlockCache().LoadCache(
			nbt_hash.CompletelyHashNumber{
				HashNumber:    nbt_hash.NBTBlockFullHash(&c.data),
				SetHashNumber: nbt_hash.ContainerSetHash(&c.data),
			},
		)
		if err != nil {
			return fmt.Errorf("makeNormal: %v", err)
		}
		if hit && !isSetHashHit {
			panic("makeNormal: Should never happened")
		}

		// 如果我们命中了集合哈希校验和
		if isSetHashHit {
			container, ok := structure.Block.(*nbt_parser_block.Container)
			if !ok {
				panic("makeNormal: Should never happened")
			}

			err = nbt_assigner_utils.ItemTransition(c.console, c.cache, *container, c.data)
			if err != nil {
				return fmt.Errorf("makeNormal: %v", err)
			}

			return nil
		}
	}

	// Step 2: 构造物品树 (仅限复杂物品或需要处理的子方块)
	itemTypeIndex := game_interface.ItemType(0)
	itemTypes := make(map[uint64]game_interface.ItemType)
	itemGroups := make(map[uint64][]nbt_parser_block.ItemWithSlot)
	for _, item := range c.data.NBT.Items {
		if !item.Item.IsComplex() {
			continue
		}
		hashNumber := nbt_hash.NBTItemNBTHash(item.Item)
		itemGroups[hashNumber] = append(itemGroups[hashNumber], item)
		if _, ok := itemTypes[hashNumber]; !ok {
			itemTypes[hashNumber] = itemTypeIndex
			itemTypeIndex++
		}
	}

	// Step 3.1: 找出部分命中和没有命中的需要处理的子方块 (找出集合)
	allSubBlocks := make([]int, 0)
	allSubBlocksSet := make(map[uint64]bool)
	subBlockPartHit := make([]int, 0)
	subBlockNotHit := make([]int, 0)
	for index, item := range c.data.NBT.Items {
		underlying := item.Item.UnderlyingItem().(*nbt_parser_item.DefaultItem)
		if underlying.Block.SubBlock == nil {
			continue
		}

		hashNumber := nbt_hash.NBTItemNBTHash(item.Item)
		if _, ok := allSubBlocksSet[hashNumber]; ok {
			continue
		}
		allSubBlocksSet[hashNumber] = true

		_, hit, partHit := c.cache.NBTBlockCache().CheckCache(nbt_hash.CompletelyHashNumber{
			HashNumber:    nbt_hash.NBTBlockFullHash(underlying.Block.SubBlock),
			SetHashNumber: nbt_hash.ContainerSetHash(underlying.Block.SubBlock),
		})

		if hit && partHit {
			subBlockPartHit = append(subBlockPartHit, index)
		}
		if !hit && !partHit {
			subBlockNotHit = append(subBlockNotHit, index)
		}
		allSubBlocks = append(allSubBlocks, index)
	}

	// Step 3.2: 处理部分命中的子方块 (容器)
	for _, index := range subBlockPartHit {
		item := c.data.NBT.Items[index]
		underlying := item.Item.UnderlyingItem().(*nbt_parser_item.DefaultItem)

		wantContainer, ok := underlying.Block.SubBlock.(*nbt_parser_block.Container)
		if !ok {
			panic("makeNormal: Should never happened")
		}

		structure, _, partHit, err := c.cache.NBTBlockCache().LoadCache(nbt_hash.CompletelyHashNumber{
			HashNumber:    nbt_hash.NBTBlockFullHash(wantContainer),
			SetHashNumber: nbt_hash.ContainerSetHash(wantContainer),
		})
		if err != nil {
			return fmt.Errorf("makeNormal: %v", err)
		}
		if !partHit {
			panic("makeNormal: Should never happened")
		}

		container, ok := structure.Block.(*nbt_parser_block.Container)
		if !ok {
			panic("makeNormal: Should never happened")
		}

		err = nbt_assigner_utils.ItemTransition(c.console, c.cache, *container, *wantContainer)
		if err != nil {
			return fmt.Errorf("makeNormal: %v", err)
		}

		err = c.cache.NBTBlockCache().StoreCache(wantContainer, protocol.BlockPos{0, 0, 0})
		if err != nil {
			return fmt.Errorf("makeNormal: %v", err)
		}
	}

	// Step 3.3: 处理没有命中的子方块
	for _, index := range subBlockNotHit {
		item := c.data.NBT.Items[index]
		underlying := item.Item.UnderlyingItem().(*nbt_parser_item.DefaultItem)
		_, _, _, err := nbt_assigner_interface.PlaceNBTBlock(c.console, c.cache, underlying.Block.SubBlock)
		if err != nil {
			return fmt.Errorf("makeNormal: %v", err)
		}
	}

	// Step 4: 生成当前容器
	err := nbt_assigner_utils.SpawnNewEmptyBlock(
		c.console,
		c.cache,
		nbt_assigner_utils.EmptyBlockData{
			Name:                  c.data.BlockName(),
			States:                c.data.BlockStates(),
			IsCanOpenConatiner:    true,
			ConsiderOpenDirection: c.data.ConsiderOpenDirection(),
			ShulkerFacing:         c.data.NBT.ShulkerFacing,
			BlockCustomName:       c.data.NBT.CustomName,
		},
	)
	if err != nil {
		return fmt.Errorf("makeNormal: %v", err)
	}

	// Step 5: 将子方块放入容器
	if len(allSubBlocks) > 0 {
		// 清空物品栏
		_, err := api.Commands().SendWSCommandWithResp("clear")
		if err != nil {
			return fmt.Errorf("makeNormal: %v", err)
		}

		// 占用所有物品栏，
		// 因为我们无法确保数据匹配
		for index := range 36 {
			c.console.UseInventorySlot(nbt_console.RequesterUser, resources_control.SlotID(index), true)
		}

		// 通过 Pick block 得到所有的子方块
		subBlocksPtr := 0
		for {
			// 这代表当前轮次已经完成了
			if subBlocksPtr >= len(allSubBlocks) {
				break
			}

			// 我们可以在一个时刻内使用 structure 命令
			// 加载多个子方块，这一步可以使用协程优化。
			// spaces 是这些可以使用的方块的位置
			spaces := c.console.FindMutipleSpaceToPlaceNewBlock(false)
			// 我们计算出当前轮次需要处理的子方块
			offset := min(len(spaces), len(allSubBlocks)-subBlocksPtr)
			currentRound := allSubBlocks[subBlocksPtr : subBlocksPtr+offset]
			// waiters 用于等待所有 structure 命令完成
			waiters := make([]chan struct{}, 0)

			// 我们以协程的方式生成当前轮次所需的所有子方块
			for spaceIndex, itemIndex := range currentRound {
				underlying := c.data.NBT.Items[itemIndex].Item.UnderlyingItem()
				subBlock := underlying.(*nbt_parser_item.DefaultItem).Block.SubBlock

				structure, hit, partHit := c.cache.NBTBlockCache().CheckCache(nbt_hash.CompletelyHashNumber{
					HashNumber:    nbt_hash.NBTBlockFullHash(subBlock),
					SetHashNumber: nbt_hash.ContainerSetHash(subBlock),
				})
				if !hit || partHit {
					panic("makeNormal: Should never happened")
				}

				newWaiter := make(chan struct{})
				waiters = append(waiters, newWaiter)
				subBlocksPtr++

				go func() {
					err := api.StructureBackup().RevertStructure(
						structure.UniqueID,
						c.console.BlockPosByIndex(spaces[spaceIndex]),
					)
					if err == nil {
						c.console.UseHelperBlock(nbt_console.RequesterUser, spaces[spaceIndex], block_helper.ComplexBlock{
							KnownStates: true,
							Name:        structure.Block.BlockName(),
							States:      structure.Block.BlockStates(),
						})
					}
					close(newWaiter)
				}()
			}

			// 等待所有 structure 命令完成
			for _, waiter := range waiters {
				<-waiter
			}

			// 然后，以阻塞的方式 Pick 到当前轮次生成的所有子方块
			for spaceIndex, itemIndex := range currentRound {
				index := spaces[spaceIndex]

				err = c.console.CanReachOrMove(c.console.BlockPosByIndex(index))
				if err != nil {
					return fmt.Errorf("makeNormal: %v", err)
				}

				success, currentSlot, err := api.BotClick().PickBlock(c.console.BlockPosByIndex(index), true)
				if err != nil || !success {
					_ = c.console.ChangeAndUpdateHotbarSlotID(nbt_console.DefaultHotbarSlot)
				}
				if err != nil {
					return fmt.Errorf("makeNormal: %v", err)
				}
				if !success {
					underlying := c.data.NBT.Items[itemIndex].Item.UnderlyingItem()
					subBlock := underlying.(*nbt_parser_item.DefaultItem).Block.SubBlock
					return fmt.Errorf("makeNormal: Failed to get sub block %#v by pick block", subBlock)
				}
				c.console.UpdateHotbarSlotID(currentSlot)
			}
		}

		// 现在所有子方块都被 Pick Block 到背包了
		allItemStack, inventoryExisted := api.Resources().Inventories().GetAllItemStack(0)
		if !inventoryExisted {
			panic("makeNormal: Should never happened")
		}

		// 打开操作台中心处容器
		success, err := c.console.OpenContainerByIndex(nbt_console.ConsoleIndexCenterBlock)
		if err != nil {
			return fmt.Errorf("makeNormal: %v", err)
		}
		if !success {
			return fmt.Errorf("makeNormal: Failed to open the container %#v when move sub block in it", c.data)
		}

		// 将背包中的每个子方块移动到对应的父节点处
		transaction := api.ItemStackOperation().OpenTransaction()
		for srcSlot, value := range allItemStack {
			if value.Stack.NetworkID == 0 || value.Stack.NetworkID == -1 {
				continue
			}

			newItem, err := nbt_parser_interface.ParseItemNetwork(
				value.Stack,
				api.Resources().ConstantPacket().ItemNameByNetworkID(value.Stack.NetworkID),
			)
			if err != nil {
				_ = api.ContainerOpenAndClose().CloseContainer()
				return fmt.Errorf("makeNormal: %v", err)
			}

			hashNumber := nbt_hash.NBTItemNBTHash(newItem)
			dstSlot := itemGroups[hashNumber][0].Slot

			_ = transaction.MoveToContainer(srcSlot, resources_control.SlotID(dstSlot), 1)
		}

		// 提交更改
		success, _, _, err = transaction.Commit()
		if err != nil {
			_ = api.ContainerOpenAndClose().CloseContainer()
			return fmt.Errorf("makeNormal: %v", err)
		}
		if !success {
			_ = api.ContainerOpenAndClose().CloseContainer()
			return fmt.Errorf("makeNormal: The server rejected the stack request action when move sub block in it")
		}

		// 关闭容器
		err = api.ContainerOpenAndClose().CloseContainer()
		if err != nil {
			return fmt.Errorf("makeNormal: %v", err)
		}
	}

	// Step 6.1: 计算出哪些物品是需要制作的非子方块复杂物品
	complexItemExcludeSubBlock := make([]nbt_parser_interface.Item, 0)
	for _, value := range itemGroups {
		if _, ok := value[0].Item.(*nbt_parser_item.DefaultItem); ok {
			continue
		}
		complexItemExcludeSubBlock = append(complexItemExcludeSubBlock, value[0].Item)
	}

	// Step 6.2: 制作非子方块的复杂物品
	variousItems := nbt_assigner_interface.MakeNBTItemMethod(c.console, c.cache, complexItemExcludeSubBlock...)
	for _, item := range variousItems {
		for {
			resultSlot, err := item.Make()
			if err != nil {
				return fmt.Errorf("makeNormal: %v", err)
			}
			if len(resultSlot) == 0 {
				break
			}

			success, err := c.console.OpenContainerByIndex(nbt_console.ConsoleIndexCenterBlock)
			if err != nil {
				return fmt.Errorf("makeNormal: %v", err)
			}
			if !success {
				return fmt.Errorf("makeNormal: Failed to open the container %#v when make complex item", c.data)
			}

			transaction := api.ItemStackOperation().OpenTransaction()
			for hashNumber, slotID := range resultSlot {
				dstSlot := resources_control.SlotID(itemGroups[hashNumber][0].Slot)
				_ = transaction.MoveToContainer(slotID, dstSlot, 1)
			}

			success, _, _, err = transaction.Commit()
			if err != nil {
				_ = api.ContainerOpenAndClose().CloseContainer()
				return fmt.Errorf("makeNormal: %v", err)
			}
			if !success {
				_ = api.ContainerOpenAndClose().CloseContainer()
				return fmt.Errorf("makeNormal: The server rejected item stack request action when make complex item")
			}
			for _, slotID := range resultSlot {
				c.console.UseInventorySlot(nbt_console.RequesterUser, slotID, false)
			}

			err = api.ContainerOpenAndClose().CloseContainer()
			if err != nil {
				return fmt.Errorf("makeNormal: %v", err)
			}
		}
	}

	// Step 7.1: 检测是否需要物品分裂
	needItemCopy := false
	for _, value := range itemGroups {
		if len(value) > 1 {
			needItemCopy = true
			break
		}
		if value[0].Item.ItemCount() > 1 {
			needItemCopy = true
			break
		}
	}

	// Step 7.2: 物品分裂 (复杂物品复制)
	if needItemCopy {
		// 清理背包
		_, err = api.Commands().SendWSCommandWithResp("clear")
		if err != nil {
			return fmt.Errorf("makeNormal: %v", err)
		}
		c.console.CleanInventory()

		// 打开容器
		success, err := c.console.OpenContainerByIndex(nbt_console.ConsoleIndexCenterBlock)
		if err != nil {
			return fmt.Errorf("makeNormal: %v", err)
		}
		if !success {
			return fmt.Errorf("makeNormal: Failed to open the container %#v when do item copy", c.data)
		}

		// 将容器中现存的所有物品拿回
		transaction := api.ItemStackOperation().OpenTransaction()
		for _, value := range itemGroups {
			_ = transaction.MoveToInventory(
				resources_control.SlotID(value[0].Slot),
				resources_control.SlotID(value[0].Slot),
				1,
			)
		}

		// 提交更改
		success, _, _, err = transaction.Commit()
		if err != nil {
			_ = api.ContainerOpenAndClose().CloseContainer()
			return fmt.Errorf("makeNormal: %v", err)
		}
		if !success {
			_ = api.ContainerOpenAndClose().CloseContainer()
			return fmt.Errorf("makeNormal: Failed to move item from container %#v when do item copy", c.data)
		}

		// 关闭容器
		err = api.ContainerOpenAndClose().CloseContainer()
		if err != nil {
			return fmt.Errorf("makeNormal: %v", err)
		}

		// 构造基物品
		baseItems := make([]game_interface.ItemInfoWithSlot, 0)
		for _, value := range itemGroups {
			baseItems = append(baseItems, game_interface.ItemInfoWithSlot{
				Slot: resources_control.SlotID(value[0].Slot),
				ItemInfo: game_interface.ItemInfo{
					Count:    1,
					ItemType: itemTypes[nbt_hash.NBTItemNBTHash(value[0].Item)],
				},
			})
		}

		// 构造蓝图
		targetItems := make([]*game_interface.ItemInfo, 27)
		for _, value := range itemGroups {
			for _, val := range value {
				targetItems[val.Slot] = &game_interface.ItemInfo{
					Count:    val.Item.ItemCount(),
					ItemType: itemTypes[nbt_hash.NBTItemNBTHash(val.Item)],
				}
			}
		}

		// 物品分裂
		err = api.ItemCopy().CopyItem(
			game_interface.UseItemOnBlocks{
				HotbarSlotID: c.console.HotbarSlotID(),
				BotPos:       c.console.Position(),
				BlockPos:     c.console.Center(),
				BlockName:    c.data.BlockName(),
				BlockStates:  c.data.BlockStates(),
			},
			baseItems, targetItems,
		)
		if err != nil {
			api.Commands().SendWSCommandWithResp("clear")
			c.console.CleanInventory()
			return fmt.Errorf("makeNormal: %v", err)
		}

		// 清理背包
		_, err = api.Commands().SendWSCommandWithResp("clear")
		if err != nil {
			return fmt.Errorf("makeNormal: %v", err)
		}
		c.console.CleanInventory()
	}

	// Step 8.1: 填充剩余物品
	for _, item := range c.data.NBT.Items {
		if item.Item.IsComplex() {
			continue
		}
		underlying := item.Item.UnderlyingItem().(*nbt_parser_item.DefaultItem)

		err = api.Replaceitem().ReplaceitemInContainerAsync(
			c.console.Center(),
			game_interface.ReplaceitemInfo{
				Name:     item.Item.ItemName(),
				Count:    item.Item.ItemCount(),
				MetaData: item.Item.ItemMetadata(),
				Slot:     resources_control.SlotID(item.Slot),
			},
			utils.MarshalItemComponent(underlying.Enhance.ItemComponent),
		)
		if err != nil {
			return fmt.Errorf("makeNormal: %v", err)
		}
	}

	// Step 8.2: 等待更改
	err = api.Commands().AwaitChangesGeneral()
	if err != nil {
		return fmt.Errorf("makeNormal: %v", err)
	}

	// Step 9.1: 找出所有需要修改物品名称或需要附魔的物品
	enchOrRenameList := make([]int, 0)
	for index, value := range c.data.NBT.Items {
		if value.Item.NeedEnchOrRename() {
			enchOrRenameList = append(enchOrRenameList, index)
		}
	}

	// Step 9.2: 将需要修改物品名称或需要附魔的物品移动到背包
	if len(enchOrRenameList) > 0 {
		_, err = api.Commands().SendWSCommandWithResp("clear")
		if err != nil {
			return fmt.Errorf("makeNormal: %v", err)
		}
		c.console.CleanInventory()

		success, err := c.console.OpenContainerByIndex(nbt_console.ConsoleIndexCenterBlock)
		if err != nil {
			return fmt.Errorf("makeNormal: %v", err)
		}
		if !success {
			return fmt.Errorf("makeNormal: Failed to open the container %#v when do ench or rename operation", c.data)
		}

		transaction := api.ItemStackOperation().OpenTransaction()
		for _, index := range enchOrRenameList {
			item := c.data.NBT.Items[index]
			_ = transaction.MoveToInventory(
				resources_control.SlotID(item.Slot),
				resources_control.SlotID(item.Slot+9),
				item.Item.ItemCount(),
			)
		}

		success, _, _, err = transaction.Commit()
		if err != nil {
			_ = api.ContainerOpenAndClose().CloseContainer()
			return fmt.Errorf("makeNormal: %v", err)
		}
		if !success {
			_ = api.ContainerOpenAndClose().CloseContainer()
			return fmt.Errorf("makeNormal: The server rejected the stack request action when do ench or rename operation")
		}

		err = api.ContainerOpenAndClose().CloseContainer()
		if err != nil {
			return fmt.Errorf("makeNormal: %v", err)
		}
	}

	// Step 9.3: 物品附魔或重命名操作
	if len(enchOrRenameList) > 0 {
		multipleItems := [27]*nbt_parser_interface.Item{}
		for _, index := range enchOrRenameList {
			item := c.data.NBT.Items[index]
			multipleItems[item.Slot] = &item.Item
		}
		err = nbt_assigner_interface.EnchAndRenameMultiple(c.console, multipleItems)
		if err != nil {
			return fmt.Errorf("makeNormal: %v", err)
		}
	}

	// Step 9.4: 将物品移动回容器
	if len(enchOrRenameList) > 0 {
		success, err := c.console.OpenContainerByIndex(nbt_console.ConsoleIndexCenterBlock)
		if err != nil {
			return fmt.Errorf("makeNormal: %v", err)
		}
		if !success {
			return fmt.Errorf("makeNormal: Failed to open the container %#v when finish ench or rename operation", c.data)
		}

		transaction := api.ItemStackOperation().OpenTransaction()
		for _, index := range enchOrRenameList {
			item := c.data.NBT.Items[index]
			_ = transaction.MoveToContainer(
				resources_control.SlotID(item.Slot+9),
				resources_control.SlotID(item.Slot),
				item.Item.ItemCount(),
			)
		}

		success, _, _, err = transaction.Commit()
		if err != nil {
			_ = api.ContainerOpenAndClose().CloseContainer()
			return fmt.Errorf("makeNormal: %v", err)
		}
		if !success {
			_ = api.ContainerOpenAndClose().CloseContainer()
			return fmt.Errorf("makeNormal: The server rejected the stack request action when finish ench or rename operation")
		}

		err = api.ContainerOpenAndClose().CloseContainer()
		if err != nil {
			return fmt.Errorf("makeNormal: %v", err)
		}
	}

	// Step 10: 返回值
	return nil
}
