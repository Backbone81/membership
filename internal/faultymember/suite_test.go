package faultymember_test

import (
	"net"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/membership/internal/encoding"
	"github.com/backbone81/membership/internal/faultymember"
)

var (
	TestAddress  = encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024)
	TestAddress2 = encoding.NewAddress(net.IPv4(11, 12, 13, 14), 1024)
	TestAddress3 = encoding.NewAddress(net.IPv4(21, 22, 23, 24), 1024)
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FaultyMember Suite")
}

func GetFromListByIndex(list *faultymember.List, index int) encoding.Member {
	var counter int
	var result encoding.Member
	list.ForEach(func(member encoding.Member) bool {
		if counter == index {
			result = member
			return false
		}
		counter++
		return true
	})
	return result
}
