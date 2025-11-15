package scheduler_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/membership/internal/scheduler"
)

// As Ginkgo does not yet support testing/synctest, we need to capture t during test suite initialization and make it
// available to our Ginkgo tests. Keep an eye on https://github.com/onsi/ginkgo/issues/1601 and remove this hack
// when Ginkgo provides support for it.
var testingT *testing.T

func TestSuite(t *testing.T) {
	testingT = t
	RegisterFailHandler(Fail)
	RunSpecs(t, "Scheduler Suite")
}

// TestTarget provides a target implementation for testing the scheduler without actually running the full membership
// algorithm.
type TestTarget struct {
	DirectPingTimes          []time.Time
	IndirectPingTimes        []time.Time
	EndOfProtocolPeriodTimes []time.Time
	RequestListTimes         []time.Time
	RTT                      time.Duration
}

// TestTarget implements scheduler.Target.
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

func (t *TestTarget) ExpectedRoundTripTime() time.Duration {
	return t.RTT
}
