package nbt_block

import (
	"fmt"

	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol"
	"github.com/OmineDev/flowers-for-machines/game_control/game_interface"
	"github.com/OmineDev/flowers-for-machines/nbt_assigner/nbt_cache"
	"github.com/OmineDev/flowers-for-machines/nbt_assigner/nbt_console"
	nbt_assigner_utils "github.com/OmineDev/flowers-for-machines/nbt_assigner/utils"
	nbt_parser_block "github.com/OmineDev/flowers-for-machines/nbt_parser/block"
	nbt_parser_item "github.com/OmineDev/flowers-for-machines/nbt_parser/item"
	"github.com/OmineDev/flowers-for-machines/utils"
)

// 唱片机
type JukeBox struct {
	console *nbt_console.Console
	cache   *nbt_cache.NBTCacheSystem
	data    nbt_parser_block.JukeBox
}

func (JukeBox) Offset() protocol.BlockPos {
	return protocol.BlockPos{0, 0, 0}
}

func (j *JukeBox) Make() error {
	var defaultItem *nbt_parser_item.DefaultItem
	api := j.console.API()

	// 生成唱片机
	err := nbt_assigner_utils.SpawnNewEmptyBlock(
		j.console,
		j.cache,
		nbt_assigner_utils.EmptyBlockData{
			Name:               j.data.BlockName(),
			States:             j.data.BlockStates(),
			IsCanOpenConatiner: false,
			BlockCustomName:    j.data.NBT.CustomName,
		},
	)
	if err != nil {
		return fmt.Errorf("Make: %v", err)
	}

	// 如果唱片机中没有唱片，则应当直接返回值
	if j.data.NBT.HaveDisc {
		underlying := j.data.NBT.Disc.UnderlyingItem()
		defaultItem = underlying.(*nbt_parser_item.DefaultItem)
	} else {
		return nil
	}

	// 如果唱片可以直接使用命令放置
	if !j.data.NBT.Disc.NeedEnchOrRename() {
		err = api.Replaceitem().ReplaceitemInContainerAsync(
			j.console.Center(),
			game_interface.ReplaceitemInfo{
				Name:     j.data.NBT.Disc.ItemName(),
				Count:    j.data.NBT.Disc.ItemCount(),
				MetaData: j.data.NBT.Disc.ItemMetadata(),
				Slot:     0,
			},
			utils.MarshalItemComponent(defaultItem.Enhance.ItemComponent),
		)
		if err != nil {
			return fmt.Errorf("Make: %v", err)
		}

		err = api.Commands().AwaitChangesGeneral()
		if err != nil {
			return fmt.Errorf("Make: %v", err)
		}

		return nil
	}

	// 获取唱片到物品栏
	err = api.Replaceitem().ReplaceitemInInventory(
		"@s",
		game_interface.ReplacePathHotbarOnly,
		game_interface.ReplaceitemInfo{
			Name:     j.data.NBT.Disc.ItemName(),
			Count:    j.data.NBT.Disc.ItemCount(),
			MetaData: j.data.NBT.Disc.ItemMetadata(),
			Slot:     j.console.HotbarSlotID(),
		},
		utils.MarshalItemComponent(defaultItem.Enhance.ItemComponent),
		true,
	)
	if err != nil {
		return fmt.Errorf("Make: %v", err)
	}
	j.console.UseInventorySlot(nbt_console.RequesterUser, j.console.HotbarSlotID(), true)

	// 唱片改名
	if len(defaultItem.Enhance.DisplayName) != 0 {
		index, err := j.console.FindOrGenerateNewAnvil()
		if err != nil {
			return fmt.Errorf("Make: %v", err)
		}

		success, err := j.console.OpenContainerByIndex(index)
		if err != nil {
			return fmt.Errorf("Make: %v", err)
		}
		if !success {
			return fmt.Errorf("Make: Failed to open the anvil")
		}

		success, _, _, err = api.ItemStackOperation().OpenTransaction().
			RenameInventoryItem(j.console.HotbarSlotID(), defaultItem.Enhance.DisplayName).
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
	} else {
		panic("Make: Should never happened")
	}

	// 传送到操作台中心
	err = j.console.CanReachOrMove(j.console.Center())
	if err != nil {
		return fmt.Errorf("Make: %v", err)
	}

	// 放置唱片
	err = api.BotClick().ClickBlock(game_interface.UseItemOnBlocks{
		HotbarSlotID: j.console.HotbarSlotID(),
		BotPos:       j.console.Position(),
		BlockPos:     j.console.Center(),
		BlockName:    j.data.BlockName(),
		BlockStates:  j.data.BlockStates(),
	})
	if err != nil {
		return fmt.Errorf("Make: %v", err)
	}

	return nil
}
