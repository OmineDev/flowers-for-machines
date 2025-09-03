package resources_control

import (
	"context"
	"fmt"
	"sync"

	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol"
	"github.com/OmineDev/flowers-for-machines/utils"
)

type (
	// ItemStackRequestID 指示每个物品堆栈操作请求的唯一 ID，
	// 它是以 -1 为首项，公差为 -2 的等差数列
	ItemStackRequestID int32
	// ContainerID 是容器的 ID
	ContainerID uint8
)

type (
	// ExpectedNewItem 描述单个物品堆栈在经历一次物品堆栈操作后，
	// 其最终应当拥有的一些数据信息。应当说明的是，这些数据信息不
	// 会由服务器告知，它应当是客户端内部处理的
	ExpectedNewItem struct {
		// ItemType 指示我们应当如何更新物品的一些基本字段
		ItemType ItemNewType
		// BlockRuntimeID 指示我们应当如何更新物品对应的方块运行时 ID 数据
		BlockRuntimeID ItemNewBlockRuntimeID
		// NBT 指示我们应当如何更新物品的 NBT 字段
		NBT ItemNewNBTData
		// Component 指示我们应该如何更新物品的 Legacy 物品组件字段
		Component ItemNewComponent
	}

	// ItemNewType 描述物品的一些基本字段应如何更新
	ItemNewType struct {
		// UseNetworkID 指示是否需要采用下方的 NetworkID 更新物品的数值网络 ID
		UseNetworkID bool
		// NetworkID 是该物品的数值网络 ID，它在单个 MC 版本中不会变化。
		// 它可正亦可负，具体取决于其所关注的物品堆栈实例
		NetworkID int32
		// UseMetadata 指示是否需要采用下方的 UseMetadata 更新物品的元数据
		UseMetadata bool
		// Metadata 指示这跟物品的新元数据。特别地，耐久数据不在本处设置，敬请参阅
		// [ItemNewNBTData.UseOriginDamage]、[ItemNewNBTData.ChangeDamage] 和
		// [ItemNewNBTData.DamageDelta]
		Metadata uint32
	}

	// ItemNewBlockRuntimeID 描述物品对应的方块运行时数据应该如何更新
	ItemNewBlockRuntimeID struct {
		// UseBlockRuntimeID 指示是否需要采用下方的 BlockRuntimeID 更新物品的方块运行时 ID
		UseBlockRuntimeID bool
		// BlockRuntimeID 指示这个物品获得的新方块运行时 ID 数据
		BlockRuntimeID int32
	}

	// ItemNewNBTData 描述物品的新 NBT 字段如何更新
	ItemNewNBTData struct {
		// UseNBTData 指示是否需要采用下方的 NBTData 更新物品的 NBT 数据
		UseNBTData bool
		// UseOriginDamage 指示在采用下方的 NBTData 时是否保留原有的 Damage
		// 标签的数据。如果原本就不存在，或 UseNBTData 为假，那么最终将不会进
		// 行任何额外的操作。另外，UseOriginDamage 似乎只对存在耐久的物品有效
		UseOriginDamage bool
		// NBTData 指示经过相应的物品堆栈操作后，其 NBT 字段的最终状态。
		// 应当保证 NBTData 是非 nil 的，尽管 NBTData 的长度可能为 0。
		// 需要说明的是，物品名称的 NBT 字段无需在此处更改，它会被自动维护
		NBTData map[string]any

		// ChangeRepairCost 指示是否需要更新物品的 RepairCost 字段。
		// 应当说明的是，RepairCost 被用于铁砧的惩罚机制
		ChangeRepairCost bool
		// RepairCostDelta 是要修改的 RepairCost 的增量，可以为负。
		// 当且仅当 ChangeRepairCost 为真时有效，并且其将在 NBTData 被使用后再应用
		RepairCostDelta int32

		// ChangeDamage 指示是否需要更新的 Damage 字段。
		// ChangeDamage 和 UseOriginDamage 不应同时为真。
		// 当 ChangeDamage 为真时，确保物品最终得到一个 Damage 标签
		ChangeDamage bool
		// DamageDelta 是要修改的 Damage 的增量，可以为负。
		// 当且仅当 ChangeDamage 为真时有效，并且其将在 NBTData 被使用后再应用
		DamageDelta int32
	}

	// ItemNewComponent 描述物品的 Legacy 物品组件应当如何更新
	ItemNewComponent struct {
		// UseCanPlaceOn 指示是否需要采用下方的 CanPlaceOn 更新物品的 can place on 物品组件
		UseCanPlaceOn bool
		// CanPlaceOn 指示物品在 can place on 上的新物品组件数据
		CanPlaceOn []string
		// UseCanDestroy 指示是否需要采用下方的 UseCanDestroy 更新物品的 can destroy 物品组件
		UseCanDestroy bool
		// CanDestroy 指示物品在 can destroy 上的新物品组件数据
		CanDestroy []string
	}
)

