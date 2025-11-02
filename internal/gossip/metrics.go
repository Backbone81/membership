package gossip

import "github.com/prometheus/client_golang/prometheus"

var (
	AddMessageTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "membership_gossip_add_message_total",
			Help: "Total number of gossip messages added. " +
				"You can calculate the active gossip messages by subtracting membership_gossip_remove_message_total from membership_gossip_add_message_total.",
		},
	)

	RemoveMessageTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "membership_gossip_remove_message_total",
			Help: "Total number of gossip messages removed. " +
				"You can calculate the active gossip messages by subtracting membership_gossip_remove_message_total from membership_gossip_add_message_total.",
		},
	)
)

// RegisterMetrics registers all metrics collectors with the given prometheus registerer.
func RegisterMetrics(registerer prometheus.Registerer) error {
	metrics := []prometheus.Collector{
		AddMessageTotal,
		RemoveMessageTotal,
	}
	for _, metric := range metrics {
		if err := registerer.Register(metric); err != nil {
			return err
		}
	}
	return nil
}
