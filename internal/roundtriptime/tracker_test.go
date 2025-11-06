package roundtriptime_test

import (
	"testing"
	"time"

	"github.com/backbone81/membership/internal/roundtriptime"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tracker", func() {
	It("should correctly set the count", func() {
		tracker := roundtriptime.NewTracker(roundtriptime.WithCount(123))
		Expect(tracker.Config().Count).To(Equal(123))
	})

	It("should correctly set the percentile", func() {
		tracker := roundtriptime.NewTracker(roundtriptime.WithPercentile(0.123))
		Expect(tracker.Config().Percentile).To(Equal(0.123))
	})

	It("should correctly set the alpha", func() {
		tracker := roundtriptime.NewTracker(roundtriptime.WithAlpha(0.987))
		Expect(tracker.Config().Alpha).To(Equal(0.987))
	})

	It("should correctly set the default", func() {
		tracker := roundtriptime.NewTracker(roundtriptime.WithDefault(123 * time.Millisecond))
		Expect(tracker.Config().Default).To(Equal(123 * time.Millisecond))
	})

	It("should correctly set the minimum", func() {
		tracker := roundtriptime.NewTracker(roundtriptime.WithMinimum(888 * time.Millisecond))
		Expect(tracker.Config().Minimum).To(Equal(888 * time.Millisecond))
	})

	It("should correctly set the maximum", func() {
		tracker := roundtriptime.NewTracker(roundtriptime.WithMaximum(444 * time.Millisecond))
		Expect(tracker.Config().Maximum).To(Equal(444 * time.Millisecond))
	})

	It("should return the default value on a newly created tracker", func() {
		tracker := roundtriptime.NewTracker(roundtriptime.WithDefault(123 * time.Millisecond))
		tracker.UpdateCalculated()
		Expect(tracker.GetCalculated()).To(Equal(123 * time.Millisecond))
	})

	It("should calculate the RTT correctly without smoothing", func() {
		tracker := roundtriptime.NewTracker(roundtriptime.WithAlpha(1))
		for i := range tracker.Config().Count {
			tracker.AddObserved(time.Duration(i+1) * time.Millisecond)
		}
		tracker.UpdateCalculated()
		Expect(tracker.GetCalculated()).To(Equal(time.Duration(float64(tracker.Config().Count)*tracker.Config().Percentile) * time.Millisecond))
	})

	It("should calculate the RTT correctly with smoothing", func() {
		tracker := roundtriptime.NewTracker(roundtriptime.WithAlpha(0.3))

		// Fill low values
		for range tracker.Config().Count {
			tracker.AddObserved(1 * time.Millisecond)
		}
		tracker.UpdateCalculated()
		firstValue := tracker.GetCalculated()

		// Fill high values
		for range tracker.Config().Count {
			tracker.AddObserved(1 * time.Second)
		}
		tracker.UpdateCalculated()
		secondValue := tracker.GetCalculated()

		Expect(secondValue).To(BeNumerically(">", firstValue))
		Expect(secondValue).To(BeNumerically("<", 1*time.Second))
	})
})

func BenchmarkTracker_AddObserved(b *testing.B) {
	tracker := roundtriptime.NewTracker()
	for b.Loop() {
		tracker.AddObserved(100 * time.Millisecond)
	}
}

func BenchmarkTracker_UpdateCalculated(b *testing.B) {
	tracker := roundtriptime.NewTracker()
	for b.Loop() {
		tracker.UpdateCalculated()
	}
}

func BenchmarkTracker_GetCalculated(b *testing.B) {
	tracker := roundtriptime.NewTracker()
	for b.Loop() {
		tracker.GetCalculated()
	}
}
