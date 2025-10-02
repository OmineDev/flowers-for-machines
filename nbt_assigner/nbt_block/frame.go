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
	nbt_parser_block "github.com/OmineDev/flowers-for-machines/nbt_parser/block"
	nbt_hash "github.com/OmineDev/flowers-for-machines/nbt_parser/hash"
	nbt_parser_item "github.com/OmineDev/flowers-for-machines/nbt_parser/item"
	"github.com/OmineDev/flowers-for-machines/utils"
)

// 物品展示框
type Frame struct {
	console *nbt_console.Console
	cache   *nbt_cache.NBTCacheSystem
	data    nbt_parser_block.Frame
}

func (Frame) Offset() protocol.BlockPos {
	return protocol.BlockPos{0, 0, 0}
}

// processComplex 处理复杂的物品
func (f *Frame) processComplex() (canUseCommand bool, resultSlot resources_control.SlotID, err error) {
	api := f.console.API()
	underlying := f.data.NBT.Item.UnderlyingItem()
	defaultItem := underlying.(*nbt_parser_item.DefaultItem)

	// 子方块
	if defaultItem.Block.SubBlock != nil {
		if !defaultItem.Block.SubBlock.NeedSpecialHandle() {
			return true, 0, nil
		}
		_, _, _, err = nbt_assigner_interface.PlaceNBTBlock(f.console, f.cache, defaultItem.Block.SubBlock)
		if err != nil {
			return false, 0, fmt.Errorf("processComplex: %v", err)
		}

		_, hit, partHit, err := f.cache.NBTBlockCache().LoadCache(nbt_hash.CompletelyHashNumber{
			HashNumber:    nbt_hash.NBTBlockFullHash(defaultItem.Block.SubBlock),
			SetHashNumber: nbt_hash.ContainerSetHash(defaultItem.Block.SubBlock),
		})
		if err != nil {
			return false, 0, fmt.Errorf("processComplex: %v", err)
		}
		if !hit || partHit {
			panic("processComplex: Should never happened")
		}

		_, err = f.console.API().Commands().SendWSCommandWithResp("clear")
		if err != nil {
			return false, 0, fmt.Errorf("processComplex: %v", err)
		}
		f.console.CleanInventory()

		success, currentSlot, err := api.BotClick().PickBlock(f.console.Center(), true)
		if err != nil || !success {
			_ = f.console.ChangeAndUpdateHotbarSlotID(nbt_console.DefaultHotbarSlot)
		}
		if err != nil {
			return false, 0, fmt.Errorf("processComplex: %v", err)
		}
		if !success {
			return false, 0, fmt.Errorf("processComplex: Failed to pick block due to unknown reason")
		}
		f.console.UpdateHotbarSlotID(currentSlot)
		f.console.UseInventorySlot(nbt_console.RequesterUser, currentSlot, true)

		return false, currentSlot, nil
	}

	// 复杂 NBT 物品制作
	methods := nbt_assigner_interface.MakeNBTItemMethod(f.console, f.cache, f.data.NBT.Item)
	if len(methods) != 1 {
		panic("Make: Should never happened")
	}
	resultSlotMapping, err := methods[0].Make()
	if err != nil {
		return false, 0, fmt.Errorf("processComplex: %v", err)
	}
	if len(resultSlotMapping) != 1 {
		panic("Make: Should never happened")
	}

	// 将复杂 NBT 物品移动到快捷栏
	for _, slotID := range resultSlotMapping {
		resultSlot = slotID
	}
	if resultSlot > 8 {
		err = api.Replaceitem().ReplaceitemInInventory(
			"@s",
			game_interface.ReplacePathHotbarOnly,
			game_interface.ReplaceitemInfo{
				Name:     "minecraft:air",
				Count:    1,
				MetaData: 0,
				Slot:     f.console.HotbarSlotID(),
			},
			"",
			true,
		)
		if err != nil {
			return false, 0, fmt.Errorf("processComplex: %v", err)
		}
		f.console.UseInventorySlot(nbt_console.RequesterUser, f.console.HotbarSlotID(), false)

		success, err := api.ContainerOpenAndClose().OpenInventory()
		if err != nil {
			return false, 0, fmt.Errorf("processComplex: %v", err)
		}
		if !success {
			return false, 0, fmt.Errorf("processComplex: %v", err)
		}

		success, _, _, err = api.ItemStackOperation().OpenTransaction().
			MoveBetweenInventory(resultSlot, f.console.HotbarSlotID(), 1).
			Commit()
		if err != nil {
			_ = api.ContainerOpenAndClose().CloseContainer()
			return false, 0, fmt.Errorf("processComplex: %v", err)
		}
		if !success {
			_ = api.ContainerOpenAndClose().CloseContainer()
			return false, 0, fmt.Errorf("processComplex: The server rejected the stack request action")
		}

		err = api.ContainerOpenAndClose().CloseContainer()
		if err != nil {
			return false, 0, fmt.Errorf("processComplex: %v", err)
		}

		resultSlot = f.console.HotbarSlotID()
	}

	return false, resultSlot, nil
}

