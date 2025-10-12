package membership

import (
	"sync"
	"time"

	"github.com/go-logr/logr"
)

// Scheduler is driving the timed actions of the membership list. This allows us to separate the algorithm from temporal
// constraints and enables tests of the main algorithm to run without delays.
type Scheduler struct {
	logger            logr.Logger
	config            SchedulerConfig
	target            SchedulerTarget
	waitGroup         sync.WaitGroup
	shutdown          chan struct{}
	listRequestTicker *time.Ticker
}

// SchedulerTarget is the interface which the implementation of the membership algorithm must implement to be driven
// by the scheduler.
type SchedulerTarget interface {
	// DirectPing is the start of the protocol period which is executing the direct ping.
	DirectPing() error

	// IndirectPing is after some time elapsed and indirect pings need to be executed.
	IndirectPing() error

	// EndOfProtocolPeriod is the end of the protocol period where suspects and faulty members need to be declared.
	EndOfProtocolPeriod() error

	// RequestList fetches the full member list from a randomly chosen member.
	RequestList() error
}

// SchedulerConfig is the configuration the scheduler is using.
type SchedulerConfig struct {
	// Logger is the logger to use for outputting status information.
	Logger logr.Logger

	// ProtocolPeriod is the time for a full cycle of direct ping followed by indirect pings. If there is no response
	// from the target member within that time, we have to assume the member to have failed.
	// Note that the protocol period must be at least three times the round-trip time.
	ProtocolPeriod time.Duration

	// DirectPingTimeout is the time to wait for a direct ping response. If there is no response within this duration,
	// we need to start indirect pings.
	DirectPingTimeout time.Duration

	// MaxSleepDuration is the maximum duration which the scheduler is allowed to sleep. Making this value bigger will
	// result in delays during shutdown. Making this value smaller will wake the scheduler more often to check for
	// a shutdown in progress which will cause more load during runtime.
	MaxSleepDuration time.Duration

	// ListRequestInterval is the time interval in which a full member list is requested from a randomly selected member.
	ListRequestInterval time.Duration
}

// SchedulerDefaultConfig provides a scheduler configuration with sane defaults for most situations.
var SchedulerDefaultConfig = SchedulerConfig{
	ProtocolPeriod:      1 * time.Second,
	DirectPingTimeout:   100 * time.Millisecond,
	MaxSleepDuration:    100 * time.Millisecond,
	ListRequestInterval: 1 * time.Minute,
}

// NewScheduler creates a new scheduler with the given configuration.
func NewScheduler(target SchedulerTarget, config SchedulerConfig) *Scheduler {
	return &Scheduler{
		logger:            config.Logger,
		config:            config,
		target:            target,
		shutdown:          make(chan struct{}),
	}
}

// Startup executes the scheduler. It will trigger the membership list algorithm until Shutdown is called.
func (s *Scheduler) Startup() error {
	s.logger.Info("Scheduler startup")
	s.listRequestTicker = time.NewTicker(s.config.ListRequestInterval)
	s.waitGroup.Add(2)
	go s.protocolPeriodTask()
	go s.requestListTask()
	return nil
}

// Shutdown stops the scheduler. It will block until all callbacks have completed.
func (s *Scheduler) Shutdown() error {
	s.logger.Info("Scheduler shutdown")
	close(s.shutdown)
	s.listRequestTicker.Stop()
	s.waitGroup.Wait()
	return nil
}

// protocolPeriodTask is driving the membership list algorithm.
func (s *Scheduler) protocolPeriodTask() {
	s.logger.Info("Protocol period background task started")
	defer s.logger.Info("Protocol period background task finished")
	defer s.waitGroup.Done()

	directPingAt := time.Now()
	for {
		if err := s.target.DirectPing(); err != nil {
			s.logger.Error(err, "Scheduled direct ping.")
		}

		indirectPingAt := directPingAt.Add(s.config.DirectPingTimeout)
		if indirectPingAt.Sub(time.Now()) < s.config.DirectPingTimeout/2 {
			s.logger.Info(
				"WARNING: The time between the direct ping and indirect ping is less than 50% of the configured timeout. " +
					"This is a strong indication that the system is overloaded. " +
					"Members declared as suspect or faulty by this member are probably false positives.",
			)
		}
		if !s.waitUntil(indirectPingAt) {
			return
		}
		if err := s.target.IndirectPing(); err != nil {
			s.logger.Error(err, "Scheduled indirect ping.")
		}

		endOfProtocolPeriodAt := directPingAt.Add(s.config.ProtocolPeriod)
		if !s.waitUntil(endOfProtocolPeriodAt) {
			return
		}
		if err := s.target.EndOfProtocolPeriod(); err != nil {
			s.logger.Error(err, "End of protocol period.")
		}

		directPingAt = directPingAt.Add(s.config.ProtocolPeriod)
	}
}

// waitUntil will sleep until the given time is reached. It will wake up in between to check if a shutdown is in
// progress. If a shutdown is in progress, it will return false. If the time was reached, it will return true.
// A warning will be logged when the timestamp is already in the past. This is usually an indication for an overloaded
// system.
func (s *Scheduler) waitUntil(timestamp time.Time) bool {
	if s.shutdownInProgress() {
		return false
	}

	if timestamp.Before(time.Now()) {
		s.logger.Info(
			"WARNING: The scheduled time within the protocol period has already passed. " +
				"This is a strong indication that the system is overloaded. " +
				"Members declared as suspect or faulty by this member are probably false positives.",
		)
		return true
	}

	now := time.Now()
	for now.Before(timestamp) {
		timeToWait := timestamp.Sub(now)
		time.Sleep(min(timeToWait, s.config.MaxSleepDuration))
		now = time.Now()

		if s.shutdownInProgress() {
			return false
		}
	}
	return true
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

// requestListTask periodically requests the full member list from a random member.
func (s *Scheduler) requestListTask() {
	s.logger.Info("Member list request background task started")
	defer s.logger.Info("Member list request background task finished")
	defer s.waitGroup.Done()

	for {
		select {
		case <-s.shutdown:
			return
		case <-s.listRequestTicker.C:
			if err := s.target.RequestList(); err != nil {
				s.logger.Error(err, "Scheduled list request.")
			}
		}
	}
}
