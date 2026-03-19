## grafanactl datasources generic

Generic datasource operations (auto-detects type)

### Synopsis

Operations for any datasource type. The datasource type is auto-detected via the Grafana API.

### Options

```
  -h, --help   help for generic
```

### Options inherited from parent commands

```
      --agent            Enable agent mode (JSON output, no color). Auto-detected from CLAUDECODE, CLAUDE_CODE, CURSOR_AGENT, GITHUB_COPILOT, AMAZON_Q, or GRAFANACTL_AGENT_MODE env vars.
      --config string    Path to the configuration file to use
      --context string   Name of the context to use
      --no-color         Disable color output
      --no-truncate      Disable table column truncation (auto-enabled when stdout is piped)
  -v, --verbose count    Verbose mode. Multiple -v options increase the verbosity (maximum: 3).
```

### SEE ALSO

* [grafanactl datasources](grafanactl_datasources.md)	 - Manage Grafana datasources
* [grafanactl datasources generic query](grafanactl_datasources_generic_query.md)	 - Execute a query against any datasource (auto-detects type)

