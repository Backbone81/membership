package membership

import "github.com/prometheus/client_golang/prometheus"

var (
	MembersAddedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "membership_list_members_added_total",
			Help: "Total number of members added.",
		},
	)
	MembersRemovedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "membership_list_members_removed_total",
			Help: "Total number of members removed.",
		},
	)
)

// RegisterMetrics registers all metrics collectors with the given prometheus registerer.
func RegisterMetrics(registerer prometheus.Registerer) error {
	metrics := []prometheus.Collector{
		MembersAddedTotal,
		MembersRemovedTotal,
	}
	for _, metric := range metrics {
		if err := registerer.Register(metric); err != nil {
			return err
		}
	}
	return nil
}
