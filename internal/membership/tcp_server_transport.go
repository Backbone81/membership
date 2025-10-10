package membership

import (
	"errors"
	"io"
	"net"
	"sync"

	"github.com/go-logr/logr"
)

type TCPServerTransport struct {
	logger    logr.Logger
	config    TCPServerTransportConfig
	list      *List
	listener  net.Listener
	waitGroup sync.WaitGroup
}

type TCPServerTransportConfig struct {
	Logger logr.Logger
	Host   string
}

func NewTCPServerTransport(list *List, config TCPServerTransportConfig) *TCPServerTransport {
	return &TCPServerTransport{
		logger: config.Logger,
		config: config,
		list:   list,
	}
}

func (t *TCPServerTransport) Startup() error {
	t.logger.Info("TCP server transport startup")
	listener, err := net.Listen("tcp", t.config.Host)
	if err != nil {
		return err
	}
	t.listener = listener

	t.waitGroup.Add(1)
	go t.backgroundTask()
	return nil
}

func (t *TCPServerTransport) Shutdown() error {
	t.logger.Info("TCP server transport shutdown")
	if err := t.listener.Close(); err != nil {
		return err
	}
	t.waitGroup.Wait()
	return nil
}

func (t *TCPServerTransport) backgroundTask() {
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

func (t *TCPServerTransport) handleConnection(connection net.Conn) {
	if err := t.handleConnectionImpl(connection); err != nil {
		t.logger.Error(err, "Handling TCP connection")
	}
	if err := connection.Close(); err != nil {
		t.logger.Error(err, "Closing TCP connection")
	}
}

func (t *TCPServerTransport) handleConnectionImpl(connection net.Conn) error {
	defer t.waitGroup.Done()
	buffer := make([]byte, 1024)

	// First let's read the datagram length which is an uint32.
	if _, err := io.ReadFull(connection, buffer[:4]); err != nil {
		return err
	}
	datagramLength := int(Endian.Uint32(buffer[:4]))

	// Let's read the datagram payload.
	if len(buffer) < 4+datagramLength {
		// TODO: Simply growing the buffer to the size given in the first four bytes makes us prone to denial of service
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

	if err := t.list.DispatchDatagram(buffer[4 : 4+datagramLength]); err != nil {
		t.logger.Error(err, "Dispatching TCP message")
	}
	return nil
}
