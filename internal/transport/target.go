package transport

// Target is the interface which the target needs to implement for processing incoming network messages.
type Target interface {
	DispatchDatagram(buffer []byte) error
}
