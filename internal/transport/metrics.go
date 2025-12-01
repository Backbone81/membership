package transport

import "github.com/prometheus/client_golang/prometheus"

var (
	TransmitBytes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "membership_list_transport_transmit_bytes_total",
			Help: "Total number of bytes transmitted.",
		},
		[]string{"transport"},
	)
	TransmitErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "membership_list_transport_transmit_errors_total",
			Help: "Total number of errors during transmit.",
		},
		[]string{"transport"},
	)
	ReceiveBytes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "membership_list_transport_receive_bytes_total",
			Help: "Total number of bytes received.",
		},
		[]string{"transport"},
	)
	ReceiveErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "membership_list_transport_receive_errors_total",
			Help: "Total number of errors during receive.",
		},
		[]string{"transport"},
	)
	Encryptions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "membership_list_transport_encryptions_total",
			Help: "Total number of encryption operations performed.",
		},
		[]string{"transport"},
	)
	Decryptions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "membership_list_transport_decryptions_total",
			Help: "Total number of decryption operations performed.",
		},
		[]string{"transport"},
	)
)

// RegisterMetrics registers all metrics collectors with the given prometheus registerer.
func RegisterMetrics(registerer prometheus.Registerer) error {
	metrics := []prometheus.Collector{
		TransmitBytes,
		TransmitErrors,
		ReceiveBytes,
		ReceiveErrors,
		Encryptions,
		Decryptions,
	}
	for _, metric := range metrics {
		if err := registerer.Register(metric); err != nil {
			return err
		}
	}
	return nil
}
