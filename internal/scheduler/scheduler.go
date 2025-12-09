package scheduler

import (
	"math"
	"sync"
	"time"

	"github.com/go-logr/logr"
)

// Scheduler is driving the timed actions of the membership list. This allows us to separate the algorithm from temporal
// constraints and enables tests of the main algorithm to run without delays.
//
// Scheduler is safe for concurrent use by multiple goroutines. But you need to make sure that Shutdown is only called
// after Startup and you should call Startup and Shutdown only once. Create a new Scheduler if you need to restart.
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

	if config.RoundTripTimeTracker == nil {
		panic("you must provide a round trip time tracker")
	}

	return &Scheduler{
		logger:   config.Logger,
		config:   config,
		target:   target,
		shutdown: make(chan struct{}),
	}
}

// Config returns the config of the scheduler.
func (s *Scheduler) Config() Config {
	return s.config
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
	s.listRequestTicker.Stop()
	close(s.shutdown)
	s.waitGroup.Wait()
	return nil
}

// protocolPeriodTask is driving the membership list algorithm.
//

func (s *Scheduler) protocolPeriodTask() {
	s.logger.Info("Protocol period background task started")
	defer s.logger.Info("Protocol period background task finished")

	var lastExpectedRoundTripTime time.Duration
	for {
		startOfProtocolPeriod := time.Now()
		s.measure("direct_ping", func() error {
			if err := s.target.DirectPing(); err != nil {
				s.logger.Error(err, "Scheduled direct ping.")
				return err
			}
			return nil
		})

		// Adjust the timeout for the direct ping to what we observed can be expected. We always use the current value
		// for the timeout, but we also want to create a log entry, when the timeout changes significantly. Therefore,
		// we only log when we move at least 10% away of the last time we logged.
		currExpectedRoundTripTime := s.config.RoundTripTimeTracker.GetCalculated()
		ExpectedRTTSeconds.Set(currExpectedRoundTripTime.Seconds())
		logThreshold := lastExpectedRoundTripTime / 10
		if math.Abs(float64(currExpectedRoundTripTime)-float64(lastExpectedRoundTripTime)) > float64(logThreshold) {
			s.logger.Info(
				"Direct ping timeout adjusted",
				"was", lastExpectedRoundTripTime,
				"is", currExpectedRoundTripTime,
			)
			lastExpectedRoundTripTime = currExpectedRoundTripTime
		}

		if !s.waitFor(currExpectedRoundTripTime) {
			return
		}
		s.measure("indirect_ping", func() error {
			if err := s.target.IndirectPing(); err != nil {
				s.logger.Error(err, "Scheduled indirect ping.")
				return err
			}
			return nil
		})

		if !s.waitFor(s.config.ProtocolPeriod - currExpectedRoundTripTime) {
			return
		}
		s.measure("end_of_protocol_period", func() error {
			s.config.RoundTripTimeTracker.UpdateCalculated()
			if err := s.target.EndOfProtocolPeriod(); err != nil {
				s.logger.Error(err, "End of protocol period.")
				return err
			}
			return nil
		})

		gotProtocolPeriod := time.Since(startOfProtocolPeriod)
		if gotProtocolPeriod > s.config.ProtocolPeriod*11/10 {
			s.logger.Info(
				"WARNING: The protocol period was more than 10% longer than expected. "+
					"This is a strong indication that the system is overloaded. "+
					"Members declared as suspect or faulty by this member are probably false positives.",
				"want-duration", s.config.ProtocolPeriod,
				"got-duration", gotProtocolPeriod,
			)
		}
	}
}

// waitFor will sleep until the given time is reached. It will wake up in between to check if a shutdown is in
// progress. If a shutdown is in progress, it will return false. If the time was reached, it will return true.
// A warning will be logged when the timestamp is already in the past. This is usually an indication for an overloaded
// system.
func (s *Scheduler) waitFor(timeout time.Duration) bool {
	if s.shutdownInProgress() {
		return false
	}

	timestamp := time.Now().Add(timeout)

	now := time.Now()
	for now.Before(timestamp) {
		timeToWait := timestamp.Sub(now)
		time.Sleep(min(timeToWait, s.config.MaxSleepDuration))
		now = time.Now()

		// In case a shutdown is already triggered, we do not want to wait for the full duration of waitUntil. We
		// therefore repeatedly check for shutdown and then continue to sleep for a bit. This allows for a fast
		// shutdown.
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
	s.measure("request_list", func() error {
		if err := s.target.RequestList(); err != nil {
			s.logger.Error(err, "Startup list request.")
			return err
		}
		return nil
	})

	for {
		select {
		case <-s.shutdown:
			return
		case <-s.listRequestTicker.C:
			s.measure("Request list completed", func() error {
				if err := s.target.RequestList(); err != nil {
					s.logger.Error(err, "Scheduled list request.")
					return err
				}
				return nil
			})
		}
	}
}

// measure executes the given function and measures the time needed. It will log the given message with the measured
// duration.
func (s *Scheduler) measure(operation string, f func() error) {
	start := time.Now()
	if err := f(); err != nil {
		OperationErrorsTotal.WithLabelValues(operation).Inc()
	}
	OperationDurationSeconds.WithLabelValues(operation).Observe(time.Since(start).Seconds())
	OperationsTotal.WithLabelValues(operation).Inc()
}
