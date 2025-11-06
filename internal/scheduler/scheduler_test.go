package scheduler_test

import (
	"time"

	"github.com/backbone81/membership/internal/scheduler"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Scheduler", func() {
	It("should correctly schedule", Serial, func() {
		var target TestTarget
		protocolPeriod := 10 * time.Millisecond
		directPingTimeout := 3 * time.Millisecond
		listRequestInterval := 3 * protocolPeriod
		maxSleepDuration := 100 * time.Microsecond
		target.RTT = directPingTimeout

		myScheduler := scheduler.New(
			&target,
			scheduler.WithLogger(GinkgoLogr),
			scheduler.WithProtocolPeriod(protocolPeriod),
			scheduler.WithMaxSleepDuration(maxSleepDuration),
			scheduler.WithListRequestInterval(listRequestInterval),
		)
		Expect(myScheduler.Startup()).To(Succeed())
		time.Sleep(10*protocolPeriod + protocolPeriod/2)
		Expect(myScheduler.Shutdown()).To(Succeed())

		Expect(len(target.DirectPingTimes)).To(BeNumerically("~", 10, 1))
		Expect(len(target.IndirectPingTimes)).To(BeNumerically("~", 10, 1))
		Expect(len(target.EndOfProtocolPeriodTimes)).To(BeNumerically("~", 10, 1))
		Expect(len(target.RequestListTimes)).To(BeNumerically("~", 3, 1))

		for i := range len(target.DirectPingTimes) - 1 {
			Expect(target.IndirectPingTimes[i].Sub(target.DirectPingTimes[i])).To(BeNumerically("~", directPingTimeout, 1*time.Millisecond))
			Expect(target.EndOfProtocolPeriodTimes[i].Sub(target.DirectPingTimes[i])).To(BeNumerically("~", protocolPeriod, 1*time.Millisecond))
			Expect(target.DirectPingTimes[i+1].Sub(target.DirectPingTimes[i])).To(BeNumerically("~", protocolPeriod, 1*time.Millisecond))
		}
	})
})
