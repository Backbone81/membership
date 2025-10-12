package transport

import (
	"errors"
	"io"
	"net"
	"sync"

	"github.com/backbone81/membership/internal/membership"
	"github.com/go-logr/logr"
)

// TCPServer provides reliable transport for receiving data from members.
type TCPServer struct {
	logger      logr.Logger
	target      Target
	bindAddress string
	listener    net.Listener
	waitGroup   sync.WaitGroup
}

// NewTCPServer creates a new TCPServer transport.
func NewTCPServer(logger logr.Logger, target Target, bindAddress string) *TCPServer {
	return &TCPServer{
		logger:      logger,
		target:      target,
		bindAddress: bindAddress,
	}
}

// Startup starts the server and listens for incoming connections.
func (t *TCPServer) Startup() error {
	t.logger.Info("TCP server transport startup")
	listener, err := net.Listen("tcp", t.bindAddress)
	if err != nil {
		return err
	}
	t.listener = listener

	t.waitGroup.Add(1)
	go t.backgroundTask()
	return nil
}

// Shutdown ends the server and waits for all connections to be closed.
func (t *TCPServer) Shutdown() error {
	t.logger.Info("TCP server transport shutdown")
	if err := t.listener.Close(); err != nil {
		return err
	}
	t.waitGroup.Wait()
	return nil
}

// backgroundTask is accepting connections and creating go routines to handle them.
func (t *TCPServer) backgroundTask() {
	t.logger.Info("TCP server transport background task started")
	defer t.logger.Info("TCP server transport background task finished")

	defer t.waitGroup.Done()
	for {
		connection, err := t.listener.Accept()
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				t.logger.Error(err, "Accepting TCP connection.")
			}
			return
		}

		t.waitGroup.Add(1)
		go t.handleConnection(connection)
	}
}

// handleConnection is handling a single connection.
func (t *TCPServer) handleConnection(connection net.Conn) {
	if err := t.handleConnectionImpl(connection); err != nil {
		t.logger.Error(err, "Handling TCP connection")
	}
	if err := connection.Close(); err != nil {
		t.logger.Error(err, "Closing TCP connection")
	}
}

func (t *TCPServer) handleConnectionImpl(connection net.Conn) error {
	defer t.waitGroup.Done()
	buffer := make([]byte, 1024)

	// First let's read the datagram length which is an uint32.
	if _, err := io.ReadFull(connection, buffer[:4]); err != nil {
		return err
	}
	datagramLength := int(membership.Endian.Uint32(buffer[:4]))

	// Let's read the datagram payload.
	if len(buffer) < 4+datagramLength {
		// TODO: Simply growing the buffer to the length given in the first four bytes makes us prone to denial of service
		// attacks, when the port is exposed on the internet. Arbitrary payload might cause memory allocations of up
		// to 4 GB which might lead to an out of memory kill. We should only grow the buffer if we know that the message
		// was sent by a trustworthy member. Find a way to incorporate encryption/signing here, to prevent such an
		// issue.
		newBuffer := make([]byte, 4+datagramLength)
		copy(newBuffer, buffer[:4])
		buffer = newBuffer
	}
	if _, err := io.ReadFull(connection, buffer[4:4+datagramLength]); err != nil {
		return err
	}

	if err := t.target.DispatchDatagram(buffer[4 : 4+datagramLength]); err != nil {
		t.logger.Error(err, "Dispatching TCP message")
	}
	return nil
}
