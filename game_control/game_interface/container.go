package game_interface

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol"
	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol/packet"
	"github.com/OmineDev/flowers-for-machines/game_control/resources_control"
	"github.com/OmineDev/flowers-for-machines/mapping"
)

const (
	// 描述容器打开的最长截止时间。
	// 当超过此时间后，将不再等待
	DefaultTimeoutContainerOpen = time.Second / 20 * 3
	// 描述容器打开失败后要重试的最大次数
	MaxRetryContainerOpen = 35
)

// ContainerOpenAndClose 是基于 Resources 实现的容器打开和关闭控制系统
type ContainerOpenAndClose struct {
	api      *ResourcesWrapper
	commands *Commands
	botClick *BotClick
	mu       *sync.Mutex
	occupy   *sync.Mutex
}

// NewContainerOpenAndClose 基于 api、commands 和
// botClick 创建并返回一个新的 ContainerOpenAndClose
func NewContainerOpenAndClose(
	api *ResourcesWrapper,
	commands *Commands,
	botClick *BotClick,
) *ContainerOpenAndClose {
	return &ContainerOpenAndClose{
		api:      api,
		commands: commands,
		botClick: botClick,
		mu:       new(sync.Mutex),
		occupy:   new(sync.Mutex),
	}
}

// openContainer ..
func (c *ContainerOpenAndClose) openContainer(
	expectedContainerID resources_control.ContainerID,
	changeToTargetSlot func() error,
	openFunc func() error,
) (success bool, err error) {
	var terminalErr error
	api := c.api.Container()

	for {
		c.occupy.Lock()
		c.mu.Lock()
		if !c.occupy.TryLock() {
			break
		}
		c.mu.Unlock()
	}
	defer func() {
		if !success {
			c.occupy.Unlock()
		}
		c.mu.Unlock()
	}()

	doOnce := new(sync.Once)
	channel := make(chan struct{})
	api.SetContainerOpenCallback(
		expectedContainerID,
		func(connCloseErr error) {
			doOnce.Do(func() {
				if connCloseErr != nil {
					terminalErr = connCloseErr
				} else {
					success = true
				}
				close(channel)
			})
		},
	)
	api.SetContainerCloseCallback(
		func(isServerSide bool, connCloseErr error) {
			c.mu.Lock()
			defer c.mu.Unlock()
			c.occupy.TryLock()
			c.occupy.Unlock()
		},
	)

	if changeToTargetSlot != nil {
		err = changeToTargetSlot()
		if err != nil {
			return false, fmt.Errorf("openContainer: %v", err)
		}
	}

	for range MaxRetryContainerOpen {
		var shouldBreak bool

		err = openFunc()
		if err != nil {
			return false, fmt.Errorf("openContainer: %v", err)
		}

		timer := time.NewTimer(DefaultTimeoutContainerOpen)
		defer timer.Stop()
		select {
		case <-timer.C:
		case <-channel:
			shouldBreak = true
		}

		if shouldBreak {
			break
		}
	}

	if terminalErr != nil {
		return false, fmt.Errorf("openContainer: %v", err)
	}
	return
}

// OpenContainer 打开 container 所指示的容器。
// action 中已自然包含打开容器时所使用的物品栏。
//
// changeToTargetSlot 指示打开容器前是否需要
// 切换快捷栏为 container 中所指示的物品栏。
//
// 通常地，如果您已经保证了快捷栏的位置，那么
// 可以安全的将 changeToTargetSlot 填写为假。
//
// 另外，在打开容器后必须将容器关闭，否则再次调用
// OpenContainer 将会阻塞。
// 欲安全的关闭容器，请使用 CloseContainer。
//
// 可以确保 OpenContainer 在逻辑上是线程安全的
func (c *ContainerOpenAndClose) OpenContainer(
	container UseItemOnBlocks,
	changeToTargetSlot bool,
) (success bool, err error) {
	var prepareFunc func() error
	var expectedContainerID resources_control.ContainerID = mapping.ContainerIDUnknown

	// Special process
	if strings.Contains(container.BlockName, "barrel") {
		expectedContainerID = protocol.ContainerBarrel
	}
	if strings.Contains(container.BlockName, "shulker") {
		expectedContainerID = protocol.ContainerShulkerBox
	}

	openFunc := func() error {
		return c.botClick.ClickBlock(container)
	}
	if changeToTargetSlot {
		prepareFunc = func() error {
			return c.botClick.ChangeSelectedHotbarSlot(container.HotbarSlotID)
		}
	}

	success, err = c.openContainer(expectedContainerID, prepareFunc, openFunc)
	if err != nil {
		return false, fmt.Errorf("OpenContainer: %v", err)
	}
	return
}

// OpenInventory 用于打开背包。
// 可以确保 OpenContainer 在逻辑上是线程安全的
func (c *ContainerOpenAndClose) OpenInventory() (success bool, err error) {
	success, err = c.openContainer(
		mapping.ContainerIDUnknown,
		nil,
		func() error {
			return c.api.WritePacket(&packet.Interact{
				ActionType:            packet.InteractActionOpenInventory,
				TargetEntityRuntimeID: c.api.BotInfo.EntityRuntimeID,
			})
		},
	)
	if err != nil {
		return false, fmt.Errorf("OpenContainer: %v", err)
	}
	return
}

// CloseContainer 关闭已打开的容器，
// 可以确保它在逻辑上是线程安全的
func (c *ContainerOpenAndClose) CloseContainer() error {
	var terminalErr error
	c.mu.Lock()
	defer c.mu.Unlock()

	containerData, _, existed := c.api.Container().ContainerData()
	if !existed {
		return nil
	}
	if c.occupy.TryLock() {
		c.occupy.Unlock()
		return nil
	}

	doOnce := new(sync.Once)
	channel := make(chan struct{})
	c.api.Container().SetContainerCloseCallback(
		func(isServerSide bool, connCloseErr error) {
			doOnce.Do(func() {
				terminalErr = connCloseErr
				close(channel)
			})
		},
	)

	err := c.api.WritePacket(&packet.ContainerClose{WindowID: containerData.WindowID})
	if err != nil {
		return fmt.Errorf("CloseContainer: %v", err)
	}

	<-channel
	c.occupy.Unlock()

	if terminalErr != nil {
		return fmt.Errorf("CloseContainer: %v", terminalErr)
	}
	return nil
}
