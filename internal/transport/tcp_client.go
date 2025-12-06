package transport

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
	"math"
	"net"

	"github.com/backbone81/membership/internal/encoding"
	"github.com/backbone81/membership/internal/encryption"
)

// TCPClient provides reliable transport for sending data to a member.
//
// TCPClient is not safe for concurrent use by multiple goroutines. Callers must serialize access to all methods.
// As this client is always called under the lock of the membership.List we have that serialization there.
type TCPClient struct {
	gcm        cipher.AEAD
	ciphertext []byte
}

// TCPClient implements Transport.
var _ Transport = (*TCPClient)(nil)

// NewTCPClient creates a new TCPClient transport.
func NewTCPClient(key encryption.Key) (*TCPClient, error) {
	aesCipher, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCMWithRandomNonce(aesCipher)
	if err != nil {
		return nil, err
	}

	return &TCPClient{
		gcm:        gcm,
		ciphertext: make([]byte, 0, 1024),
	}, nil
}

// Send transmits the given buffer to the member with the given address.
func (c *TCPClient) Send(address encoding.Address, buffer []byte) error {
	if err := c.send(address, buffer); err != nil {
		return fmt.Errorf("TCP client transport send: %w", err)
	}
	return nil
}

func (c *TCPClient) send(address encoding.Address, plaintext []byte) error {
	// Make sure we are not exceeding the maximum datagram length with the given buffer.
	if len(plaintext) > math.MaxUint32 {
		return errors.New("buffer length exceeds maximum datagram length")
	}

	var lengthBuffer [4]byte
	encoding.Endian.PutUint32(lengthBuffer[:], uint32(len(plaintext))) //nolint:gosec // we already checked before
	c.ciphertext = c.gcm.Seal(c.ciphertext[:0], nil, lengthBuffer[:], nil)
	c.ciphertext = c.gcm.Seal(c.ciphertext, nil, plaintext, nil)
	Encryptions.WithLabelValues("tcp_client").Add(2)

	connection, err := net.Dial("tcp", address.String())
	if err != nil {
		return fmt.Errorf("connecting to remote host at %q: %w", address, err)
	}
	defer connection.Close() //nolint:errcheck

	n, err := connection.Write(c.ciphertext)
	TransmitBytes.WithLabelValues("tcp_client").Add(float64(n))
	if err != nil {
		TransmitErrors.WithLabelValues("tcp_client").Inc()
		return fmt.Errorf("sending the datagram payload: %w", err)
	}
	return nil
}
