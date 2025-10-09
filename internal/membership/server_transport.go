package membership

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/go-logr/logr"
)

type ServerTransport struct {
	config     ServerTransportConfig
	list       *List
	logger     logr.Logger
	connection *net.UDPConn
	waitGroup  sync.WaitGroup
}

type ServerTransportConfig struct {
	Host                string
	ReceiveBufferLength int
	Logger              logr.Logger
}

func NewServerTransport(list *List, config ServerTransportConfig) *ServerTransport {
	return &ServerTransport{
		config: config,
		list:   list,
		logger: config.Logger,
	}
}

func (t *ServerTransport) Startup() error {
	t.logger.Info("Server transport startup")
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

func (t *ServerTransport) Shutdown() error {
	t.logger.Info("Server transport shutdown")
	if err := t.connection.Close(); err != nil {
		return err
	}
	t.waitGroup.Wait()
	return nil
}

func (t *ServerTransport) backgroundTask() {
	t.logger.Info("Server transport background task started")
	defer t.logger.Info("Server transport background task finished")

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
