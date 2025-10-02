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

const (
	// BannerMaxSlotCanUse 指示单次旗帜制作轮次中，
	// 可以使用的最多的物品栏数量。
	// 因为一共有 36 个物品栏，所以该值为 36
	BannerMaxSlotCanUse uint8 = 36
	// BannerMaxBannerToMake 是在仅制作旗帜时，
	// 单次能制作的最多的旗帜数量。
	// 因为一共有 36 个物品栏，所以最终可以制作
	// 最多 36 个旗帜
	BannerMaxBannerToMake uint8 = 36
	// ShieldMaxBannerToMake 是在制作带图案的盾牌时，
	// 单次能制作的最多的旗帜数量。考虑到旗帜和盾牌各
	// 占用一格，并且一共 36 个物品栏，所以最终可以制
	// 作最多 36/2 = 18 个旗帜
	ShieldMaxBannerToMake uint8 = 18
)

// 旗帜
type Banner struct {
	api             *nbt_console.Console
	items           []nbt_parser_item.Banner
	maxSlotCanUse   uint8
	maxBannerToMake uint8
}

// planner 计算并给出本次可以制作的旗帜，
// 以及制作它需要用到的染料和旗帜图案。
// 保证给出的旗帜列表包含尽可能多的旗帜
func (b *Banner) planner() (
	bannerToMake []int,
	colorToUse []int32,
	colorToUseMapping map[int32]int,
	patternToUse []string,
	patternToUseMapping map[string]int,
) {
	// These fields means the banner we can make,
	// and the color and pattern we needed.
	usedBannersCount := 0
	usedBanners := make([]bool, len(b.items))
	usedColors := make(map[int32]bool)
	usedPatterns := make(map[string]bool)
	// Do planning (Algorithm: Greedy)
	for {
		// These two fileds holds the best banner for this round
		bestOneNewColors := make(map[int32]bool)
		bestOneNewPatterns := make(map[string]bool)
		// Same as above, but they are numbers
		bestOneIndex := -1
		bestOneNewColorsCount := 25
		bestOneNewPatternsCount := 25
		// Iter all unfinished banner to get the best one
		for index, banner := range b.items {
			// If this banner is used, then we don't need to iter it
			if usedBanners[index] {
				continue
			}
			// Prepare
			isOminousBanner := (banner.NBT.Type == nbt_parser_general.BannerTypeOminous)
			currentNewColors := make(map[int32]bool)
			currentNewPatterns := make(map[string]bool)
			// Get new colors and new patterns of current banner
			for _, pattern := range banner.NBT.Patterns {
				// Ominous banner check
				if pattern.Pattern == mapping.BannerPatternOminous {
					isOminousBanner = true
					break
				}
				// Set new colors
				if _, ok := usedColors[pattern.Color]; !ok {
					currentNewColors[pattern.Color] = true
				}
				// Set new patterns
				if _, ok := mapping.BannerPatternToItemName[pattern.Pattern]; !ok {
					// This pattern don't need pattern item
					continue
				}
				if _, ok := usedPatterns[pattern.Pattern]; !ok {
					currentNewPatterns[pattern.Pattern] = true
				}
			}
			// Ominous banner is not consider in planner func
			if isOminousBanner {
				continue
			}
			// Try to find the banner that need new color and patterns at least
			if len(currentNewColors) < bestOneNewColorsCount && len(currentNewPatterns) < bestOneNewPatternsCount {
				bestOneIndex = index
				bestOneNewColorsCount = len(currentNewColors)
				bestOneNewPatternsCount = len(currentNewPatterns)
				bestOneNewColors = currentNewColors
				bestOneNewPatterns = currentNewPatterns
			}
		}
		// bestOneIndex is -1 means all banner is finished
		if bestOneIndex == -1 {
			break
		}
		// after is the slot count we need in total (include the best banner)
		after := len(usedColors) + bestOneNewColorsCount     // Colors
		after += len(usedPatterns) + bestOneNewPatternsCount // Patterns
		after += usedBannersCount + 1                        // Banners
		// Check if we can make more banner or not
		if after > int(b.maxSlotCanUse) || usedBannersCount+1 > int(b.maxBannerToMake) {
			break
		}
		// The loop is not break, so we can apply changes, and current best banner can be make
		for key := range bestOneNewColors {
			usedColors[key] = true
		}
		for key := range bestOneNewPatterns {
			usedPatterns[key] = true
		}
		usedBanners[bestOneIndex] = true
		usedBannersCount += 1
	}
	// Prepare
	colorToUseMapping = make(map[int32]int)
	patternToUseMapping = make(map[string]int)
	// Convert usedColors and usedPatterns to mapping,
	// and make usedBanners to slice.
	for index, value := range usedBanners {
		if value {
			bannerToMake = append(bannerToMake, index)
		}
	}
	for key := range usedColors {
		colorToUseMapping[key] = len(colorToUse)
		colorToUse = append(colorToUse, key)
	}
	for key := range usedPatterns {
		patternToUseMapping[key] = len(patternToUse)
		patternToUse = append(patternToUse, key)
	}
	// Return
	return
}

