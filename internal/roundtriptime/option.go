package roundtriptime

import "time"

// Option is the data type all Tracker options need to implement.
type Option func(config *Config)

func WithCount(count int) Option {
	return func(config *Config) {
		config.Count = max(1, count)
	}
}

func WithPercentile(percentile float64) Option {
	return func(config *Config) {
		config.Percentile = min(1, max(0, percentile))
	}
}

func WithAlpha(alpha float64) Option {
	return func(config *Config) {
		config.Alpha = min(1, max(0, alpha))
	}
}

func WithDefault(duration time.Duration) Option {
	return func(config *Config) {
		config.Default = max(0, duration)
	}
}

func WithMinimum(duration time.Duration) Option {
	return func(config *Config) {
		config.Minimum = max(0, duration)
	}
}

func WithMaximum(duration time.Duration) Option {
	return func(config *Config) {
		config.Maximum = max(0, duration)
	}
}
