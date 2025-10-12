// Package scheduler provides functionality for driving the membership list algorithm with the desired timing. It
// triggers direct and indirect pings and triggers the end of the protocol period. This allows for a cleaner separation
// of concerns and makes it possible to drive the membership list algorithm without scheduling in the context of tests.
package scheduler
