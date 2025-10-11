package membership

// Message is the interface all network messages need to implement. This is needed to allow the gossip queue to work
// with those messages.
type Message interface {
	AppendToBuffer(buffer []byte) ([]byte, int, error)
	FromBuffer(buffer []byte) (int, error)
}
