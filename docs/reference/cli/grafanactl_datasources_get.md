## grafanactl datasources get

Get details of a specific datasource

### Synopsis

Get detailed information about a specific datasource by its UID.

```
grafanactl datasources get UID [flags]
```

### Examples

```

	# Get datasource details
	grafanactl datasources get my-prometheus

	# Output as JSON
	grafanactl datasources get my-prometheus -o json
```

### Options

```
  -h, --help            help for get
  -o, --output string   Output format. One of: json, yaml (default "yaml")
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

