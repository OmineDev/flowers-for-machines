package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol"
	"github.com/df-mc/worldupgrader/itemupgrader"
)

// ItemComponent 是一个物品的物品组件数据
type ItemComponent struct {
	// 控制此物品/方块 (在冒险模式下) 可以使用/放置在其上的方块类型
	CanPlaceOn []string
	// 控制此物品/方块 (在冒险模式下) 可以破坏的方块类型。
	// 此效果不会改变原本的破坏速度和破坏后掉落物
	CanDestroy []string
	// 阻止该物品被从玩家的物品栏
	// 移除、丢弃或用于合成
	LockInInventory bool
	// 阻止该物品被从玩家物品栏的该槽位
	// 移动、移除、丢弃或用于合成
	LockInSlot bool
	// 使该物品在玩家死亡时不会掉落
	KeepOnDeath bool
}

// NeedFormat ..
func (i ItemComponent) NeedFormat() bool {
	if len(i.CanPlaceOn) > 0 || len(i.CanDestroy) > 0 {
		return true
	}
	if i.LockInInventory || i.LockInSlot {
		return true
	}
	if i.KeepOnDeath {
		return true
	}
	return false
}

// Format ..
func (i ItemComponent) Format(prefix string) string {
	result := ""

	if canPlaceOnCount := len(i.CanPlaceOn); canPlaceOnCount > 0 {
		result += prefix + fmt.Sprintf("冒险放置 (合计 %d 个): \n", canPlaceOnCount)
		for _, canPlaceOn := range i.CanPlaceOn {
			result += prefix + "\t- " + canPlaceOn
		}
	} else {
		result += prefix + "冒险放置: 不存在\n"
	}

	if canDestroyCount := len(i.CanDestroy); canDestroyCount > 0 {
		result += prefix + fmt.Sprintf("冒险破坏 (合计 %d 个): \n", canDestroyCount)
		for _, canDestroy := range i.CanDestroy {
			result += prefix + "\t- " + canDestroy
		}
	} else {
		result += prefix + "冒险破坏: 不存在\n"
	}

	if i.LockInInventory {
		result += prefix + "物品锁定: 物品锁定在背包\n"
	} else if i.LockInSlot {
		result += prefix + "物品锁定: 物品锁定在物品栏\n"
	} else {
		result += prefix + "物品锁定: 无\n"
	}
	result += prefix + fmt.Sprintf("在死亡时保留: %s\n", FormatBool(i.KeepOnDeath))

	return result
}

// Marshal ..
func (i *ItemComponent) Marshal(io protocol.IO) {
	protocol.FuncSliceUint16Length(io, &i.CanPlaceOn, io.String)
	protocol.FuncSliceUint16Length(io, &i.CanDestroy, io.String)
	io.Bool(&i.LockInInventory)
	io.Bool(&i.LockInSlot)
	io.Bool(&i.KeepOnDeath)
}

// ParseItemComponent 从 nbtMap 解析一个物品的物品组件数据
func ParseItemComponent(nbtMap map[string]any) (result ItemComponent) {
	list, ok := nbtMap["CanDestroy"].([]any)
	if ok {
		for _, value := range list {
			val, ok := value.(string)
			if !ok {
				continue
			}

			val = strings.ToLower(val)
			if !strings.HasPrefix(val, "minecraft:") {
				val = "minecraft:" + val
			}

			newItem := itemupgrader.Upgrade(
				itemupgrader.ItemMeta{
					Name: val,
					Meta: 0,
				},
			)
			result.CanDestroy = append(result.CanDestroy, newItem.Name)
		}
	}

	list, ok = nbtMap["CanPlaceOn"].([]any)
	if ok {
		for _, value := range list {
			val, ok := value.(string)
			if !ok {
				continue
			}

			val = strings.ToLower(val)
			if !strings.HasPrefix(val, "minecraft:") {
				val = "minecraft:" + val
			}

			newItem := itemupgrader.Upgrade(
				itemupgrader.ItemMeta{
					Name: val,
					Meta: 0,
				},
			)
			result.CanPlaceOn = append(result.CanPlaceOn, newItem.Name)
		}
	}

	tag, ok := nbtMap["tag"].(map[string]any)
	if !ok {
		return
	}

	itemLock, _ := tag["minecraft:item_lock"].(byte)
	switch itemLock {
	case 1:
		result.LockInSlot = true
	case 2:
		result.LockInInventory = true
	}

	keepOnDeath, _ := tag["minecraft:keep_on_death"].(byte)
	if keepOnDeath == 1 {
		result.KeepOnDeath = true
	}

	return
}

// ParseItemComponentNetwork 从 item 解析一个物品的物品组件数据
func ParseItemComponentNetwork(item protocol.ItemStack) (result ItemComponent) {
	result.CanDestroy = make([]string, len(item.CanBreak))
	result.CanPlaceOn = make([]string, len(item.CanBePlacedOn))
	copy(result.CanDestroy, item.CanBreak)
	copy(result.CanPlaceOn, item.CanBePlacedOn)

	if item.NBTData == nil {
		return
	}

	itemLock, _ := item.NBTData["minecraft:item_lock"].(byte)
	switch itemLock {
	case 1:
		result.LockInSlot = true
	case 2:
		result.LockInInventory = true
	}

	keepOnDeath, _ := item.NBTData["minecraft:keep_on_death"].(byte)
	if keepOnDeath == 1 {
		result.KeepOnDeath = true
	}

	return
}

// MarshalItemComponent 将 component 序列化为 MC 命令中的物品组件字符串
func MarshalItemComponent(component ItemComponent) string {
	type Blocks struct {
		Blocks []string `json:"blocks"`
	}
	type Mode struct {
		Mode string `json:"mode"`
	}
	type Component struct {
		CanPlaceOn  *Blocks   `json:"can_place_on,omitempty"`
		CanDestroy  *Blocks   `json:"can_destroy,omitempty"`
		ItemLock    *Mode     `json:"item_lock,omitempty"`
		KeepOnDeath *struct{} `json:"keep_on_death,omitempty"`
	}

	c := Component{}
	if len(component.CanDestroy) > 0 {
		c.CanDestroy = &Blocks{Blocks: component.CanDestroy}
	}
	if len(component.CanPlaceOn) > 0 {
		c.CanPlaceOn = &Blocks{Blocks: component.CanPlaceOn}
	}
	if component.LockInInventory {
		c.ItemLock = &Mode{Mode: "lock_in_inventory"}
	}
	if component.LockInSlot {
		c.ItemLock = &Mode{Mode: "lock_in_slot"}
	}
	if component.KeepOnDeath {
		c.KeepOnDeath = &struct{}{}
	}

	resultBytes, _ := json.Marshal(c)
	result := string(resultBytes)
	if result == "{}" {
		return ""
	}

	return result
}
