package gossip

// Option is the function signature for all queue options to implement.
type Option func(config *Config)

func WithMaxTransmissionCount(count int) Option {
	count = max(1, count)
	return func(config *Config) {
		config.MaxTransmissionCount = count
	}
}

func WithPreAllocationCount(count int) Option {
	count = max(1, count)
	return func(config *Config) {
		config.PreAllocationCount = count
	}
}
