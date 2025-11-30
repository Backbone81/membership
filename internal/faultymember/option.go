package faultymember

// Option is the function signature for all list options to implement.
type Option func(config *Config)

func WithMaxListRequestCount(count int) Option {
	count = max(1, count)
	return func(config *Config) {
		config.MaxListRequestCount = count
	}
}

func WithPreAllocationCount(count int) Option {
	count = max(1, count)
	return func(config *Config) {
		config.PreAllocationCount = count
	}
}
