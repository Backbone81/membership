package scheduler

import "github.com/prometheus/client_golang/prometheus"

var (
	OperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "membership_scheduler_operations_total",
			Help: "Total number of scheduler operations executed.",
		},
		[]string{"operation"}, // direct_ping, indirect_ping, end_of_period, request_list
	)
	OperationErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "membership_scheduler_operation_errors_total",
			Help: "Total number of scheduler operation errors.",
		},
		[]string{"operation"},
	)
	OperationDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "membership_scheduler_operation_duration_seconds",
			Help:    "Duration of scheduler operations in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)
	ExpectedRTTSeconds = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "membership_scheduler_expected_rtt_seconds",
			Help: "Current expected round-trip time in seconds.",
		},
	)
)

// RegisterMetrics registers all metrics collectors with the given prometheus registerer.
func RegisterMetrics(registerer prometheus.Registerer) error {
	metrics := []prometheus.Collector{
		OperationsTotal,
		OperationErrorsTotal,
		OperationDurationSeconds,
		ExpectedRTTSeconds,
	}
	for _, metric := range metrics {
		if err := registerer.Register(metric); err != nil {
			return err
		}
	}
	return nil
}
