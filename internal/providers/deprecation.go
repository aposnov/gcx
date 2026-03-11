package providers

import (
	"fmt"
	"os"

	"github.com/grafana/grafanactl/internal/agent"
	"github.com/spf13/cobra"
)

// WarnDeprecated prints a deprecation warning to stderr if not in agent or JSON output mode.
// It should be called from PersistentPreRun hooks on provider top-level commands.
//
// The warning is suppressed when:
//   - Agent mode is active (risk mitigation for CI/agent workflows).
//   - The --json flag has been set (risk mitigation for scripts parsing stdout).
//
// newCmd is the equivalent unified command the user should migrate to,
// e.g. "grafanactl resources list slo".
func WarnDeprecated(cmd *cobra.Command, newCmd string) {
	if agent.IsAgentMode() || jsonFlagActive(cmd) {
		return
	}

	fmt.Fprintf(os.Stderr, "Warning: '%s' is deprecated, use '%s' instead\n",
		cmd.CommandPath(), newCmd)
}

// jsonFlagActive returns true when any command in the ancestry chain has the
// --json flag set (i.e. the user explicitly passed --json on the command line).
// This avoids an import cycle with cmd/grafanactl/root while still detecting
// the flag regardless of where it was defined in the command tree.
func jsonFlagActive(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		if f := c.Flags().Lookup("json"); f != nil && f.Changed {
			return true
		}
		if f := c.PersistentFlags().Lookup("json"); f != nil && f.Changed {
			return true
		}
	}
	return false
}
