package packetlossjoin

import (
	"math"
	"net"

	"github.com/go-logr/logr"

	"github.com/backbone81/membership/internal/encoding"
	"github.com/backbone81/membership/internal/membership"
	"github.com/backbone81/membership/internal/transport"
	"github.com/backbone81/membership/internal/utility"
)

// PacketLossJoin measures the cluster size during member joins with an unreliable transport.
func PacketLossJoin(logger logr.Logger) error {
	logger.Info("The cluster size during member joins with an unreliable transport.")
	for memberCount := range utility.ClusterSize(17, 17, 17) {
		memoryTransport := transport.NewMemory()

		if err := runProtocol(logger, memoryTransport, memberCount); err != nil {
			return err
		}
	}
	return nil
}

func runProtocol(logger logr.Logger, memoryTransport *transport.Memory, memberCount int) error {
	var observedMinMemberCount []int
	var lists []*membership.List
	for protocolPeriod := range 1024 {
		if protocolPeriod < memberCount {
			// We join one member in every protocol period until the max member count is reached.
			address := encoding.NewAddress(net.IPv4(255, 255, 255, 255), protocolPeriod+1)
			options := []membership.Option{
				membership.WithLogger(logr.Discard()),
				membership.WithAdvertisedAddress(address),
				membership.WithUDPClient(&transport.Unreliable{
					Transport:   memoryTransport.Client(),
					Reliability: 0.9,
				}),
				membership.WithTCPClient(&transport.Unreliable{
					Transport:   memoryTransport.Client(),
					Reliability: 0.9,
				}),
			}
			for _, list := range lists {
				options = append(options,
					membership.WithBootstrapMember(list.AdvertiseAddress()),
				)
			}
			newList := membership.NewList(options...)
			memoryTransport.AddTarget(address, newList)
			lists = append(lists, newList)
		}

		for _, list := range lists {
			if err := list.DirectPing(); err != nil {
				return err
			}
		}
		if err := memoryTransport.FlushAllPendingSends(); err != nil {
			return err
		}

		for _, list := range lists {
			if err := list.IndirectPing(); err != nil {
				return err
			}
		}
		if err := memoryTransport.FlushAllPendingSends(); err != nil {
			return err
		}

		for _, list := range lists {
			if err := list.EndOfProtocolPeriod(); err != nil {
				return err
			}
		}

		minMemberCount := math.MaxInt
		for _, list := range lists {
			minMemberCount = min(minMemberCount, list.Len())
		}
		observedMinMemberCount = append(observedMinMemberCount, minMemberCount)

		logger.Info(
			"Cluster members",
			"cluster-size", memberCount,
			"protocol-period", protocolPeriod+1,
			"min-cluster-size", minMemberCount+1,
		)
		if hasStableTail(observedMinMemberCount) {
			return nil
		}
	}
	return nil
}

func hasStableTail(s []int) bool {
	if len(s) < 5 {
		return false
	}

	// We stop as soon as we have a stable state for at least 5 protocol periods.
	for i := len(s) - 5; i < len(s); i++ {
		if s[i] != s[len(s)-1] {
			return false
		}
	}
	return true
}
