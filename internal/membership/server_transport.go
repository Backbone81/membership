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
		if err := t.dispatchMessages(buffer[:n]); err != nil {
			log.Printf("ERROR: %v\n", err)
		}
	}
}

func (t *ServerTransport) dispatchMessages(buffer []byte) error {
	for len(buffer) > 0 {
		messageType, messageTypeN, err := MessageTypeFromBuffer(buffer)
		if err != nil {
			return err
		}

		switch messageType {
		case MessageTypeDirectPing:
			var message MessageDirectPing
			n, err := message.FromBuffer(buffer[messageTypeN:])
			if err != nil {
				return err
			}
			buffer = buffer[messageTypeN+n:]
			t.config.List.HandleDirectPing(message)
		case MessageTypeDirectAck:
			var message MessageDirectAck
			n, err := message.FromBuffer(buffer[messageTypeN:])
			if err != nil {
				return err
			}
			buffer = buffer[messageTypeN+n:]
			t.config.List.HandleDirectAck(message)
		case MessageTypeIndirectPing:
			var message MessageIndirectPing
			n, err := message.FromBuffer(buffer[messageTypeN:])
			if err != nil {
				return err
			}
			buffer = buffer[messageTypeN+n:]
			t.config.List.HandleIndirectPing(message)
		case MessageTypeIndirectAck:
			var message MessageIndirectAck
			n, err := message.FromBuffer(buffer[messageTypeN:])
			if err != nil {
				return err
			}
			buffer = buffer[messageTypeN+n:]
			t.config.List.HandleIndirectAck(message)
		case MessageTypeSuspect:
			var message MessageSuspect
			n, err := message.FromBuffer(buffer[messageTypeN:])
			if err != nil {
				return err
			}
			buffer = buffer[messageTypeN+n:]
			t.config.List.HandleSuspect(message)
		case MessageTypeAlive:
			var message MessageAlive
			n, err := message.FromBuffer(buffer[messageTypeN:])
			if err != nil {
				return err
			}
			buffer = buffer[messageTypeN+n:]
			t.config.List.HandleAlive(message)
		case MessageTypeFaulty:
			var message MessageFaulty
			n, err := message.FromBuffer(buffer[messageTypeN:])
			if err != nil {
				return err
			}
			buffer = buffer[messageTypeN+n:]
			t.config.List.HandleFaulty(message)
		default:
			log.Printf("ERROR: Unknown message type %d.", messageType)
		}
	}
	return nil
}
