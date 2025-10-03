package membership

import (
	"sync"
	"time"
)

// Scheduler is driving the timed actions of the membership list. This allows us to separate the algorithm from temporal
// constraints.
type Scheduler struct {
	config    SchedulerConfig
	waitGroup sync.WaitGroup
	shutdown  chan struct{}
}

// SchedulerConfig is the configuration the scheduler is using.
type SchedulerConfig struct {
	// DirectPingTimeout is the time to wait for a direct ping response. If there is no response within this duration,
	// we need to start indirect pings.
	DirectPingTimeout time.Duration

	// ProtocolPeriod is the time for a full cycle of direct ping followed by indirect pings. If there is no response
	// from the target node within that time, we have to assume the node to have failed.
	// Note that the protocol period must be at least three times the round-trip time.
	ProtocolPeriod time.Duration
}

// SchedulerDefaultConfig provides a scheduler configuration with sane defaults for most situations.
var SchedulerDefaultConfig = SchedulerConfig{
	DirectPingTimeout: 100 * time.Millisecond,
	ProtocolPeriod:    1 * time.Second,
}

// NewScheduler creates a new scheduler with the given configuration.
func NewScheduler(config SchedulerConfig) *Scheduler {
	// TODO: Validate config parameters.

	return &Scheduler{
		config:   config,
		shutdown: make(chan struct{}),
	}
}

// Startup executes the scheduler. It will trigger the membership list algorithm until Shutdown is called.
func (s *Scheduler) Startup() error {
	s.waitGroup.Add(1)
	go s.backgroundTask()
	return nil
}

// Shutdown stops the scheduler. It will block until all callbacks have completed.
func (s *Scheduler) Shutdown() error {
	close(s.shutdown)
	s.waitGroup.Wait()
	return nil
}

// backgroundTask is driving the membership list algorithm.
func (s *Scheduler) backgroundTask() {
	defer s.waitGroup.Done()
	for {
		if s.shutdownInProgress() {
			// We check for shutdown once before we start the next period. We make sure that an algorithm period always
			// runs to completion before shutdown.
			return
		}

		startOfProtocolPeriod := time.Now()
		s.directPingHandler()
		time.Sleep(time.Until(startOfProtocolPeriod.Add(s.config.DirectPingTimeout)))
		s.indirectPingHandler()
		time.Sleep(time.Until(startOfProtocolPeriod.Add(s.config.ProtocolPeriod)))
		s.endOfPeriodHandler()
	}
}

// shutdownInProgress reports if a shutdown is currently in progress.
func (s *Scheduler) shutdownInProgress() bool {
	select {
	case <-s.shutdown:
		return true
	default:
		return false
	}
}

func (s *Scheduler) directPingHandler() {
	// TODO: call membership list
}

func (s *Scheduler) indirectPingHandler() {
	// TODO: call membership list
}

func (s *Scheduler) endOfPeriodHandler() {
	// TODO: call membership list
}
