package membership

import (
	"github.com/backbone81/membership/internal/encoding"
	intmembership "github.com/backbone81/membership/internal/membership"
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

	list := intmembership.NewList(
		intmembership.WithLogger(config.Logger),
		intmembership.WithProtocolPeriod(config.ProtocolPeriod),
		intmembership.WithDirectPingTimeout(config.DirectPingTimeout),
		intmembership.WithBootstrapMembers(config.BootstrapMembers),
		intmembership.WithAdvertisedAddress(config.AdvertisedAddress),
		intmembership.WithMaxDatagramLength(config.MaxDatagramLength),
		intmembership.WithUDPClient(inttransport.NewUDPClient(config.MaxDatagramLength)),
		intmembership.WithTCPClient(inttransport.NewTCPClient()),
		intmembership.WithMemberAddedCallback(config.MemberAddedCallback),
		intmembership.WithMemberRemovedCallback(config.MemberRemovedCallback),
	)
	udpServerTransport := inttransport.NewUDPServer(config.Logger, list, config.BindAddress, config.MaxDatagramLength)
	tcpServerTransport := inttransport.NewTCPServer(config.Logger, list, config.BindAddress)
	scheduler := intscheduler.New(
		list,
		intscheduler.WithLogger(config.Logger),
		intscheduler.WithProtocolPeriod(config.ProtocolPeriod),
		intscheduler.WithDirectPingTimeout(config.DirectPingTimeout),
		intscheduler.WithMaxSleepDuration(config.MaxSleepDuration),
		intscheduler.WithListRequestInterval(config.ListRequestInterval),
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
	if err := l.Startup(); err != nil {
		return err
	}
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
	if err := l.Shutdown(); err != nil {
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
