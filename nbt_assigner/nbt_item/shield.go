package nbt_item

import (
	"fmt"
	"slices"

	"github.com/OmineDev/flowers-for-machines/game_control/game_interface"
	"github.com/OmineDev/flowers-for-machines/game_control/resources_control"
	"github.com/OmineDev/flowers-for-machines/mapping"
	"github.com/OmineDev/flowers-for-machines/nbt_assigner/nbt_console"
	nbt_parser_general "github.com/OmineDev/flowers-for-machines/nbt_parser/general"
	nbt_hash "github.com/OmineDev/flowers-for-machines/nbt_parser/hash"
	nbt_parser_interface "github.com/OmineDev/flowers-for-machines/nbt_parser/interface"
	nbt_parser_item "github.com/OmineDev/flowers-for-machines/nbt_parser/item"
	"github.com/OmineDev/flowers-for-machines/utils"

	"github.com/mitchellh/mapstructure"
)

// 盾牌
type Shield struct {
	api   *nbt_console.Console
	items []nbt_parser_item.Shield
}

func (s *Shield) Append(item ...nbt_parser_interface.Item) {
	for _, value := range item {
		val, ok := value.(*nbt_parser_item.Shield)
		if !ok {
			continue
		}
		s.items = append(s.items, *val)
	}
}

