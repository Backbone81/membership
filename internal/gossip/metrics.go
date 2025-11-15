package gossip

import "github.com/prometheus/client_golang/prometheus"

var (
	MessagesAddedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "membership_gossip_messages_added_total",
			Help: "Total number of gossip messages added.",
		},
	)
	MessagesOverwrittenTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "membership_gossip_messages_overwritten_total",
			Help: "Total number of gossip messages overwritten due to higher precedence.",
		},
	)
	MessagesRemovedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "membership_gossip_messages_removed_total",
			Help: "Total number of gossip messages removed after exceeding maximum transmission count.",
		},
	)
	MessagesByTypeTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "membership_gossip_messages_by_type_total",
			Help: "Total number of gossip messages added to the queue, labeled by message type.",
		},
		[]string{"type"},
	)
	QueueCapacityMessages = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "membership_gossip_queue_capacity_messages",
			Help: "Current capacity of the gossip queue ring buffer, measured in messages.",
		},
	)
	QueueGrowthsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "membership_gossip_queue_growths_total",
			Help: "Total number of times the ring buffer grew to accommodate more messages.",
		},
	)
)

// RegisterMetrics registers all metrics collectors with the given prometheus registerer.
func RegisterMetrics(registerer prometheus.Registerer) error {
	metrics := []prometheus.Collector{
		MessagesAddedTotal,
		MessagesOverwrittenTotal,
		MessagesRemovedTotal,
		MessagesByTypeTotal,
		QueueCapacityMessages,
		QueueGrowthsTotal,
	}
	for _, metric := range metrics {
		if err := registerer.Register(metric); err != nil {
			return err
		}
	}
	return nil
}
