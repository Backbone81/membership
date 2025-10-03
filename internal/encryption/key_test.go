package encryption_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/membership/internal/encryption"
)

var _ = Describe("Key", func() {
	It("should create random keys", func() {
		key := encryption.NewRandomKey()
		Expect(key.Data()).ToNot(BeNil())
		Expect(key.Data()).To(HaveLen(encryption.KeySize))

		for range 1000 {
			otherKey := encryption.NewRandomKey()
			Expect(key.Data()).ToNot(Equal(otherKey.Data()))
		}
	})

	It("should convert to string and back again", func() {
		key := encryption.NewRandomKey()

		hexEncodedKey := key.String()
		decodedKey, err := encryption.ParseKeyFromHexString(hexEncodedKey)
		Expect(err).ToNot(HaveOccurred())

		Expect(key.Data()).To(Equal(decodedKey.Data()))
	})

	DescribeTable("should fail to parse invalid keys",
		func(hexEncodedKey string) {
			Expect(encryption.ParseKeyFromHexString(hexEncodedKey)).Error().To(HaveOccurred())
		},
		Entry("an empty string", ""),
		Entry("an invalid hex string", "foo"),
		Entry("a key which is too short", encryption.NewRandomKey().String()[:encryption.KeySize*2-1]),
		Entry("a key which is too long", encryption.NewRandomKey().String()+"x"),
	)
})

func BenchmarkNewRandomKey(b *testing.B) {
	for b.Loop() {
		_ = encryption.NewRandomKey()
	}
}

func BenchmarkParseKeyFromHexString(b *testing.B) {
	hexEncodedString := encryption.NewRandomKey().String()
	for b.Loop() {
		_, err := encryption.ParseKeyFromHexString(hexEncodedString)
		if err != nil {
			b.Fatal(err)
		}
	}
}
