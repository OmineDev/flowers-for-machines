package item_stack_transaction

import (
	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol"
	"github.com/OmineDev/flowers-for-machines/game_control/game_interface/item_stack_operation"
	"github.com/OmineDev/flowers-for-machines/game_control/resources_control"
	"github.com/OmineDev/flowers-for-machines/mapping"
)

// MoveItem 将 source 处的物品移动到 destination 处，
// 且只移动 count 个物品。
//
// 该操作是支持内联的，它会与所有相邻的支持内联的操作一
// 起被内联到单个物品堆栈操作请求中
func (i *ItemStackTransaction) MoveItem(
	source resources_control.SlotLocation,
	destination resources_control.SlotLocation,
	count uint8,
) *ItemStackTransaction {
	i.operations = append(i.operations, item_stack_operation.Move{
		Source:      source,
		Destination: destination,
		Count:       int32(count),
	})
	return i
}

// MoveBetweenInventory 将背包中 source 处的物品移动到 destination 处，
// 且只移动 count 个物品。
//
// 此操作需要保证背包已被打开，或者已打开的容器中可以在背包中移动物品。
//
// 该操作是支持内联的，它会与所有相邻的支持内联的操作一起被内联到单个
// 物品堆栈操作请求中
func (i *ItemStackTransaction) MoveBetweenInventory(
	source resources_control.SlotID,
	destination resources_control.SlotID,
	count uint8,
) *ItemStackTransaction {
	return i.MoveItem(
		resources_control.SlotLocation{
			WindowID: protocol.WindowIDInventory,
			SlotID:   source,
		},
		resources_control.SlotLocation{
			WindowID: protocol.WindowIDInventory,
			SlotID:   destination,
		},
		count,
	)
}

// MoveBetweenContainer 将已打开容器中 source 处的物品
// 移动到已打开容器的 destination 处，且只移动 count 个物品。
//
// 此操作需要保证目前已经打开了一个容器，否则效果将会与
// MoveBetweenInventory 等同。
//
// 该操作是支持内联的，它会与所有相邻的支持内联的操作一起被内联
// 到单个物品堆栈操作请求中
func (i *ItemStackTransaction) MoveBetweenContainer(
	source resources_control.SlotID,
	destination resources_control.SlotID,
	count uint8,
) *ItemStackTransaction {
	data, _, _ := i.api.Container().ContainerData()
	windowID := resources_control.WindowID(data.WindowID)
	return i.MoveItem(
		resources_control.SlotLocation{
			WindowID: windowID,
			SlotID:   source,
		},
		resources_control.SlotLocation{
			WindowID: windowID,
			SlotID:   destination,
		},
		count,
	)
}

// MoveToContainer 将背包中 source 处的物品移动到已打开容器的
// destination 处，且只移动 count 个物品。
//
// 此操作需要保证目前已经打开了一个容器，否则效果将会与
// MoveBetweenInventory 等同。
//
// 该操作是支持内联的，它会与所有相邻的支持内联的操作一起被内联
// 到单个物品堆栈操作请求中
func (i *ItemStackTransaction) MoveToContainer(
	source resources_control.SlotID,
	destination resources_control.SlotID,
	count uint8,
) *ItemStackTransaction {
	data, _, _ := i.api.Container().ContainerData()
	return i.MoveItem(
		resources_control.SlotLocation{
			WindowID: protocol.WindowIDInventory,
			SlotID:   source,
		},
		resources_control.SlotLocation{
			WindowID: resources_control.WindowID(data.WindowID),
			SlotID:   destination,
		},
		count,
	)
}

// MoveToInventory 将已打开容器中 source 处的物品移动到
// 背包的 destination 处，且只移动 count 个物品。
//
// 此操作需要保证目前已经打开了一个容器，否则效果将会与
// MoveBetweenInventory 等同。
//
// 该操作是支持内联的，它会与所有相邻的支持内联的操作一起
// 被内联到单个物品堆栈操作请求中
func (i *ItemStackTransaction) MoveToInventory(
	source resources_control.SlotID,
	destination resources_control.SlotID,
	count uint8,
) *ItemStackTransaction {
	data, _, _ := i.api.Container().ContainerData()
	return i.MoveItem(
		resources_control.SlotLocation{
			WindowID: resources_control.WindowID(data.WindowID),
			SlotID:   source,
		},
		resources_control.SlotLocation{
			WindowID: protocol.WindowIDInventory,
			SlotID:   destination,
		},
		count,
	)
}

