package game_interface

import (
	"fmt"
	"sync"
	"time"

	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol"
	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol/packet"
	"github.com/OmineDev/flowers-for-machines/game_control/resources_control"

	"github.com/TriM-Organization/bedrock-world-operator/block"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	// 描述 Pick Block 请求的最长截止时间。
	// 这与 packet.BlockPickRequest 相关。
	// 当超过此时间后，将视为该请求未被接受
	DefaultTimeoutBlockPick = time.Second
	// 描述 Pick Block 失败后要重试的最大次数
	MaxRetryBlockPick = 3
	// 用作放置方块时的依赖性方块。
	//
	// 部分方块需要客户端以点击方块的形式来放置，
	// 例如告示牌和不同朝向的潜影盒。
	// 这里则选择了绿宝石块作为被点击的方块。
	//
	// SuperScript 最喜欢绿宝石块了！
	BasePlaceBlock string = "minecraft:emerald_block"
)

// UseItemOnBlocks 是机器人在使用手
// 持物品对方块进行操作时的通用结构体
type UseItemOnBlocks struct {
	HotbarSlotID resources_control.SlotID // 指代机器人当前已选择的快捷栏编号
	BotPos       mgl32.Vec3               // 指代机器人操作该方块时的位置
	BlockPos     protocol.BlockPos        // 指代被操作方块的位置
	BlockName    string                   // 指代被操作方块的名称
	BlockStates  map[string]any           // 指代被操作方块的方块状态
}

// BotClick 是基于 ResourcesWrapper
// 和 Commands 实现的已简化的点击实现。
//
// 由于点击操作与机器人手持物品强相关，
// 本处也集成了切换手持物品的实现。
//
// 另外，考虑到 Pick Block 操作的语义
// 也与 点击方块 有关，因此其也被集成在
// 此，尽管它使用了完全不同的数据包
type BotClick struct {
	r *ResourcesWrapper
	c *Commands
	s *SetBlock
}

// NewBotClick 基于 wrapper、commands 和 setblock 创建并返回一个新的 BotClick
func NewBotClick(wrapper *ResourcesWrapper, commands *Commands, setblock *SetBlock) *BotClick {
	return &BotClick{r: wrapper, c: commands, s: setblock}
}

// 切换客户端的手持物品栏为 hotBarSlotID 。
// 若提供的 hotBarSlotID 大于 8 ，则会重定向为 0
func (b *BotClick) ChangeSelectedHotbarSlot(hotbarSlotID resources_control.SlotID) error {
	if hotbarSlotID > 8 {
		hotbarSlotID = 0
	}

	err := b.r.WritePacket(&packet.PlayerHotBar{
		SelectedHotBarSlot: uint32(hotbarSlotID),
		WindowID:           0,
		SelectHotBarSlot:   true,
	})
	if err != nil {
		return fmt.Errorf("ChangeSelectedHotbarSlot: %v", err)
	}

	return nil
}

// clickBlock ..
func (b *BotClick) clickBlock(
	request UseItemOnBlocks,
	blockFace int32,
	position mgl32.Vec3,
) error {
	// Step 1: 取得被点击方块的方块运行时 ID
	blockRuntimeID, found := block.StateToRuntimeID(request.BlockName, request.BlockStates)
	if !found {
		return fmt.Errorf(
			"clickBlock: Can't found the block runtime ID of block %#v (block states = %#v)",
			request.BlockName, request.BlockStates,
		)
	}

	// Step 2: 取得当前手持物品的信息
	item, inventoryExisted := b.r.Inventories().GetItemStack(0, request.HotbarSlotID)
	if !inventoryExisted {
		return fmt.Errorf("clickBlock: Should never happened")
	}

	// Step 3: 发送点击操作
	err := b.r.WritePacket(&packet.InventoryTransaction{
		LegacyRequestID:    0,
		LegacySetItemSlots: []protocol.LegacySetItemSlot(nil),
		Actions:            []protocol.InventoryAction{},
		TransactionData: &protocol.UseItemTransactionData{
			LegacyRequestID:    0,
			LegacySetItemSlots: nil,
			Actions:            nil,
			ActionType:         protocol.UseItemActionClickBlock,
			BlockPosition:      request.BlockPos,
			BlockFace:          blockFace,
			HotBarSlot:         int32(request.HotbarSlotID),
			HeldItem:           *item,
			Position:           position,
			BlockRuntimeID:     blockRuntimeID,
		},
	})
	if err != nil {
		return fmt.Errorf("clickBlock: %v", err)
	}

	// Step 4: 额外操作 (自 v1.20.50 以来的必须更改)
	//
	// !!! NOTE - MUST SEND AUTH INPUT TWICE !!!
	// await changes and send auth
	// input to submit changes
	for range 2 {
		err = b.c.AwaitChangesGeneral()
		if err != nil {
			return fmt.Errorf("clickBlock: %v", err)
		}
		err = b.r.WritePacket(&packet.PlayerAuthInput{
			InputData: packet.InputFlagStartFlying,
			Position:  request.BotPos,
		})
		if err != nil {
			return fmt.Errorf("clickBlock: %v", err)
		}
	}

	return nil
}

