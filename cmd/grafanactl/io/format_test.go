package io_test

import (
	goio "io"
	"testing"

	cmdio "github.com/grafana/grafanactl/cmd/grafanactl/io"
	"github.com/grafana/grafanactl/internal/agent"
	"github.com/grafana/grafanactl/internal/format"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBindFlags_AgentModeOverridesDefaultFormat(t *testing.T) {
	tests := []struct {
		name           string
		agentMode      bool
		defaultFormat  string
		explicitOutput string // simulates -o flag; empty = use default
		wantFormat     string
	}{
		{
			name:       "agent mode forces json when no command default set",
			agentMode:  true,
			wantFormat: "json",
		},
		{
			name:          "agent mode forces json when command sets text default",
			agentMode:     true,
			defaultFormat: "text",
			wantFormat:    "json",
		},
		{
			name:           "explicit -o yaml overrides agent mode json default",
			agentMode:      true,
			defaultFormat:  "text",
			explicitOutput: "yaml",
			wantFormat:     "yaml",
		},
		{
			name:          "no agent mode uses command default format",
			agentMode:     false,
			defaultFormat: "yaml",
			wantFormat:    "yaml",
		},
		{
			name:       "no agent mode uses json when no command default set",
			agentMode:  false,
			wantFormat: "json",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			agent.SetFlag(tc.agentMode)
			t.Cleanup(func() { agent.SetFlag(false) })

			opts := &cmdio.Options{}
			if tc.defaultFormat != "" {
				opts.DefaultFormat(tc.defaultFormat)
			}

			// Register a dummy text codec so "text" is a valid format.
			opts.RegisterCustomCodec("text", &dummyCodec{})

			flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
			opts.BindFlags(flags)

			if tc.explicitOutput != "" {
				require.NoError(t, flags.Set("output", tc.explicitOutput))
			}

			assert.Equal(t, tc.wantFormat, opts.OutputFormat)
		})
	}
}

// dummyCodec satisfies format.Codec for testing.
type dummyCodec struct{}

func (*dummyCodec) Encode(_ goio.Writer, _ any) error { return nil }
func (*dummyCodec) Decode(_ goio.Reader, _ any) error { return nil }
func (*dummyCodec) Format() format.Format             { return "text" }