// MoveToCraftingTable 将背包中 source 处的物品移动
// 到合成栏的 destination 处，且只移动 count 个物品。
//
// 此操作需要保证背包已被打开，或打开了工作台，
// 否则整个事务将会失败。
//
// 该操作是支持内联的，它会与所有相邻的支持内联的操作一
// 起被内联到单个物品堆栈操作请求中
func (i *ItemStackTransaction) MoveToCraftingTable(
	source resources_control.SlotID,
	destination resources_control.SlotID,
	count uint8,
) *ItemStackTransaction {
	return i.MoveItem(
		resources_control.SlotLocation{
			WindowID: protocol.WindowIDInventory,
			SlotID:   source,
		},
		resources_control.SlotLocation{
			WindowID: protocol.WindowIDCrafting,
			SlotID:   destination,
		},
		count,
	)
}

// MoveFromCraftingTable 将已放入合成栏 source 处的物品
// 移动到背包的 destination 处，且只移动 count 个物品。
//
// 此操作需要保证背包已被打开，或打开了工作台，否则整个事务
// 将会失败。
//
// 该操作是支持内联的，它会与所有相邻的支持内联的操作一起被
// 内联到单个物品堆栈操作请求中
func (i *ItemStackTransaction) MoveFromCraftingTable(
	source resources_control.SlotID,
	destination resources_control.SlotID,
	count uint8,
) *ItemStackTransaction {
	return i.MoveItem(
		resources_control.SlotLocation{
			WindowID: protocol.WindowIDCrafting,
			SlotID:   source,
		},
		resources_control.SlotLocation{
			WindowID: protocol.WindowIDInventory,
			SlotID:   destination,
		},
		count,
	)
}

// SwapItem 交换 source 处和 destination 处的物品。
//
// 该操作是支持内联的，它会与所有相邻的支持内联的操作
// 一起被内联到单个物品堆栈操作请求中
func (i *ItemStackTransaction) SwapItem(
	source resources_control.SlotLocation,
	destination resources_control.SlotLocation,
) *ItemStackTransaction {
	i.operations = append(i.operations, item_stack_operation.Swap{
		Source:      source,
		Destination: destination,
	})
	return i
}

// SwapBetweenInventory 交换背包中 source
// 处和背包中 destination 处的物品。
//
// 此操作需要保证背包已被打开，或者已打开
// 的容器中可以在背包中移动物品。
//
// 该操作是支持内联的，它会与所有相邻的支
// 持内联的操作一起被内联到单个物品堆栈操
// 作请求中
func (i *ItemStackTransaction) SwapBetweenInventory(
	source resources_control.SlotID,
	destination resources_control.SlotID,
) *ItemStackTransaction {
	return i.SwapItem(
		resources_control.SlotLocation{
			WindowID: protocol.WindowIDInventory,
			SlotID:   source,
		},
		resources_control.SlotLocation{
			WindowID: protocol.WindowIDInventory,
			SlotID:   destination,
		},
	)
}

// SwapInventoryBetweenContainer 交换背包中 source
// 处和已打开容器 destination 处的物品。
//
// 此操作需要保证目前已经打开了一个容器，否则效果将会与
// SwapInventoryItem 等同。
//
// 该操作是支持内联的，它会与所有相邻的支持内联的操作一起
// 被内联到单个物品堆栈操作请求中
func (i *ItemStackTransaction) SwapInventoryBetweenContainer(
	source resources_control.SlotID,
	destination resources_control.SlotID,
) *ItemStackTransaction {
	data, _, _ := i.api.Container().ContainerData()
	return i.SwapItem(
		resources_control.SlotLocation{
			WindowID: protocol.WindowIDInventory,
			SlotID:   source,
		},
		resources_control.SlotLocation{
			WindowID: resources_control.WindowID(data.WindowID),
			SlotID:   destination,
		},
	)
}

// DropItem 将 slot 处的物品丢出，且只丢出 count 个。
//
// 该操作是支持内联的，它会与所有相邻的支持内联的操作一
// 起被内联到单个物品堆栈操作请求中
func (i *ItemStackTransaction) DropItem(slot resources_control.SlotLocation, count uint8) *ItemStackTransaction {
	i.operations = append(i.operations, item_stack_operation.Drop{
		Path:  slot,
		Count: count,
	})
	return i
}

// DropInventoryItem 将背包中 slot 处的
// 物品丢出，且只丢出 count 个。
//
// 此操作需要保证背包已被打开，
// 或者已打开的容器中可以在背包中移动物品。
//
// 该操作是支持内联的，它会与所有相邻的支
// 持内联的操作一起被内联到单个物品堆栈操
// 作请求中
func (i *ItemStackTransaction) DropInventoryItem(slot resources_control.SlotID, count uint8) *ItemStackTransaction {
	return i.DropItem(
		resources_control.SlotLocation{
			WindowID: protocol.WindowIDInventory,
			SlotID:   slot,
		},
		count,
	)
}

