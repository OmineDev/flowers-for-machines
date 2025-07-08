package nbt_console

import (
	"fmt"

	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol"
	"github.com/OmineDev/flowers-for-machines/game_control/game_interface"
	"github.com/OmineDev/flowers-for-machines/game_control/resources_control"

	"github.com/go-gl/mathgl/mgl32"
)

// API 返回操作台的底层游戏交互接口
func (c *Console) API() *game_interface.GameInterface {
	return c.api
}

// Dimension 返回操作台所在的维度 ID
func (c Console) Dimension() uint8 {
	return c.dimension
}

// Center 返回操作台中心处的方块坐标
func (c Console) Center() protocol.BlockPos {
	return c.center
}

// Center 返回机器人当前的坐标。不保证是最准确的，
// 因为可能机器人可能会由于受到方块挤压而发生了一定的偏移
func (c Console) Position() mgl32.Vec3 {
	return mgl32.Vec3{
		float32(c.position[0]) + 0.5,
		float32(c.position[1]) + 1.5,
		float32(c.position[2]) + 0.5,
	}
}

// UpdatePosition 设置机器人当前所处的坐标
func (c *Console) UpdatePosition(blockPos protocol.BlockPos) {
	c.position = blockPos
}

// HotbarSlotID 返回机器人当前所手持物品的快捷栏槽位索引
func (c Console) HotbarSlotID() resources_control.SlotID {
	return c.currentHotBar
}

// UpdateHotbarSlotID 设置机器人当前所手持物品栏的槽位索引
func (c *Console) UpdateHotbarSlotID(slotID resources_control.SlotID) {
	c.currentHotBar = slotID
}

// ChangeAndUpdateHotbarSlotID 将机器人的手持物品栏
// 切换为 slotID 并同时将此更改广播到操作台的底层实现
func (c *Console) ChangeAndUpdateHotbarSlotID(slotID resources_control.SlotID) error {
	err := c.api.BotClick().ChangeSelectedHotbarSlot(slotID)
	if err != nil {
		return fmt.Errorf("ChangeAndUpdateHotbarSlotID: %v", err)
	}
	c.currentHotBar = slotID
	return nil
}

// ChangeConsolePosition 切换操作台的位置。
//
// dimensionID 是新位置所在维度的 ID，
// center 是新位置的方块坐标。
//
// 如果返回了错误，则在下次成功调用此函数前，
// 操作台都不应该被使用，否则其他操作的结果
// 将会是未定义的
func (c *Console) ChangeConsolePosition(dimensionID uint8, center protocol.BlockPos) error {
	err := c.initConsole(dimensionID, center)
	if err != nil {
		return fmt.Errorf("ChangeConsolePosition: %v", err)
	}
	return nil
}
