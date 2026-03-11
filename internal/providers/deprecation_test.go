package providers_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/grafana/grafanactl/internal/agent"
	"github.com/grafana/grafanactl/internal/providers"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureStderr redirects os.Stderr to a buffer for the duration of f,
// then restores it and returns what was written.
func captureStderr(t *testing.T, f func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	require.NoError(t, err)

	origStderr := os.Stderr
	os.Stderr = w

	f()

	w.Close()
	os.Stderr = origStderr

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)

	return buf.String()
}

// clearAgentEnvForDeprecation unsets all agent-mode env vars so host env
// (e.g. CLAUDECODE=1 in Claude Code) does not leak into test cases.
func clearAgentEnvForDeprecation(t *testing.T) {
	t.Helper()

	for _, env := range []string{
		"GRAFANACTL_AGENT_MODE",
		"CLAUDECODE",
		"CLAUDE_CODE",
		"CURSOR_AGENT",
		"GITHUB_COPILOT",
		"AMAZON_Q",
	} {
		t.Setenv(env, "")
	}

	agent.ResetForTesting()
}

// newCmdWithPath builds a minimal cobra.Command hierarchy that satisfies
// cmd.CommandPath() to look like "grafanactl slo definitions list".
func newCmdWithPath(use string) *cobra.Command {
	root := &cobra.Command{Use: "grafanactl"}
	parent := &cobra.Command{Use: use}
	child := &cobra.Command{Use: "definitions"}
	leaf := &cobra.Command{Use: "list"}

	root.AddCommand(parent)
	parent.AddCommand(child)
	child.AddCommand(leaf)

	return leaf
}

func TestWarnDeprecated(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		setFlag     *bool
		jsonChanged bool
		wantOutput  bool
		newCmd      string
	}{
		{
			name:       "prints warning when no agent mode and no json flag",
			wantOutput: true,
			newCmd:     "grafanactl resources list slo",
		},
		{
			name:       "suppresses warning in agent mode via env var",
			envVars:    map[string]string{"GRAFANACTL_AGENT_MODE": "1"},
			wantOutput: false,
			newCmd:     "grafanactl resources list slo",
		},
		{
			name:       "suppresses warning when CLAUDE_CODE env var set",
			envVars:    map[string]string{"CLAUDE_CODE": "1"},
			wantOutput: false,
			newCmd:     "grafanactl resources list slo",
		},
		{
			name:        "suppresses warning when --json flag is active",
			jsonChanged: true,
			wantOutput:  false,
			newCmd:      "grafanactl resources list slo",
		},
		{
			name:        "prints warning when --json flag exists but not changed",
			jsonChanged: false,
			wantOutput:  true,
			newCmd:      "grafanactl resources list slo",
		},
		{
			name:       "warning message includes newCmd",
			wantOutput: true,
			newCmd:     "grafanactl resources list checks",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clearAgentEnvForDeprecation(t)

			for k, v := range tc.envVars {
				t.Setenv(k, v)
			}

			agent.ResetForTesting()

			if tc.setFlag != nil {
				agent.SetFlag(*tc.setFlag)
			}

			// Build a leaf command so cmd.CommandPath() returns a realistic path.
			cmd := newCmdWithPath("slo")

			// Add a --json flag and optionally mark it as changed.
			cmd.Flags().Bool("json", false, "JSON output")
			if tc.jsonChanged {
				require.NoError(t, cmd.Flags().Set("json", "true"))
			}

			got := captureStderr(t, func() {
				providers.WarnDeprecated(cmd, tc.newCmd)
			})

			if tc.wantOutput {
				assert.Contains(t, got, "is deprecated")
				assert.Contains(t, got, tc.newCmd)
			} else {
				assert.Empty(t, got)
			}
		})
	}
}

func TestWarnDeprecatedMessageFormat(t *testing.T) {
	clearAgentEnvForDeprecation(t)

	cmd := newCmdWithPath("slo")

	got := captureStderr(t, func() {
		providers.WarnDeprecated(cmd, "grafanactl resources list slo")
	})

	// Verify the exact format: "Warning: '<command-path>' is deprecated, use '<newCmd>' instead\n"
	want := fmt.Sprintf(
		"Warning: '%s' is deprecated, use '%s' instead\n",
		cmd.CommandPath(),
		"grafanactl resources list slo",
	)
	assert.Equal(t, want, got)
}

func TestWarnDeprecatedJSONFlagInAncestor(t *testing.T) {
	// The json flag may be defined on a parent command; WarnDeprecated should
	// still suppress the warning if any ancestor has it set.
	clearAgentEnvForDeprecation(t)

	root := &cobra.Command{Use: "grafanactl"}
	parent := &cobra.Command{Use: "slo"}
	child := &cobra.Command{Use: "definitions"}
	leaf := &cobra.Command{Use: "list"}

	root.AddCommand(parent)
	parent.AddCommand(child)
	child.AddCommand(leaf)

	// Put the --json flag on the child (not on leaf directly).
	child.Flags().Bool("json", false, "JSON output")
	require.NoError(t, child.Flags().Set("json", "true"))

	got := captureStderr(t, func() {
		providers.WarnDeprecated(leaf, "grafanactl resources list slo")
	})

	assert.Empty(t, got, "warning should be suppressed when a parent command has --json set")
}
