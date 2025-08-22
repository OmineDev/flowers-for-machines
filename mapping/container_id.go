package mapping

import (
	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol"
)

const (
	// ContainerIDCanNotOpen 是一个特殊值，用于描述目标容器无法被打开，
	// 因此尝试描述该容器的容器 ID 似乎是没有意义的
	ContainerIDCanNotOpen = 254
	// ContainerIDUnknown 指示该容器的容器 ID 仍然是未知的，
	// 需要通过测试才能了解实际情况
	ContainerIDUnknown = 255
)

// ContainerTypeWithSlot 用于描述一个容器 ID 对应的槽位表示。
// 大多数容器的 ID 只取决于这个容器的类型，
// 但部分情况下，部分容器的部分槽位的容器 ID 是不同的
type ContainerTypeWithSlot struct {
	// ContainerType 是这个容器的类型
	ContainerType int
	// 该容器的容器 ID 可能与槽位索引有关。
	// 如果是无关的，则 SlotID 应当置为 0。
	//
	// 需要说明的是，置为 0 不代表是无关的，
	// 但无关需要置为 0。
	//
	// 你可通过 ContainerNeedSlotIDMapping
	// 来查找具体存在哪些容器具有这样的特性，
	// 然后再根据得到的结果将此字段 (SlotID)
	// 置为 0
	SlotID uint8
}

// ContainerNeedSlotIDMapping 记载了容器 ID 同时取决于 ContainerType 和 SlotID 的容器。
// 该映射是没有彻底完成的，这意味着仍然存在部分方块满足上面的叙述，但没有出现在下表中。
// 因此，修补该表并使得其完整仍然是一个正在进行的议题
var ContainerNeedSlotIDMapping = map[int]bool{
	protocol.ContainerTypeFurnace:       true,
	protocol.ContainerTypeBrewingStand:  true,
	protocol.ContainerTypeAnvil:         true,
	protocol.ContainerTypeLoom:          true,
	protocol.ContainerTypeBlastFurnace:  true,
	protocol.ContainerTypeSmoker:        true,
	protocol.ContainerTypeSmithingTable: true,
}