// DropItem 将已打开容器 slot 处的物品丢出，且只丢出 count 个。
//
// 此操作需要保证目前已经打开了一个容器，否则效果将会与
// DropInventoryItem 等同。
//
// 该操作是支持内联的，它会与所有相邻的支持内联的操作一起被内联
// 到单个物品堆栈操作请求中
func (i *ItemStackTransaction) DropContainerItem(slot resources_control.SlotID, count uint8) *ItemStackTransaction {
	data, _, _ := i.api.Container().ContainerData()
	return i.DropItem(
		resources_control.SlotLocation{
			WindowID: resources_control.WindowID(data.WindowID),
			SlotID:   slot,
		},
		count,
	)
}

// GetCreativeItem 从创造物品栏获取 创造物品网络 ID 为
// creativeItemNetworkID 的物品到 slot 处，
// 且只移动 count 个物品。
//
// 该操作不支持内联，但它仍然可以被紧缩在单个的物品堆栈操
// 作请求的数据包中
func (i *ItemStackTransaction) GetCreativeItem(
	creativeItemNetworkID uint32,
	slot resources_control.SlotLocation,
	count uint8,
) *ItemStackTransaction {
	i.operations = append(i.operations, item_stack_operation.CreativeItem{
		CINI:  creativeItemNetworkID,
		Path:  slot,
		Count: count,
	})
	return i
}

// GetCreativeItemToInventory 从创造物品栏获取创造物品网络
// ID 为 creativeItemNetworkID 的物品到背包中的 slot 处，
// 且只移动 count 个物品。
//
// 该操作不支持内联，但它仍然可以被紧缩在单个的物品堆栈操作请
// 求的数据包中
func (i *ItemStackTransaction) GetCreativeItemToInventory(
	creativeItemNetworkID uint32,
	slot resources_control.SlotID,
	count uint8,
) *ItemStackTransaction {
	return i.GetCreativeItem(
		creativeItemNetworkID,
		resources_control.SlotLocation{
			WindowID: protocol.WindowIDInventory,
			SlotID:   slot,
		},
		count,
	)
}

// RenameItem 将 slot 处的物品全部重命名为 newName
//
// 重命名操作是通过铁砧完成的，这意味着您需要确保铁砧已被打开，
// 且铁砧内没有放置任何物品。
//
// 如果操作成功，则物品将回到原位。
//
// 该操作不支持内联，但它仍然可以被紧缩在单个的物品堆栈操作请
// 求的数据包中
func (i *ItemStackTransaction) RenameItem(slot resources_control.SlotLocation, newName string) *ItemStackTransaction {
	i.operations = append(i.operations, item_stack_operation.Renaming{
		Path:    slot,
		NewName: newName,
	})
	return i
}

// RenameInventoryItem 将背包中 slot 处的物品全部重命名为 newName。
//
// 重命名操作是通过铁砧完成的，这意味着您需要确保铁砧已被打开，
// 且铁砧内没有放置任何物品。如果操作成功，则物品将回到原位。
//
// 与 RenameItem 的不同之处在于，它只能操作背包中的物品，
// 因此您需要确保背包已被打开。
//
// 该操作不支持内联，但它仍然可以被紧缩在单个的物品堆栈操作请求的数据包中
func (i *ItemStackTransaction) RenameInventoryItem(slot resources_control.SlotID, newName string) *ItemStackTransaction {
	return i.RenameItem(
		resources_control.SlotLocation{
			WindowID: protocol.WindowIDInventory,
			SlotID:   slot,
		},
		newName,
	)
}

// Looming 将 patternSlot 处的旗帜放入织布机中，
// 并通过使用 dyeSlot 处的染料合成新旗帜。
//
// patternName 是织布时使用的图案，patternSlot
// 则指示该图案物品的位置。如果无需使用图案，
// 请将 patternName 和 patternSlot 都置为默认的
// 零值。
//
// resultItem 指示期望得到的旗帜的部分数据。
// 如果操作成功，则新旗帜将回到原位。
//
// 该操作不支持内联，但它仍然可以被紧缩在单个的物品
// 堆栈操作请求的数据包中
func (i *ItemStackTransaction) Looming(
	patternName string,
	patternSlot resources_control.SlotLocation,
	bannerSlot resources_control.SlotLocation,
	dyeSlot resources_control.SlotLocation,
	resultItem resources_control.ExpectedNewItem,
) *ItemStackTransaction {
	_, usePattern := mapping.BannerPatternToItemName[patternName]

	i.operations = append(i.operations, item_stack_operation.Looming{
		UsePattern:  usePattern,
		PatternName: patternName,
		PatternPath: patternSlot,
		BannerPath:  bannerSlot,
		DyePath:     dyeSlot,
		ResultItem:  resultItem,
	})

	return i
}

