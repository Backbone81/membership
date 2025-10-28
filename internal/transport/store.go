package transport

import "github.com/backbone81/membership/internal/encoding"

// Store provides a client transport which stores all data and always reports success. This is useful for
// tests, when we need to check if some specific data was transmitted.
type Store struct {
	Addresses []encoding.Address
	Buffers   [][]byte
}

// Store implements Transport
var _ Transport = (*Store)(nil)

func (s *Store) Send(address encoding.Address, buffer []byte) error {
	s.Addresses = append(s.Addresses, address)
	s.Buffers = append(s.Buffers, buffer)
	return nil
}
