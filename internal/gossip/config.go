package gossip

// Config is the configuration for the gossip queue.
type Config struct {
	MaxTransmissionCount int
	PreAllocationCount   int
}

// DefaultConfig provides a default configuration for the gossip queue with sane defaults for most situations.
var DefaultConfig = Config{
	MaxTransmissionCount: 10,
	PreAllocationCount:   128,
}
