package nbt_block

import (
	"fmt"
	"strings"

	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol"
	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol/packet"
	"github.com/OmineDev/flowers-for-machines/game_control/game_interface"
	"github.com/OmineDev/flowers-for-machines/mapping"
	"github.com/OmineDev/flowers-for-machines/nbt_assigner/block_helper"
	"github.com/OmineDev/flowers-for-machines/nbt_assigner/nbt_console"
	nbt_parser_block "github.com/OmineDev/flowers-for-machines/nbt_parser/block"
	"github.com/OmineDev/flowers-for-machines/utils"

	"github.com/go-gl/mathgl/mgl32"
)

// 告示牌
type Sign struct {
	console *nbt_console.Console
	data    nbt_parser_block.Sign
}

func (s *Sign) replaceitem(itemName string, block bool) error {
	err := s.console.API().Replaceitem().ReplaceitemInInventory(
		"@s",
		game_interface.ReplacePathHotbarOnly,
		game_interface.ReplaceitemInfo{
			Name:     itemName,
			Count:    1,
			MetaData: 0,
			Slot:     s.console.HotbarSlotID(),
		},
		"",
		block,
	)
	if err != nil {
		return fmt.Errorf("replaceitem: %v", err)
	}

	if itemName == "minecraft:air" {
		s.console.UseInventorySlot(nbt_console.RequesterUser, s.console.HotbarSlotID(), false)
	} else {
		s.console.UseInventorySlot(nbt_console.RequesterUser, s.console.HotbarSlotID(), true)
	}

	return nil
}

func (Sign) Offset() protocol.BlockPos {
	return protocol.BlockPos{0, 0, 0}
}

func (s *Sign) Make() error {
	api := s.console.API()

	// 前置准备
	helperSignBlock := "minecraft:cherry_hanging_sign"
	helperBlockStates := `["attached_bit"=false,"facing_direction"=4,"ground_sign_direction"=0,"hanging"=false]`
	if !strings.Contains(s.data.BlockName(), "hanging") {
		helperSignBlock = "minecraft:cherry_wall_sign"
		helperBlockStates = `["facing_direction"=4]`
	}

	// 初始化
	blockAction := game_interface.UseItemOnBlocks{
		HotbarSlotID: s.console.HotbarSlotID(),
		BotPos:       s.console.Position(),
		BlockPos:     s.console.Center(),
		BlockName:    helperSignBlock,
		BlockStates:  utils.ParseBlockStatesString(helperBlockStates),
	}

	// 清空手持物品栏以防止稍后在手持蜜脾的情况下点击告示牌，
	// 因为用 蜜脾 点击告示牌会导致告示牌被封装
	err := s.replaceitem("minecraft:air", false)
	if err != nil {
		return fmt.Errorf("Make: %v", err)
	}

	// 生成一个新的告示牌
	err = s.console.API().SetBlock().SetBlock(s.console.Center(), "minecraft:air", "[]")
	if err != nil {
		return fmt.Errorf("Make: %v", err)
	}
	s.console.UseHelperBlock(nbt_console.RequesterUser, nbt_console.ConsoleIndexCenterBlock, block_helper.Air{})
	err = s.console.API().SetBlock().SetBlock(s.console.Center(), helperSignBlock, helperBlockStates)
	if err != nil {
		return fmt.Errorf("Make: %v", err)
	}
	s.console.UseHelperBlock(nbt_console.RequesterUser, nbt_console.ConsoleIndexCenterBlock, block_helper.ComplexBlock{
		KnownStates: false,
		Name:        helperSignBlock,
	})

	// 打开告示牌并写入文本数据
	if len(s.data.NBT.FrontText.Text) > 0 || len(s.data.NBT.BackText.Text) > 0 {
		// 打开告示牌
		err = s.console.CanReachOrMove(s.console.Center())
		if err != nil {
			return fmt.Errorf("Make: %v", err)
		}
		err = api.BotClick().ClickBlock(blockAction)
		if err != nil {
			return fmt.Errorf("Make: %v", err)
		}

		// 确定告示牌 NBT 数据
		nbtMap := make(map[string]any)
		if len(s.data.NBT.FrontText.Text) > 0 {
			nbtMap["FrontText"] = map[string]any{"Text": s.data.NBT.FrontText.Text}
		}
		if len(s.data.NBT.BackText.Text) > 0 {
			nbtMap["BackText"] = map[string]any{"Text": s.data.NBT.BackText.Text}
		}

		// 写入告示牌 NBT 数据
		err = api.Resources().WritePacket(&packet.BlockActorData{
			Position: s.console.Center(),
			NBTData:  nbtMap,
		})
		if err != nil {
			return fmt.Errorf("Make: %v", err)
		}
	}

	// 确定告示牌各面颜色
	frontColor, _ := utils.DecodeVarRGBA(s.data.NBT.FrontText.SignTextColor)
	backColor, _ := utils.DecodeVarRGBA(s.data.NBT.BackText.SignTextColor)

	// 告示牌染色
	for index, color := range [2][3]uint8{frontColor, backColor} {
		if color == [3]uint8{0, 0, 0} {
			continue
		}

		dyeName, ok := mapping.RGBToDyeItemName[color]
		if !ok {
			panic("Make: Should never happened")
		}
		err = s.replaceitem(dyeName, true)
		if err != nil {
			return fmt.Errorf("Make: %v", err)
		}

		clickPos := s.console.Center()
		if index != 0 {
			clickPos[0] += 1
		}
		err = api.BotClick().ClickBlockWitchPosition(
			blockAction,
			mgl32.Vec3{float32(clickPos[0]), float32(clickPos[1]), float32(clickPos[2])},
		)
		if err != nil {
			return fmt.Errorf("Make: %v", err)
		}
	}

	// 获取发光墨囊，如果需要的话
	if s.data.NBT.FrontText.IgnoreLighting == 1 || s.data.NBT.BackText.IgnoreLighting == 1 {
		err = s.replaceitem("minecraft:glow_ink_sac", true)
		if err != nil {
			return fmt.Errorf("Make: %v", err)
		}
	}

	// 告示牌发光效果
	for index, value := range [2]nbt_parser_block.SignText{s.data.NBT.FrontText, s.data.NBT.BackText} {
		if value.IgnoreLighting == 0 {
			continue
		}
		clickPos := s.console.Center()
		if index != 0 {
			clickPos[0] += 1
		}
		err = api.BotClick().ClickBlockWitchPosition(
			blockAction,
			mgl32.Vec3{float32(clickPos[0]), float32(clickPos[1]), float32(clickPos[2])},
		)
		if err != nil {
			return fmt.Errorf("Make: %v", err)
		}
	}

	// 告示牌涂蜡
	if s.data.NBT.IsWaxed == 1 {
		err = s.replaceitem("minecraft:honeycomb", true)
		if err != nil {
			return fmt.Errorf("Make: %v", err)
		}
		err = api.BotClick().ClickBlock(blockAction)
		if err != nil {
			return fmt.Errorf("Make: %v", err)
		}
	}

	// 覆写告示牌方块状态
	err = api.SetBlock().SetBlock(s.console.Center(), s.data.BlockName(), s.data.BlockStatesString())
	if err != nil {
		return fmt.Errorf("Make: %v", err)
	}
	s.console.UseHelperBlock(nbt_console.RequesterUser, nbt_console.ConsoleIndexCenterBlock, block_helper.ComplexBlock{
		KnownStates: true,
		Name:        s.data.BlockName(),
		States:      s.data.BlockStates(),
	})

	return nil
}
