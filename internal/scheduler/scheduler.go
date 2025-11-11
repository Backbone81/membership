package scheduler

import (
	"math"
	"sync"
	"time"

	"github.com/go-logr/logr"
)

// Scheduler is driving the timed actions of the membership list. This allows us to separate the algorithm from temporal
// constraints and enables tests of the main algorithm to run without delays.
type Scheduler struct {
	logger            logr.Logger
	config            Config
	target            Target
	waitGroup         sync.WaitGroup
	shutdown          chan struct{}
	listRequestTicker *time.Ticker
}

// New creates a new scheduler with the given configuration. Provide options to customize default config.
func New(target Target, options ...Option) *Scheduler {
	config := DefaultConfig
	for _, option := range options {
		option(&config)
	}
	return &Scheduler{
		logger:   config.Logger,
		config:   config,
		target:   target,
		shutdown: make(chan struct{}),
	}
}

// Startup executes the scheduler. It will trigger the membership list algorithm until Shutdown is called.
func (s *Scheduler) Startup() error {
	s.logger.Info("Scheduler startup")
	s.listRequestTicker = time.NewTicker(s.config.ListRequestInterval)
	s.waitGroup.Go(func() {
		s.protocolPeriodTask()
	})
	s.waitGroup.Go(func() {
		s.requestListTask()
	})
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

	var lastExpectedRoundTripTime time.Duration
	directPingAt := time.Now()
	for {
		s.measure("Direct ping completed", func() {
			if err := s.target.DirectPing(); err != nil {
				s.logger.Error(err, "Scheduled direct ping.")
			}
		})

		// Adjust the timeout for the direct ping to what we observed can be expected. We always use the current value
		// for the timeout, but we also want to create a log entry, when the timeout changes significantly. Therefore,
		// we only log when we move at least 10% away of the last time we logged.
		currExpectedRoundTripTime := s.target.ExpectedRoundTripTime()
		logThreshold := lastExpectedRoundTripTime / 10
		if math.Abs(float64(currExpectedRoundTripTime)-float64(lastExpectedRoundTripTime)) > float64(logThreshold) {
			s.logger.Info(
				"Direct ping timeout adjusted",
				"was", lastExpectedRoundTripTime,
				"is", currExpectedRoundTripTime,
			)
			lastExpectedRoundTripTime = currExpectedRoundTripTime
		}

		indirectPingAt := directPingAt.Add(currExpectedRoundTripTime)
		if time.Until(indirectPingAt) < currExpectedRoundTripTime/2 {
			s.logger.Info(
				"WARNING: The time between the direct ping and indirect ping is less than 50% of the expected round trip time. " +
					"This is a strong indication that the system is overloaded. " +
					"Members declared as suspect or faulty by this member are probably false positives.",
			)
		}
		if !s.waitUntil(indirectPingAt) {
			return
		}
		s.measure("Indirect ping completed", func() {
			if err := s.target.IndirectPing(); err != nil {
				s.logger.Error(err, "Scheduled indirect ping.")
			}
		})

		endOfProtocolPeriodAt := directPingAt.Add(s.config.ProtocolPeriod)
		if !s.waitUntil(endOfProtocolPeriodAt) {
			return
		}
		s.measure("End of protocol period completed", func() {
			if err := s.target.EndOfProtocolPeriod(); err != nil {
				s.logger.Error(err, "End of protocol period.")
			}
		})

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

	// Let's do a list request right at startup to have the full member list available as soon as possible.
	s.measure("Request list completed", func() {
		if err := s.target.RequestList(); err != nil {
			s.logger.Error(err, "Startup list request.")
		}
	})

	for {
		select {
		case <-s.shutdown:
			return
		case <-s.listRequestTicker.C:
			s.measure("Request list completed", func() {
				if err := s.target.RequestList(); err != nil {
					s.logger.Error(err, "Scheduled list request.")
				}
			})
		}
	}
}

// measure executes the given function and measures the time needed. It will log the given message with the measured
// duration.
func (s *Scheduler) measure(message string, f func()) {
	start := time.Now()
	f()
	s.logger.V(4).Info(message, "duration", time.Since(start))
}
