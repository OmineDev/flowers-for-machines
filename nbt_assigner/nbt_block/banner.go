package nbt_block

import (
	"fmt"
	"strings"

	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol"
	"github.com/OmineDev/flowers-for-machines/game_control/game_interface"
	"github.com/OmineDev/flowers-for-machines/game_control/resources_control"
	"github.com/OmineDev/flowers-for-machines/nbt_assigner/block_helper"
	nbt_assigner_interface "github.com/OmineDev/flowers-for-machines/nbt_assigner/interface"
	"github.com/OmineDev/flowers-for-machines/nbt_assigner/nbt_cache"
	"github.com/OmineDev/flowers-for-machines/nbt_assigner/nbt_console"
	nbt_parser_block "github.com/OmineDev/flowers-for-machines/nbt_parser/block"
	nbt_parser_general "github.com/OmineDev/flowers-for-machines/nbt_parser/general"
	nbt_parser_item "github.com/OmineDev/flowers-for-machines/nbt_parser/item"
)

// 旗帜
type Banner struct {
	console *nbt_console.Console
	cache   *nbt_cache.NBTCacheSystem
	data    nbt_parser_block.Banner
}

func (Banner) Offset() protocol.BlockPos {
	return protocol.BlockPos{0, 0, 0}
}

func (b *Banner) Make() error {
	var resultSlotID resources_control.SlotID
	api := b.console.API()

	// 前置准备
	blockFacing := 1
	helperBannerBlock := "minecraft:standing_banner"
	if strings.Contains(b.data.BlockName(), "wall") {
		blockFacing = 2
		helperBannerBlock = "minecraft:wall_banner"
	}

	// 清空操作台中心处的方块
	err := b.console.API().SetBlock().SetBlock(b.console.Center(), "minecraft:air", "[]")
	if err != nil {
		return fmt.Errorf("Make: %v", err)
	}
	b.console.UseHelperBlock(nbt_console.RequesterUser, nbt_console.ConsoleIndexCenterBlock, block_helper.Air{})

	// 这是一个需要通过织布机来制作的旗帜。
	// 否则，这是一个可以直接通过命令得到的旗帜
	if len(b.data.NBT.Patterns) > 0 || b.data.NBT.Type == nbt_parser_general.BannerTypeOminous {
		// 取得生成旗帜所需要的旗帜物品
		bannerItem := nbt_assigner_interface.MakeNBTItemMethod(
			b.console,
			b.cache,
			&nbt_parser_item.Banner{
				DefaultItem: nbt_parser_item.DefaultItem{
					Basic: nbt_parser_item.ItemBasicData{
						Name:     "minecraft:banner",
						Count:    1,
						Metadata: int16(b.data.NBT.Base),
					},
				},
				NBT: nbt_parser_item.BannerNBT{
					Patterns: b.data.NBT.Patterns,
					Type:     b.data.NBT.Type,
				},
			},
		)
		if len(bannerItem) != 1 {
			panic("Make: Should never happened")
		}

		// 制作旗帜物品
		resultSlot, err := bannerItem[0].Make()
		if err != nil {
			return fmt.Errorf("Make: %v", err)
		}
		if len(resultSlot) != 1 {
			panic("Make: Should never happened")
		}

		// 确定旗帜所在的槽位
		for _, slotID := range resultSlot {
			resultSlotID = slotID
		}
	} else {
		err = b.console.API().Replaceitem().ReplaceitemInInventory(
			"@s",
			game_interface.ReplacePathHotbarOnly,
			game_interface.ReplaceitemInfo{
				Name:     "minecraft:banner",
				Count:    1,
				MetaData: int16(b.data.NBT.Base),
				Slot:     b.console.HotbarSlotID(),
			},
			"",
			true,
		)
		if err != nil {
			return fmt.Errorf("Make: %v", err)
		}
		resultSlotID = b.console.HotbarSlotID()
	}

	// 如果目标物品所在位置不在快捷栏中，
	// 则将该物品移动到快捷栏中
	if resultSlotID > 8 {
		err = api.Replaceitem().ReplaceitemInInventory(
			"@s",
			game_interface.ReplacePathHotbarOnly,
			game_interface.ReplaceitemInfo{
				Name:     "minecraft:air",
				Count:    1,
				MetaData: 0,
				Slot:     b.console.HotbarSlotID(),
			},
			"",
			true,
		)
		if err != nil {
			return fmt.Errorf("Make: %v", err)
		}
		b.console.UseInventorySlot(nbt_console.RequesterUser, b.console.HotbarSlotID(), false)

		success, err := api.ContainerOpenAndClose().OpenInventory()
		if err != nil {
			return fmt.Errorf("Make: %v", err)
		}
		if !success {
			return fmt.Errorf("Make: Failed to open the inventory")
		}

		success, _, _, err = api.ItemStackOperation().OpenTransaction().
			MoveBetweenInventory(resultSlotID, b.console.HotbarSlotID(), 1).
			Commit()
		if err != nil {
			_ = api.ContainerOpenAndClose().CloseContainer()
			return fmt.Errorf("Make: %v", err)
		}
		if !success {
			_ = api.ContainerOpenAndClose().CloseContainer()
			return fmt.Errorf("Make: The server rejected the stack request action")
		}

		err = api.ContainerOpenAndClose().CloseContainer()
		if err != nil {
			return fmt.Errorf("Make: %v", err)
		}

		resultSlotID = b.console.HotbarSlotID()
	}

	// 切换物品栏，如果需要的话
	if resultSlotID != b.console.HotbarSlotID() {
		err = b.console.ChangeAndUpdateHotbarSlotID(resultSlotID)
		if err != nil {
			return fmt.Errorf("Make: %v", err)
		}
	}

	// 前往操作台中心处
	err = b.console.CanReachOrMove(b.console.Center())
	if err != nil {
		return fmt.Errorf("Make: %v", err)
	}

	// 放置旗帜
	_, offsetPos, err := api.BotClick().PlaceBlockHighLevel(
		b.console.Center(),
		b.console.Position(),
		b.console.HotbarSlotID(),
		uint8(blockFacing),
	)
	if err != nil {
		return fmt.Errorf("Make: %v", err)
	}
	b.console.UseHelperBlock(nbt_console.RequesterUser, nbt_console.ConsoleIndexCenterBlock, block_helper.ComplexBlock{
		KnownStates: false,
		Name:        helperBannerBlock,
	})
	*b.console.NearBlockByIndex(nbt_console.ConsoleIndexCenterBlock, offsetPos) = block_helper.NearBlock{
		Name: game_interface.BasePlaceBlock,
	}

	// 覆写旗帜的方块状态
	err = api.SetBlock().SetBlock(b.console.Center(), b.data.BlockName(), b.data.BlockStatesString())
	if err != nil {
		return fmt.Errorf("Make: %v", err)
	}
	b.console.UseHelperBlock(nbt_console.RequesterUser, nbt_console.ConsoleIndexCenterBlock, block_helper.ComplexBlock{
		KnownStates: true,
		Name:        b.data.BlockName(),
		States:      b.data.BlockStates(),
	})

	return nil
}
