package transport

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"io"
	"net"
	"strconv"
	"sync"

	"github.com/backbone81/membership/internal/encryption"
	"github.com/go-logr/logr"

	"github.com/backbone81/membership/internal/encoding"
)

// TCPServer provides reliable transport for receiving data from members.
//
// TCPServer is safe for concurrent use by multiple goroutines. Access to shared state is internally synchronized.
type TCPServer struct {
	logger      logr.Logger
	target      Target
	bindAddress string
	listener    net.Listener
	waitGroup   sync.WaitGroup

	mutex   sync.Mutex
	gcms    []cipher.AEAD
	buffers [][]byte
}

// NewTCPServer creates a new TCPServer transport.
func NewTCPServer(logger logr.Logger, target Target, bindAddress string, keys []encryption.Key) (*TCPServer, error) {
	var gcms []cipher.AEAD
	for _, key := range keys {
		aesCipher, err := aes.NewCipher(key[:])
		if err != nil {
			return nil, err
		}

		gcm, err := cipher.NewGCMWithRandomNonce(aesCipher)
		if err != nil {
			return nil, err
		}
		gcms = append(gcms, gcm)
	}
	return &TCPServer{
		logger:      logger,
		target:      target,
		bindAddress: bindAddress,
		gcms:        gcms,
		buffers:     make([][]byte, 0, 16),
	}, nil
}

// Startup starts the server and listens for incoming connections.
func (t *TCPServer) Startup() error {
	t.logger.Info("TCP server transport startup")
	listener, err := net.Listen("tcp", t.bindAddress)
	if err != nil {
		return err
	}
	t.listener = listener

	t.waitGroup.Go(func() {
		t.backgroundTask()
	})
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

// Addr returns the address the server is listening on.
func (t *TCPServer) Addr() (encoding.Address, error) {
	host, port, err := net.SplitHostPort(t.listener.Addr().String())
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

// backgroundTask is accepting connections and creating go routines to handle them.
func (t *TCPServer) backgroundTask() {
	t.logger.Info("TCP server transport background task started")
	defer t.logger.Info("TCP server transport background task finished")

	for {
		connection, err := t.listener.Accept()
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				t.logger.Error(err, "Accepting TCP connection.")
			}
			return
		}

		t.waitGroup.Go(func() {
			t.handleConnection(connection)
		})
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
	buffer := t.allocateBuffer()
	defer t.releaseBuffer(buffer)

	// First let's read the datagram length which is an uint32.
	n, err := io.ReadFull(connection, buffer[:4+encryption.Overhead])
	ReceiveBytes.WithLabelValues("tcp_server").Add(float64(n))
	if err != nil {
		ReceiveErrors.WithLabelValues("tcp_server").Inc()
		return err
	}
	datagramLength, err := t.decryptMessageLength(buffer[:n])
	if err != nil {
		return err
	}

	// Let's read the datagram payload.
	if len(buffer) < datagramLength+encryption.Overhead {
		buffer = make([]byte, datagramLength+encryption.Overhead)
	}
	n, err = io.ReadFull(connection, buffer[:datagramLength+encryption.Overhead])
	ReceiveBytes.WithLabelValues("tcp_server").Add(float64(n))
	if err != nil {
		ReceiveErrors.WithLabelValues("tcp_server").Inc()
		return err
	}

	if err := t.decryptAndDispatch(buffer[:n]); err != nil {
		t.logger.Error(err, "Dispatching TCP message")
	}
	return nil
}

func (t *TCPServer) decryptMessageLength(buffer []byte) (int, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	var plaintext [4]byte
	var joinedErr error
	for _, gcm := range t.gcms {
		var err error
		// Note that we are using the plaintext buffer for decrypting the buffer. In case the
		// decryption fails, the plaintext buffer is overwritten with garbage, so we cannot directly decrypt into
		// buffer.
		_, err = gcm.Open(plaintext[:0], nil, buffer, nil)
		Decryptions.WithLabelValues("tcp_server").Add(1)
		if err != nil {
			// Decryption failed. Let's try the next key.
			joinedErr = errors.Join(joinedErr, err)
			continue
		}
		return int(encoding.Endian.Uint32(plaintext[:4])), nil
	}
	joinedErr = errors.Join(joinedErr, errors.New("no encryption key could decrypt the network message"))
	return 0, joinedErr
}

func (t *TCPServer) decryptAndDispatch(buffer []byte) error {
	plaintext := t.allocateBuffer()
	defer t.releaseBuffer(plaintext)

	t.mutex.Lock()
	defer t.mutex.Unlock()

	if len(plaintext) < len(buffer) {
		plaintext = make([]byte, 0, len(buffer))
	}

	var joinedErr error
	for _, gcm := range t.gcms {
		var err error
		// Note that we are using the plaintext buffer for decrypting the buffer. In case the decryption fails, the
		// plaintext buffer is overwritten with garbage, so we cannot directly decrypt into buffer.
		plaintext, err = gcm.Open(plaintext[:0], nil, buffer, nil)
		Decryptions.WithLabelValues("tcp_server").Add(1)
		if err != nil {
			// Decryption failed. Let's try the next key.
			joinedErr = errors.Join(joinedErr, err)
			continue
		}
		return t.target.DispatchDatagram(plaintext)
	}
	joinedErr = errors.Join(joinedErr, errors.New("no encryption key could decrypt the network message"))
	return joinedErr
}

func (t *TCPServer) allocateBuffer() []byte {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if len(t.buffers) < 1 {
		return make([]byte, 16*1024)
	}

	result := t.buffers[len(t.buffers)-1]
	t.buffers = t.buffers[:len(t.buffers)-1]
	return result[:cap(result)]
}

func (t *TCPServer) releaseBuffer(buffer []byte) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.buffers = append(t.buffers, buffer)
}