type (
	// ItemStackResponseMapping 是一个由容器 ID 到库存窗口 ID 的映射。
	// 由于服务器返回的物品堆栈响应按 ContainerID 来返回更改的物品堆栈，
	// 因此本处的资源处理器定义了下面的运行时映射，以便于操作
	ItemStackResponseMapping map[ContainerID]WindowID

	// ItemStackOperationManager 是所有物品堆栈操作的管理者
	ItemStackOperationManager struct {
		// mu 阻止可能的并发读写
		mu *sync.Mutex
		// ctx 指示底层 Raknet 连接是否已被关闭
		ctx context.Context
		// currentItemStackRequestID 是目前物品堆栈请求的累计 RequestID 计数
		currentItemStackRequestID int32
		// itemStackMapping 存放每个物品堆栈操作请求中的 ItemStackResponseMapping
		itemStackMapping map[ItemStackRequestID]ItemStackResponseMapping
		// itemStackUpdater 存放每个物品堆栈操作请求中相关物品的更新函数
		itemStackUpdater map[ItemStackRequestID]map[SlotLocation]ExpectedNewItem
		// itemStackCallback 存放所有物品堆栈操作请求的回调函数
		itemStackCallback map[ItemStackRequestID]func(response *protocol.ItemStackResponse, connCloseErr error)
	}
)

// NewItemStackOperationManager 根据 ctx 创建并返回一个新的 ItemStackOperationManager
func NewItemStackOperationManager(ctx context.Context) *ItemStackOperationManager {
	return &ItemStackOperationManager{
		mu:                        new(sync.Mutex),
		ctx:                       ctx,
		currentItemStackRequestID: 1,
		itemStackMapping:          make(map[ItemStackRequestID]ItemStackResponseMapping),
		itemStackUpdater:          make(map[ItemStackRequestID]map[SlotLocation]ExpectedNewItem),
		itemStackCallback:         make(map[ItemStackRequestID]func(response *protocol.ItemStackResponse, connCloseErr error)),
	}
}

// NewRequestID 返回一个可以独立使用的新 RequestID
func (i *ItemStackOperationManager) NewRequestID() ItemStackRequestID {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.currentItemStackRequestID -= 2
	return ItemStackRequestID(i.currentItemStackRequestID)
}

// AddNewRequest 设置一个即将发送的物品堆栈操作请求的钩子函数。
// mapping 是由容器 ID 到库存窗口 ID 的映射；
//
// updater 存放每个物品堆栈操作请求中所涉及的特定物品的更新方式。
// 需要说明的是，它不必为单个物品堆栈请求中所涉及的所有物品都设置 ExpectedNewItem。
// 就目前而言，只有 NBT 会因物品堆栈操作而发生变化的物品需要这么操作。
//
// callback 是收到服务器响应后应该执行的回调函数。
// 特别地，如果底层 Raknet 连接关闭，则传入 callback 的 connCloseErr 不为 nil
func (i *ItemStackOperationManager) AddNewRequest(
	requestID ItemStackRequestID,
	mapping ItemStackResponseMapping,
	updater map[SlotLocation]ExpectedNewItem,
	callback func(response *protocol.ItemStackResponse, connCloseErr error),
) {
	i.mu.Lock()
	defer i.mu.Unlock()

	select {
	case <-i.ctx.Done():
		go callback(nil, fmt.Errorf("AddNewRequest: Add new request on closed connection"))
		return
	default:
	}

	i.itemStackMapping[requestID] = mapping
	i.itemStackCallback[requestID] = callback
	if len(updater) > 0 {
		i.itemStackUpdater[requestID] = updater
	}
}

// handleConnClose ..
func (i *ItemStackOperationManager) handleConnClose(err error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	for requestID, cb := range i.itemStackCallback {
		go cb(nil, err)
		i.itemStackCallback[requestID] = nil
	}

	i.currentItemStackRequestID = 1
	i.itemStackMapping = nil
	i.itemStackUpdater = nil
	i.itemStackCallback = nil
}

// UpdateNetworkItem 通过 serverResponse 和 clientExpected 共同评估 item 的新值。
// slotLocation 指示该物品的位置。应当说明的是，相关修改将直接在 item 上进行。
// UpdateItem 保证 clientExpected 所指示的数据不会被意外重用，内部实现将使用深拷贝
func UpdateNetworkItem(
	item *protocol.ItemInstance,
	slotLocation SlotLocation,
	serverResponse protocol.StackResponseSlotInfo,
	clientExpected map[SlotLocation]ExpectedNewItem,
) {
	item.Stack.Count = uint16(serverResponse.Count)
	item.StackNetworkID = serverResponse.StackNetworkID

	if clientExpected != nil {
		updater, ok := clientExpected[slotLocation]
		if ok {
			UpdateItemClientSide(item, slotLocation, updater)
		}
	}

	UpdateDisplay(item, serverResponse)
}

