package nbt_console

import (
	"fmt"
	"math"
	"time"

	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol"
	"github.com/OmineDev/flowers-for-machines/game_control/game_interface"
	"github.com/OmineDev/flowers-for-machines/game_control/resources_control"
	"github.com/OmineDev/flowers-for-machines/nbt_assigner/block_helper"
	"github.com/OmineDev/flowers-for-machines/utils"
)

// Console 是机器人导入 NBT 方块所使用的操作台。
// 它目前被定义为以操作台中心为中心的 11*5*11
// 的全空气区域
type Console struct {
	// api 是与租赁服进行交互的若干接口
	api *game_interface.GameInterface

	// dimension 是操作台所在的维度 ID
	dimension uint8
	// center 是操作台的中心位置
	center protocol.BlockPos
	// position 是机器人目前所在的方块位置
	position protocol.BlockPos
	// currentHotBar 是机器人当前的手持物品栏
	currentHotBar resources_control.SlotID

	// airSlotInInventory 记录机器人槽位中
	// 哪些地方的槽位已经存在物品，而哪些没有。
	// 为真指示这是一个物品，否则它是一个空气
	airSlotInInventory [36]bool

	// helperBlocks 是操作台中心及其
	// 不远处等分布均匀的 8 个帮助方块。
	// 通过记录这 9 个方块的实际情况，
	// 有助于减少部分操作的实际耗时
	helperBlocks [9]*block_helper.BlockHelper
	// nearBlocks 是操作台中心方块及另
	// 外 8 个帮助方块相邻的方块。
	//
	// 如果认为操作台中心方块和另外 8 个
	// 帮助方块是 master 方块，那么对于
	// 第二层数组，可以通过 nearBlockMapping
	// 确定它们各自相邻其 master 方块的位置变化。
	//
	// 另外，nearBlockMappingInv 是
	// nearBlockMapping 的逆映射
	nearBlocks [9][6]*block_helper.BlockHelper

	// inventoryUseCallback 存放了一系列回调函数，
	// 用于其他实现在修改机器人背包物品栏数据时通知
	// 这件事给其他可能的使用者
	inventoryUseCallback []func(requester string, slotID resources_control.SlotID)
	// blocksUseCallback 存放了一系列回调函数，
	// 用于其他实现在修改操作台上帮助方块(或中心
	// 方块)时通知这件事给其他可能的使用者
	blocksUseCallback []func(requester string, index int)
}

// NewConsole 根据交互接口 api 和操作台中心 center
// 创建并返回一个新的操作台实例。
//
// NewConsole 会将机器人切换为创造模式，清空物品栏，
// 然后重置手持物品栏为 5 并将机器人传送至操作台的中
// 心方块处。在传送完成后，NewConsole 将试图初始化操
// 作台的地板方块。
//
// NewConsole 的调用者有责任确保操作台位于 dimensionID
// 所指示的维度上，并且以操作台中心方块为中心处的 11*5*11
// 的区域全为空气且没有任何实体
func NewConsole(api *game_interface.GameInterface, dimensionID uint8, center protocol.BlockPos) (result *Console, err error) {
	c := &Console{api: api}

	err = c.initConsole(dimensionID, center)
	if err != nil {
		return nil, fmt.Errorf("NewConsole: %v", err)
	}

	return c, nil
}

