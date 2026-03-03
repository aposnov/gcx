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

	# Output as JSON
	grafanactl query -d <datasource-uid> -e 'up' -o json
```

### Options

```
      --config string       Path to the configuration file to use
      --context string      Name of the context to use
  -d, --datasource string   Datasource UID (required unless default-prometheus-datasource is configured)
      --end string          End time (RFC3339, Unix timestamp, or relative like 'now')
  -e, --expr string         Query expression (PromQL for prometheus, LogQL for loki)
  -h, --help                help for query
  -o, --output string       Output format. One of: graph, json, table, yaml (default "table")
      --start string        Start time (RFC3339, Unix timestamp, or relative like 'now-1h')
      --step string         Query step (e.g., '15s', '1m')
  -t, --type string         Datasource type (prometheus, loki) (default "prometheus")
```

### Options inherited from parent commands

```
      --no-color        Disable color output
  -v, --verbose count   Verbose mode. Multiple -v options increase the verbosity (maximum: 3).
```

### SEE ALSO

* [grafanactl](grafanactl.md)	 - 

