package query

import (
	"errors"
	"fmt"
	"io"
	"time"

	cmdconfig "github.com/grafana/grafanactl/cmd/grafanactl/config"
	cmdio "github.com/grafana/grafanactl/cmd/grafanactl/io"
	"github.com/grafana/grafanactl/internal/format"
	"github.com/grafana/grafanactl/internal/grafana"
	"github.com/grafana/grafanactl/internal/query/loki"
	"github.com/grafana/grafanactl/internal/query/prometheus"
	"github.com/grafana/grafanactl/internal/query/pyroscope"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type queryOpts struct {
	IO          cmdio.Options
	Datasource  string
	Query       string
	From        string
	To          string
	Step        string
	ProfileType string // Pyroscope-specific
	MaxNodes    int64  // Pyroscope-specific
}

func (opts *queryOpts) setup(flags *pflag.FlagSet) {
	opts.IO.RegisterCustomCodec("table", &queryTableCodec{})
	opts.IO.RegisterCustomCodec("wide", &queryWideCodec{})
	opts.IO.RegisterCustomCodec("graph", &queryGraphCodec{})
	opts.IO.DefaultFormat("table")
	opts.IO.BindFlags(flags)

	flags.StringVarP(&opts.Datasource, "datasource", "d", "", "Datasource UID (required unless default-{type}-datasource is configured)")
	flags.StringVarP(&opts.Query, "expr", "e", "", "Query expression (PromQL for prometheus, LogQL for loki, label selector for pyroscope)")
	flags.StringVar(&opts.From, "from", "", "Start time (RFC3339, Unix timestamp, or relative like 'now-1h')")
	flags.StringVar(&opts.To, "to", "", "End time (RFC3339, Unix timestamp, or relative like 'now')")
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
	grafanactl query -d <datasource-uid> -e 'rate(http_requests_total[5m])' --from now-1h --to now --step 1m

	# Loki log query (instant)
	grafanactl query -d <loki-uid> -e '{job="varlogs"}'

	# Loki log query (range)
	grafanactl query -d <loki-uid> -e '{name="private-datasource-connect"}' --from now-1h --to now

	# Loki metric query (log rate)
	grafanactl query -d <loki-uid> -e 'sum(rate({job="varlogs"}[5m]))' --from now-1h --to now --step 1m

	# Pyroscope profile query (requires --profile-type)
	grafanactl query -d <pyroscope-uid> -e '{service_name="frontend"}' --profile-type process_cpu:cpu:nanoseconds:cpu:nanoseconds --from now-1h --to now

	# Output as JSON
	grafanactl query -d <datasource-uid> -e 'up' -o json

	# Loki logs with all labels (wide format)
	grafanactl query -d <loki-uid> -e '{job="varlogs"}' --from now-1h --to now -o wide`,
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

			// Resolve datasource UID
			fullCfg, err := configOpts.LoadConfig(ctx)
			if err != nil {
				return err
			}
			datasourceUID := opts.Datasource
			if datasourceUID == "" {
				curCtx := fullCfg.GetCurrentContext()
				promDefault := curCtx.DefaultPrometheusDatasource
				lokiDefault := curCtx.DefaultLokiDatasource

				switch {
				case promDefault != "" && lokiDefault != "":
					return errors.New("both default-prometheus-datasource and default-loki-datasource are configured; use -d to specify which datasource to query")
				case promDefault != "":
					datasourceUID = promDefault
				case lokiDefault != "":
					datasourceUID = lokiDefault
				default:
					return errors.New("datasource UID is required: use -d flag or configure default-prometheus-datasource or default-loki-datasource")
				}
			}

			// Fetch datasource to determine type
			gClient, err := grafana.ClientFromContext(fullCfg.GetCurrentContext())
			if err != nil {
				return fmt.Errorf("failed to create Grafana client: %w", err)
			}
			dsResp, err := gClient.Datasources.GetDataSourceByUID(datasourceUID)
			if err != nil {
				return fmt.Errorf("failed to get datasource %q: %w", datasourceUID, err)
			}
			dsType := dsResp.Payload.Type

			now := time.Now()
			start, err := ParseTime(opts.From, now)
			if err != nil {
				return fmt.Errorf("invalid --from time: %w", err)
			}

			end, err := ParseTime(opts.To, now)
			if err != nil {
				return fmt.Errorf("invalid --to time: %w", err)
			}

			step, err := ParseDuration(opts.Step)
			if err != nil {
				return fmt.Errorf("invalid step: %w", err)
			}

			switch dsType {
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

				switch opts.IO.OutputFormat {
				case "table":
					return loki.FormatQueryTable(cmd.OutOrStdout(), resp)
				case "wide":
					return loki.FormatQueryTableWide(cmd.OutOrStdout(), resp)
				default:
					return opts.IO.Encode(cmd.OutOrStdout(), resp)
				}

			case "pyroscope":
				if opts.ProfileType == "" {
					return errors.New("profile type is required for pyroscope queries (use --profile-type)")
				}

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
				return fmt.Errorf("datasource type %q is not supported (supported: prometheus, loki, pyroscope)", dsType)
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

type queryWideCodec struct{}

func (c *queryWideCodec) Format() format.Format {
	return "wide"
}

func (c *queryWideCodec) Encode(w io.Writer, data any) error {
	switch resp := data.(type) {
	case *prometheus.QueryResponse:
		return prometheus.FormatTable(w, resp)
	case *loki.QueryResponse:
		return loki.FormatQueryTableWide(w, resp)
	default:
		return errors.New("invalid data type for query wide codec")
	}
}

func (c *queryWideCodec) Decode(io.Reader, any) error {
	return errors.New("query wide codec does not support decoding")
}
