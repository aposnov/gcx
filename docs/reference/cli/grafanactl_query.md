## grafanactl query

Execute queries against Grafana datasources

### Synopsis

Execute queries against Grafana datasources via the unified query API.

```
grafanactl query [flags]
```

### Examples

```

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
	grafanactl query -d <datasource-uid> -e 'up' -o json
```

### Options

```
      --config string         Path to the configuration file to use
      --context string        Name of the context to use
  -d, --datasource string     Datasource UID (required unless default-{type}-datasource is configured)
      --end string            End time (RFC3339, Unix timestamp, or relative like 'now')
  -e, --expr string           Query expression (PromQL for prometheus, LogQL for loki, label selector for pyroscope)
  -h, --help                  help for query
      --max-nodes int         Maximum nodes in flame graph (pyroscope only) (default 1024)
  -o, --output string         Output format. One of: graph, json, table, yaml (default "table")
      --profile-type string   Profile type ID for pyroscope queries (e.g., 'process_cpu:cpu:nanoseconds:cpu:nanoseconds')
      --start string          Start time (RFC3339, Unix timestamp, or relative like 'now-1h')
      --step string           Query step (e.g., '15s', '1m')
  -t, --type string           Datasource type (prometheus, loki, pyroscope) (default "prometheus")
```

### Options inherited from parent commands

```
      --agent           Enable agent mode (JSON output, no color). Auto-detected from CLAUDE_CODE, CURSOR_AGENT, GITHUB_COPILOT, AMAZON_Q, or GRAFANACTL_AGENT_MODE env vars.
      --no-color        Disable color output
  -v, --verbose count   Verbose mode. Multiple -v options increase the verbosity (maximum: 3).
```

### SEE ALSO

* [grafanactl](grafanactl.md)	 - 

