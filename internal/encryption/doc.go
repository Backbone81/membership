// Package encryption provides encryption and decryption functionality for securing the data transmitted by the
// membership list and for making sure that only members with the same shared key are allowed to communicate with
// each other. Tampering of data during transmission is also detected.
//
// This package implements encryption with AES-256 in Galois Counter Mode.
//
// Note that the default nonce length of 96 bits is used. This might have an impact on large setups where collisions on
// nonces might be more likely to happen. This topic should be re-evaluated when such setups are actually happening.
// A configurable nonce sice might be the solution for this issue.
package encryption
