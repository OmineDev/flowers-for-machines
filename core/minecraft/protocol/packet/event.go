package packet

import (
	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol"
)

// Event is sent by the server to send an event with additional data. It is typically sent to the client for
// telemetry reasons, much like the SimpleEvent packet.
type Event struct {
	// EntityRuntimeID is the runtime ID of the player. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	EntityRuntimeID uint64
	// UsePlayerID ...
	// TODO: Figure out what UsePlayerID is for.
	UsePlayerID byte
	// Event is the event that is transmitted.
	Event protocol.Event
}

// ID ...
func (*Event) ID() uint32 {
	return IDEvent
}

func (pk *Event) Marshal(io protocol.IO) {
	io.Varuint64(&pk.EntityRuntimeID)
	io.EventType(&pk.Event)
	io.Uint8(&pk.UsePlayerID)
	pk.Event.Marshal(io)
}
