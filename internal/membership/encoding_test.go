package membership_test

import (
	"math"
	"net"
	"testing"

	"github.com/backbone81/membership/internal/membership"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Encoding", func() {
	Context("IP", func() {
		It("should append to nil buffer", func() {
			buffer, _, err := membership.AppendIPToBuffer(nil, net.IPv4(1, 2, 3, 4))
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())
		})

		It("should append to buffer", func() {
			var localBuffer [10]byte
			buffer, _, err := membership.AppendIPToBuffer(localBuffer[:0], net.IPv4(1, 2, 3, 4))
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())
		})

		It("should read from buffer", func() {
			appendIP := net.IPv4(1, 2, 3, 4)
			buffer, appendN, err := membership.AppendIPToBuffer(nil, appendIP)
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())

			readIP, readN, err := membership.IPFromBuffer(buffer)
			Expect(err).ToNot(HaveOccurred())
			Expect(readIP).ToNot(BeNil())

			Expect(appendN).To(Equal(readN))
			Expect(appendIP).To(Equal(readIP))
		})

		It("should fail to read from nil buffer", func() {
			Expect(membership.IPFromBuffer(nil)).Error().To(HaveOccurred())
		})

		It("should fail to read from buffer which is too small", func() {
			buffer, _, err := membership.AppendIPToBuffer(nil, net.IPv4(1, 2, 3, 4))
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())

			for i := len(buffer) - 1; i >= 0; i-- {
				Expect(membership.IPFromBuffer(buffer[:i])).Error().To(HaveOccurred())
			}
		})
	})

	Context("Port", func() {
		It("should append to nil buffer", func() {
			buffer, _, err := membership.AppendPortToBuffer(nil, 1024)
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())
		})

		It("should append to buffer", func() {
			var localBuffer [10]byte
			buffer, _, err := membership.AppendPortToBuffer(localBuffer[:0], 1024)
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())
		})

		DescribeTable("should append to buffer with valid port numbers",
			func(port int) {
				buffer, _, err := membership.AppendPortToBuffer(nil, port)
				Expect(err).ToNot(HaveOccurred())
				Expect(buffer).ToNot(BeNil())
			},
			Entry("first port", 1),
			Entry("privileged port", 80),
			Entry("user port", 3000),
		)

		DescribeTable("should fail to append to buffer with invalid port numbers",
			func(port int) {
				Expect(membership.AppendPortToBuffer(nil, port)).Error().To(HaveOccurred())
			},
			Entry("negative port", -10),
			Entry("any port", 0),
			Entry("too big of a port", math.MaxUint16+1),
		)

		It("should read from buffer", func() {
			appendPort := 1024
			buffer, appendN, err := membership.AppendPortToBuffer(nil, appendPort)
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())

			readPort, readN, err := membership.PortFromBuffer(buffer)
			Expect(err).ToNot(HaveOccurred())

			Expect(appendN).To(Equal(readN))
			Expect(appendPort).To(Equal(readPort))
		})

		It("should fail to read from nil buffer", func() {
			Expect(membership.PortFromBuffer(nil)).Error().To(HaveOccurred())
		})

		It("should fail to read from buffer which is too small", func() {
			buffer, _, err := membership.AppendPortToBuffer(nil, 1024)
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())

			for i := len(buffer) - 1; i >= 0; i-- {
				Expect(membership.PortFromBuffer(buffer[:i])).Error().To(HaveOccurred())
			}
		})
	})

	Context("SequenceNumber", func() {
		It("should append to nil buffer", func() {
			buffer, _, err := membership.AppendSequenceNumberToBuffer(nil, 1024)
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())
		})

		It("should append to buffer", func() {
			var localBuffer [10]byte
			buffer, _, err := membership.AppendSequenceNumberToBuffer(localBuffer[:0], 1024)
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())
		})

		DescribeTable("should append to buffer with valid sequence numbers",
			func(port int) {
				buffer, _, err := membership.AppendSequenceNumberToBuffer(nil, port)
				Expect(err).ToNot(HaveOccurred())
				Expect(buffer).ToNot(BeNil())
			},
			Entry("zero", 0),
			Entry("small positive", 80),
			Entry("big positive", 3000),
		)

		DescribeTable("should fail to append to buffer with invalid sequence numbers",
			func(port int) {
				Expect(membership.AppendSequenceNumberToBuffer(nil, port)).Error().To(HaveOccurred())
			},
			Entry("negative", -10),
			Entry("too big of a sequence number", math.MaxUint16+1),
		)

		It("should read from buffer", func() {
			appendSequenceNumber := 1024
			buffer, appendN, err := membership.AppendSequenceNumberToBuffer(nil, appendSequenceNumber)
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())

			readSequenceNumber, readN, err := membership.SequenceNumberFromBuffer(buffer)
			Expect(err).ToNot(HaveOccurred())

			Expect(appendN).To(Equal(readN))
			Expect(appendSequenceNumber).To(Equal(readSequenceNumber))
		})

		It("should fail to read from nil buffer", func() {
			Expect(membership.SequenceNumberFromBuffer(nil)).Error().To(HaveOccurred())
		})

		It("should fail to read from buffer which is too small", func() {
			buffer, _, err := membership.AppendSequenceNumberToBuffer(nil, 1024)
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())

			for i := len(buffer) - 1; i >= 0; i-- {
				Expect(membership.SequenceNumberFromBuffer(buffer[:i])).Error().To(HaveOccurred())
			}
		})
	})

	Context("IncarnationNumber", func() {
		It("should append to nil buffer", func() {
			buffer, _, err := membership.AppendIncarnationNumberToBuffer(nil, 1024)
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())
		})

		It("should append to buffer", func() {
			var localBuffer [10]byte
			buffer, _, err := membership.AppendIncarnationNumberToBuffer(localBuffer[:0], 1024)
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())
		})

		DescribeTable("should append to buffer with valid incarnation numbers",
			func(port int) {
				buffer, _, err := membership.AppendIncarnationNumberToBuffer(nil, port)
				Expect(err).ToNot(HaveOccurred())
				Expect(buffer).ToNot(BeNil())
			},
			Entry("zero", 0),
			Entry("small positive", 80),
			Entry("big positive", 3000),
		)

		DescribeTable("should fail to append to buffer with invalid incarnation numbers",
			func(port int) {
				Expect(membership.AppendIncarnationNumberToBuffer(nil, port)).Error().To(HaveOccurred())
			},
			Entry("negative", -10),
			Entry("too big of an incarnation number", math.MaxUint16+1),
		)

		It("should read from buffer", func() {
			appendIncarnationNumber := 1024
			buffer, appendN, err := membership.AppendIncarnationNumberToBuffer(nil, appendIncarnationNumber)
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())

			readIncarnationNumber, readN, err := membership.IncarnationNumberFromBuffer(buffer)
			Expect(err).ToNot(HaveOccurred())

			Expect(appendN).To(Equal(readN))
			Expect(appendIncarnationNumber).To(Equal(readIncarnationNumber))
		})

		It("should fail to read from nil buffer", func() {
			Expect(membership.IncarnationNumberFromBuffer(nil)).Error().To(HaveOccurred())
		})

		It("should fail to read from buffer which is too small", func() {
			buffer, _, err := membership.AppendIncarnationNumberToBuffer(nil, 1024)
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())

			for i := len(buffer) - 1; i >= 0; i-- {
				Expect(membership.IncarnationNumberFromBuffer(buffer[:i])).Error().To(HaveOccurred())
			}
		})
	})
})

