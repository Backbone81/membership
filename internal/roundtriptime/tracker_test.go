package roundtriptime_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/membership/internal/roundtriptime"
)

var _ = Describe("Tracker", func() {
	It("should correctly set the count", func() {
		tracker := roundtriptime.NewTracker(roundtriptime.WithCount(123))
		Expect(tracker.Config().Count).To(Equal(123))
	})

	It("should clamp count to minimum 1", func() {
		tracker := roundtriptime.NewTracker(roundtriptime.WithCount(0))
		Expect(tracker.Config().Count).To(Equal(1))
	})

	It("should correctly set the percentile", func() {
		tracker := roundtriptime.NewTracker(roundtriptime.WithPercentile(0.123))
		Expect(tracker.Config().Percentile).To(Equal(0.123))
	})

	It("should clamp percentile to [0, 1] range", func() {
		tracker1 := roundtriptime.NewTracker(roundtriptime.WithPercentile(-0.5))
		Expect(tracker1.Config().Percentile).To(Equal(0.0))

		tracker2 := roundtriptime.NewTracker(roundtriptime.WithPercentile(1.5))
		Expect(tracker2.Config().Percentile).To(Equal(1.0))
	})

	It("should correctly set the alpha", func() {
		tracker := roundtriptime.NewTracker(roundtriptime.WithAlpha(0.987))
		Expect(tracker.Config().Alpha).To(Equal(0.987))
	})

	It("should clamp alpha to [0, 1] range", func() {
		tracker1 := roundtriptime.NewTracker(roundtriptime.WithAlpha(-0.5))
		Expect(tracker1.Config().Alpha).To(Equal(0.0))

		tracker2 := roundtriptime.NewTracker(roundtriptime.WithAlpha(1.5))
		Expect(tracker2.Config().Alpha).To(Equal(1.0))
	})

	It("should correctly set the default", func() {
		tracker := roundtriptime.NewTracker(roundtriptime.WithDefault(123 * time.Millisecond))
		Expect(tracker.Config().Default).To(Equal(123 * time.Millisecond))
	})

	It("should correctly set the minimum", func() {
		tracker := roundtriptime.NewTracker(roundtriptime.WithMinimum(2 * time.Millisecond))
		Expect(tracker.Config().Minimum).To(Equal(2 * time.Millisecond))
	})

	It("should correctly set the maximum", func() {
		tracker := roundtriptime.NewTracker(roundtriptime.WithMaximum(444 * time.Millisecond))
		Expect(tracker.Config().Maximum).To(Equal(444 * time.Millisecond))
	})

	It("should adjust minimum when maximum is smaller", func() {
		tracker := roundtriptime.NewTracker(
			roundtriptime.WithMinimum(100*time.Millisecond),
			roundtriptime.WithMaximum(50*time.Millisecond),
		)
		Expect(tracker.Config().Minimum).To(Equal(50 * time.Millisecond))
		Expect(tracker.Config().Maximum).To(Equal(50 * time.Millisecond))
	})

	It("should adjust default when below minimum", func() {
		tracker := roundtriptime.NewTracker(
			roundtriptime.WithDefault(10*time.Millisecond),
			roundtriptime.WithMinimum(50*time.Millisecond),
		)
		Expect(tracker.Config().Default).To(Equal(50 * time.Millisecond))
	})

	It("should adjust default when above maximum", func() {
		tracker := roundtriptime.NewTracker(
			roundtriptime.WithDefault(100*time.Millisecond),
			roundtriptime.WithMaximum(50*time.Millisecond),
		)
		Expect(tracker.Config().Default).To(Equal(50 * time.Millisecond))
	})

	It("should clamp calculated RTT to minimum", func() {
		tracker := roundtriptime.NewTracker(
			roundtriptime.WithMinimum(100*time.Millisecond),
			roundtriptime.WithAlpha(1.0), // No smoothing
		)

		// Add values well below minimum
		for range tracker.Config().Count {
			tracker.AddObserved(1 * time.Millisecond)
		}
		tracker.UpdateCalculated()

		Expect(tracker.GetCalculated()).To(Equal(100 * time.Millisecond))
	})

	It("should clamp calculated RTT to maximum", func() {
		tracker := roundtriptime.NewTracker(
			roundtriptime.WithMaximum(100*time.Millisecond),
			roundtriptime.WithAlpha(1.0), // No smoothing
		)

		// Add values well above maximum
		for range tracker.Config().Count {
			tracker.AddObserved(10 * time.Second)
		}
		tracker.UpdateCalculated()

		Expect(tracker.GetCalculated()).To(Equal(100 * time.Millisecond))
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

	It("should correctly overwrite old observations when buffer is full", func() {
		tracker := roundtriptime.NewTracker(
			roundtriptime.WithCount(5),
			roundtriptime.WithAlpha(1.0),      // No smoothing
			roundtriptime.WithPercentile(1.0), // Use maximum value
		)

		// Fill buffer with low values
		for i := 0; i < 5; i++ {
			tracker.AddObserved(10 * time.Millisecond)
		}
		tracker.UpdateCalculated()
		Expect(tracker.GetCalculated()).To(Equal(10 * time.Millisecond))

		// Add more values (should overwrite old ones)
		for i := 0; i < 5; i++ {
			tracker.AddObserved(100 * time.Millisecond)
		}
		tracker.UpdateCalculated()
		Expect(tracker.GetCalculated()).To(Equal(100 * time.Millisecond))
	})

	It("should reset to initial state", func() {
		tracker := roundtriptime.NewTracker(
			roundtriptime.WithDefault(50*time.Millisecond),
			roundtriptime.WithAlpha(1.0),
		)

		for range tracker.Config().Count {
			tracker.AddObserved(200 * time.Millisecond)
		}
		tracker.UpdateCalculated()
		Expect(tracker.GetCalculated()).To(Equal(200 * time.Millisecond))

		tracker.Reset()
		tracker.UpdateCalculated()

		Expect(tracker.GetCalculated()).To(Equal(50 * time.Millisecond))
	})

	It("should handle 0th percentile (minimum)", func() {
		tracker := roundtriptime.NewTracker(
			roundtriptime.WithCount(3),
			roundtriptime.WithPercentile(0.0),
			roundtriptime.WithAlpha(1.0),
		)

		tracker.AddObserved(100 * time.Millisecond)
		tracker.AddObserved(200 * time.Millisecond)
		tracker.AddObserved(300 * time.Millisecond)
		tracker.UpdateCalculated()

		// Should return minimum value
		Expect(tracker.GetCalculated()).To(Equal(100 * time.Millisecond))
	})

	It("should handle 100th percentile (maximum)", func() {
		tracker := roundtriptime.NewTracker(
			roundtriptime.WithCount(3),
			roundtriptime.WithPercentile(1.0),
			roundtriptime.WithAlpha(1.0),
		)

		tracker.AddObserved(100 * time.Millisecond)
		tracker.AddObserved(200 * time.Millisecond)
		tracker.AddObserved(300 * time.Millisecond)
		tracker.UpdateCalculated()

		// Should return maximum value
		Expect(tracker.GetCalculated()).To(Equal(300 * time.Millisecond))
	})

	It("should handle median (50th percentile)", func() {
		tracker := roundtriptime.NewTracker(
			roundtriptime.WithCount(5),
			roundtriptime.WithPercentile(0.5),
			roundtriptime.WithAlpha(1.0),
		)

		tracker.AddObserved(10 * time.Millisecond)
		tracker.AddObserved(20 * time.Millisecond)
		tracker.AddObserved(30 * time.Millisecond)
		tracker.AddObserved(40 * time.Millisecond)
		tracker.AddObserved(50 * time.Millisecond)
		tracker.UpdateCalculated()

		// Median of [10, 20, 30, 40, 50] = 30
		Expect(tracker.GetCalculated()).To(Equal(30 * time.Millisecond))
	})

	It("should not smooth with alpha=0 (uses only old value)", func() {
		tracker := roundtriptime.NewTracker(
			roundtriptime.WithDefault(100*time.Millisecond),
			roundtriptime.WithAlpha(0.0),
		)

		for range tracker.Config().Count {
			tracker.AddObserved(500 * time.Millisecond)
		}
		tracker.UpdateCalculated()

		Expect(tracker.GetCalculated()).To(Equal(100 * time.Millisecond))
	})

	It("should fully update with alpha=1 (ignores old value)", func() {
		tracker := roundtriptime.NewTracker(
			roundtriptime.WithDefault(100*time.Millisecond),
			roundtriptime.WithMaximum(500*time.Millisecond),
			roundtriptime.WithAlpha(1.0),
		)

		for range tracker.Config().Count {
			tracker.AddObserved(500 * time.Millisecond)
		}
		tracker.UpdateCalculated()

		Expect(tracker.GetCalculated()).To(Equal(500 * time.Millisecond))
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
