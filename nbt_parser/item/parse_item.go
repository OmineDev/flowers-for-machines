package nbt_parser_item

import (
	"fmt"

	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol"
	"github.com/OmineDev/flowers-for-machines/mapping"
	nbt_parser_interface "github.com/OmineDev/flowers-for-machines/nbt_parser/interface"
)

// ParseItemNormal 从 nbtMap 解析一个 NBT 物品。
// nbtMap 是含有这个物品 tag 标签的父复合标签。
//
// nameChecker 是一个可选的函数，用于检查 name 所
// 指示的物品名称是否可通过指令获取。如果不能，则返
// 回的 canGetByCommand 为假。
//
// 无论 canGetByCommand 的值是多少，如果解析没有发
// 生错误，则 item 不会为空。
//
// 另外，如果没有这样的 nameChecker 函数，则可以将其
// 简单的置为 nil
func ParseItemNormal(
	nameChecker func(name string) bool,
	nbtMap map[string]any,
) (item nbt_parser_interface.Item, canGetByCommand bool, err error) {
	defaultItem := DefaultItem{
		NameChecker: nameChecker,
	}

	err = defaultItem.ParseNormal(nbtMap)
	if err != nil {
		return nil, false, fmt.Errorf("ParseItemNormal: %v", err)
	}

	if nameChecker != nil {
		canGetByCommand = nameChecker(defaultItem.ItemName())
	} else {
		canGetByCommand = true
	}

	itemType, ok := mapping.SupportItemsPool[defaultItem.ItemName()]
	if !ok {
		return &defaultItem, canGetByCommand, nil
	}

	switch itemType {
	case mapping.SupportNBTItemTypeBook:
		item = &Book{DefaultItem: defaultItem}
	case mapping.SupportNBTItemTypeBanner:
		item = &Banner{DefaultItem: defaultItem}
	case mapping.SupportNBTItemTypeShield:
		item = &Shield{DefaultItem: defaultItem}
	default:
		panic("ParseItemNormal: Should never happened")
	}

	err = item.ParseNormal(nbtMap)
	if err != nil {
		return nil, false, fmt.Errorf("ParseItemNormal: %v", err)
	}
	return item, canGetByCommand, nil
}

// ParseItemNetwork 解析网络传输上的物品堆栈实例 item。
// itemName 是这个物品堆栈实例的名称
func ParseItemNetwork(itemStack protocol.ItemStack, itemName string) (item nbt_parser_interface.Item, err error) {
	var defaultItem DefaultItem

	err = defaultItem.ParseNetwork(itemStack, itemName)
	if err != nil {
		return nil, fmt.Errorf("ParseItemNetwork: %v", err)
	}

	itemType, ok := mapping.SupportItemsPool[defaultItem.ItemName()]
	if !ok {
		return &defaultItem, nil
	}

	switch itemType {
	case mapping.SupportNBTItemTypeBook:
		item = &Book{DefaultItem: defaultItem}
	case mapping.SupportNBTItemTypeBanner:
		item = &Banner{DefaultItem: defaultItem}
	case mapping.SupportNBTItemTypeShield:
		item = &Shield{DefaultItem: defaultItem}
	default:
		panic("ParseItemNetwork: Should never happened")
	}

	err = item.ParseNetwork(itemStack, itemName)
	if err != nil {
		return nil, fmt.Errorf("ParseItemNetwork: %v", err)
	}
	return item, nil
}

func init() {
	nbt_parser_interface.ParseItemNormal = ParseItemNormal
	nbt_parser_interface.ParseItemNetwork = ParseItemNetwork
}
