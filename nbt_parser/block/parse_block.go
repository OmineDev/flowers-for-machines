package nbt_parser_block

import (
	"fmt"
	"strings"

	"github.com/OmineDev/flowers-for-machines/mapping"
	nbt_parser_interface "github.com/OmineDev/flowers-for-machines/nbt_parser/interface"

	"github.com/df-mc/worldupgrader/blockupgrader"
)

// ParseBlock 从方块实体数据 blockNBT 解析一个方块实体。
// blockName 和 blockStates 分别指示这个方块实体的名称和方块状态。
//
// nameChecker 是一个可选的函数，用于检查 name 所指示的物品名称是
// 否可通过指令获取。如果不能，则 nameChecker 返回假。
//
// nameChecker 对于大多数方块的解析可能没有帮助，但它可以帮助验证
// 容器内的物品是否是可以通过指令获取的物品。
//
// 另外，如果没有这样的 nameChecker 函数，则可以将其简单的置为 nil
func ParseBlock(
	nameChecker func(name string) bool,
	blockName string,
	blockStates map[string]any,
	blockNBT map[string]any,
) (block nbt_parser_interface.Block, err error) {
	name := strings.ToLower(blockName)
	if !strings.HasPrefix(name, "minecraft:") {
		name = "minecraft:" + name
	}

	newBlock := blockupgrader.Upgrade(blockupgrader.BlockState{
		Name:       name,
		Properties: blockStates,
	})
	defaultBlock := DefaultBlock{
		Name:        newBlock.Name,
		States:      newBlock.Properties,
		NameChecker: nameChecker,
	}

	blockType, ok := mapping.SupportBlocksPool[newBlock.Name]
	if !ok {
		return &defaultBlock, nil
	}
	defaultBlock.States = nbt_parser_interface.DeepCopyAndFixStates(
		blockType,
		defaultBlock.BlockName(),
		defaultBlock.BlockStates(),
	)

	switch blockType {
	case mapping.SupportNBTBlockTypeCommandBlock:
		block = &CommandBlock{DefaultBlock: defaultBlock}
	case mapping.SupportNBTBlockTypeContainer:
		block = &Container{DefaultBlock: defaultBlock}
	case mapping.SupportNBTBlockTypeSign:
		block = &Sign{DefaultBlock: defaultBlock}
	case mapping.SupportNBTBlockTypeFrame:
		block = &Frame{DefaultBlock: defaultBlock}
	case mapping.SupportNBTBlockTypeStructureBlock:
		block = &StructureBlock{DefaultBlock: defaultBlock}
	case mapping.SupportNBTBlockTypeBanner:
		block = &Banner{DefaultBlock: defaultBlock}
	case mapping.SupportNBTBlockTypeLectern:
		block = &Lectern{DefaultBlock: defaultBlock}
	case mapping.SupportNBTBlockTypeJukeBox:
		block = &JukeBox{DefaultBlock: defaultBlock}
	case mapping.SupportNBTBlockTypeBrewingStand:
		block = &BrewingStand{DefaultBlock: defaultBlock}
	case mapping.SupportNBTBlockTypeCrafter:
		block = &Crafter{DefaultBlock: defaultBlock}
	default:
		panic("ParseNBTBlock: Should never happened")
	}

	err = block.Parse(blockNBT)
	if err != nil {
		return nil, fmt.Errorf("ParseNBTBlock: %v", err)
	}
	return block, nil
}

func init() {
	nbt_parser_interface.ParseBlock = ParseBlock
}
