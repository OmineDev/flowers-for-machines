package item_stack_operation

import "github.com/OmineDev/flowers-for-machines/core/minecraft/protocol"

const (
	IDItemStackOperationMove uint8 = iota
	IDItemStackOperationSwap
	IDItemStackOperationDrop
	IDItemStackOperationCreativeItem
	IDItemStackOperationHighLevelRenaming
	IDItemStackOperationHighLevelLooming
	IDItemStackOperationHighLevelCrafting
	IDItemStackOperationHighLevelTrimming
)

// ItemStackOperation 指示所有实现了它的物品操作
type ItemStackOperation interface {
	// CanInline 指示该物品操作是否可以内联到单个物品堆栈操作请求中。
	// 如果不能，则该物品操作则应被内联到同一个数据包的另外一个请求中
	CanInline() bool
	// ID 指示该物品操作的编号，它是自定义的
	ID() uint8
	// Make 基于运行时数据 runtiemData，
	// 返回目标物品操作的标准物品堆栈请求的动作
	Make(runtiemData MakingRuntime) []protocol.StackRequestAction
}
