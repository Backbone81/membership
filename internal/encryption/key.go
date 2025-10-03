package encryption

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// KeySize is the size for keys for encryption and decryption in bytes.
const KeySize = 32

// Key is a key for encrypting and decrypting data.
type Key [KeySize]byte

// Key implements fmt.Stringer.
var _ fmt.Stringer = (*Key)(nil)

// String returns the key as a hexadecimal string.
func (k Key) String() string {
	return hex.EncodeToString(k.Data())
}

// Data returns the data which makes up the key as a slice of bytes.
func (k *Key) Data() []byte {
	return k[:]
}

// NewRandomKey creates a new random key for encryption and decryption.
func NewRandomKey() Key {
	var result Key
	n, err := rand.Read(result.Data())
	if err != nil || n != KeySize {
		panic("failed to create a new random encryption key")
	}
	return result
}

// ParseKeyFromHexString parses the given hexadecimal string into a key.
func ParseKeyFromHexString(hexString string) (Key, error) {
	var result Key
	n, err := hex.Decode(result.Data(), ([]byte)(hexString))
	if err != nil {
		return Key{}, fmt.Errorf("decoding encryption key from hex string: %w", err)
	}
	if n != KeySize {
		return Key{}, fmt.Errorf("invalid encryption key size: expected %d, got %d", KeySize, n)
	}
	return result, nil
}
