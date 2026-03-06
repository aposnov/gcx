## grafanactl datasources pyroscope profile-types

List available profile types

### Synopsis

List available profile types from a Pyroscope datasource.

```
grafanactl datasources pyroscope profile-types [flags]
```

### Examples

```

	# List profile types (use datasource UID, not name)
	grafanactl datasources pyroscope profile-types -d <datasource-uid>

	# Output as JSON
	grafanactl datasources pyroscope profile-types -d <datasource-uid> -o json
```

### Options

```
  -d, --datasource string   Datasource UID (required unless default-pyroscope-datasource is configured)
  -h, --help                help for profile-types
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

