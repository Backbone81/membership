package main

import (
	"fmt"
	"math"
	"time"

	"github.com/backbone81/membership/internal/utility"
)

func main() {
	for members := 1; members <= 64*1024; members *= 2 {
		fmt.Printf(
			"%d members, %f ping chance, %s failure detection\n",
			members,
			utility.PingTargetProbability(members, 0.8),
			utility.FailureDetectionDuration(1*time.Second, members, 0.8),
		)
	}
	fmt.Printf(
		"infinite members, %f ping chance, %s failure detection\n",
		utility.PingTargetProbabilityLimes(0.8),
		utility.FailureDetectionDurationLimes(1*time.Second, 0.8),
	)
	for networkReliability := 1.0; networkReliability > 0.98; networkReliability -= 0.001 {
		fmt.Printf("false positive chance network reliability %f: %f\n", networkReliability, utility.FailureDetectionFalsePositiveProbability(networkReliability, 1.0))
	}
	for members := 1; members <= 64*1024; members *= 2 {
		fmt.Printf(
			"%d members, %d protocol periods required\n",
			members,
			int(math.Ceil(utility.DisseminationPeriods(1.0, members))),
		)
	}
}
