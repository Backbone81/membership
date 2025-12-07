package infection

import (
	"log"
	"math/rand"
	"os"

	"github.com/go-logr/stdr"
	"github.com/spf13/cobra"
)

var memberCount int

type Member struct {
	NextMembers []int
}

// failurePropagationCmd represents the allDetection command.
var failurePropagationCmd = &cobra.Command{
	Use:   "infection",
	Short: "Simulates information dissemination through infection.",
	Long: `Simulates information spreading in an infection style mechanic in clusters.
Measures the number of iterations for each member to learn about the information.
Note that this simulation does not run the membership list, but is some general purpose simulation for comparison.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := stdr.New(log.New(os.Stdout, "", log.LstdFlags))

		// Build up our cluster
		members := make([]Member, memberCount)
		for i := range members {
			member := &members[i]
			member.NextMembers = make([]int, 0, memberCount-1)
			for j := range members {
				if i == j {
					continue
				}
				member.NextMembers = append(member.NextMembers, j)
			}
			rand.Shuffle(len(member.NextMembers), func(i, j int) {
				member.NextMembers[i], member.NextMembers[j] = member.NextMembers[j], member.NextMembers[i]
			})
		}

		// The first member in the cluster has some specific information
		information := make(map[int]int, 1024)
		information[0] = 0
		for iteration := 1; iteration < 1000; iteration++ {
			// Every member spreads the information to one random other member.
			for i := range members {
				member := &members[i]

				nextMemberIndex := iteration % (len(members) - 1)
				if nextMemberIndex == 0 {
					rand.Shuffle(len(member.NextMembers), func(i, j int) {
						member.NextMembers[i], member.NextMembers[j] = member.NextMembers[j], member.NextMembers[i]
					})
				}

				if _, ok := information[i]; !ok {
					// The member does not have the information yet. No way to infect somebody.
					continue
				}

				if _, ok := information[member.NextMembers[nextMemberIndex]]; !ok {
					// The target was not yet infected, we infect it now.
					information[member.NextMembers[nextMemberIndex]] = iteration
				}
			}
			if len(information) == len(members) {
				// We are done when all members have the information.
				break
			}
		}

		// print the result.
		var maxIteration int
		for i := range members {
			maxIteration = max(maxIteration, information[i])
			logger.Info("Information received", "member", i, "iteration", information[i])
		}
		logger.Info("Maximum iterations needed", "iterations", maxIteration)
		return nil
	},
}

func RegisterSubCommand(command *cobra.Command) {
	command.AddCommand(failurePropagationCmd)

	failurePropagationCmd.PersistentFlags().IntVar(
		&memberCount,
		"member-count",
		512,
		"The number of members to simulate.",
	)
}
