package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol"
	"github.com/OmineDev/flowers-for-machines/game_control/game_interface"
	"github.com/OmineDev/flowers-for-machines/game_control/resources_control"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/pterm/pterm"
)

func SystemTestingItemStackOperation() {
	tA := time.Now()

	// Test round 1
	{
		api.Commands().SendSettingsCommand("clear", true)
		api.Commands().AwaitChangesGeneral()
		api.Commands().SendSettingsCommand("give @s apple 25", true) // Slot 0 (0 -> 25)
		api.Commands().AwaitChangesGeneral()
		api.Commands().SendSettingsCommand("give @s diamond_sword 1", true) // Slot 1 (0 -> 1)
		api.Commands().AwaitChangesGeneral()
		api.Commands().SendSettingsCommand("give @s red_flower 20", true) // Slot 2 (0 -> 20)
		api.Commands().AwaitChangesGeneral()

		success, _, _, _ := api.ItemStackOperation().OpenTransaction().
			MoveBetweenInventory(0, 3, 24). // Slot 3 (0 -> 24), Slot 0 (25 -> 1)
			MoveBetweenInventory(0, 3, 1).  // Slot 0 -> Slot 3
			MoveBetweenInventory(1, 4, 1).  // Slot 1 -> Slot 4
			MoveBetweenInventory(3, 0, 25). // Slot 3 -> Slot 0
			MoveBetweenInventory(0, 3, 10). // Slot 3 (0 -> 10), Slot 0 (25 -> 15)
			MoveBetweenInventory(3, 5, 5).  // Slot 5 (0 -> 5); Slot 3 (10 -> 5)
			MoveToContainer(4, 0, 1).       // Slot 4 (Inventory) -> Slot 0 (Chest)
			MoveToContainer(5, 1, 5).       // Slot 5 (Inventory) -> Slot 1 (Chest)
			MoveToContainer(3, 1, 5).       // Slot 3 (Inventory) -> Slot 1 (Chest)
			MoveToContainer(2, 2, 20).      // Slot 2 (Inventory) -> Slot 2 (Chest)
			MoveToInventory(2, 8, 6).       // Slot 2 (Chest, 20 -> 14) -> Slot 8 (Inventory, 0 -> 6)
			DropInventoryItem(8, 3).        // Slot 8 (6 -> 3)
			MoveToContainer(8, 2, 3).       // Slot 8 (Inventory, 3 -> 0) -> Slot 2 (Chest, 14 -> 17)
			DropInventoryItem(0, 15).       // Slot 0 (15 -> 0)
			Commit()
		if !success {
			panic("SystemTestingItemStackOperation: Failed on test round 1")
		}
	}

	// Test round 2
	{
		err := api.ContainerOpenAndClose().CloseContainer()
		if err != nil {
			panic(fmt.Sprintf("SystemTestingItemStackOperation: Test round 2 failed due to %v (stage 1)", err))
		}

		api.Commands().SendSettingsCommand("clear", true)
		api.Commands().AwaitChangesGeneral()
		api.Commands().SendSettingsCommand("give @s apple 25", true) // Slot 0 (0 -> 25)
		api.Commands().AwaitChangesGeneral()
		api.Commands().SendSettingsCommand("give @s diamond_sword 1", true) // Slot 1 (0 -> 1)
		api.Commands().AwaitChangesGeneral()
		api.Commands().SendSettingsCommand("give @s red_flower 20", true) // Slot 2 (0 -> 20)
		api.Commands().AwaitChangesGeneral()

		states, err := api.SetBlock().SetAnvil([3]int32{0, 0, 0}, true)
		if err != nil {
			panic(fmt.Sprintf("SystemTestingItemStackOperation: Test round 2 failed due to %v (stage 2)", err))
		}

		success, err := api.ContainerOpenAndClose().OpenContainer(
			game_interface.UseItemOnBlocks{
				HotbarSlotID: 2,
				BotPos:       mgl32.Vec3{0, 0, 0},
				BlockPos:     [3]int32{0, 0, 0},
				BlockName:    "anvil",
				BlockStates:  states,
			},
			true,
		)
		if err != nil {
			panic(fmt.Sprintf("SystemTestingItemStackOperation: Test round 2 failed due to %v (stage 3)", err))
		}
		if !success {
			panic("SystemTestingItemStackOperation: Failed on test round 2")
		}

		success, _, _, _ = api.ItemStackOperation().OpenTransaction().
			MoveBetweenInventory(0, 3, 25).                       // Slot 0 -> Slot 3
			MoveBetweenInventory(1, 4, 1).                        // Slot 1 -> Slot 4
			MoveBetweenInventory(2, 5, 20).                       // Slot 2 -> Slot 5
			RenameInventoryItem(3, "SYSTEM TESTING A").           // Hacking Attempt
			RenameInventoryItem(4, "SYSTEM TESTING B").           // Hacking Attempt
			SwapBetweenInventory(3, 4).                           // Slot 3 <-> Slot 4
			SwapBetweenInventory(3, 5).                           // Slot 3 <-> Slot 5
			RenameInventoryItem(5, "INLINE").                     // Hacking Attempt
			RenameInventoryItem(5, "INLINE A").                   // Hacking Attempt
			RenameInventoryItem(3, "INLINE B").                   // Hacking Attempt
			RenameInventoryItem(4, "INLINE C").                   // Hacking Attempt
			RenameInventoryItem(3, "§r§fflowers for m[A]chines"). // Real Name
			RenameInventoryItem(4, "§r§fAPPLE").                  // Real Name
			RenameInventoryItem(5, "§r§fSWORD").                  // Real Name
			SwapBetweenInventory(3, 4).                           // Slot 3 <-> Slot 4
			SwapBetweenInventory(4, 5).                           // Slot 4 <-> Slot 5
			MoveToContainer(3, 1, 25).                            // Slot 3 (Inventory) -> Slot 1 (Anvil)
			DropContainerItem(1, 25).                             // Slot 1 (Anvil, 25 -> 0)
			SwapBetweenInventory(4, 5).                           // Slot 4 <-> Slot 5
			SwapBetweenInventory(4, 5).                           // Slot 4 <-> Slot 5
			SwapBetweenInventory(5, 4).                           // Slot 5 <-> Slot 4
			DropInventoryItem(5, 1).                              // Slot 5 (1 -> 0)
			MoveBetweenInventory(4, 5, 10).                       // Slot 4 (20 -> 10) -> Slot 5 (0 -> 10)
			MoveToContainer(4, 1, 10).                            // Slot 4 (Inventory) -> Slot 1 (Anvil)
			SwapInventoryBetweenContainer(5, 1).                  // Slot 5 (Anvil) <-> Slot 1 (Anvil)
			MoveToInventory(1, 4, 10).                            // Slot 1 (Anvil) -> Slot 4 (Inventory)
			DropInventoryItem(4, 10).                             // Slot 4 (10 -> 0)
			DropInventoryItem(5, 10).                             // Slot 5 (10 -> 0)
			Commit()
		if !success {
			panic("SystemTestingItemStackOperation: Failed on test round 2")
		}

		err = api.ContainerOpenAndClose().CloseContainer()
		if err != nil {
			panic(fmt.Sprintf("SystemTestingItemStackOperation: Test round 2 failed due to %v (stage 4)", err))
		}
	}

	// Test round 3
	{
		success, err := api.ContainerOpenAndClose().OpenInventory()
		if err != nil {
			panic(fmt.Sprintf("SystemTestingItemStackOperation: Test round 3 failed due to %v (stage 1)", err))
		}
		if !success {
			panic("SystemTestingItemStackOperation: Failed on test round 3")
		}

		api.Commands().SendSettingsCommand("clear", true)
		api.Commands().AwaitChangesGeneral()
		success, _, _, _ = api.ItemStackOperation().OpenTransaction().
			GetCreativeItemToInventory(1, 0, 64).
			GetCreativeItemToInventory(2, 1, 64).
			GetCreativeItemToInventory(1570, 8, 1).
			DropInventoryItem(0, 64).
			DropInventoryItem(1, 64).
			DropInventoryItem(8, 1).
			Commit()
		if !success {
			panic("SystemTestingItemStackOperation: Failed on test round 3")
		}

		err = api.ContainerOpenAndClose().CloseContainer()
		if err != nil {
			panic(fmt.Sprintf("SystemTestingItemStackOperation: Test round 3 failed due to %v (stage 2)", err))
		}
	}

	// Test round 4
	{
		api.Commands().SendSettingsCommand("clear", true)
		api.Commands().AwaitChangesGeneral()
		api.Commands().SendSettingsCommand("give @s banner 1 10", true) // Slot 0
		api.Commands().AwaitChangesGeneral()
		api.Commands().SendSettingsCommand("give @s yellow_dye 20", true) // Slot 1
		api.Commands().AwaitChangesGeneral()
		api.Commands().SendSettingsCommand("give @s mojang_banner_pattern 1", true) // Slot 2
		api.Commands().AwaitChangesGeneral()
		api.Commands().SendSettingsCommand("give @s red_dye 20", true) // Slot 3
		api.Commands().AwaitChangesGeneral()
		api.Commands().SendSettingsCommand("give @s skull_banner_pattern", true) // Slot 4
		api.Commands().AwaitChangesGeneral()
		api.Commands().SendSettingsCommand("give @s bordure_indented_banner_pattern", true) // Slot 5
		api.Commands().AwaitChangesGeneral()
		api.Commands().SendSettingsCommand("give @s banner 1 11", true) // Slot 6
		api.Commands().AwaitChangesGeneral()
		api.Commands().SendSettingsCommand("give @s light_blue_dye 20", true) // Slot 7
		api.Commands().AwaitChangesGeneral()
		api.Commands().SendSettingsCommand("give @s shield 2", true) // Slot 8
		api.Commands().AwaitChangesGeneral()

		err := api.SetBlock().SetBlock(protocol.BlockPos{0, 0, 0}, "loom", `["direction"=0]`)
		if err != nil {
			panic(fmt.Sprintf("SystemTestingItemStackOperation: Test round 4 failed due to %v (stage 1)", err))
		}

		success, err := api.ContainerOpenAndClose().OpenContainer(
			game_interface.UseItemOnBlocks{
				HotbarSlotID: 2,
				BotPos:       mgl32.Vec3{0, 0, 0},
				BlockPos:     protocol.BlockPos{0, 0, 0},
				BlockName:    "loom",
				BlockStates: map[string]any{
					"direction": int32(0),
				},
			},
			false,
		)
		if err != nil {
			panic(fmt.Sprintf("SystemTestingItemStackOperation: Test round 4 failed due to %v (stage 2)", err))
		}
		if !success {
			panic("SystemTestingItemStackOperation: Failed on test round 4")
		}

		success, _, _, _ = api.ItemStackOperation().OpenTransaction().
			LoomingFromInventory("bo", 0, 0, 1, resources_control.ExpectedNewItem{}).  // Banner 1 (1)
			LoomingFromInventory("moj", 2, 0, 3, resources_control.ExpectedNewItem{}). // Banner 1 (2)
			LoomingFromInventory("sku", 4, 0, 1, resources_control.ExpectedNewItem{}). // Banner 1 (3)
			LoomingFromInventory("sku", 4, 0, 1, resources_control.ExpectedNewItem{}). // Banner 1 (4)
			LoomingFromInventory("sku", 4, 0, 1, resources_control.ExpectedNewItem{}). // Banner 1 (5)
			LoomingFromInventory("sku", 4, 0, 1, resources_control.ExpectedNewItem{}). // Banner 1 (6)
			LoomingFromInventory("cbo", 5, 6, 3, resources_control.ExpectedNewItem{}). // Banner 2 (1)
			LoomingFromInventory("bo", 0, 6, 1, resources_control.ExpectedNewItem{}).  // Banner 2 (2)
			LoomingFromInventory("moj", 2, 6, 7, resources_control.ExpectedNewItem{}). // Banner 2 (3)
			LoomingFromInventory("sku", 4, 6, 3, resources_control.ExpectedNewItem{}). // Banner 2 (4)
			LoomingFromInventory("cbo", 5, 6, 1, resources_control.ExpectedNewItem{}). // Banner 2 (5)
			LoomingFromInventory("bo", 0, 6, 1, resources_control.ExpectedNewItem{}).  // Banner 2 (6)
			Commit()
		if !success {
			panic("SystemTestingItemStackOperation: Failed on test round 4")
		}

		err = api.ContainerOpenAndClose().CloseContainer()
		if err != nil {
			panic(fmt.Sprintf("SystemTestingItemStackOperation: Test round 4 failed due to %v (stage 3)", err))
		}

		success, err = api.ContainerOpenAndClose().OpenInventory()
		if err != nil {
			panic(fmt.Sprintf("SystemTestingItemStackOperation: Test round 4 failed due to %v (stage 4)", err))
		}
		if !success {
			panic("SystemTestingItemStackOperation: Failed on test round 4")
		}

		success, _, _, _ = api.ItemStackOperation().OpenTransaction().
			MoveToCraftingTable(6, 28, 1).
			MoveToCraftingTable(8, 29, 1).
			Crafting(2418, 10, 1, resources_control.ExpectedNewItem{}).
			MoveToCraftingTable(9, 28, 1). // Hacking attempt
			MoveToCraftingTable(0, 31, 1). // Hacking attempt
			Crafting(2418, 0, 1, resources_control.ExpectedNewItem{}).
			MoveBetweenInventory(10, 8, 1).
			DropInventoryItem(8, 1).
			DropInventoryItem(0, 1).
			Commit()
		if !success {
			panic("SystemTestingItemStackOperation: Failed on test round 4")
		}

		err = api.ContainerOpenAndClose().CloseContainer()
		if err != nil {
			panic(fmt.Sprintf("SystemTestingItemStackOperation: Test round 4 failed due to %v (stage 5)", err))
		}
	}

	// Test round 5
	{
		api.Commands().SendSettingsCommand("clear", true)
		api.Commands().AwaitChangesGeneral()
		api.Commands().SendSettingsCommand("give @s enchanted_golden_apple 10", true)
		api.Commands().AwaitChangesGeneral()

		success, err := api.ContainerOpenAndClose().OpenInventory()
		if err != nil {
			panic(fmt.Sprintf("SystemTestingItemStackOperation: Test round 5 failed due to %v (stage 1)", err))
		}
		if !success {
			panic("SystemTestingItemStackOperation: Failed on test round 5")
		}

		_, _, _, err = api.ItemStackOperation().OpenTransaction().
			MoveBetweenInventory(0, 1, 5).
			MoveBetweenInventory(1, 0, 13).
			Commit()
		if err == nil {
			panic("SystemTestingItemStackOperation: Failed on test round 5")
		}
		if !strings.Contains(fmt.Sprintf("%v", err), "(origin count = 5, delta = -13, result count = -8)") {
			panic("SystemTestingItemStackOperation: Failed on test round 5")
		}

		err = api.ContainerOpenAndClose().CloseContainer()
		if err != nil {
			panic(fmt.Sprintf("SystemTestingItemStackOperation: Test round 5 failed due to %v (stage 2)", err))
		}
	}

	// Test round 6
	{
		api.Commands().SendSettingsCommand("clear", true)
		api.Commands().AwaitChangesGeneral()
		api.Commands().SendSettingsCommand("give @s enchanted_golden_apple 10", true)
		api.Commands().AwaitChangesGeneral()

		states, err := api.SetBlock().SetAnvil(protocol.BlockPos{0, 0, 0}, true)
		if err != nil {
			panic(fmt.Sprintf("SystemTestingItemStackOperation: Test round 6 failed due to %v (stage 1)", err))
		}

		success, err := api.ContainerOpenAndClose().OpenContainer(
			game_interface.UseItemOnBlocks{
				HotbarSlotID: 2,
				BotPos:       mgl32.Vec3{0, 0, 0},
				BlockPos:     [3]int32{0, 0, 0},
				BlockName:    "minecraft:anvil",
				BlockStates:  states,
			},
			false,
		)
		if err != nil {
			panic(fmt.Sprintf("SystemTestingItemStackOperation: Test round 6 failed due to %v (stage 2)", err))
		}
		if !success {
			panic("SystemTestingItemStackOperation: Failed on test round 6")
		}

		success, _, _, err = api.ItemStackOperation().OpenTransaction().
			RenameInventoryItem(0, "sb").
			DropInventoryItem(0, 10).
			Commit()
		if err != nil {
			panic(fmt.Sprintf("SystemTestingItemStackOperation: Test round 6 failed due to %v (stage 3)", err))
		}
		if !success {
			panic("SystemTestingItemStackOperation: Failed on test round 6")
		}

		err = api.ContainerOpenAndClose().CloseContainer()
		if err != nil {
			panic(fmt.Sprintf("SystemTestingItemStackOperation: Test round 6 failed due to %v (stage 4)", err))
		}
	}

	// Test round 7
	{
		api.Commands().SendSettingsCommand("clear", true)
		api.Commands().AwaitChangesGeneral()
		api.Commands().SendSettingsCommand("give @s diamond_helmet 3", true) // Slot 0, 1, 2
		api.Commands().AwaitChangesGeneral()
		api.Commands().SendSettingsCommand("give @s diamond 4", true) // Slot 3
		api.Commands().AwaitChangesGeneral()
		api.Commands().SendSettingsCommand("give @s iron_ingot 2", true) // Slot 4
		api.Commands().AwaitChangesGeneral()
		api.Commands().SendSettingsCommand("give @s tide_armor_trim_smithing_template 4", true) // Slot 5
		api.Commands().AwaitChangesGeneral()
		api.Commands().SendSettingsCommand("give @s eye_armor_trim_smithing_template 2", true) // Slot 6
		api.Commands().AwaitChangesGeneral()

		err := api.SetBlock().SetBlock(protocol.BlockPos{0, 0, 0}, "minecraft:smithing_table", `[]`)
		if err != nil {
			panic(fmt.Sprintf("SystemTestingItemStackOperation: Test round 7 failed due to %v (stage 1)", err))
		}

		success, err := api.ContainerOpenAndClose().OpenContainer(
			game_interface.UseItemOnBlocks{
				HotbarSlotID: 2,
				BotPos:       mgl32.Vec3{0, 0, 0},
				BlockPos:     [3]int32{0, 0, 0},
				BlockName:    "minecraft:smithing_table",
				BlockStates:  map[string]any{},
			},
			false,
		)
		if err != nil {
			panic(fmt.Sprintf("SystemTestingItemStackOperation: Test round 7 failed due to %v (stage 2)", err))
		}
		if !success {
			panic("SystemTestingItemStackOperation: Failed on test round 7")
		}

		success, _, _, err = api.ItemStackOperation().OpenTransaction().
			TrimmingFromInventory(0, 3, 5, resources_control.ExpectedNewItem{}).
			TrimmingFromInventory(1, 3, 5, resources_control.ExpectedNewItem{}).
			TrimmingFromInventory(2, 3, 5, resources_control.ExpectedNewItem{}).
			TrimmingFromInventory(0, 4, 5, resources_control.ExpectedNewItem{}).
			TrimmingFromInventory(0, 3, 6, resources_control.ExpectedNewItem{}).
			TrimmingFromInventory(1, 4, 6, resources_control.ExpectedNewItem{}).
			DropInventoryItem(0, 1).
			DropInventoryItem(1, 1).
			DropInventoryItem(2, 1).
			Commit()
		if err != nil {
			panic(fmt.Sprintf("SystemTestingItemStackOperation: Test round 7 failed due to %v (stage 3)", err))
		}
		if !success {
			panic("SystemTestingItemStackOperation: Failed on test round 7")
		}

		err = api.ContainerOpenAndClose().CloseContainer()
		if err != nil {
			panic(fmt.Sprintf("SystemTestingItemStackOperation: Test round 7 failed due to %v (stage 4)", err))
		}
	}

	pterm.Success.Printfln("SystemTestingItemStackOperation: PASS (Time used = %v)", time.Since(tA))
}