// makeNormal 使用织布机制作不含灾厄旗帜的旗帜。
// 向 makeNormal 传入的参数应当是 planner 的返回值
func (b *Banner) makeNormal(
	bannerToMake []int,
	colorToUse []int32,
	colorToUseMapping map[int32]int,
	patternToUse []string,
	patternToUseMapping map[string]int,
) (resultSlot map[uint64]resources_control.SlotID, err error) {
	api := b.api.API()
	slot := resources_control.SlotID(0)

	// Make expected new item
	resultBanners := make([]resources_control.ExpectedNewItem, len(bannerToMake))
	for idx, index := range bannerToMake {
		banner := b.items[index]
		nbtPatterns := make([]any, 0)

		for _, pattern := range banner.NBT.Patterns {
			var singlePattern map[string]any

			err := mapstructure.Decode(&pattern, &singlePattern)
			if err != nil {
				return nil, fmt.Errorf("makeNormal: %v", err)
			}

			nbtPatterns = append(nbtPatterns, singlePattern)
		}

		bannerNBT := map[string]any{
			"Patterns": nbtPatterns,
			"Type":     nbt_parser_general.BannerTypeNormal,
		}
		if banner.DefaultItem.Enhance.ItemComponent.KeepOnDeath {
			bannerNBT["minecraft:keep_on_death"] = byte(1)
		}

		resultBanners[idx] = resources_control.ExpectedNewItem{
			ItemType: resources_control.ItemNewType{
				UseNetworkID: true,
				NetworkID:    int32(api.Resources().ConstantPacket().ItemByName("minecraft:banner").RuntimeID),
				UseMetadata:  true,
				Metadata:     uint32(banner.ItemMetadata()),
			},
			BlockRuntimeID: resources_control.ItemNewBlockRuntimeID{
				UseBlockRuntimeID: true,
				BlockRuntimeID:    0,
			},
			NBT: resources_control.ItemNewNBTData{
				UseNBTData:       true,
				UseOriginDamage:  false,
				NBTData:          bannerNBT,
				ChangeRepairCost: false,
				ChangeDamage:     false,
			},
			Component: resources_control.ItemNewComponent{
				UseCanPlaceOn: true,
				CanPlaceOn:    banner.Enhance.ItemComponent.CanPlaceOn,
				UseCanDestroy: true,
				CanDestroy:    banner.Enhance.ItemComponent.CanDestroy,
			},
		}
	}

	// Get all dyes
	for _, value := range colorToUse {
		dyeName, ok := mapping.BannerColorToDyeName[value]
		if !ok {
			panic("makeNormal: Should never happened")
		}

		err := api.Replaceitem().ReplaceitemInInventory(
			"@s",
			game_interface.ReplacePathInventory,
			game_interface.ReplaceitemInfo{
				Name:     dyeName,
				Count:    64,
				MetaData: 0,
				Slot:     slot,
			},
			"",
			false,
		)
		if err != nil {
			return nil, fmt.Errorf("makeNormal: %v", err)
		}

		b.api.UseInventorySlot(nbt_console.RequesterUser, slot, true)
		slot++
	}
	// Get all paterns
	for _, value := range patternToUse {
		patternName, ok := mapping.BannerPatternToItemName[value]
		if !ok {
			panic("makeNormal: Should never happened")
		}

		err := api.Replaceitem().ReplaceitemInInventory(
			"@s",
			game_interface.ReplacePathInventory,
			game_interface.ReplaceitemInfo{
				Name:     patternName,
				Count:    64,
				MetaData: 0,
				Slot:     slot,
			},
			"",
			false,
		)
		if err != nil {
			return nil, fmt.Errorf("makeNormal: %v", err)
		}

		b.api.UseInventorySlot(nbt_console.RequesterUser, slot, true)
		slot++
	}
	// Get all banners
	for _, index := range bannerToMake {
		banner := b.items[index].DefaultItem

		err := api.Replaceitem().ReplaceitemInInventory(
			"@s",
			game_interface.ReplacePathInventory,
			game_interface.ReplaceitemInfo{
				Name:     banner.ItemName(),
				Count:    1,
				MetaData: banner.ItemMetadata(),
				Slot:     slot,
			},
			utils.MarshalItemComponent(banner.Enhance.ItemComponent),
			false,
		)
		if err != nil {
			return nil, fmt.Errorf("makeNormal: %v", err)
		}

		b.api.UseInventorySlot(nbt_console.RequesterUser, slot, true)
		slot++
	}
	// Await changes
	err = api.Commands().AwaitChangesGeneral()
	if err != nil {
		return nil, fmt.Errorf("makeNormal: %v", err)
	}

	// Find or generate new loom
	index, err := b.api.FindOrGenerateNewLoom()
	if err != nil {
		return nil, fmt.Errorf("makeNormal: %v", err)
	}
	// Open loom
	success, err := b.api.OpenContainerByIndex(index)
	if err != nil {
		return nil, fmt.Errorf("makeNormal: %v", err)
	}
	if !success {
		return nil, fmt.Errorf("makeNormal: Failed to open the loom block")
	}
	defer api.ContainerOpenAndClose().CloseContainer()

	// Open transaction and do operation
	transaction := api.ItemStackOperation().OpenTransaction()
	resultSlot = make(map[uint64]resources_control.SlotID)
	for idx, index := range bannerToMake {
		// Get banner item
		banner := b.items[index]
		// Compute offset and banner slot
		offsetPattern := len(colorToUse)
		bannerSlot := resources_control.SlotID(len(colorToUse) + len(patternToUse) + idx)
		// Update result slot
		resultSlot[nbt_hash.NBTItemNBTHash(&banner)] = bannerSlot
		// Add loom operation
		for _, pattern := range banner.NBT.Patterns {
			_ = transaction.LoomingFromInventory(
				pattern.Pattern,
				resources_control.SlotID(offsetPattern+patternToUseMapping[pattern.Pattern]),
				bannerSlot,
				resources_control.SlotID(colorToUseMapping[pattern.Color]),
				resultBanners[idx],
			)
		}
	}

	// Commit changes
	success, _, _, err = transaction.Commit()
	if err != nil {
		return nil, fmt.Errorf("makeNormal: %v", err)
	}
	if !success {
		return nil, fmt.Errorf("makeNormal: Looming operation rejected by the server")
	}

	// Check hash only
	for idx, index := range bannerToMake {
		banner := b.items[index]
		bannerSlot := resources_control.SlotID(len(colorToUse) + len(patternToUse) + idx)

		bannerWeGet, inventoryExisted := api.Resources().Inventories().GetItemStack(0, bannerSlot)
		if !inventoryExisted {
			panic("Make: Should never happened")
		}

		if bannerWeGet.Stack.NetworkID != int32(api.Resources().ConstantPacket().ItemByName("minecraft:banner").RuntimeID) {
			panic("Make: Should never happened")
		}
		newBanner, err := nbt_parser_interface.ParseItemNetwork(bannerWeGet.Stack, "minecraft:banner")
		if err != nil {
			return nil, fmt.Errorf("Make: %v", err)
		}

		if nbt_hash.NBTItemNBTHash(newBanner) != nbt_hash.NBTItemNBTHash(&banner) {
			panic("Make: Should never happened")
		}
	}

	// Return
	return resultSlot, nil
}

