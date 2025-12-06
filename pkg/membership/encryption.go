package membership

import "github.com/backbone81/membership/internal/encryption"

type Key = encryption.Key

var NewRandomKey = encryption.NewRandomKey

var ParseKeyFromHexString = encryption.ParseKeyFromHexString