// UpdateItemClientSide 通过 clientExpected 评估 item 的新值。
// slotLocation 指示该物品的位置。应当说明的是，相关修改将直接在 item 上进行。
// UpdateItemData 保证 clientExpected 所指示的数据不会被意外重用，内部实现将使用深拷贝
func UpdateItemClientSide(
	item *protocol.ItemInstance,
	slotLocation SlotLocation,
	clientExpected ExpectedNewItem,
) {
	// Prepare
	var originDamageExist bool
	var originDamage int32
	if item.Stack.NBTData != nil {
		originDamage, originDamageExist = item.Stack.NBTData["Damage"].(int32)
	}

	// Update to new network ID and new metadata
	if clientExpected.ItemType.UseNetworkID {
		item.Stack.ItemType.NetworkID = clientExpected.ItemType.NetworkID
	}
	if clientExpected.ItemType.UseMetadata {
		item.Stack.ItemType.MetadataValue = clientExpected.ItemType.Metadata
	}
	// Update to new block runtime ID
	if clientExpected.BlockRuntimeID.UseBlockRuntimeID {
		item.Stack.BlockRuntimeID = clientExpected.BlockRuntimeID.BlockRuntimeID
	}
	// Update to new can place on
	if clientExpected.Component.UseCanPlaceOn {
		item.Stack.CanBePlacedOn = make([]string, len(clientExpected.Component.CanPlaceOn))
		copy(item.Stack.CanBePlacedOn, clientExpected.Component.CanPlaceOn)
	}
	// Update to new can destroy
	if clientExpected.Component.UseCanDestroy {
		item.Stack.CanBreak = make([]string, len(clientExpected.Component.CanDestroy))
		copy(item.Stack.CanBreak, clientExpected.Component.CanDestroy)
	}

	// Update to new NBT data
	if clientExpected.NBT.UseNBTData {
		item.Stack.NBTData = utils.DeepCopyNBT(clientExpected.NBT.NBTData)
		// Use origin damage if needed
		if clientExpected.NBT.UseOriginDamage && originDamageExist {
			item.Stack.NBTData["Damage"] = originDamage
		}
	}
	// Update to new damage
	if clientExpected.NBT.ChangeDamage {
		if item.Stack.NBTData == nil {
			item.Stack.NBTData = make(map[string]any)
		}
		damage, _ := item.Stack.NBTData["Damage"].(int32)
		damage += clientExpected.NBT.DamageDelta
		item.Stack.NBTData["Damage"] = damage
	}
	// Update to new repair cost
	if clientExpected.NBT.ChangeRepairCost {
		if item.Stack.NBTData == nil {
			item.Stack.NBTData = make(map[string]any)
		}
		repairCost, _ := item.Stack.NBTData["RepairCost"].(int32)
		repairCost += clientExpected.NBT.RepairCostDelta
		item.Stack.NBTData["RepairCost"] = repairCost
	}
}

// UpdateItemDisplay 通过 serverResponse 评估 item 的新自定义物品名数据。
// slotLocation 指示该物品的位置。应当说明的是，相关修改将直接在 item 上进行
func UpdateDisplay(
	item *protocol.ItemInstance,
	serverResponse protocol.StackResponseSlotInfo,
) {
	// 物品没有 NBT 数据，但有物品名称
	if len(item.Stack.NBTData) == 0 && len(serverResponse.CustomName) > 0 {
		item.Stack.NBTData = map[string]any{
			"display": map[string]any{
				"Name": serverResponse.CustomName,
			},
		}
	}

	// 物品存在 NBT 数据
	if len(item.Stack.NBTData) > 0 {
		_, displayExisted := item.Stack.NBTData["display"].(map[string]any)

		// 不存在自定义物品名 且 不存在 display
		if len(serverResponse.CustomName) == 0 && !displayExisted {
			return
		}

		// 存在自定义物品名 且 (不)存在 display
		if len(serverResponse.CustomName) > 0 {
			// 存在自定义物品名 且 存在 display
			if displayExisted {
				item.Stack.NBTData["display"].(map[string]any)["Name"] = serverResponse.CustomName
				return
			}
			// 存在自定义物品名 且 不存在 display
			item.Stack.NBTData["display"] = map[string]any{
				"Name": serverResponse.CustomName,
			}
			return
		}

		// 不存在自定义物品名 且 存在 display
		delete(item.Stack.NBTData["display"].(map[string]any), "Name")
		if len(item.Stack.NBTData["display"].(map[string]any)) == 0 {
			delete(item.Stack.NBTData, "display")
		}
	}
}