/*
让客户端点击 request 所指代的方块，
并且指定当次交互时玩家的位置为 position 。

position 不一定需要是真实的，
客户端可以上传欺骗性的数据，
服务器不会对它们进行验证。

该函数在通常情况下被用于十分精细的操作，
例如为告示牌的特定面附加发光效果。

此函数不会自动切换物品栏，但会等待租赁服响应更改
*/
func (b *BotClick) ClickBlockWitchPosition(
	request UseItemOnBlocks,
	position mgl32.Vec3,
) error {
	err := b.clickBlock(request, 0, position)
	if err != nil {
		return fmt.Errorf("ClickBlockWitchPosition: %v", err)
	}
	return nil
}

/*
让客户端点击 request 所指代的方块。

你可以对容器使用这样的操作，这会使得容器被打开。

你亦可以对物品展示框使用这样的操作，
这会使得物品被放入或令展示框内的物品旋转。

此函数不会自动切换物品栏，但会等待租赁服响应更改
*/
func (b *BotClick) ClickBlock(request UseItemOnBlocks) error {
	err := b.clickBlock(request, 0, mgl32.Vec3{})
	if err != nil {
		return fmt.Errorf("ClickBlock: %v", err)
	}
	return nil
}

// 使用快捷栏 hotbarSlotID 进行一次空点击操作。
// realPosition 指示机器人在操作时的实际位置。
// 此函数不会自动切换物品栏，但会等待租赁服响应更改
func (b *BotClick) ClickAir(hotbarSlot resources_control.SlotID, realPosition mgl32.Vec3) error {
	// Step 1: 获取手持物品栏物品数据信息
	item, inventoryExisted := b.r.Inventories().GetItemStack(0, hotbarSlot)
	if !inventoryExisted {
		return fmt.Errorf("ClickAir: Should never happened")
	}

	// Step 2: 发送点击数据包
	err := b.r.WritePacket(
		&packet.InventoryTransaction{
			TransactionData: &protocol.UseItemTransactionData{
				ActionType: protocol.UseItemActionClickAir,
				HotBarSlot: int32(hotbarSlot),
				HeldItem:   *item,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("ClickAir: %v", err)
	}

	// Step 3: 额外操作 (自 v1.20.50 以来的必须更改)
	err = b.c.AwaitChangesGeneral()
	if err != nil {
		return fmt.Errorf("ClickAir: %v", err)
	}
	err = b.r.WritePacket(&packet.PlayerAuthInput{
		InputData: packet.InputFlagStartFlying,
		Position:  realPosition,
	})
	if err != nil {
		return fmt.Errorf("ClickAir: %v", err)
	}

	return nil
}

/*
PlaceBlock 使客户端创建一个新方块。

request 指代实际被点击的方块，但这并不代表新方块被创建的位置。
我们通过点击 request 处的方块，并指定点击的面为 blockFace ，
然后租赁服根据这些信息，在另外相应的位置创建这些新的方块。

此函数不会自动切换物品栏，但会等待租赁服响应更改
*/
func (b *BotClick) PlaceBlock(
	request UseItemOnBlocks,
	blockFace int32,
) error {
	err := b.clickBlock(request, blockFace, mgl32.Vec3{})
	if err != nil {
		return fmt.Errorf("PlaceBlock: %v", err)
	}
	return nil
}

// PlaceBlockHighLevel 是对 PlaceBlock 的进一步封装。
//
// 它通过方块点击的方式，直接在 blockPos 处创建朝向为 facing
// 的方块。其中 hotBarSlot 指代要放置的方块在快捷栏的位置。
// 而 botPos 则指示当前方块放置操作时，机器人所处的实际位置。
//
// 应当说明的是，PlaceBlockHighLevel 的调用者有义务保证
// 调用 PlaceBlockHighLevel 前，机器人以及出现在 blockPos 所指示的位置。
//
// clickPos 指示为了生成目标方块而使用的基方块，它与 blockPos 是不等价的，
// 但可以确保 clickPos 和 blockPos 是相邻的。
// offsetPos 是 clickPos 相对于 blockPos 的偏移量。
//
// 这意味着，您有义务确保 blockPos 的相邻方块没有被使用，否则它们可能被替换。
// 最后，当您使用完 blockPos 处的方块后，您可以清除 clickPos 处的方块
//
// 值得注意的是，facing 必须是 0 到 5 之间的整数，
// 否则调用 PlaceBlockHighLevel 将返回错误。
//
// 最后，您希望要创建的方块可以是潜影盒，亦可以是旗帜。
// 另外，此函数会等待方块放置完成，但不会自动切换物品栏
func (b *BotClick) PlaceBlockHighLevel(
	blockPos protocol.BlockPos,
	botPos mgl32.Vec3,
	hotBarSlot resources_control.SlotID,
	facing uint8,
) (clickPos protocol.BlockPos, offsetPos protocol.BlockPos, err error) {
	if facing > 5 {
		return clickPos, clickPos, fmt.Errorf("PlaceBlockHighLevel: Given facing (%d) is not meet 0 <= facing <= 5", facing)
	}

	switch facing {
	case 0:
		offsetPos = protocol.BlockPos{0, 1, 0}
	case 1:
		offsetPos = protocol.BlockPos{0, -1, 0}
	case 2:
		offsetPos = protocol.BlockPos{0, 0, 1}
	case 3:
		offsetPos = protocol.BlockPos{0, 0, -1}
	case 4:
		offsetPos = protocol.BlockPos{1, 0, 0}
	case 5:
		offsetPos = protocol.BlockPos{-1, 0, 0}
	}
	clickPos = protocol.BlockPos{
		blockPos[0] + offsetPos[0],
		blockPos[1] + offsetPos[1],
		blockPos[2] + offsetPos[2],
	}

	err = b.s.SetBlock(blockPos, "air", "[]")
	if err != nil {
		return clickPos, offsetPos, fmt.Errorf("PlaceBlockHighLevel: %v", err)
	}
	err = b.s.SetBlock(clickPos, BasePlaceBlock, "[]")
	if err != nil {
		return clickPos, offsetPos, fmt.Errorf("PlaceBlockHighLevel: %v", err)
	}

	err = b.PlaceBlock(
		UseItemOnBlocks{
			HotbarSlotID: hotBarSlot,
			BotPos:       botPos,
			BlockPos:     clickPos,
			BlockName:    BasePlaceBlock,
			BlockStates:  map[string]any{},
		},
		int32(facing),
	)
	if err != nil {
		return clickPos, offsetPos, fmt.Errorf("PlaceBlockHighLevel: %v", err)
	}
	err = b.c.AwaitChangesGeneral()
	if err != nil {
		return clickPos, offsetPos, fmt.Errorf("PlaceBlockHighLevel: %v", err)
	}

	return clickPos, offsetPos, nil
}

// PickBlock 获取 pos 处的方块到物品栏。
// assignNBTData 指示是否需要携带该方块的 NBT 数据。
//
// 返回的 success 指示操作是否成功；
// slot 指示物品最终所在的物品栏位置。
//
// 应当说明的是，
// 手持物品栏会因此切换到 slot 所指示的位置
func (b *BotClick) PickBlock(
	pos protocol.BlockPos,
	assignNBTData bool,
) (
	success bool,
	slot resources_control.SlotID,
	err error,
) {
	doOnce := new(sync.Once)
	channel := make(chan struct{})
	packetListener := b.r.PacketListener()

	for range MaxRetryBlockPick {
		var terminalErr error

		uniqueID, err := packetListener.ListenPacket(
			[]uint32{packet.IDPlayerHotBar},
			func(p packet.Packet, connCloseErr error) {
				doOnce.Do(func() {
					if connCloseErr != nil {
						terminalErr = connCloseErr
					} else {
						slot = resources_control.SlotID(p.(*packet.PlayerHotBar).SelectedHotBarSlot)
						success = true
					}
					close(channel)
				})
			},
		)
		if err != nil {
			return false, 0, fmt.Errorf("PickBlock: %v", err)
		}

		err = b.r.WritePacket(&packet.BlockPickRequest{
			Position:    pos,
			AddBlockNBT: assignNBTData,
			HotBarSlot:  byte(slot),
		})
		if err != nil {
			packetListener.DestroyListener(uniqueID)
			return false, 0, fmt.Errorf("PickBlock: %v", err)
		}

		timer := time.NewTimer(DefaultTimeoutBlockPick)
		defer timer.Stop()
		select {
		case <-timer.C:
		case <-channel:
		}
		packetListener.DestroyListener(uniqueID)

		if terminalErr != nil {
			return false, 0, fmt.Errorf("PickBlock: %v", terminalErr)
		}
		if success {
			break
		}
	}

	return success, slot, nil
}
