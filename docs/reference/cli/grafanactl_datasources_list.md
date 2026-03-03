## grafanactl datasources list

List all datasources

### Synopsis

List all datasources configured in Grafana.

```
grafanactl datasources list [flags]
```

### Examples

```

	# List all datasources
	grafanactl datasources list

	# List only Prometheus datasources
	grafanactl datasources list --type prometheus

	# Output as JSON
	grafanactl datasources list -o json
```

### Options

```
  -h, --help            help for list
  -o, --output string   Output format. One of: json, table, yaml (default "table")
  -t, --type string     Filter by datasource type (e.g., prometheus, loki)
```

### Options inherited from parent commands

```
      --config string    Path to the configuration file to use
      --context string   Name of the context to use
      --no-color         Disable color output
  -v, --verbose count    Verbose mode. Multiple -v options increase the verbosity (maximum: 3).
```

### SEE ALSO

* [grafanactl datasources](grafanactl_datasources.md)	 - Manage Grafana datasources

