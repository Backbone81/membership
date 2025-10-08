package membership

import (
	"fmt"
	"log"
	"net"
	"sync"
)

type ServerTransport struct {
	config     ServerTransportConfig
	connection *net.UDPConn
	waitGroup  sync.WaitGroup
}

type ServerTransportConfig struct {
	Host                string
	ReceiveBufferLength int
	List                *List
}

func NewServerTransport(config ServerTransportConfig) *ServerTransport {
	return &ServerTransport{
		config: config,
	}
}

func (t *ServerTransport) Startup() error {
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
	if err := t.connection.Close(); err != nil {
		return err
	}
	t.waitGroup.Wait()
	return nil
}

func (t *ServerTransport) backgroundTask() {
	defer t.waitGroup.Done()
	buffer := make([]byte, t.config.ReceiveBufferLength)
	for {
		n, _, err := t.connection.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("ERROR: %v\n", err)
			return
		}
		if n < 1 {
			continue
		}
		if err := t.config.List.DispatchDatagram(buffer[:n]); err != nil {
			log.Printf("ERROR: %v\n", err)
		}
	}
}