func BenchmarkAppendIPToBuffer(b *testing.B) {
	var buffer [1024]byte
	ip := net.IPv4(1, 2, 3, 4)
	for b.Loop() {
		if _, _, err := membership.AppendIPToBuffer(buffer[:0], ip); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIPFromBuffer(b *testing.B) {
	buffer, _, err := membership.AppendIPToBuffer(nil, net.IPv4(1, 2, 3, 4))
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		if _, _, err := membership.IPFromBuffer(buffer); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAppendPortToBuffer(b *testing.B) {
	var buffer [1024]byte
	for b.Loop() {
		if _, _, err := membership.AppendPortToBuffer(buffer[:0], 1024); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPortFromBuffer(b *testing.B) {
	buffer, _, err := membership.AppendPortToBuffer(nil, 1024)
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		if _, _, err := membership.PortFromBuffer(buffer); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAppendSequenceNumberToBuffer(b *testing.B) {
	var buffer [1024]byte
	for b.Loop() {
		if _, _, err := membership.AppendSequenceNumberToBuffer(buffer[:0], 1024); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSequenceNumberFromBuffer(b *testing.B) {
	buffer, _, err := membership.AppendSequenceNumberToBuffer(nil, 1024)
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		if _, _, err := membership.SequenceNumberFromBuffer(buffer); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAppendIncarnationNumberToBuffer(b *testing.B) {
	var buffer [1024]byte
	for b.Loop() {
		if _, _, err := membership.AppendIncarnationNumberToBuffer(buffer[:0], 1024); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIncarnationNumberFromBuffer(b *testing.B) {
	buffer, _, err := membership.AppendIncarnationNumberToBuffer(nil, 1024)
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		if _, _, err := membership.IncarnationNumberFromBuffer(buffer); err != nil {
			b.Fatal(err)
		}
	}
}
