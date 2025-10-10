package membership

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/go-logr/logr"
)

type UDPServerTransport struct {
	logger     logr.Logger
	config     UDPServerTransportConfig
	list       *List
	connection *net.UDPConn
	waitGroup  sync.WaitGroup
}

type UDPServerTransportConfig struct {
	Logger              logr.Logger
	Host                string
	ReceiveBufferLength int
}

func NewUDPServerTransport(list *List, config UDPServerTransportConfig) *UDPServerTransport {
	return &UDPServerTransport{
		logger: config.Logger,
		config: config,
		list:   list,
	}
}

func (t *UDPServerTransport) Startup() error {
	t.logger.Info("UDP server transport startup")
	addr, err := net.ResolveUDPAddr("udp", t.config.Host)
	if err != nil {
		return fmt.Errorf("resolving host: %w", err)
	}

	connection, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("listening for UDP: %w", err)
	}
	t.connection = connection

	t.waitGroup.Add(1)
	go t.backgroundTask()
	return nil
}

func (t *UDPServerTransport) Shutdown() error {
	t.logger.Info("UDP server transport shutdown")
	if err := t.connection.Close(); err != nil {
		return err
	}
	t.waitGroup.Wait()
	return nil
}

func (t *UDPServerTransport) backgroundTask() {
	t.logger.Info("UDP server transport background task started")
	defer t.logger.Info("UDP server transport background task finished")

	defer t.waitGroup.Done()
	buffer := make([]byte, t.config.ReceiveBufferLength)
	for {
		n, _, err := t.connection.ReadFromUDP(buffer)
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				t.logger.Error(err, "Reading UDP message.")
			}
			return
		}
		if n < 1 {
			continue
		}
		if err := t.list.DispatchDatagram(buffer[:n]); err != nil {
			t.logger.Error(err, "Dispatching UDP message.")
		}
	}
}
