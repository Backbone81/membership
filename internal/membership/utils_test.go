package membership_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/membership/internal/membership"
)

var _ = Describe("Utils", func() {
	DescribeTable("IncarnationLessThan should provide correct results",
		func(lhs int, rhs int, want bool) {
			Expect(membership.IncarnationLessThan(lhs, rhs)).To(Equal(want))
		},
		Entry("lhs after rhs", 1, 0, false),
		Entry("lhs before rhs", 0, 1, true),
		Entry("lhs identical to rhs", 0, 0, false),
		Entry("rhs wraps around", (1<<16)-1, 0, true),
		Entry("rhs far behind (almost full wrap)", 0, (1<<16)-1, false),
		Entry("lhs on max, rhs on min", (1<<16)-1, 0, true),
		Entry("exactly half-space forward", 0, 1<<15, false),
		Entry("exactly half-space backward", 1<<15, 0, false),
		Entry("just under half-space", 0, (1<<15)-1, true),
		Entry("just over half-space", 0, (1<<15)+1, false),
	)
})
