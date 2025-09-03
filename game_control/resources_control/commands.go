package resources_control

import (
	"context"
	"fmt"
	"sync"

	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol/packet"

	"github.com/google/uuid"
)

// CommandRequestCallback 是简单的指令回调维护器
type CommandRequestCallback struct {
	mu       *sync.Mutex
	ctx      context.Context
	callback map[uuid.UUID]func(p *packet.CommandOutput, connCloseErr error)
}

// NewCommandRequestCallback 根据 ctx 创建并返回一个新的 CommandRequestCallback
func NewCommandRequestCallback(ctx context.Context) *CommandRequestCallback {
	return &CommandRequestCallback{
		mu:       new(sync.Mutex),
		ctx:      ctx,
		callback: make(map[uuid.UUID]func(p *packet.CommandOutput, connCloseErr error)),
	}
}

// SetCommandRequestCallback 设置当收到请求 ID 为 requestID 的命令请求的响应后，
// 应当执行的回调函数 f。其中，p 指示服务器发送的针对此命令请求的响应体。
// 特别地，如果底层 Raknet 连接关闭，则传入 f 的 connCloseErr 不为 nil
func (c *CommandRequestCallback) SetCommandRequestCallback(
	requestID uuid.UUID,
	f func(p *packet.CommandOutput, connCloseErr error),
) {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-c.ctx.Done():
		go f(nil, fmt.Errorf("SetCommandRequestCallback: Set callback on closed connection"))
	default:
		c.callback[requestID] = f
	}
}

// DeleteCommandRequestCallback 清除请求
// ID 为 requestID 的命令请求的回调函数。
// 此函数应当只在命令请求超时的时候被调用
func (c *CommandRequestCallback) DeleteCommandRequestCallback(requestID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-c.ctx.Done():
		return
	default:
		delete(c.callback, requestID)
	}
}

// onCommandOutput ..
func (c *CommandRequestCallback) onCommandOutput(p *packet.CommandOutput) {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-c.ctx.Done():
		return
	default:
		cb, ok := c.callback[p.CommandOrigin.UUID]
		if ok {
			go cb(p, nil)
		}
	}
}

// handleConnClose ..
func (c *CommandRequestCallback) handleConnClose(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for requestID, cb := range c.callback {
		go cb(nil, err)
		c.callback[requestID] = nil
	}

	c.callback = nil
}