// initConsole 初始化操作台。
// 它是一个内部实现细节，不应被其他人所使用
func (c *Console) initConsole(dimensionID uint8, center protocol.BlockPos) error {
	api := c.api

	// Check center
	deltaX := int(math.Abs(float64(center[0])))
	deltaY := int(math.Abs(float64(center[1])))
	deltaZ := int(math.Abs(float64(center[2])))
	if deltaX*deltaX+deltaY*deltaY+deltaZ*deltaZ < 900 {
		return fmt.Errorf("initConsole: The bot can not appear around position (0,0,0) and it must be at least 30 blocks away from here")
	}

	// Reflush console info
	*c = Console{
		api:                  c.api,
		dimension:            dimensionID,
		center:               center,
		position:             center,
		currentHotBar:        DefaultHotbarSlot,
		airSlotInInventory:   [36]bool{},
		helperBlocks:         [9]*block_helper.BlockHelper{},
		nearBlocks:           [9][6]*block_helper.BlockHelper{},
		inventoryUseCallback: c.inventoryUseCallback,
		blocksUseCallback:    c.blocksUseCallback,
	}

	// Change gamemode and hotbar slot
	err := api.Commands().SendSettingsCommand("gamemode 1", true)
	if err != nil {
		return fmt.Errorf("initConsole: %v", err)
	}
	err = api.Commands().SendSettingsCommand("clear", true)
	if err != nil {
		return fmt.Errorf("initConsole: %v", err)
	}
	err = api.BotClick().ChangeSelectedHotbarSlot(c.currentHotBar)
	if err != nil {
		return fmt.Errorf("initConsole: %v", err)
	}

	// Teleport to target area
	err = api.Commands().SendSettingsCommand(
		fmt.Sprintf("execute in %s run tp %d %d %d", utils.DimensionNameByID(dimensionID), c.center[0], c.center[1], c.center[2]),
		true,
	)
	if err != nil {
		return fmt.Errorf("initConsole: %v", err)
	}

	timer := time.NewTimer(DefaultTimeoutInitConsole)
	defer timer.Stop()

	// Waiting bot to go to the target area
	for {
		select {
		case <-timer.C:
			return fmt.Errorf("initConsole: Can not teleport to the target area (timeout)")
		default:
		}

		uniqueID, err := api.StructureBackup().BackupOffset(
			protocol.BlockPos{c.center[0] - 5, c.center[1], c.center[2] - 5},
			protocol.BlockPos{10, 0, 10},
		)
		if err != nil {
			continue
		}

		resp, err := api.Commands().SendWSCommandWithResp(
			fmt.Sprintf(
				`structure delete "%v"`,
				utils.MakeUUIDSafeString(uniqueID),
			),
		)
		if err != nil {
			return fmt.Errorf("initConsole: %v", err)
		}
		if resp.SuccessCount > 0 {
			break
		}
	}

	// Init console blocks
	{
		// Clean area frist
		_, err = api.Commands().SendWSCommandWithResp(
			fmt.Sprintf(
				"execute as @s at @s positioned %d ~ ~ positioned ~ %d ~ positioned ~ ~ %d run fill ~-5 ~-2 ~-5 ~5 ~2 ~5 air",
				c.center[0], c.center[1], c.center[2],
			),
		)
		if err != nil {
			return fmt.Errorf("initConsole: %v", err)
		}
		// Teleport again
		err = api.Commands().SendSettingsCommand(
			fmt.Sprintf("execute in overworld run tp %d %d %d", c.center[0], c.center[1], c.center[2]),
			true,
		)
		if err != nil {
			return fmt.Errorf("initConsole: %v", err)
		}
		// Filling floor blocks
		_, err = api.Commands().SendWSCommandWithResp(
			fmt.Sprintf(
				"execute as @s at @s positioned %d ~ ~ positioned ~ %d ~ positioned ~ ~ %d run fill ~-5 ~-1 ~-5 ~5 ~-1 ~5 %s",
				c.center[0], c.center[1], c.center[2], BaseBackground,
			),
		)
		if err != nil {
			return fmt.Errorf("initConsole: %v", err)
		}
	}

	// Sync console block info
	for index := range 9 {
		var airBlock block_helper.BlockHelper = block_helper.Air{}
		c.helperBlocks[index] = &airBlock
	}
	for index := range 9 {
		for idx := range 6 {
			var airBlock block_helper.BlockHelper = block_helper.Air{}
			c.nearBlocks[index][idx] = &airBlock
		}
	}
	for index := range 9 {
		var floorBlock block_helper.BlockHelper = block_helper.NearBlock{
			Name: BaseBackground,
		}
		*c.nearBlocks[index][nearBlockMappingInv[[3]int32{0, -1, 0}]] = floorBlock
	}

	return nil
}