func (f *Frame) Make() error {
	var canUseCommand bool
	var resultSlot resources_control.SlotID
	var err error
	api := f.console.API()

	// 如果这是一个复杂的物品
	if f.data.NBT.Item.IsComplex() {
		canUseCommand, resultSlot, err = f.processComplex()
		if err != nil {
			return fmt.Errorf("Make: %v", err)
		}
	} else {
		canUseCommand = true
	}

	// canUseCommand 指示可以先使用命令获取目标物品
	if canUseCommand {
		underlying := f.data.NBT.Item.UnderlyingItem()
		defaultItem := underlying.(*nbt_parser_item.DefaultItem)

		err = f.console.API().Replaceitem().ReplaceitemInInventory(
			"@s",
			game_interface.ReplacePathHotbarOnly,
			game_interface.ReplaceitemInfo{
				Name:     f.data.NBT.Item.ItemName(),
				Count:    1,
				MetaData: f.data.NBT.Item.ItemMetadata(),
				Slot:     f.console.HotbarSlotID(),
			},
			utils.MarshalItemComponent(defaultItem.Enhance.ItemComponent),
			true,
		)
		if err != nil {
			return fmt.Errorf("Make: %v", err)
		}

		f.console.UseInventorySlot(nbt_console.RequesterUser, f.console.HotbarSlotID(), true)
		resultSlot = f.console.HotbarSlotID()
	}

	// 切换物品栏，如果需要的话
	if resultSlot != f.console.HotbarSlotID() {
		err = f.console.ChangeAndUpdateHotbarSlotID(resultSlot)
		if err != nil {
			return fmt.Errorf("Make: %v", err)
		}
	}

	// 如果这个物品需要重命名或附魔
	if f.data.NBT.Item.NeedEnchOrRename() {
		underlying := f.data.NBT.Item.UnderlyingItem()
		defaultItem := underlying.(*nbt_parser_item.DefaultItem)

		// 附魔处理
		for _, ench := range defaultItem.Enhance.EnchList {
			err = api.Commands().SendSettingsCommand(fmt.Sprintf("enchant @s %d %d", ench.ID, ench.Level), true)
			if err != nil {
				return fmt.Errorf("Make: %v", err)
			}
		}
		if len(defaultItem.Enhance.EnchList) > 0 {
			err = api.Commands().AwaitChangesGeneral()
			if err != nil {
				return fmt.Errorf("Make: %v", err)
			}
		}

		// 物品改名处理
		if len(defaultItem.Enhance.DisplayName) > 0 {
			index, err := f.console.FindOrGenerateNewAnvil()
			if err != nil {
				return fmt.Errorf("Make: %v", err)
			}

			success, err := f.console.OpenContainerByIndex(index)
			if err != nil {
				return fmt.Errorf("Make: %v", err)
			}
			if !success {
				return fmt.Errorf("Make: Failed to open the anvil who at %#v", f.console.BlockPosByIndex(index))
			}

			success, _, _, err = api.ItemStackOperation().OpenTransaction().
				RenameInventoryItem(resultSlot, defaultItem.Enhance.DisplayName).
				Commit()
			if err != nil {
				_ = api.ContainerOpenAndClose().CloseContainer()
				return fmt.Errorf("Make: %v", err)
			}
			if !success {
				_ = api.ContainerOpenAndClose().CloseContainer()
				return fmt.Errorf("Make: The server rejected the renaming operation")
			}

			err = api.ContainerOpenAndClose().CloseContainer()
			if err != nil {
				return fmt.Errorf("Make: %v", err)
			}
		}
	}

	// 将操作台中心处的方块设置为空气
	err = f.console.API().SetBlock().SetBlock(f.console.Center(), "minecraft:air", "[]")
	if err != nil {
		return fmt.Errorf("Make: %v", err)
	}
	f.console.UseHelperBlock(nbt_console.RequesterUser, nbt_console.ConsoleIndexCenterBlock, block_helper.Air{})

	// 检查物品展示框的地板
	needFloorBlock := false
	nearBlock := f.console.NearBlockByIndex(nbt_console.ConsoleIndexCenterBlock, protocol.BlockPos{0, -1, 0})
	switch (*nearBlock).(type) {
	case block_helper.Air, block_helper.ComplexBlock:
		needFloorBlock = true
	}

	// 如果物品展示框没有地板，则生成地板
	if needFloorBlock {
		floorblockPos := f.console.NearBlockPosByIndex(nbt_console.ConsoleIndexCenterBlock, protocol.BlockPos{0, -1, 0})
		err = api.SetBlock().SetBlock(floorblockPos, game_interface.BasePlaceBlock, "[]")
		if err != nil {
			return fmt.Errorf("Make: %v", err)
		}
		*f.console.NearBlockByIndex(nbt_console.ConsoleIndexCenterBlock, protocol.BlockPos{0, -1, 0}) = block_helper.NearBlock{
			Name: game_interface.BasePlaceBlock,
		}
	}

	// 放置物品展示框
	err = f.console.API().SetBlock().SetBlock(
		f.console.Center(),
		f.data.BlockName(),
		`["facing_direction"=1,"item_frame_map_bit"=false,"item_frame_photo_bit"=false]`,
	)
	if err != nil {
		return fmt.Errorf("Make: %v", err)
	}
	f.console.UseHelperBlock(nbt_console.RequesterUser, nbt_console.ConsoleIndexCenterBlock, block_helper.ComplexBlock{
		KnownStates: false,
		Name:        f.data.BlockName(),
	})

	// 前往操作台中心处
	err = f.console.CanReachOrMove(f.console.Center())
	if err != nil {
		return fmt.Errorf("Make: %v", err)
	}

	// 点击物品展示框
	for range 1 + int(f.data.NBT.ItemRotation/45) {
		err = api.BotClick().ClickBlock(game_interface.UseItemOnBlocks{
			HotbarSlotID: f.console.HotbarSlotID(),
			BotPos:       f.console.Position(),
			BlockPos:     f.console.Center(),
			BlockName:    f.data.BlockName(),
			BlockStates: map[string]any{
				"facing_direction":     int32(1),
				"item_frame_map_bit":   byte(0),
				"item_frame_photo_bit": byte(0),
			},
		})
		if err != nil {
			return fmt.Errorf("Make: %v", err)
		}
	}

	// 覆写物品展示框的方块状态
	err = api.SetBlock().SetBlock(f.console.Center(), f.data.BlockName(), f.data.BlockStatesString())
	if err != nil {
		return fmt.Errorf("Make: %v", err)
	}
	f.console.UseHelperBlock(nbt_console.RequesterUser, nbt_console.ConsoleIndexCenterBlock, block_helper.ComplexBlock{
		KnownStates: true,
		Name:        f.data.BlockName(),
		States:      f.data.BlockStates(),
	})

	return nil
}