func (s *Shield) Make() (resultSlot map[uint64]resources_control.SlotID, err error) {
	api := s.api.API()
	if len(s.items) == 0 {
		return nil, nil
	}

	// Prepare
	bannerNBTHashToIndex := make(map[uint64]int)
	bannerItems := make([]nbt_parser_item.Banner, len(s.items))
	banners := Banner{
		api:             s.api,
		items:           make([]nbt_parser_item.Banner, len(s.items)),
		maxSlotCanUse:   BannerMaxSlotCanUse,
		maxBannerToMake: ShieldMaxBannerToMake,
	}

	// Step 1: Convert shields to banners
	for index, shield := range s.items {
		banner := nbt_parser_item.Banner{
			DefaultItem: nbt_parser_item.DefaultItem{
				Basic: nbt_parser_item.ItemBasicData{
					Name:     "minecraft:banner",
					Count:    1,
					Metadata: int16(shield.NBT.Base),
				},
			},
			NBT: nbt_parser_item.BannerNBT{
				Patterns: shield.NBT.Patterns,
				Type:     nbt_parser_general.BannerTypeNormal,
			},
		}

		for _, val := range shield.NBT.Patterns {
			if val.Pattern == mapping.BannerPatternOminous {
				banner.NBT.Type = nbt_parser_general.BannerTypeOminous
				banner.NBT.Patterns = nil
				break
			}
		}

		banners.items[index] = banner
		bannerItems[index] = banner
		bannerNBTHashToIndex[nbt_hash.NBTItemNBTHash(&banner)] = index
	}

	// Step 2: Make banners
	// Note that due to we consider non-complex banners (like banners without any pattern)
	// when we construct the Make func of banner, so here we can call banners.Make directly.
	// However, the define of Make is refer to make complex item (with NBT data).
	resultBanner, err := banners.Make()
	if err != nil {
		return nil, fmt.Errorf("Make: %v", err)
	}
	if len(resultBanner) == 0 {
		return nil, nil
	}

	// Note: Bind bannerSlots[i], bannerHashes[i], shieldSlots[i] and shieldExpectedItem[i]
	bannerSlots := make([]resources_control.SlotID, 0)
	bannerHashes := make([]uint64, 0)
	shieldSlots := make([]resources_control.SlotID, 0)
	shieldExpectedItem := make([]resources_control.ExpectedNewItem, 0)

	// Step 3.1: Iter banners and expected their shields
	for hashNumber, slot := range resultBanner {
		bannerSlots = append(bannerSlots, slot)
		bannerHashes = append(bannerHashes, hashNumber)

		originItemIndex := bannerNBTHashToIndex[hashNumber]
		banner := bannerItems[originItemIndex]
		shield := s.items[originItemIndex]

		expectedItem := resources_control.ExpectedNewItem{
			ItemType: resources_control.ItemNewType{
				UseNetworkID: true,
				NetworkID:    int32(api.Resources().ConstantPacket().ItemByName("minecraft:shield").RuntimeID),
				UseMetadata:  true,
				Metadata:     0,
			},
			BlockRuntimeID: resources_control.ItemNewBlockRuntimeID{
				UseBlockRuntimeID: true,
				BlockRuntimeID:    0,
			},
			NBT: resources_control.ItemNewNBTData{
				UseNBTData:       true,
				UseOriginDamage:  false,
				NBTData:          make(map[string]any),
				ChangeRepairCost: false,
				ChangeDamage:     false,
			},
			Component: resources_control.ItemNewComponent{
				UseCanPlaceOn: true,
				CanPlaceOn:    shield.Enhance.ItemComponent.CanPlaceOn,
				UseCanDestroy: true,
				CanDestroy:    shield.Enhance.ItemComponent.CanDestroy,
			},
		}

		if shield.DefaultItem.Enhance.ItemComponent.KeepOnDeath {
			expectedItem.NBT.NBTData["minecraft:keep_on_death"] = byte(1)
		}
		if banner.NBT.Type == nbt_parser_general.BannerTypeOminous {
			expectedItem.NBT.NBTData = map[string]any{
				"Base": int32(15),
				"Patterns": []any{
					map[string]any{
						"Color":   int32(15),
						"Pattern": mapping.BannerPatternOminous,
					},
				},
			}
		} else {
			expectedItem.NBT.NBTData["Base"] = int32(banner.ItemMetadata())
			expectedItem.NBT.NBTData["Damage"] = int32(shield.ItemMetadata())

			nbtPatterns := make([]any, 0)
			for _, pattern := range banner.NBT.Patterns {
				var nbtPattern map[string]any
				err = mapstructure.Decode(&pattern, &nbtPattern)
				if err != nil {
					return nil, fmt.Errorf("Make: %v", err)
				}
				nbtPatterns = append(nbtPatterns, nbtPattern)
			}

			if len(nbtPatterns) > 0 {
				expectedItem.NBT.NBTData["Patterns"] = nbtPatterns
			}
		}

		shieldExpectedItem = append(shieldExpectedItem, expectedItem)
	}

	occupySlots := slices.Clone(bannerSlots)

	// Step 3.2: Find slot to place shield and do replaceitem
	for index := range bannerSlots {
		bannerHashNumber := bannerHashes[index]
		originItemIndex := bannerNBTHashToIndex[bannerHashNumber]
		shield := s.items[originItemIndex]

		inventorySlot := s.api.FindInventorySlot(occupySlots)
		shieldSlots = append(shieldSlots, inventorySlot)

		err = api.Replaceitem().ReplaceitemInInventory(
			"@s",
			game_interface.ReplacePathInventory,
			game_interface.ReplaceitemInfo{
				Name:     "minecraft:shield",
				Count:    1,
				MetaData: shield.ItemMetadata(),
				Slot:     inventorySlot,
			},
			utils.MarshalItemComponent(shield.Enhance.ItemComponent),
			false,
		)
		if err != nil {
			return nil, fmt.Errorf("Make: %v", err)
		}

		s.api.UseInventorySlot(nbt_console.RequesterUser, inventorySlot, true)
		occupySlots = append(occupySlots, inventorySlot)
	}

	// Step 3.3: Wait replaceitem to finish
	err = api.Commands().AwaitChangesGeneral()
	if err != nil {
		return nil, fmt.Errorf("Make: %v", err)
	}

	// Step 4: Open inventory
	success, err := api.ContainerOpenAndClose().OpenInventory()
	if err != nil {
		return nil, fmt.Errorf("Make: %v", err)
	}
	if !success {
		return nil, fmt.Errorf("Make: Failed to open the inventory")
	}
	defer api.ContainerOpenAndClose().CloseContainer()

	// Step 5: Open transaction and do crafting
	transaction := api.ItemStackOperation().OpenTransaction()
	for index := range bannerSlots {
		transaction.
			MoveToCraftingTable(bannerSlots[index], 28, 1).
			MoveToCraftingTable(shieldSlots[index], 29, 1).
			Crafting(2418, shieldSlots[index], 1, shieldExpectedItem[index])
	}

	// Step 6: Commit changes
	success, _, _, err = transaction.Commit()
	if err != nil {
		return nil, fmt.Errorf("Make: %v", err)
	}
	if !success {
		return nil, fmt.Errorf("Make: The server rejected the crafting stack request actions")
	}

	finishedShieldIndex := make(map[int]bool)
	resultSlot = make(map[uint64]resources_control.SlotID)

	// Step 7: Compute resultSlot and the shield we finished
	for index, shieldSlot := range shieldSlots {
		bannerHashNumber := bannerHashes[index]
		originItemIndex := bannerNBTHashToIndex[bannerHashNumber]
		shield := s.items[originItemIndex]
		resultSlot[nbt_hash.NBTItemNBTHash(&shield)] = shieldSlot
		finishedShieldIndex[originItemIndex] = true
	}

	// Setp 8: Check hash only
	for index, shieldSlot := range shieldSlots {
		bannerHashNumber := bannerHashes[index]
		originItemIndex := bannerNBTHashToIndex[bannerHashNumber]
		shield := s.items[originItemIndex]

		shieldWeGet, inventoryExisted := api.Resources().Inventories().GetItemStack(0, shieldSlot)
		if !inventoryExisted {
			panic("Make: Should never happened")
		}

		if shieldWeGet.Stack.NetworkID != int32(api.Resources().ConstantPacket().ItemByName("minecraft:shield").RuntimeID) {
			panic("Make: Should never happened")
		}
		newShield, err := nbt_parser_interface.ParseItemNetwork(shieldWeGet.Stack, "minecraft:shield")
		if err != nil {
			return nil, fmt.Errorf("Make: %v", err)
		}

		if nbt_hash.NBTItemNBTHash(newShield) != nbt_hash.NBTItemNBTHash(&shield) {
			panic("Make: Should never happened")
		}
	}

	// Step 8: Remove the shield we finished
	newItems := make([]nbt_parser_item.Shield, 0)
	for index, value := range s.items {
		if finishedShieldIndex[index] {
			continue
		}
		newItems = append(newItems, value)
	}
	s.items = newItems

	// Step 9: Return
	return resultSlot, nil
}
