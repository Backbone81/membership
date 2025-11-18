package membership

import (
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

func NewList(options ...Option) *List {
	config := DefaultConfig
	for _, option := range options {
		option(&config)
	}

	rttTracker := roundtriptime.NewTracker(
		// The maximum round trip time is derived from 90% of the protocol period to allow for some leeway, then divided
		// by three, because a ping and indirect ping require three round trips in total to complete.
		roundtriptime.WithMaximum(config.ProtocolPeriod * 90 / 100 / 3),
	)
	list := intmembership.NewList(
		intmembership.WithLogger(config.Logger),
		intmembership.WithBootstrapMembers(config.BootstrapMembers),
		intmembership.WithAdvertisedAddress(config.AdvertisedAddress),
		intmembership.WithMaxDatagramLengthSend(config.MaxDatagramLengthSend),
		intmembership.WithUDPClient(inttransport.NewUDPClient(config.MaxDatagramLengthSend)),
		intmembership.WithTCPClient(inttransport.NewTCPClient()),
		intmembership.WithMemberAddedCallback(config.MemberAddedCallback),
		intmembership.WithMemberRemovedCallback(config.MemberRemovedCallback),
		intmembership.WithSafetyFactor(config.SafetyFactor),
		intmembership.WithShutdownMemberCount(config.ShutdownMemberCount),
		intmembership.WithDirectPingMemberCount(config.DirectPingMemberCount),
		intmembership.WithIndirectPingMemberCount(config.IndirectPingMemberCount),
		intmembership.WithRoundTripTimeTracker(rttTracker),
	)
	udpServerTransport := inttransport.NewUDPServer(config.Logger, list, config.BindAddress, config.MaxDatagramLengthReceive)
	tcpServerTransport := inttransport.NewTCPServer(config.Logger, list, config.BindAddress)
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
	return &newList
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

func (l *List) Get() []encoding.Address {
	return l.list.Get()
}
