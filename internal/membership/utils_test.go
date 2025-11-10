package membership_test

import (
	"github.com/backbone81/membership/internal/membership"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Utils", func() {
	DescribeTable("IsIncarnationNumberNewer should provide correct results",
		func(old int, new int, want bool) {
			Expect(membership.IsIncarnationNumberNewer(old, new)).To(Equal(want))
		},
		Entry("old after new", 1, 0, false),
		Entry("old before new", 0, 1, true),
		Entry("old identical to new", 0, 0, false),
		Entry("new wraps around", (1<<16)-1, 0, true),
		Entry("new far behind (almost full wrap)", 0, (1<<16)-1, false),
		Entry("old on max, new on min", (1<<16)-1, 0, true),
		Entry("exactly half-space forward", 0, 1<<15, false),
		Entry("exactly half-space backward", 1<<15, 0, false),
		Entry("just under half-space", 0, (1<<15)-1, true),
		Entry("just over half-space", 0, (1<<15)+1, false),
	)
})
