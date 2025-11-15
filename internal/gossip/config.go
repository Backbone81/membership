package gossip

// Config is the configuration for the gossip queue.
type Config struct {
	// MaxTransmissionCount is the maximum number of times messages in the gossip queue are transmitted before they are
	// dropped.
	MaxTransmissionCount int

	// PreAllocationCount is the number of elements the gossip queue should pre-allocate to reduce the number of grow
	// events during runtime.
	PreAllocationCount int
}

// DefaultConfig provides a default configuration for the gossip queue with sane defaults for most situations.
var DefaultConfig = Config{
	MaxTransmissionCount: 8,
	PreAllocationCount:   64,
}
