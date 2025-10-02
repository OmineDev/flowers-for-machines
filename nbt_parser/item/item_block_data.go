package nbt_parser_item

import (
	"fmt"

	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol"
	"github.com/OmineDev/flowers-for-machines/mapping"
	nbt_parser_interface "github.com/OmineDev/flowers-for-machines/nbt_parser/interface"
	"github.com/OmineDev/flowers-for-machines/utils"

	"github.com/TriM-Organization/bedrock-world-operator/block"
	"github.com/df-mc/worldupgrader/blockupgrader"
)

// ItemBlockData 指示该物品是一个方块，
// 或该物品可以作为一个方块进行放置
type ItemBlockData struct {
	// Name 是这个方块的名称
	Name string
	// States 是这个方块的方块状态
	States map[string]any
	// SubBlock 是这个方块的附加数据。
	//
	// 如果这个方块是已被支持的 NBT 方块，
	// 且已被确认是需要进行特殊处理的子方块，
	// 则 SubBlock 非空。
	//
	// 这意味着，纵使这个物品是一个带有 NBT
	// 数据的容器，但如果被判定为不需要特殊
	// 处理，则 SubBlock 仍然解析为空
	SubBlock nbt_parser_interface.Block
}

// HaveSubBlockData 验证 tag 是否指向有效的子方块数据荷载。
//
// 一个特殊情况是，可能在首次解析时，目标子方块的子方块被证明
// 为不需要进行特殊处理，于是它的子方块数据被剔除。
//
// 但由于这个子方块对应的物品形式含有物品名称，于是在验证导入
// 完整性时，由于 len(tag) > 0 而尝试解析子方块，并得到一个
// 需要特殊处理的子方块 (而原始情况却是不需要进行处理)。
//
// 这是因为，此时的 tag 虽然不包含子方块的数据，但包含其他 NBT
// 字段，如这个物品的物品组件或名称数据等。于是，在尝试解析它为
// 子方块时，子方块会得到全零值的字段，但这不代表它是不需要被特
// 殊处理的情况——于是再次解析所得的产物跟原产物不等价，自环发生。
//
// 所以，HaveSubBlockData 深拷贝 tag 并移除所有可能因导入而产生
// 的额外的 NBT 字段，然后再检查 tag 的长度是否大于 0 并返回结果
func HaveSubBlockData(tag map[string]any) bool {
	newTag := utils.DeepCopyNBT(tag)
	delete(newTag, "Damage")
	delete(newTag, "RepairCost")
	delete(newTag, "display")
	delete(newTag, "ench")
	delete(newTag, "minecraft:keep_on_death")
	delete(newTag, "minecraft:item_lock")
	return len(newTag) > 0
}

// ParseItemBlock ..
func ParseItemBlock(
	nameChecker func(name string) bool,
	itemName string,
	nbtMap map[string]any,
) (result ItemBlockData, err error) {
	var blockMap map[string]any
	var haveBlock bool

	// Step 1: Get data from nbtMap
	blockMap, haveBlock = nbtMap["Block"].(map[string]any)
	tag, _ := nbtMap["tag"].(map[string]any)

	// Step 2: Get block type of this sub block
	blockName, ok := mapping.ItemNameToBlockName[itemName]
	if !ok {
		return
	}
	blockType, ok := mapping.SupportBlocksPool[blockName]
	if !ok {
		panic("ParseItemBlock: Should never happened")
	}
	if !mapping.SubBlocksPool[blockType] {
		return
	}

	// Step 3: If exist, then we use the states directly
	if haveBlock {
		name, _ := blockMap["name"].(string)
		states, _ := blockMap["states"].(map[string]any)
		version, _ := blockMap["version"].(int32)

		newBlock := blockupgrader.Upgrade(blockupgrader.BlockState{
			Name:       name,
			Properties: states,
			Version:    version,
		})

		result.Name = newBlock.Name
		result.States = newBlock.Properties
	} else {
		rid, found := block.StateToRuntimeID(blockName, map[string]any{})
		if !found {
			panic("ParseItemBlock: Should never happened")
		}

		name, states, found := block.RuntimeIDToState(rid)
		if !found {
			panic("ParseItemBlock: Should never happened")
		}

		result.Name = name
		result.States = states
	}

	// Step 4: Fix block states, and check have sub block data
	result.States = nbt_parser_interface.DeepCopyAndFixStates(blockType, result.Name, result.States)
	if !HaveSubBlockData(tag) {
		return
	}

	// Step 5: Parse sub block data
	subBlock, err := nbt_parser_interface.ParseBlock(nameChecker, result.Name, result.States, utils.DeepCopyNBT(tag))
	if err != nil {
		return ItemBlockData{}, fmt.Errorf("ParseItemBlock: %v", err)
	}
	if subBlock.NeedSpecialHandle() {
		result.SubBlock = subBlock
	}

	// Step 6: Return
	return
}

// ParseItemBlockNetwork ..
func ParseItemBlockNetwork(itemName string, item protocol.ItemStack) (result ItemBlockData, err error) {
	// Step 1: Get block type of this sub block
	blockName, ok := mapping.ItemNameToBlockName[itemName]
	if !ok {
		return
	}
	blockType, ok := mapping.SupportBlocksPool[blockName]
	if !ok {
		panic("ParseItemBlockNetwork: Should never happened")
	}
	if !mapping.SubBlocksPool[blockType] {
		return
	}

	// Step 2: If exist, then we use the states directly
	if item.BlockRuntimeID != 0 {
		name, states, found := block.RuntimeIDToState(uint32(item.BlockRuntimeID))
		if !found {
			panic("ParseItemBlockNetwork: Should never happened")
		}
		result.Name = name
		result.States = states
	} else {
		rid, found := block.StateToRuntimeID(blockName, map[string]any{})
		if !found {
			panic("ParseItemBlockNetwork: Should never happened")
		}

		name, states, found := block.RuntimeIDToState(rid)
		if !found {
			panic("ParseItemBlockNetwork: Should never happened")
		}

		result.Name = name
		result.States = states
	}

	// Step 3: Fix block states, and check have sub block data
	result.States = nbt_parser_interface.DeepCopyAndFixStates(blockType, result.Name, result.States)
	if !HaveSubBlockData(item.NBTData) {
		return
	}

	// Step 4: Parse sub block data
	subBlock, err := nbt_parser_interface.ParseBlock(nil, result.Name, result.States, utils.DeepCopyNBT(item.NBTData))
	if err != nil {
		return ItemBlockData{}, fmt.Errorf("ParseItemBlockNetwork: %v", err)
	}
	if subBlock.NeedSpecialHandle() {
		result.SubBlock = subBlock
	}

	// Step 5: Return
	return
}
