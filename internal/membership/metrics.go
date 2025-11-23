package membership

import "github.com/prometheus/client_golang/prometheus"

var (
	MembersByState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "membership_list_members",
			Help: "Current number of members by state (alive, suspect, faulty).",
		},
		[]string{"state"},
	)
	MemberStateTransitionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "membership_list_member_state_transitions_total",
			Help: "Total number of member state transitions.",
		},
		[]string{"transition"},
	)
	MessagesReceivedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "membership_list_messages_received_total",
			Help: "Total number of member state transitions.",
		},
		[]string{"type"},
	)
)

// RegisterMetrics registers all metrics collectors with the given prometheus registerer.
func RegisterMetrics(registerer prometheus.Registerer) error {
	metrics := []prometheus.Collector{
		MembersByState,
		MemberStateTransitionsTotal,
		MessagesReceivedTotal,
	}
	for _, metric := range metrics {
		if err := registerer.Register(metric); err != nil {
			return err
		}
	}
	return nil
}
