package encryption

import (
	"crypto/aes"
	"crypto/cipher"
)

const (
	// BlockSize is the number of bytes encrypted and decrypted as one block.
	BlockSize = aes.BlockSize

	// Overhead is the number of bytes the ciphertext is longer than the plaintext.
	Overhead = 12 + 16 // nonce size + tag size taken from gcmWithRandomNonce.Overhead()
)

// Encrypt encrypts the given plaintext with the given key.
// Note that encryption is done in-place on the plaintext buffer and overwrites its content.
func Encrypt(key Key, plaintext []byte) ([]byte, error) {
	aesCipher, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	mode, err := cipher.NewGCMWithRandomNonce(aesCipher)
	if err != nil {
		return nil, err
	}
	mode.Overhead()
	return mode.Seal(plaintext[:0], nil, plaintext, nil), nil
}

// Decrypt decrypts the given ciphertext with the given key.
// Note that decryption is done in-place on the ciphertext buffer and overwrites its content.
func Decrypt(key Key, ciphertext []byte) ([]byte, error) {
	aesCipher, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	mode, err := cipher.NewGCMWithRandomNonce(aesCipher)
	if err != nil {
		return nil, err
	}

	return mode.Open(ciphertext[:0], nil, ciphertext, nil)
}
