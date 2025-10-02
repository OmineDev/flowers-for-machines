package nbt_console

import (
	"fmt"
	"math/rand"

	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol"
	"github.com/OmineDev/flowers-for-machines/game_control/game_interface"
	"github.com/OmineDev/flowers-for-machines/nbt_assigner/block_helper"
)

// FindAir 从操作台的帮助方块中寻找一个空气方块。
// includeCenter 指示要查找的方块是否也包括操作台
// 中心处的方块。
//
// 返回的 index 可用于 BlockByIndex，
// 而返回的 offset 可用于 BlockByOffset。
//
// 如果返回的 block 不为空，则说明找到，
// 否则没有找到。找到的方块可以通过修改
// 其指向的值从而将它变成其他方块
func (c Console) FindAir(includeCenter bool) (index int, offset protocol.BlockPos, block *block_helper.BlockHelper) {
	for index, value := range c.helperBlocks {
		if !includeCenter && index == 0 {
			continue
		}
		if _, ok := (*value).(block_helper.Air); ok {
			return index, helperBlockMapping[index], value
		}
	}
	return 0, protocol.BlockPos{}, nil
}

// FindAir 从操作台的帮助方块中寻找一个铁砧方块。
// includeCenter 指示要查找的方块是否也包括操作
// 台中心处的方块。
//
// 返回的 index 可用于 BlockByIndex，
// 而返回的 offset 可用于 BlockByOffset。
//
// 如果返回的 block 不为空，则说明找到，
// 否则没有找到。找到的方块可以通过修改
// 其指向的值从而将它变成其他方块
func (c Console) FindAnvil(includeCenter bool) (index int, offset protocol.BlockPos, block *block_helper.BlockHelper) {
	for index, value := range c.helperBlocks {
		if !includeCenter && index == 0 {
			continue
		}
		if _, ok := (*value).(block_helper.AnvilBlockHelper); ok {
			return index, helperBlockMapping[index], value
		}
	}
	return 0, protocol.BlockPos{}, nil
}

// FindLoom 从操作台的帮助方块中寻找一个织布机方块。
// includeCenter 指示要查找的方块是否也包括操作台
// 中心处的方块。
//
// 返回的 index 可用于 BlockByIndex，
// 而返回的 offset 可用于 BlockByOffset。
//
// 如果返回的 block 不为空，则说明找到，
// 否则没有找到。找到的方块可以通过修改
// 其指向的值从而将它变成其他方块
func (c Console) FindLoom(includeCenter bool) (index int, offset protocol.BlockPos, block *block_helper.BlockHelper) {
	for index, value := range c.helperBlocks {
		if !includeCenter && index == 0 {
			continue
		}
		if _, ok := (*value).(block_helper.LoomBlockHelper); ok {
			return index, helperBlockMapping[index], value
		}
	}
	return 0, protocol.BlockPos{}, nil
}

// FindNonAnvilAndNonLoom 从操作台的帮助方块
// 中寻找一个既不是铁砧，也不是织布机的方块。
//
// 这意味目标方块将可以是空气、容器或其他方块。
//
// includeCenter 指示要查找的方块是否也包括操
// 作台中心处的方块。
//
// 返回的 index 可用于 BlockByIndex，
// 而返回的 offset 可用于 BlockByOffset。
//
// FindNonAnvilAndNonLoom 在设计上认为一定
// 可以找到目标的方块。
//
// 找到的方块可以通过修改其指向的值从而将它变成其他方块
func (c Console) FindNonAnvilAndNonLoom(includeCenter bool) (index int, offset protocol.BlockPos, block *block_helper.BlockHelper) {
	idxs := make([]int, 0)

	for index, value := range c.helperBlocks {
		if !includeCenter && index == 0 {
			continue
		}
		switch (*value).(type) {
		case block_helper.AnvilBlockHelper, block_helper.LoomBlockHelper:
		default:
			idxs = append(idxs, index)
		}
	}

	if len(idxs) == 0 {
		panic("FindNonAnvilAndNonLoom: Should never happened")
	}

	randIndex := rand.Intn(len(idxs))
	index = idxs[randIndex]
	offset = helperBlockMapping[index]
	block = c.helperBlocks[index]

	return
}

