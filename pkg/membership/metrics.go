package membership

import (
	intscheduler "github.com/backbone81/membership/internal/scheduler"
	"github.com/prometheus/client_golang/prometheus"

	intgossip "github.com/backbone81/membership/internal/gossip"
	intmembership "github.com/backbone81/membership/internal/membership"
)

// RegisterMetrics registers all metrics collectors with the given prometheus registerer.
func RegisterMetrics(registerer prometheus.Registerer) error {
	if err := intmembership.RegisterMetrics(registerer); err != nil {
		return err
	}
	if err := intgossip.RegisterMetrics(registerer); err != nil {
		return err
	}
	if err := intscheduler.RegisterMetrics(registerer); err != nil {
		return err
	}
	return nil
}
