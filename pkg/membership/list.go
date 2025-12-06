package membership

import (
	"errors"

	"github.com/backbone81/membership/internal/encoding"
	intmembership "github.com/backbone81/membership/internal/membership"
	"github.com/backbone81/membership/internal/roundtriptime"
	intscheduler "github.com/backbone81/membership/internal/scheduler"
	inttransport "github.com/backbone81/membership/internal/transport"
)

type List struct {
	list               *intmembership.List
	scheduler          *intscheduler.Scheduler
	udpServerTransport *inttransport.UDPServer
	tcpServerTransport *inttransport.TCPServer
}

//nolint:funlen
func NewList(options ...Option) (*List, error) {
	config := DefaultConfig
	for _, option := range options {
		option(&config)
	}
	if len(config.EncryptionKeys) < 1 {
		return nil, errors.New("encryption key missing")
	}

	// The maximum round trip time is derived from 90% of the protocol period to allow for some leeway, then divided
	// by three, because a ping and indirect ping require three round trips in total to complete.
	maxRTT := config.ProtocolPeriod * 90 / 100 / 3
	defaultRTT := maxRTT / 2
	rttTracker := roundtriptime.NewTracker(
		roundtriptime.WithDefault(defaultRTT),
		roundtriptime.WithMaximum(maxRTT),
	)
	udpClientTransport, err := inttransport.NewUDPClient(config.MaxDatagramLengthSend, config.EncryptionKeys[0])
	if err != nil {
		return nil, err
	}
	tcpClientTransport, err := inttransport.NewTCPClient(config.EncryptionKeys[0])
	if err != nil {
		return nil, err
	}
	list := intmembership.NewList(
		intmembership.WithLogger(config.Logger),
		intmembership.WithBootstrapMembers(config.BootstrapMembers),
		intmembership.WithAdvertisedAddress(config.AdvertisedAddress),
		intmembership.WithMaxDatagramLengthSend(config.MaxDatagramLengthSend),
		intmembership.WithUDPClient(udpClientTransport),
		intmembership.WithTCPClient(tcpClientTransport),
		intmembership.WithMemberAddedCallback(config.MemberAddedCallback),
		intmembership.WithMemberRemovedCallback(config.MemberRemovedCallback),
		intmembership.WithSafetyFactor(config.SafetyFactor),
		intmembership.WithShutdownMemberCount(config.ShutdownMemberCount),
		intmembership.WithDirectPingMemberCount(config.DirectPingMemberCount),
		intmembership.WithIndirectPingMemberCount(config.IndirectPingMemberCount),
		intmembership.WithRoundTripTimeTracker(rttTracker),
		intmembership.WithReconnectBootstrapMembers(config.ReconnectBootstrapMembers),
	)
	udpServerTransport, err := inttransport.NewUDPServer(config.Logger, list, config.BindAddress, config.MaxDatagramLengthReceive, config.EncryptionKeys)
	if err != nil {
		return nil, err
	}
	tcpServerTransport, err := inttransport.NewTCPServer(config.Logger, list, config.BindAddress, config.EncryptionKeys)
	if err != nil {
		return nil, err
	}
	scheduler := intscheduler.New(
		list,
		intscheduler.WithLogger(config.Logger),
		intscheduler.WithProtocolPeriod(config.ProtocolPeriod),
		intscheduler.WithMaxSleepDuration(config.MaxSleepDuration),
		intscheduler.WithListRequestInterval(config.ListRequestInterval),
		intscheduler.WithRoundTripTimeTracker(rttTracker),
	)

	newList := List{
		list:               list,
		udpServerTransport: udpServerTransport,
		tcpServerTransport: tcpServerTransport,
		scheduler:          scheduler,
	}
	return &newList, nil
}

func (l *List) Startup() error {
	if err := l.udpServerTransport.Startup(); err != nil {
		return err
	}
	if err := l.tcpServerTransport.Startup(); err != nil {
		return err
	}
	if err := l.scheduler.Startup(); err != nil {
		return err
	}
	return nil
}

func (l *List) Shutdown() error {
	if err := l.scheduler.Shutdown(); err != nil {
		return err
	}
	if err := l.tcpServerTransport.Shutdown(); err != nil {
		return err
	}
	if err := l.udpServerTransport.Shutdown(); err != nil {
		return err
	}
	if err := l.list.BroadcastShutdown(); err != nil {
		return err
	}
	return nil
}

func (l *List) Len() int {
	return l.list.Len()
}

func (l *List) ForEach(fn func(encoding.Address) bool) {
	l.list.ForEach(fn)
}
