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
		config.Percentile = max(0, min(percentile, 1))
	}
}

func WithAlpha(alpha float64) Option {
	return func(config *Config) {
		config.Alpha = max(0, min(alpha, 1))
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
