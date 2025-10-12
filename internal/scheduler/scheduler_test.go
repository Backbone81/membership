package scheduler_test

import (
	"time"

	"github.com/backbone81/membership/internal/scheduler"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestTarget provides a target implementation for testing the scheduler without actually running the full membership
// algorithm.
type TestTarget struct {
	DirectPingTimes          []time.Time
	IndirectPingTimes        []time.Time
	EndOfProtocolPeriodTimes []time.Time
	RequestListTimes         []time.Time
}

// TestTarget implements scheduler.Target
var _ scheduler.Target = (*TestTarget)(nil)

func (t *TestTarget) DirectPing() error {
	t.DirectPingTimes = append(t.DirectPingTimes, time.Now())
	return nil
}

func (t *TestTarget) IndirectPing() error {
	t.IndirectPingTimes = append(t.IndirectPingTimes, time.Now())
	return nil
}

func (t *TestTarget) EndOfProtocolPeriod() error {
	t.EndOfProtocolPeriodTimes = append(t.EndOfProtocolPeriodTimes, time.Now())
	return nil
}

func (t *TestTarget) RequestList() error {
	t.RequestListTimes = append(t.RequestListTimes, time.Now())
	return nil
}

var _ = Describe("Scheduler", func() {
	It("should correctly schedule", func() {
		var target TestTarget
		protocolPeriod := 10 * time.Millisecond
		directPingTimeout := 3 * time.Millisecond
		listRequestInterval := 3 * protocolPeriod
		maxSleepDuration := 100 * time.Microsecond

		myScheduler := scheduler.New(
			&target,
			scheduler.WithProtocolPeriod(protocolPeriod),
			scheduler.WithDirectPingTimeout(directPingTimeout),
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
