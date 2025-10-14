package encoding

import (
	"encoding/binary"
)

// Endian is the endianness membership uses for serializing/deserializing integers to network messages.
var Endian = binary.LittleEndian
