package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/backbone81/membership/internal/encoding"
	"github.com/backbone81/membership/internal/encryption"
	"github.com/backbone81/membership/pkg/membership"
	"github.com/go-logr/stdr"
)

func main() {
	if err := execute(); err != nil {
		log.Fatalln(err)
	}
}

func execute() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger := stdr.New(log.New(os.Stdout, "", log.LstdFlags))

	bindAddress := encoding.NewAddress(net.IPv4(127, 0, 0, 1), 3000)
	bootstrapMemberAddress := encoding.NewAddress(net.IPv4(127, 0, 0, 1), 3001)
	membershipList, err := membership.NewList(
		membership.WithLogger(logger),
		membership.WithBootstrapMembers([]encoding.Address{bootstrapMemberAddress}),
		membership.WithAdvertisedAddress(bindAddress),
		membership.WithBindAddress(bindAddress.String()),
		membership.WithEncryptionKey(encryption.NewRandomKey()),
	)
	if err != nil {
		return err
	}

	if err := membershipList.Startup(); err != nil {
		return err
	}
	<-ctx.Done()
	if err := membershipList.Shutdown(); err != nil {
		return err
	}
	return nil
}