// ContainerIDMapping 保存了一个 ContainerTypeWithSlot 到容器 ID 的映射。
// 该映射是没有彻底完成的，部分字段仍然需要通过大量测试才能将该表修补完整
var ContainerIDMapping = map[ContainerTypeWithSlot]uint8{
	{ContainerType: protocol.ContainerTypeInventory}: protocol.ContainerCombinedHotBarAndInventory, // 12 (inventory)
	{ContainerType: protocol.ContainerTypeContainer}: protocol.ContainerLevelEntity,                // 7 (chest, trapped_chest)
	{ContainerType: protocol.ContainerTypeDispenser}: protocol.ContainerLevelEntity,                // 7 (dispenser)
	{ContainerType: protocol.ContainerTypeDropper}:   protocol.ContainerLevelEntity,                // 7 (dropper)
	{ContainerType: protocol.ContainerTypeHopper}:    protocol.ContainerLevelEntity,                // 7 (hopper)
	{ContainerType: protocol.ContainerTypeCrafter}:   protocol.ContainerLevelEntity,                // 7 (crafter)
	{ContainerType: protocol.ContainerTypeArmour}:    protocol.ContainerArmor,                      // 6 (armour)

	// furnac, lit_furnace
	{ContainerType: protocol.ContainerTypeFurnace, SlotID: 0}: protocol.ContainerFurnaceIngredient, // 26
	{ContainerType: protocol.ContainerTypeFurnace, SlotID: 1}: protocol.ContainerFurnaceFuel,       // 25
	{ContainerType: protocol.ContainerTypeFurnace, SlotID: 2}: protocol.ContainerFurnaceResult,     // 27

	// brewing_stand
	{ContainerType: protocol.ContainerTypeBrewingStand, SlotID: 0}: protocol.ContainerBrewingStandInput,  // 9
	{ContainerType: protocol.ContainerTypeBrewingStand, SlotID: 1}: protocol.ContainerBrewingStandResult, // 10
	{ContainerType: protocol.ContainerTypeBrewingStand, SlotID: 2}: protocol.ContainerBrewingStandResult, // 10
	{ContainerType: protocol.ContainerTypeBrewingStand, SlotID: 3}: protocol.ContainerBrewingStandResult, // 10
	{ContainerType: protocol.ContainerTypeBrewingStand, SlotID: 4}: protocol.ContainerBrewingStandFuel,   // 11

	// anvil
	{ContainerType: protocol.ContainerTypeAnvil, SlotID: 0}: ContainerIDUnknown,           // Need do research
	{ContainerType: protocol.ContainerTypeAnvil, SlotID: 1}: protocol.ContainerAnvilInput, // 0
	{ContainerType: protocol.ContainerTypeAnvil, SlotID: 2}: ContainerIDUnknown,           // Need do research

	// loom
	{ContainerType: protocol.ContainerTypeLoom, SlotID: 9}:  protocol.ContainerLoomInput,    // 42
	{ContainerType: protocol.ContainerTypeLoom, SlotID: 10}: protocol.ContainerLoomDye,      // 43
	{ContainerType: protocol.ContainerTypeLoom, SlotID: 11}: protocol.ContainerLoomMaterial, // 44

	// smithing table
	{ContainerType: protocol.ContainerTypeSmithingTable, SlotID: 0x33}: protocol.ContainerSmithingTableInput,    // 3
	{ContainerType: protocol.ContainerTypeSmithingTable, SlotID: 0x34}: protocol.ContainerSmithingTableMaterial, // 4
	{ContainerType: protocol.ContainerTypeSmithingTable, SlotID: 0x35}: protocol.ContainerSmithingTableTemplate, // 62

	// blast_furnace (lit_blast_furnace)
	{ContainerType: protocol.ContainerTypeBlastFurnace, SlotID: 0}: protocol.ContainerBlastFurnaceIngredient, // 46
	{ContainerType: protocol.ContainerTypeBlastFurnace, SlotID: 1}: protocol.ContainerFurnaceFuel,            // 25
	{ContainerType: protocol.ContainerTypeBlastFurnace, SlotID: 2}: protocol.ContainerFurnaceResult,          // 27

	// smoker (lit_smoker)
	{ContainerType: protocol.ContainerTypeSmoker, SlotID: 0}: protocol.ContainerSmokerIngredient, // 47
	{ContainerType: protocol.ContainerTypeSmoker, SlotID: 1}: protocol.ContainerFurnaceFuel,      // 25
	{ContainerType: protocol.ContainerTypeSmoker, SlotID: 2}: protocol.ContainerFurnaceResult,    // 27

	// TODO: Guess, and need to test
	{ContainerType: protocol.ContainerTypeHorse}:    protocol.ContainerHorseEquip,    // 28 (horse)
	{ContainerType: protocol.ContainerTypeBeacon}:   protocol.ContainerBeaconPayment, // 8 (beacon)
	{ContainerType: protocol.ContainerTypeHand}:     protocol.ContainerOffhand,       // 35 (offhand)
	{ContainerType: protocol.ContainerTypeLabTable}: protocol.ContainerLabTableInput, // 41 (lab_table)

	// TODO: Need to do further research to figure out what they are
	{ContainerType: protocol.ContainerTypeWorkbench}:          ContainerIDUnknown,
	{ContainerType: protocol.ContainerTypeEnchantment}:        ContainerIDUnknown,
	{ContainerType: protocol.ContainerTypeCartChest}:          ContainerIDUnknown,
	{ContainerType: protocol.ContainerTypeCartHopper}:         ContainerIDUnknown,
	{ContainerType: protocol.ContainerTypeStructureEditor}:    ContainerIDUnknown,
	{ContainerType: protocol.ContainerTypeTrade}:              ContainerIDUnknown,
	{ContainerType: protocol.ContainerTypeCompoundCreator}:    ContainerIDUnknown,
	{ContainerType: protocol.ContainerTypeElementConstructor}: ContainerIDUnknown,
	{ContainerType: protocol.ContainerTypeMaterialReducer}:    ContainerIDUnknown,
	{ContainerType: protocol.ContainerTypeGrindstone}:         ContainerIDUnknown,
	{ContainerType: protocol.ContainerTypeStonecutter}:        ContainerIDUnknown,
	{ContainerType: protocol.ContainerTypeCartography}:        ContainerIDUnknown,
	{ContainerType: protocol.ContainerTypeJigsawEditor}:       ContainerIDUnknown,
	{ContainerType: protocol.ContainerTypeChestBoat}:          ContainerIDUnknown,

	// The following container can't be opened
	{ContainerType: protocol.ContainerTypeCauldron}:     ContainerIDCanNotOpen,
	{ContainerType: protocol.ContainerTypeCommandBlock}: ContainerIDCanNotOpen,
	{ContainerType: protocol.ContainerTypeJukebox}:      ContainerIDCanNotOpen,
	{ContainerType: protocol.ContainerTypeLectern}:      ContainerIDCanNotOpen,
	{ContainerType: protocol.ContainerTypeHUD}:          ContainerIDCanNotOpen,
	{ContainerType: protocol.ContainerTypeDecoratedPot}: ContainerIDCanNotOpen,
}
