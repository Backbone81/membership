package faultymember

// Config is the configuration for the faulty member list.
type Config struct {
	// PreAllocationCount is the number of elements the faulty member list should pre-allocate to reduce the number of
	// grow events during runtime.
	PreAllocationCount int

	// MaxListRequestCount is the maximum number of list requests faulty members are kept around before they are
	// dropped. While a faulty member is still within the first 50% of that list request count, it is still propagated
	// to other members during a full list sync. If a member is in the second 50% of that list request count, it is
	// not part of a full list sync anymore, but still kept around to avoid being re-added through full list syncs
	// with other members.
	MaxListRequestCount int
}

// DefaultConfig provides a default configuration for the faulty member list with sane defaults for most situations.
var DefaultConfig = Config{
	MaxListRequestCount: 10,
	PreAllocationCount:  64,
}
