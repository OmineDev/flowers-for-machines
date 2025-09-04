package resources_control

import (
	"context"
	"fmt"
	"sync"

	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol/packet"
	"github.com/OmineDev/flowers-for-machines/mapping"
)

const (
	ContainerStatesHaveNotOpen uint8 = iota
	ContainerStatesOpening
	ContainerStatesClosed
	ContainerStatesConnClosed
)

// ContainerManager 描述一个在内存中维护的容器实现，
// 它用于追踪和监控目前已打开容器的状态
type ContainerManager struct {
	mu     *sync.Mutex
	ctx    context.Context
	states uint8

	openingData  *packet.ContainerOpen
	containerID  ContainerID
	openCallback func(connCloseErr error)

	closingData   *packet.ContainerClose
	closeCallback func(isServerSide bool, connCloseErr error)
}

// NewContainerManager 基于 ctx 创建并返回一个新的容器管理器
func NewContainerManager(ctx context.Context) *ContainerManager {
	return &ContainerManager{
		mu:            new(sync.Mutex),
		ctx:           ctx,
		states:        ContainerStatesHaveNotOpen,
		openingData:   nil,
		containerID:   mapping.ContainerIDUnknown,
		openCallback:  nil,
		closingData:   nil,
		closeCallback: nil,
	}
}

// States 返回已打开容器的状态。
// 目前只存在 3 种状态：
//   - 0: 曾经没有打开过容器
//   - 1: 目前存在一个已被打开的容器
//   - 2: 曾经打开过容器，但是关闭了
func (c ContainerManager) States() uint8 {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-c.ctx.Done():
		return ContainerStatesConnClosed
	default:
		return c.states
	}
}

// ContainerData 获取当前已打开容器的状态。
// 返回的 existed 指示当前是否已经打开了容器。
//
// containerID 是提前预设的，这意味着其值如果
// 不是 mapping.ContainerIDUnknown 则应当优
// 先使用
func (c ContainerManager) ContainerData() (data packet.ContainerOpen, containerID ContainerID, existed bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-c.ctx.Done():
		return packet.ContainerOpen{}, 0, false
	default:
	}

	if c.states != ContainerStatesOpening {
		return packet.ContainerOpen{}, 0, false
	}
	return *c.openingData, c.containerID, true
}

// SetContainerOpenCallback 设置容器打开时应该执行的回调函数。
// 另外，设置的回调函数会在其被执行后被移除。
//
// 特别地，如果底层 Raknet 连接关闭，
// 则传入 f 的 connCloseErr 不为 nil。
//
// containerID 是提前预设的将要打开容器的容器 ID，
// 通常情况下可以安全的置为 mapping.ContainerIDUnknown (255)。
//
// 只有部分容器需要提前预设，目前已知的包含木桶
func (c *ContainerManager) SetContainerOpenCallback(
	containerID ContainerID,
	f func(connCloseErr error),
) {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-c.ctx.Done():
		go f(fmt.Errorf("SetContainerOpenCallback: Set callback on closed connection"))
	default:
		c.containerID = containerID
		c.openCallback = f
	}
}

// SetContainerCloseCallback 设置容器关闭时应该执行的回调函数。
// isServerSide 指示容器是否是由服务器强制关闭的。
//
// 特别地，如果底层 Raknet 连接关闭，
// 则传入 f 的 connCloseErr 不为 nil。
//
// 另外，设置的回调函数会在其被执行后被移除
func (c *ContainerManager) SetContainerCloseCallback(
	f func(isServerSide bool, connCloseErr error),
) {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-c.ctx.Done():
		go f(false, fmt.Errorf("SetContainerCloseCallback: Set callback on closed connection"))
	default:
		c.closeCallback = f
	}
}

// onContainerOpen ..
func (c *ContainerManager) onContainerOpen(p *packet.ContainerOpen) {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-c.ctx.Done():
		return
	default:
	}

	c.openingData = p
	c.closingData = nil
	c.states = ContainerStatesOpening

	if c.openCallback != nil {
		go c.openCallback(nil)
		c.openCallback = nil
	}
}

// ContainerClose ..
func (c *ContainerManager) onContainerClose(p *packet.ContainerClose) {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-c.ctx.Done():
		return
	default:
	}

	c.closingData = p
	c.openingData = nil
	c.containerID = mapping.ContainerIDUnknown
	c.states = ContainerStatesClosed

	if c.closeCallback != nil {
		go c.closeCallback(p.ServerSide, nil)
		c.closeCallback = nil
	}
}

// handleConnClose ..
func (c *ContainerManager) handleConnClose(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.openCallback != nil {
		go c.openCallback(err)
	}
	if c.closeCallback != nil {
		go c.closeCallback(false, err)
	}

	c.states = ContainerStatesConnClosed
	c.openingData = nil
	c.containerID = mapping.ContainerIDUnknown
	c.openCallback = nil
	c.closingData = nil
	c.closeCallback = nil
}
