package item_stack_operation

import (
	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol"
	"github.com/OmineDev/flowers-for-machines/game_control/resources_control"
)

// Trimming 指示锻造台纹饰操作
type Trimming struct {
	TrimItem   resources_control.SlotLocation
	Material   resources_control.SlotLocation
	Template   resources_control.SlotLocation
	ResultItem resources_control.ExpectedNewItem
}

func (Trimming) ID() uint8 {
	return IDItemStackOperationHighLevelTrimming
}

func (Trimming) CanInline() bool {
	return false
}

func (t Trimming) Make(runtimeData MakingRuntime) []protocol.StackRequestAction {
	data := runtimeData.(TrimmingRuntime)

	moveTrimItem := protocol.PlaceStackRequestAction{}
	moveMaterial := protocol.PlaceStackRequestAction{}
	moveTemplate := protocol.PlaceStackRequestAction{}
	moveResult := protocol.TakeStackRequestAction{}

	moveTrimItem.Count = 1
	moveTrimItem.Source = protocol.StackRequestSlotInfo{
		ContainerID:    data.MoveTrimItemSrcContainerID,
		Slot:           byte(t.TrimItem.SlotID),
		StackNetworkID: data.MoveTrimItemSrcStackNetworkID,
	}
	moveTrimItem.Destination = protocol.StackRequestSlotInfo{
		ContainerID:    protocol.ContainerSmithingTableInput,
		Slot:           0x33,
		StackNetworkID: data.TrimItemStackNetworkID,
	}

	moveMaterial.Count = 1
	moveMaterial.Source = protocol.StackRequestSlotInfo{
		ContainerID:    data.MoveMaterialSrcContainerID,
		Slot:           byte(t.Material.SlotID),
		StackNetworkID: data.MoveMaterialSrcStackNetworkID,
	}
	moveMaterial.Destination = protocol.StackRequestSlotInfo{
		ContainerID:    protocol.ContainerSmithingTableMaterial,
		Slot:           0x34,
		StackNetworkID: data.MaterialStackNetworkID,
	}

	moveTemplate.Count = 1
	moveTemplate.Source = protocol.StackRequestSlotInfo{
		ContainerID:    data.MoveTemplateSrcContainerID,
		Slot:           byte(t.Template.SlotID),
		StackNetworkID: data.MoveTemplateSrcStackNetworkID,
	}
	moveTemplate.Destination = protocol.StackRequestSlotInfo{
		ContainerID:    protocol.ContainerSmithingTableTemplate,
		Slot:           0x35,
		StackNetworkID: data.TemplateStackNetworkID,
	}

	moveResult.Count = 1
	moveResult.Source = protocol.StackRequestSlotInfo{
		ContainerID:    protocol.ContainerCreatedOutput, // [NEMC 1.20.10] 60 -> 61 (Added by Happy2018new)
		Slot:           0x32,
		StackNetworkID: data.RequestID,
	}
	moveResult.Destination = protocol.StackRequestSlotInfo{
		ContainerID:    data.MoveTrimItemSrcContainerID,
		Slot:           byte(t.TrimItem.SlotID),
		StackNetworkID: data.RequestID,
	}

	return []protocol.StackRequestAction{
		&moveTrimItem,
		&moveMaterial,
		&moveTemplate,
		&protocol.CraftRecipeStackRequestAction{
			RecipeNetworkID: data.RecipeNetworkID,
		},
		&protocol.ConsumeStackRequestAction{
			DestroyStackRequestAction: protocol.DestroyStackRequestAction{
				Count: 1,
				Source: protocol.StackRequestSlotInfo{
					ContainerID:    protocol.ContainerSmithingTableInput,
					Slot:           0x33,
					StackNetworkID: data.RequestID,
				},
			},
		},
		&protocol.ConsumeStackRequestAction{
			DestroyStackRequestAction: protocol.DestroyStackRequestAction{
				Count: 1,
				Source: protocol.StackRequestSlotInfo{
					ContainerID:    protocol.ContainerSmithingTableMaterial,
					Slot:           0x34,
					StackNetworkID: data.RequestID,
				},
			},
		},
		&protocol.ConsumeStackRequestAction{
			DestroyStackRequestAction: protocol.DestroyStackRequestAction{
				Count: 1,
				Source: protocol.StackRequestSlotInfo{
					ContainerID:    protocol.ContainerSmithingTableTemplate,
					Slot:           0x35,
					StackNetworkID: data.RequestID,
				},
			},
		},
		&moveResult,
	}
}
