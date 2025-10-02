package nbt_item

import (
	"fmt"

	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol/packet"
	"github.com/OmineDev/flowers-for-machines/game_control/game_interface"
	"github.com/OmineDev/flowers-for-machines/game_control/resources_control"
	"github.com/OmineDev/flowers-for-machines/nbt_assigner/nbt_console"
	nbt_hash "github.com/OmineDev/flowers-for-machines/nbt_parser/hash"
	nbt_parser_interface "github.com/OmineDev/flowers-for-machines/nbt_parser/interface"
	nbt_parser_item "github.com/OmineDev/flowers-for-machines/nbt_parser/item"
	"github.com/OmineDev/flowers-for-machines/utils"
)

// 成书
type Book struct {
	api   *nbt_console.Console
	items []nbt_parser_item.Book
}

func (b *Book) Append(item ...nbt_parser_interface.Item) {
	for _, value := range item {
		val, ok := value.(*nbt_parser_item.Book)
		if !ok {
			continue
		}
		b.items = append(b.items, *val)
	}
}

func (b *Book) Make() (resultSlot map[uint64]resources_control.SlotID, err error) {
	api := b.api.API()

	if len(b.items) == 0 {
		return nil, nil
	}

	for {
		var shouldRepeat bool
		currentRound := b.items[0:min(len(b.items), 9)]
		bookSlots := make([]resources_control.SlotID, 0)

		// Get writable book
		for index, book := range currentRound {
			inventorySlot := resources_control.SlotID(index)

			err = b.api.API().Replaceitem().ReplaceitemInInventory(
				"@s",
				game_interface.ReplacePathHotbarOnly,
				game_interface.ReplaceitemInfo{
					Name:     "minecraft:writable_book",
					Count:    1,
					MetaData: book.ItemMetadata(),
					Slot:     inventorySlot,
				},
				utils.MarshalItemComponent(book.Enhance.ItemComponent),
				false,
			)
			if err != nil {
				return nil, fmt.Errorf("Make: %v", err)
			}

			bookSlots = append(bookSlots, inventorySlot)
			b.api.UseInventorySlot(nbt_console.RequesterUser, inventorySlot, true)
		}

		// Wait replaceitem to finish
		err = api.Commands().AwaitChangesGeneral()
		if err != nil {
			return nil, fmt.Errorf("Make: %v", err)
		}

		// Write (and sign) book
		for index, book := range currentRound {
			wantSlotID := resources_control.SlotID(index)

			if b.api.HotbarSlotID() != wantSlotID {
				err = b.api.ChangeAndUpdateHotbarSlotID(wantSlotID)
				if err != nil {
					return nil, fmt.Errorf("Make: %v", err)
				}
			}

			err = api.BotClick().ClickAir(wantSlotID, b.api.Position())
			if err != nil {
				return nil, fmt.Errorf("Make: %v", err)
			}
			err = api.Commands().AwaitChangesGeneral()
			if err != nil {
				return nil, fmt.Errorf("Make: %v", err)
			}

			for pageNumber, content := range book.NBT.Pages {
				err = api.Resources().WritePacket(
					&packet.BookEdit{
						ActionType:    packet.BookActionReplacePage,
						InventorySlot: byte(wantSlotID),
						PageNumber:    byte(pageNumber),
						Text:          content,
					},
				)
				if err != nil {
					return nil, fmt.Errorf("Make: %v", err)
				}
			}

			if book.ItemName() == "minecraft:written_book" {
				err = api.Resources().WritePacket(
					&packet.BookEdit{
						ActionType:    packet.BookActionSign,
						InventorySlot: byte(wantSlotID),
						Title:         book.NBT.Title,
						Author:        book.NBT.Author,
					},
				)
				if err != nil {
					return nil, fmt.Errorf("Make: %v", err)
				}
			}
		}

		// Wait book action to finish
		err = api.Commands().AwaitChangesGeneral()
		if err != nil {
			return nil, fmt.Errorf("Make: %v", err)
		}

		// Check completely
		for index, book := range currentRound {
			item, inventoryExisted := api.Resources().Inventories().GetItemStack(0, bookSlots[index])
			if !inventoryExisted {
				panic("Make: Should never happened")
			}

			if item.Stack.NetworkID != int32(api.Resources().ConstantPacket().ItemByName(book.ItemName()).RuntimeID) {
				shouldRepeat = true
				break
			}

			bookWeGet, err := nbt_parser_interface.ParseItemNetwork(item.Stack, book.ItemName())
			if err != nil {
				return nil, fmt.Errorf("Make: %v", err)
			}

			if nbt_hash.NBTItemNBTHash(bookWeGet) != nbt_hash.NBTItemNBTHash(&book) {
				shouldRepeat = true
				break
			}
		}
		if shouldRepeat {
			continue
		}

		// Sync result slot
		resultSlot = make(map[uint64]resources_control.SlotID)
		for index, book := range currentRound {
			resultSlot[nbt_hash.NBTItemNBTHash(&book)] = bookSlots[index]
		}

		// Remove the book we finished
		b.items = b.items[len(currentRound):]

		// Return
		return resultSlot, nil
	}
}
