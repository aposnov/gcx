## grafanactl datasources pyroscope labels

List labels or label values

### Synopsis

List all labels or get values for a specific label from a Pyroscope datasource.

```
grafanactl datasources pyroscope labels [flags]
```

### Examples

```

	# List all labels (use datasource UID, not name)
	grafanactl datasources pyroscope labels -d <datasource-uid>

	# Get values for a specific label
	grafanactl datasources pyroscope labels -d <datasource-uid> --label service_name

	# Output as JSON
	grafanactl datasources pyroscope labels -d <datasource-uid> -o json
```

### Options

```
  -d, --datasource string   Datasource UID (required unless default-pyroscope-datasource is configured)
  -h, --help                help for labels
  -l, --label string        Get values for this label (omit to list all labels)
  -o, --output string       Output format. One of: json, table, yaml (default "table")
```

### Options inherited from parent commands

```
      --config string    Path to the configuration file to use
      --context string   Name of the context to use
      --no-color         Disable color output
  -v, --verbose count    Verbose mode. Multiple -v options increase the verbosity (maximum: 3).
```

### SEE ALSO

* [grafanactl datasources pyroscope](grafanactl_datasources_pyroscope.md)	 - Pyroscope datasource operations

