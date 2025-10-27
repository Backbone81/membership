package gossip

import "github.com/backbone81/membership/internal/encoding"

// Message is the interface all gossip network messages need to implement.
type Message interface {
	AppendToBuffer(buffer []byte) ([]byte, int, error)
	FromBuffer(buffer []byte) (int, error)
	GetAddress() encoding.Address
	GetType() encoding.MessageType
	GetIncarnationNumber() int
}