// LoomingFromInventory 将背包中 patternSlot 处的旗帜放入织布机中，
// 并通过使用背包中 dyeSlot 处的染料合成新旗帜。
//
// patternName 是织布时使用的图案，patternSlot 则指示该图案物品
// 在背包中的位置。如果使用的旗帜图案无需实际的旗帜图案物品用于合成，
// 请将 patternSlot 置为默认的零值。
//
// resultItem 指示期望得到的旗帜的部分数据。
// 如果操作成功，则新旗帜将回到原位。
//
// 该操作不支持内联，但它仍然可以被紧缩在单个的物品堆栈操作请求的数据包中
func (i *ItemStackTransaction) LoomingFromInventory(
	patternName string,
	patternSlot resources_control.SlotID,
	bannerSlot resources_control.SlotID,
	dyeSlot resources_control.SlotID,
	resultItem resources_control.ExpectedNewItem,
) *ItemStackTransaction {
	return i.Looming(
		patternName,
		resources_control.SlotLocation{
			WindowID: protocol.WindowIDInventory,
			SlotID:   patternSlot,
		},
		resources_control.SlotLocation{
			WindowID: protocol.WindowIDInventory,
			SlotID:   bannerSlot,
		},
		resources_control.SlotLocation{
			WindowID: protocol.WindowIDInventory,
			SlotID:   dyeSlot,
		},
		resultItem,
	)
}

// Crafting 用于合成一个物品。
//
// 它消耗已放入合成栏的全部物品，
// 然后制作相应的物品到背包。
//
// recipeNetworkID 是合成配方的网络 ID；
// resultSlotID 是合成后物品应当放置的位置；
// resultCount 是合成后所得物品的数量；
// resultItem 是合成后物品的最终数据
func (i *ItemStackTransaction) Crafting(
	recipeNetworkID uint32,
	resultSlotID resources_control.SlotID,
	resultCount uint8,
	resultItem resources_control.ExpectedNewItem,
) *ItemStackTransaction {
	i.operations = append(i.operations, item_stack_operation.Crafting{
		RecipeNetworkID: recipeNetworkID,
		ResultSlotID:    resultSlotID,
		ResultCount:     resultCount,
		ResultItem:      resultItem,
	})
	return i
}

// Trimming 将下面列出的物品放置在锻造台中，
// 并进行对于的锻造台的纹饰操作。
//
// - trimItemPath 处的 1 个装备
// - materialPath 处的 1 个材料
// - templatePath 处的 1 个模板
//
// resultItem 指示期望得到的锻造结果数据。
// 如果操作成功，则被锻造物品将回到原位。
//
// 该操作不支持内联，但它仍然可以被紧缩在单个
// 的物品堆栈操作请求的数据包中
func (i *ItemStackTransaction) Trimming(
	trimItemPath resources_control.SlotLocation,
	materialPath resources_control.SlotLocation,
	templatePath resources_control.SlotLocation,
	resultItem resources_control.ExpectedNewItem,
) *ItemStackTransaction {
	i.operations = append(i.operations, item_stack_operation.Trimming{
		TrimItem:   trimItemPath,
		Material:   materialPath,
		Template:   templatePath,
		ResultItem: resultItem,
	})
	return i
}

// TrimmingFromInventory 将下面列出的，
// 位于背包中的物品放置在锻造台中，
// 并进行对于的锻造台的纹饰操作。
//
// - trimItemSlot 处的 1 个装备
// - materialSlot 处的 1 个材料
// - templateSlot 处的 1 个模板
//
// resultItem 指示期望得到的锻造结果数据。
// 如果操作成功，则被锻造物品将回到原位。
//
// 该操作不支持内联，但它仍然可以被紧缩在单个
// 的物品堆栈操作请求的数据包中
func (i *ItemStackTransaction) TrimmingFromInventory(
	trimItemSlot resources_control.SlotID,
	materialSlot resources_control.SlotID,
	templateSlot resources_control.SlotID,
	resultItem resources_control.ExpectedNewItem,
) *ItemStackTransaction {
	return i.Trimming(
		resources_control.SlotLocation{
			WindowID: protocol.WindowIDInventory,
			SlotID:   trimItemSlot,
		},
		resources_control.SlotLocation{
			WindowID: protocol.WindowIDInventory,
			SlotID:   materialSlot,
		},
		resources_control.SlotLocation{
			WindowID: protocol.WindowIDInventory,
			SlotID:   templateSlot,
		},
		resultItem,
	)
}
