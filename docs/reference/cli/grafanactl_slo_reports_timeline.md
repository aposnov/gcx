## grafanactl slo reports timeline

Render SLI values over time for SLO reports.

### Synopsis

Render SLI values over time as line charts for each SLO report by
executing range queries against the Prometheus datasource associated with
each constituent SLO.

Requires that SLO destination datasources have recording rules generating
grafana_slo_sli_window metrics.

```
grafanactl slo reports timeline [UUID] [flags]
```

### Examples

```
  # Render SLI trend for all SLO reports over the past 7 days.
  grafanactl slo reports timeline

  # Render SLI trend for a specific report.
  grafanactl slo reports timeline abc123def

  # Custom time range with explicit step.
  grafanactl slo reports timeline --start now-24h --end now --step 5m

  # Output timeline data as a table.
  grafanactl slo reports timeline -o table
```

### Options

```
      --end string      End of the time range (e.g. now, RFC3339, Unix timestamp) (default "now")
  -h, --help            help for timeline
  -o, --output string   Output format. One of: graph, json, table, yaml (default "graph")
      --start string    Start of the time range (e.g. now-7d, now-24h, RFC3339, Unix timestamp) (default "now-7d")
      --step string     Query step (e.g. 5m, 1h). Defaults to auto-computed value.
```

### Options inherited from parent commands

```
      --config string    Path to the configuration file to use
      --context string   Name of the context to use
      --no-color         Disable color output
  -v, --verbose count    Verbose mode. Multiple -v options increase the verbosity (maximum: 3).
```

### SEE ALSO

* [grafanactl slo reports](grafanactl_slo_reports.md)	 - Manage SLO reports.