// FindSpaceToPlaceNewBlock 尝试从操作台
// 找到一个位置以便于使用者放置一个新的方块。
// 它可以是帮助方块、容器，或者其他方块。
//
// includeCenter 指示要查找的方块是否也包括操
// 作台中心处的方块。
//
// 返回的 index 可用于 BlockByIndex，
// 而返回的 offset 可用于 BlockByOffset。
//
// FindSpaceToPlaceNewBlock 在设计上认为
// 一定可以找到目标的方块。
//
// 找到的方块可以通过修改其指向的值从而将它变成其他方块
func (c Console) FindSpaceToPlaceNewBlock(includeCenter bool) (
	index int,
	offset protocol.BlockPos,
	block *block_helper.BlockHelper,
) {
	index, offset, block = c.FindAir(includeCenter)
	if block != nil {
		return
	}

	index, offset, block = c.FindNonAnvilAndNonLoom(includeCenter)
	if block == nil {
		panic("FindSpaceToPlaceNewBlock: Should never happened")
	}

	return
}

// FindMutipleSpaceToPlaceNewBlock 从操作台
// 找到所有可供放置新方块的位置。
//
// includeCenter 指示要查找的方块是否也包括操
// 作台中心处的方块。
//
// FindMutipleSpaceToPlaceNewBlock 在设计上
// 认为 blockIndexs 的长度必定大于 0。
//
// 找到的方块可以通过修改其指向的值从而将它变成其他方块
func (c Console) FindMutipleSpaceToPlaceNewBlock(includeCenter bool) (blockIndexs []int) {
	for index, value := range c.helperBlocks {
		if !includeCenter && index == 0 {
			continue
		}
		switch (*value).(type) {
		case block_helper.AnvilBlockHelper, block_helper.LoomBlockHelper:
		default:
			blockIndexs = append(blockIndexs, index)
		}
	}
	return
}

// FindOrGenerateNewAnvil 寻找操作台的 8 个帮助方块中
// 是否有一个是铁砧。如果没有，则生成一个铁砧及其承重方块。
// index 指示找到或生成的铁砧在操作台上的索引
func (c *Console) FindOrGenerateNewAnvil() (index int, err error) {
	var block *block_helper.BlockHelper
	var needFloorBlock bool

	index, _, block = c.FindAnvil(false)
	if block != nil {
		return
	}

	index, _, block = c.FindSpaceToPlaceNewBlock(false)
	if block == nil {
		panic("FindOrGenerateNewAnvil: Should never happened")
	}

	nearBlock := c.NearBlockByIndex(index, protocol.BlockPos{0, -1, 0})
	switch (*nearBlock).(type) {
	case block_helper.Air, block_helper.ComplexBlock:
		needFloorBlock = true
	}

	states, err := c.api.SetBlock().SetAnvil(c.BlockPosByIndex(index), needFloorBlock)
	if err != nil {
		return 0, fmt.Errorf("FindOrGenerateNewAnvil: %v", err)
	}

	anvil := block_helper.AnvilBlockHelper{States: states}
	c.UseHelperBlock(RequesterSystemCall, index, anvil)
	if needFloorBlock {
		var floorBlock block_helper.BlockHelper = block_helper.NearBlock{
			Name: game_interface.BaseAnvil,
		}
		*c.NearBlockByIndex(index, protocol.BlockPos{0, -1, 0}) = floorBlock
	}

	return index, nil
}

// FindOrGenerateNewLoom 寻找操作台的 8 个帮助方块中
// 是否有一个是织布机。如果没有，则生成一个新的织布机。
// index 指示找到或生成的织布机在操作台上的索引
func (c *Console) FindOrGenerateNewLoom() (index int, err error) {
	var block *block_helper.BlockHelper

	index, _, block = c.FindLoom(false)
	if block != nil {
		return
	}

	index, _, block = c.FindSpaceToPlaceNewBlock(false)
	if block == nil {
		panic("FindOrGenerateNewLoom: Should never happened")
	}

	loom := block_helper.LoomBlockHelper{}
	err = c.api.SetBlock().SetBlock(
		c.BlockPosByIndex(index),
		loom.BlockName(),
		loom.BlockStatesString(),
	)
	if err != nil {
		return 0, fmt.Errorf("FindOrGenerateNewLoom: %v", err)
	}
	c.UseHelperBlock(RequesterSystemCall, index, loom)

	return index, nil
}
