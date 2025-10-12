package scheduler_test

import (
	"testing"
	"time"

	"github.com/backbone81/membership/internal/scheduler"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
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
