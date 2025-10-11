package encryption_test

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/membership/internal/encryption"
)

var _ = Describe("Encryption", func() {
	It("should encrypt plaintext to ciphertext", func() {
		key := encryption.NewRandomKey()
		ciphertext, err := encryption.Encrypt(key, []byte("foo"))
		Expect(err).ToNot(HaveOccurred())
		Expect(ciphertext).ToNot(Equal([]byte("foo")))
	})

	It("should encrypt and decrypt", func() {
		key := encryption.NewRandomKey()
		plaintext := []byte("foo")

		ciphertext, err := encryption.Encrypt(key, plaintext)
		Expect(err).ToNot(HaveOccurred())

		plaintext, err = encryption.Decrypt(key, ciphertext)
		Expect(err).ToNot(HaveOccurred())

		Expect(plaintext).To(Equal([]byte("foo")))
	})

	It("should encrypt the same data in a different way because of different nonces", func() {
		key := encryption.NewRandomKey()

		ciphertext1, err := encryption.Encrypt(key, []byte("foo"))
		Expect(err).ToNot(HaveOccurred())

		ciphertext2, err := encryption.Encrypt(key, []byte("foo"))
		Expect(err).ToNot(HaveOccurred())

		Expect(ciphertext1).ToNot(Equal(ciphertext2))
	})

	It("should not decrypt with the wrong key", func() {
		key := encryption.NewRandomKey()
		plaintext := []byte("foo")

		ciphertext, err := encryption.Encrypt(key, plaintext)
		Expect(err).ToNot(HaveOccurred())

		wrongKey := encryption.NewRandomKey()
		plaintext, err = encryption.Decrypt(wrongKey, ciphertext)
		Expect(err).To(HaveOccurred())
	})

	It("should not decrypt when ciphertext has been tampered with", func() {
		key := encryption.NewRandomKey()
		plaintext := []byte("foo")

		ciphertext, err := encryption.Encrypt(key, plaintext)
		Expect(err).ToNot(HaveOccurred())

		ciphertext[3]++

		plaintext, err = encryption.Decrypt(key, ciphertext)
		Expect(err).To(HaveOccurred())
	})

	It("should correctly respect the overhead", func() {
		key := encryption.NewRandomKey()
		for plaintextLength := range 1024 {
			plaintext := bytes.Repeat([]byte{0}, plaintextLength)
			ciphertext, err := encryption.Encrypt(key, plaintext)
			Expect(err).ToNot(HaveOccurred())

			Expect(ciphertext).To(HaveLen(plaintextLength + encryption.Overhead))
		}
	})
})

func BenchmarkEncrypt(b *testing.B) {
	key := encryption.NewRandomKey()
	buffer := make([]byte, 0, 10*1024)
	for dataLength := 8; dataLength <= 1024; dataLength *= 2 {
		data := make([]byte, dataLength)
		if _, err := rand.Read(data); err != nil {
			b.Fatal(err)
		}

		b.Run(fmt.Sprintf("%d bytes", dataLength), func(b *testing.B) {
			for b.Loop() {
				buffer = append(buffer[:0], data...)
				_, err := encryption.Encrypt(key, buffer)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkDecrypt(b *testing.B) {
	key := encryption.NewRandomKey()
	buffer := make([]byte, 0, 10*1024)
	for dataLength := 8; dataLength <= 1024; dataLength *= 2 {
		data := make([]byte, dataLength)
		if _, err := rand.Read(data); err != nil {
			b.Fatal(err)
		}
		ciphertext, err := encryption.Encrypt(key, data)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(fmt.Sprintf("%d bytes", dataLength), func(b *testing.B) {
			for b.Loop() {
				buffer = append(buffer[:0], ciphertext...)
				_, err := encryption.Decrypt(key, buffer)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
