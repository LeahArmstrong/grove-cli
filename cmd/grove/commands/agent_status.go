package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/LeahArmstrong/grove-cli/internal/cli"
	"github.com/LeahArmstrong/grove-cli/plugins/docker"
)

var agentStatusJSON bool

type agentSlotOutput struct {
	Slot    int    `json:"slot"`
	Name    string `json:"worktree"`
	Project string `json:"composeProject"`
	URL     string `json:"url,omitempty"`
}

func init() {
	agentStatusCmd.Flags().BoolVarP(&agentStatusJSON, "json", "j", false, "Output as JSON")
	rootCmd.AddCommand(agentStatusCmd)
}

var agentStatusCmd = &cobra.Command{
	Use:   "agent-status",
	Short: "Show active isolated stacks",
	Long: `List all active isolated (agent) stacks, their slot numbers, and URLs.

Each isolated stack started with 'grove up --isolated' occupies a numbered slot
and runs its own containers independently.

Examples:
  grove agent-status          # Show active stacks
  grove agent-status --json   # Machine-readable output`,
	RunE: RequireGroveContext(func(cmd *cobra.Command, args []string, ctx *GroveContext) error {
		slots, err := docker.ListActiveSlots(ctx.Config)
		if err != nil {
			if agentStatusJSON {
				fmt.Println("[]")
				return nil
			}
			fmt.Println("Isolated stacks not configured for this project")
			fmt.Println("\nTo enable, add to .grove/config.toml:")
			fmt.Println("\n  [plugins.docker.external.agent]")
			fmt.Println("  enabled = true")
			fmt.Println("  services = [\"app\"]")
			fmt.Println("  template_path = \"agent-stacks/template.yml\"")
			return nil
		}

		if agentStatusJSON {
			out := make([]agentSlotOutput, 0, len(slots))
			for _, s := range slots {
				out = append(out, agentSlotOutput{
					Slot:    s.Slot,
					Name:    s.Worktree,
					Project: docker.AgentComposeProjectName(ctx.Config, s.Slot),
					URL:     docker.AgentURL(ctx.Config, s.Slot),
				})
			}
			data, err := json.MarshalIndent(out, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		w := cli.NewStdout()

		if len(slots) == 0 {
			cli.Info(w, "No active isolated stacks")
			return nil
		}

		maxSlots := 5
		if ctx.Config != nil && ctx.Config.Plugins.Docker.External != nil &&
			ctx.Config.Plugins.Docker.External.Agent != nil &&
			ctx.Config.Plugins.Docker.External.Agent.MaxSlots > 0 {
			maxSlots = ctx.Config.Plugins.Docker.External.Agent.MaxSlots
		}

		cli.Header(w, "Active isolated stacks (%d/%d slots)", len(slots), maxSlots)

		for _, s := range slots {
			project := docker.AgentComposeProjectName(ctx.Config, s.Slot)
			url := docker.AgentURL(ctx.Config, s.Slot)

			cli.Bold(w, "  Slot %d: %s", s.Slot, s.Worktree)
			cli.Label(w, "          Project:", project)
			if url != "" {
				cli.Label(w, "          URL:    ", url)
			}
		}

		return nil
	}),
}