// makeOminous 通过创造物品栏制作一个灾厄旗帜
func (b *Banner) makeOminous() (resultSlot map[uint64]resources_control.SlotID, err error) {
	api := b.api.API()
	inventorySlot := b.api.FindInventorySlot(nil)

	err = api.Replaceitem().ReplaceitemInInventory(
		"@s",
		game_interface.ReplacePathInventory,
		game_interface.ReplaceitemInfo{
			Name:     "minecraft:air",
			Slot:     inventorySlot,
			Count:    1,
			MetaData: 0,
		},
		"",
		true,
	)
	if err != nil {
		return nil, fmt.Errorf("makeOminous: %v", err)
	}
	b.api.UseInventorySlot(nbt_console.RequesterUser, inventorySlot, false)

	success, err := api.ContainerOpenAndClose().OpenInventory()
	if err != nil {
		return nil, fmt.Errorf("makeOminous: %v", err)
	}
	if !success {
		return nil, fmt.Errorf("makeOminous: Failed to open the inventory")
	}
	defer api.ContainerOpenAndClose().CloseContainer()

	cini := uint32(0)
	banners := api.Resources().ConstantPacket().CreativeItemByName("minecraft:banner")
	for _, banner := range banners {
		if banner.Item.NBTData != nil && banner.Item.NBTData["Type"] == int32(1) {
			cini = banner.CreativeItemNetworkID
			break
		}
	}
	if cini == 0 {
		panic("makeOminous: Should never happened")
	}

	success, _, _, err = api.ItemStackOperation().OpenTransaction().
		GetCreativeItemToInventory(cini, inventorySlot, 1).
		Commit()
	if err != nil {
		return nil, fmt.Errorf("makeOminous: %v", err)
	}
	if !success {
		return nil, fmt.Errorf("makeOminous: Creative item request rejected by the server")
	}
	b.api.UseInventorySlot(nbt_console.RequesterUser, inventorySlot, true)

	return map[uint64]resources_control.SlotID{
		nbt_hash.NBTItemNBTHash(&b.items[0]): inventorySlot,
	}, nil
}

func (b *Banner) Append(item ...nbt_parser_interface.Item) {
	for _, value := range item {
		val, ok := value.(*nbt_parser_item.Banner)
		if !ok {
			continue
		}
		b.items = append(b.items, *val)
	}
}

func (b *Banner) Make() (resultSlot map[uint64]resources_control.SlotID, err error) {
	if len(b.items) == 0 {
		return nil, nil
	}

	bannerToMake, colorToUse, colorToUseMapping, patternToUse, patternToUseMapping := b.planner()
	if len(bannerToMake) > 0 {
		resultSlot, err = b.makeNormal(bannerToMake, colorToUse, colorToUseMapping, patternToUse, patternToUseMapping)
	} else {
		resultSlot, err = b.makeOminous()
	}
	if err != nil {
		return nil, fmt.Errorf("Make: %v", err)
	}

	if len(bannerToMake) > 0 {
		newItems := make([]nbt_parser_item.Banner, 0)
		for index, value := range b.items {
			if slices.Contains(bannerToMake, index) {
				continue
			}
			newItems = append(newItems, value)
		}
		b.items = newItems
	} else {
		b.items = nil
	}

	return
}
