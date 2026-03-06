## grafanactl slo reports get

Get a single SLO report.

```
grafanactl slo reports get UUID [flags]
```

### Options

```
  -h, --help            help for get
  -o, --output string   Output format. One of: json, yaml (default "yaml")
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

* [grafanactl slo reports](grafanactl_slo_reports.md)	 - Manage SLO reports.

