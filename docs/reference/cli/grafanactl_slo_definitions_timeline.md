## grafanactl slo definitions timeline

Render SLI values over time as a line chart.

### Synopsis

Render SLI values over time as a line chart by executing a range query
against the Prometheus datasource associated with each SLO.

Requires that the SLO destination datasource has recording rules generating
grafana_slo_sli_window metrics.

```
grafanactl slo definitions timeline [UUID] [flags]
```

### Examples

```
  # Render SLI trend for all SLOs over the past 7 days.
  grafanactl slo definitions timeline

  # Render SLI trend for a specific SLO.
  grafanactl slo definitions timeline abc123def

  # Custom time range with explicit step.
  grafanactl slo definitions timeline --start now-24h --end now --step 5m

  # Output timeline data as a table.
  grafanactl slo definitions timeline -o table
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
      --agent            Enable agent mode (JSON output, no color). Auto-detected from CLAUDE_CODE, CURSOR_AGENT, GITHUB_COPILOT, AMAZON_Q, or GRAFANACTL_AGENT_MODE env vars.
      --config string    Path to the configuration file to use
      --context string   Name of the context to use
      --no-color         Disable color output
  -v, --verbose count    Verbose mode. Multiple -v options increase the verbosity (maximum: 3).
```

### SEE ALSO

* [grafanactl slo definitions](grafanactl_slo_definitions.md)	 - Manage SLO definitions.

