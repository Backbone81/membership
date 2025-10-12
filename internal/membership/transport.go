package membership

// Transport is the interface the transport needs to implement for transmitting data between members.
type Transport interface {
	Send(address Address, buffer []byte) error
}
