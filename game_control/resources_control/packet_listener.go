package resources_control

import (
	"context"
	"fmt"
	"sync"

	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol/packet"

	"github.com/google/uuid"
)

// singleListener 是单个数据包的监听器
type singleListener struct {
	// 该监听器的唯一标识符
	uniqueID string
	// 该监听器的回调函数
	callback func(
		p packet.Packet,
		connCloseErr error,
	)
}

// PacketListener 实现了一个可撤销监听的，
// 相对基础的数据包监听器
type PacketListener struct {
	mu                      *sync.Mutex
	ctx                     context.Context
	anyPacketListeners      []singleListener
	specificPacketListeners map[uint32][]singleListener
}

// NewPacketListener 基于 ctx 创建并返回一个新的 NewPacketListener
func NewPacketListener(ctx context.Context) *PacketListener {
	return &PacketListener{
		mu:                      new(sync.Mutex),
		ctx:                     ctx,
		anyPacketListeners:      nil,
		specificPacketListeners: make(map[uint32][]singleListener),
	}
}

// ListenPacket 监听数据包 ID 在 packetID 中的数据包，
// 并在收到这些数据包后执行回调函数 callback。
//
// 特别地，如果底层 Raknet 连接关闭，
// 则传入 callback 的 connCloseErr 不为 nil。
//
// 如果 packetID 置空，则监听所有数据包。
//
// 返回的 uniqueID 用于标识该监听器，以便于
// 后续调用 DestroyListener 以手动销毁监听器
func (p *PacketListener) ListenPacket(
	packetID []uint32,
	callback func(p packet.Packet, connCloseErr error),
) (uniqueID string, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	select {
	case <-p.ctx.Done():
		return "", fmt.Errorf("ListenPacket: Listen packet on closed connection")
	default:
	}

	uniqueID = uuid.NewString()
	listener := singleListener{
		uniqueID: uniqueID,
		callback: callback,
	}

	if len(packetID) == 0 {
		p.anyPacketListeners = append(p.anyPacketListeners, listener)
		return
	}

	for _, pkID := range packetID {
		if p.specificPacketListeners[pkID] == nil {
			p.specificPacketListeners[pkID] = make([]singleListener, 0)
		}
		p.specificPacketListeners[pkID] = append(p.specificPacketListeners[pkID], listener)
	}
	return
}

// DestroyListener 销毁唯一标识为 uniqueID 的数据包监听器。
// 如果这样的监听器不存在，则不会执行任何操作
func (p *PacketListener) DestroyListener(uniqueID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	select {
	case <-p.ctx.Done():
		return
	default:
	}

	// Any packet listener
	{
		found := false
		listenerIndex := 0

		for index, listener := range p.anyPacketListeners {
			if listener.uniqueID == uniqueID {
				found = true
				listenerIndex = index
				break
			}
		}

		if found {
			newListeners := make([]singleListener, 0)

			for index, listener := range p.anyPacketListeners {
				if index == listenerIndex {
					continue
				}
				newListeners = append(newListeners, listener)
			}

			p.anyPacketListeners = newListeners
			return
		}
	}

	// Specific packet listener
	for packetID, listeners := range p.specificPacketListeners {
		found := false
		listenerIndex := 0

		for index, listener := range listeners {
			if listener.uniqueID == uniqueID {
				found = true
				listenerIndex = index
				break
			}
		}

		if found {
			newListeners := make([]singleListener, 0)

			for index, listener := range listeners {
				if index == listenerIndex {
					continue
				}
				newListeners = append(newListeners, listener)
			}

			if len(newListeners) == 0 {
				delete(p.specificPacketListeners, packetID)
			} else {
				p.specificPacketListeners[packetID] = newListeners
			}

			return
		}
	}
}

// onPacket ..
func (p *PacketListener) onPacket(pk packet.Packet) {
	p.mu.Lock()
	defer p.mu.Unlock()

	select {
	case <-p.ctx.Done():
		return
	default:
	}

	// Any packet listener
	for _, listener := range p.anyPacketListeners {
		go listener.callback(pk, nil)
	}

	// Specific packet listener
	for _, listener := range p.specificPacketListeners[pk.ID()] {
		go listener.callback(pk, nil)
	}
}

// handleConnClose ..
func (p *PacketListener) handleConnClose(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Any packet listener
	for _, listener := range p.anyPacketListeners {
		go listener.callback(nil, err)
	}

	// Specific packet listener
	for packetID, listeners := range p.specificPacketListeners {
		for _, listener := range listeners {
			go listener.callback(nil, err)
		}
		p.specificPacketListeners[packetID] = nil
	}

	p.anyPacketListeners = nil
	p.specificPacketListeners = nil
}
