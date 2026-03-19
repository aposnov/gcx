package datasources

import (
	"github.com/grafana/grafanactl/cmd/grafanactl/datasources/query"
	"github.com/spf13/cobra"
)

func tempoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tempo",
		Short: "Tempo datasource operations",
		Long:  "Operations specific to Tempo datasources such as queries.",
	}

	cmd.AddCommand(query.TempoCmd())

	return cmd
}
