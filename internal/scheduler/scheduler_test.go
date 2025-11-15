package scheduler_test

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/backbone81/membership/internal/scheduler"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Scheduler", func() {
	It("should correctly schedule", func() {
		synctest.Test(testingT, func(t *testing.T) {
			defer GinkgoRecover()

			var target TestTarget
			target.RTT = 100 * time.Millisecond

			myScheduler := scheduler.New(
				&target,
				scheduler.WithLogger(GinkgoLogr),
			)
			Expect(myScheduler.Startup()).To(Succeed())

			// Sleeping for 60 protocol periods will trigger 61 direct pings. We need to overlap a bit (1 ms) to make
			// sure we have a reliable outcome.
			time.Sleep(60*myScheduler.Config().ProtocolPeriod + 1*time.Millisecond)
			Expect(myScheduler.Shutdown()).To(Succeed())
			synctest.Wait()

			// As the 60th protocol period will immediately move into the next direct ping, we have 61 direct pings here.
			Expect(len(target.DirectPingTimes)).To(Equal(61))
			Expect(len(target.IndirectPingTimes)).To(Equal(60))
			Expect(len(target.EndOfProtocolPeriodTimes)).To(Equal(60))
			Expect(len(target.RequestListTimes)).To(Equal(2))

			for i := range len(target.DirectPingTimes) - 1 {
				Expect(target.IndirectPingTimes[i].Sub(target.DirectPingTimes[i])).To(Equal(target.RTT))
				Expect(target.EndOfProtocolPeriodTimes[i].Sub(target.DirectPingTimes[i])).To(Equal(myScheduler.Config().ProtocolPeriod))
				Expect(target.DirectPingTimes[i+1].Sub(target.DirectPingTimes[i])).To(Equal(myScheduler.Config().ProtocolPeriod))
			}
		})
	})

	It("should handle shutdown during indirect ping wait", func() {
		synctest.Test(testingT, func(t *testing.T) {
			defer GinkgoRecover()

			var target TestTarget
			target.RTT = 100 * time.Millisecond

			myScheduler := scheduler.New(
				&target,
				scheduler.WithLogger(GinkgoLogr),
			)
			Expect(myScheduler.Startup()).To(Succeed())

			// Shutdown after DirectPing but before IndirectPing
			time.Sleep(target.RTT - 1*time.Millisecond)
			Expect(myScheduler.Shutdown()).To(Succeed())
			synctest.Wait()

			Expect(len(target.DirectPingTimes)).To(Equal(1))
			Expect(len(target.IndirectPingTimes)).To(Equal(0))
		})
	})

	It("should handle shutdown during end-of-period wait", func() {
		synctest.Test(testingT, func(t *testing.T) {
			defer GinkgoRecover()

			var target TestTarget
			target.RTT = 100 * time.Millisecond

			myScheduler := scheduler.New(
				&target,
				scheduler.WithLogger(GinkgoLogr),
			)
			Expect(myScheduler.Startup()).To(Succeed())

			// Shutdown after IndirectPing but before EndOfProtocolPeriod
			time.Sleep(myScheduler.Config().ProtocolPeriod - 1*time.Millisecond)
			Expect(myScheduler.Shutdown()).To(Succeed())
			synctest.Wait()

			Expect(len(target.DirectPingTimes)).To(Equal(1))
			Expect(len(target.IndirectPingTimes)).To(Equal(1))
			Expect(len(target.EndOfProtocolPeriodTimes)).To(Equal(0))
		})
	})
})
