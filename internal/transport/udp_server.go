package transport

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/go-logr/logr"

	"github.com/backbone81/membership/internal/encoding"
)

// UDPServer provides unreliable transport for receiving data from members.
type UDPServer struct {
	logger              logr.Logger
	target              Target
	bindAddress         string
	connection          *net.UDPConn
	waitGroup           sync.WaitGroup
	receiveBufferLength int
}

// NewUDPServer creates a new UDPServer.
func NewUDPServer(logger logr.Logger, target Target, bindAddress string, receiveBufferLength int) *UDPServer {
	return &UDPServer{
		logger:              logger,
		target:              target,
		bindAddress:         bindAddress,
		receiveBufferLength: receiveBufferLength,
	}
}

// Startup starts the server and listens for incoming data.
func (t *UDPServer) Startup() error {
	t.logger.Info("UDP server transport startup")
	addr, err := net.ResolveUDPAddr("udp", t.bindAddress)
	if err != nil {
		return fmt.Errorf("resolving host: %w", err)
	}

	connection, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("listening for UDP: %w", err)
	}
	t.connection = connection

	t.waitGroup.Go(func() {
		t.backgroundTask()
	})
	return nil
}

// Shutdown ends the server and waits for all data to be processed.
func (t *UDPServer) Shutdown() error {
	t.logger.Info("UDP server transport shutdown")
	if err := t.connection.Close(); err != nil {
		return err
	}
	t.waitGroup.Wait()
	return nil
}

// Addr returns the address the server is listening on.
func (t *UDPServer) Addr() (encoding.Address, error) {
	host, port, err := net.SplitHostPort(t.connection.LocalAddr().String())
	if err != nil {
		return encoding.Address{}, err
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return encoding.Address{}, errors.New("not an ip address")
	}
	typedPort, err := strconv.Atoi(port)
	if err != nil {
		return encoding.Address{}, err
	}

	return encoding.NewAddress(ip, typedPort), nil
}

func (t *UDPServer) backgroundTask() {
	t.logger.Info("UDP server transport background task started")
	defer t.logger.Info("UDP server transport background task finished")

	buffer := make([]byte, t.receiveBufferLength)
	for {
		n, _, err := t.connection.ReadFromUDP(buffer)
		ReceiveBytes.WithLabelValues("udp_server").Add(float64(n))
		if err != nil {
			ReceiveErrors.WithLabelValues("udp_server").Inc()
			if !errors.Is(err, net.ErrClosed) {
				t.logger.Error(err, "Reading UDP message.")
			}
			return
		}
		if n < 1 {
			continue
		}
		if err := t.target.DispatchDatagram(buffer[:n]); err != nil {
			t.logger.Error(err, "Dispatching UDP message.")
		}
	}
}
