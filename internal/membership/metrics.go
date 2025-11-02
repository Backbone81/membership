package membership

import "github.com/prometheus/client_golang/prometheus"

var (
	AddMemberTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "membership_add_member_total",
			Help: "Total number of members added." +
				"You can calculate the active members by subtracting membership_remove_member_total from membership_add_member_total.",
		},
	)

	RemoveMemberTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "membership_remove_member_total",
			Help: "Total number of members removed." +
				"You can calculate the active members by subtracting membership_remove_member_total from membership_add_member_total.",
		},
	)
)

// RegisterMetrics registers all metrics collectors with the given prometheus registerer.
func RegisterMetrics(registerer prometheus.Registerer) error {
	metrics := []prometheus.Collector{
		AddMemberTotal,
		RemoveMemberTotal,
	}
	for _, metric := range metrics {
		if err := registerer.Register(metric); err != nil {
			return err
		}
	}
	return nil
}
