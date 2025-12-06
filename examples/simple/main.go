package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-logr/stdr"

	"github.com/backbone81/membership/pkg/membership"
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

	bindAddress := membership.NewAddress(net.IPv4(127, 0, 0, 1), 3000)
	bootstrapMemberAddress := membership.NewAddress(net.IPv4(127, 0, 0, 1), 3001)
	membershipList, err := membership.NewList(
		membership.WithLogger(logger),
		membership.WithBootstrapMembers([]membership.Address{bootstrapMemberAddress}),
		membership.WithAdvertisedAddress(bindAddress),
		membership.WithBindAddress(bindAddress.String()),
		membership.WithEncryptionKey(membership.NewRandomKey()),
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
