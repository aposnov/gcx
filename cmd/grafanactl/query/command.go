package query

import (
	"errors"
	"fmt"
	"io"
	"time"

	cmdconfig "github.com/grafana/grafanactl/cmd/grafanactl/config"
	cmdio "github.com/grafana/grafanactl/cmd/grafanactl/io"
	"github.com/grafana/grafanactl/internal/format"
	"github.com/grafana/grafanactl/internal/query/loki"
	"github.com/grafana/grafanactl/internal/query/prometheus"
	"github.com/grafana/grafanactl/internal/query/pyroscope"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type queryOpts struct {
	IO          cmdio.Options
	Datasource  string
	Type        string
	Query       string
	Start       string
	End         string
	Step        string
	ProfileType string // Pyroscope-specific
	MaxNodes    int64  // Pyroscope-specific
}

func (opts *queryOpts) setup(flags *pflag.FlagSet) {
	opts.IO.RegisterCustomCodec("table", &queryTableCodec{})
	opts.IO.RegisterCustomCodec("graph", &queryGraphCodec{})
	opts.IO.DefaultFormat("table")
	opts.IO.BindFlags(flags)

	flags.StringVarP(&opts.Datasource, "datasource", "d", "", "Datasource UID (required unless default-{type}-datasource is configured)")
	flags.StringVarP(&opts.Type, "type", "t", "prometheus", "Datasource type (prometheus, loki, pyroscope)")
	flags.StringVarP(&opts.Query, "expr", "e", "", "Query expression (PromQL for prometheus, LogQL for loki, label selector for pyroscope)")
	flags.StringVar(&opts.Start, "start", "", "Start time (RFC3339, Unix timestamp, or relative like 'now-1h')")
	flags.StringVar(&opts.End, "end", "", "End time (RFC3339, Unix timestamp, or relative like 'now')")
	flags.StringVar(&opts.Step, "step", "", "Query step (e.g., '15s', '1m')")
	flags.StringVar(&opts.ProfileType, "profile-type", "", "Profile type ID for pyroscope queries (e.g., 'process_cpu:cpu:nanoseconds:cpu:nanoseconds')")
	flags.Int64Var(&opts.MaxNodes, "max-nodes", 1024, "Maximum nodes in flame graph (pyroscope only)")
}

func (opts *queryOpts) Validate() error {
	if err := opts.IO.Validate(); err != nil {
		return err
	}

	if opts.Query == "" {
		return errors.New("query expression is required (use -e or --expr)")
	}

	if opts.Type == "pyroscope" && opts.ProfileType == "" {
		return errors.New("profile type is required for pyroscope queries (use --profile-type)")
	}

	return nil
}

// Command returns the query command group.
func Command() *cobra.Command {
	configOpts := &cmdconfig.Options{}
	opts := &queryOpts{}

	cmd := &cobra.Command{
		Use:   "query",
		Short: "Execute queries against Grafana datasources",
		Long:  "Execute queries against Grafana datasources via the unified query API.",
		Example: `
	# First, find your datasource UID
	grafanactl datasources list

	# Prometheus instant query (use the UID from datasources list, not the name)
	grafanactl query -d <datasource-uid> -e 'up{job="grafana"}'

	# Prometheus range query
	grafanactl query -d <datasource-uid> -e 'rate(http_requests_total[5m])' --start now-1h --end now --step 1m

	# Loki log query (instant)
	grafanactl query -d <loki-uid> -t loki -e '{job="varlogs"}'

	# Loki log query (range)
	grafanactl query -d <loki-uid> -t loki -e '{name="private-datasource-connect"}' --start now-1h --end now

	# Loki metric query (log rate)
	grafanactl query -d <loki-uid> -t loki -e 'sum(rate({job="varlogs"}[5m]))' --start now-1h --end now --step 1m

	# Pyroscope profile query (requires --profile-type)
	grafanactl query -d <pyroscope-uid> -t pyroscope -e '{service_name="frontend"}' --profile-type process_cpu:cpu:nanoseconds:cpu:nanoseconds --start now-1h --end now

	# Output as JSON
	grafanactl query -d <datasource-uid> -e 'up' -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := opts.Validate(); err != nil {
				return err
			}

			ctx := cmd.Context()

			cfg, err := configOpts.LoadRESTConfig(ctx)
			if err != nil {
				return err
			}

			// Resolve datasource
			datasourceUID := opts.Datasource
			if datasourceUID == "" {
				fullCfg, err := configOpts.LoadConfig(ctx)
				if err != nil {
					return err
				}
				switch opts.Type {
				case "loki":
					datasourceUID = fullCfg.GetCurrentContext().DefaultLokiDatasource
				case "pyroscope":
					datasourceUID = fullCfg.GetCurrentContext().DefaultPyroscopeDatasource
				default:
					datasourceUID = fullCfg.GetCurrentContext().DefaultPrometheusDatasource
				}
			}
			if datasourceUID == "" {
				return fmt.Errorf("datasource UID is required: use -d flag or set default-%s-datasource in config", opts.Type)
			}

			now := time.Now()
			start, err := ParseTime(opts.Start, now)
			if err != nil {
				return fmt.Errorf("invalid start time: %w", err)
			}

			end, err := ParseTime(opts.End, now)
			if err != nil {
				return fmt.Errorf("invalid end time: %w", err)
			}

			step, err := ParseDuration(opts.Step)
			if err != nil {
				return fmt.Errorf("invalid step: %w", err)
			}

			switch opts.Type {
			case "prometheus":
				client, err := prometheus.NewClient(cfg)
				if err != nil {
					return fmt.Errorf("failed to create client: %w", err)
				}

				req := prometheus.QueryRequest{
					Query: opts.Query,
					Start: start,
					End:   end,
					Step:  step,
				}

				resp, err := client.Query(ctx, datasourceUID, req)
				if err != nil {
					return fmt.Errorf("query failed: %w", err)
				}

				if opts.IO.OutputFormat == "table" {
					return prometheus.FormatTable(cmd.OutOrStdout(), resp)
				}

				return opts.IO.Encode(cmd.OutOrStdout(), resp)

			case "loki":
				client, err := loki.NewClient(cfg)
				if err != nil {
					return fmt.Errorf("failed to create client: %w", err)
				}

				req := loki.QueryRequest{
					Query: opts.Query,
					Start: start,
					End:   end,
					Step:  step,
					Limit: 1000, // Default limit
				}

				resp, err := client.Query(ctx, datasourceUID, req)
				if err != nil {
					return fmt.Errorf("query failed: %w", err)
				}

				if opts.IO.OutputFormat == "table" {
					return loki.FormatQueryTable(cmd.OutOrStdout(), resp)
				}

				return opts.IO.Encode(cmd.OutOrStdout(), resp)

			case "pyroscope":
				client, err := pyroscope.NewClient(cfg)
				if err != nil {
					return fmt.Errorf("failed to create client: %w", err)
				}

				req := pyroscope.QueryRequest{
					LabelSelector: opts.Query,
					ProfileTypeID: opts.ProfileType,
					Start:         start,
					End:           end,
					MaxNodes:      opts.MaxNodes,
				}

				resp, err := client.Query(ctx, datasourceUID, req)
				if err != nil {
					return fmt.Errorf("query failed: %w", err)
				}

				if opts.IO.OutputFormat == "table" {
					return pyroscope.FormatQueryTable(cmd.OutOrStdout(), resp)
				}
				return opts.IO.Encode(cmd.OutOrStdout(), resp)

			default:
				return fmt.Errorf("datasource type %q is not supported (supported: prometheus, loki, pyroscope)", opts.Type)
			}
		},
	}

	configOpts.BindFlags(cmd.PersistentFlags())
	opts.setup(cmd.Flags())

	return cmd
}

type queryTableCodec struct{}

func (c *queryTableCodec) Format() format.Format {
	return "table"
}

func (c *queryTableCodec) Encode(w io.Writer, data any) error {
	switch resp := data.(type) {
	case *prometheus.QueryResponse:
		return prometheus.FormatTable(w, resp)
	case *loki.QueryResponse:
		return loki.FormatQueryTable(w, resp)
	case *pyroscope.QueryResponse:
		return pyroscope.FormatQueryTable(w, resp)
	default:
		return errors.New("invalid data type for query table codec")
	}
}

func (c *queryTableCodec) Decode(io.Reader, any) error {
	return errors.New("query table codec does not support decoding")
}
