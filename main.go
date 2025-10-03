package main

import (
	"fmt"
	"time"

	"github.com/backbone81/membership/internal/membership"
)

func main() {
	for members := 1; members < 1_000_000; members *= 2 {
		fmt.Printf(
			"%d members, %f ping chance, %s failure detection\n",
			members,
			membership.PingTargetProbability(members, 0.8),
			membership.FailureDetectionDuration(2*time.Second, members, 0.8),
		)
	}
	fmt.Printf(
		"infinite members, %f ping chance, %s failure detection\n",
		membership.PingTargetProbabilityLimes(0.8),
		membership.FailureDetectionDurationLimes(2*time.Second, 0.8),
	)
	for networkReliability := 1.0; networkReliability > 0.98; networkReliability -= 0.001 {
		fmt.Printf("false positive chance network reliability %f: %f\n", networkReliability, membership.FailureDetectionFalsePositiveProbability(networkReliability, 1.0))
	}
}
