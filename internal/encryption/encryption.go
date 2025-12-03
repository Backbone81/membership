package encryption

const (
	// Overhead is the number of bytes the ciphertext is longer than the plaintext.
	Overhead = 12 + 16 // nonce length + tag length taken from gcmWithRandomNonce.Overhead()
)
