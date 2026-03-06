## grafanactl slo reports pull

Pull SLO reports to disk.

```
grafanactl slo reports pull [flags]
```

### Options

```
  -h, --help                help for pull
  -d, --output-dir string   Directory to write SLO report files to (default ".")
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

