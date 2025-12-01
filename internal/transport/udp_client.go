package transport

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
	"net"

	"github.com/backbone81/membership/internal/encoding"
	"github.com/backbone81/membership/internal/encryption"
)

// UDPClient provides unreliable transport for sending data to a member.
//
// UDPClient is not safe for concurrent use by multiple goroutines. Callers must serialize access to all methods.
// As this client is always called under the lock of the membership.List we have that serialization there.
type UDPClient struct {
	maxDatagramLength int
	gcm               cipher.AEAD
	ciphertext        []byte
}

// UDPClient implements Transport.
var _ Transport = (*UDPClient)(nil)

// NewUDPClient creates a new UDPClient transport.
func NewUDPClient(maxDatagramLength int, key encryption.Key) (*UDPClient, error) {
	aesCipher, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCMWithRandomNonce(aesCipher)
	if err != nil {
		return nil, err
	}

	return &UDPClient{
		maxDatagramLength: maxDatagramLength,
		gcm:               gcm,
		ciphertext:        make([]byte, 0, maxDatagramLength),
	}, nil
}

// Send transmits the given buffer to the member with the given address. The length of the buffer is validated against
// the max datagram length provided during construction. If the length exceeds the maximum length, no data is sent
// and an error is returned.
func (c *UDPClient) Send(address encoding.Address, buffer []byte) error {
	// Note that we do not encrypt in-place here, because the buffer might grow for encryption. In that case we want to
	// hold onto the bigger buffer instead of dropping it again and allocating a bigger buffer again next time.
	c.ciphertext = c.gcm.Seal(c.ciphertext[:0], nil, buffer, nil)
	Encryptions.WithLabelValues("udp_client").Add(1)
	if err := c.send(address, c.ciphertext); err != nil {
		return fmt.Errorf("UDP client transport send: %w", err)
	}
	return nil
}

func (c *UDPClient) send(address encoding.Address, buffer []byte) error {
	if len(buffer) > c.maxDatagramLength {
		return errors.New("buffer length exceeds maximum datagram length")
	}

	connection, err := net.Dial("udp", address.String())
	if err != nil {
		return fmt.Errorf("connecting to remote host at %q: %w", address, err)
	}
	defer connection.Close()

	n, err := connection.Write(buffer)
	TransmitBytes.WithLabelValues("udp_client").Add(float64(n))
	if err != nil {
		TransmitErrors.WithLabelValues("udp_client").Inc()
		return fmt.Errorf("sending the datagram payload: %w", err)
	}
	return nil
}
