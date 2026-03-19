# Query Patterns

Advanced patterns for querying Prometheus and Loki datasources with grafanactl.

## Datasource UID Resolution

**CRITICAL**: Always use datasource UID, never the name.

### Finding Datasource UIDs

```bash
# List all datasources
grafanactl datasources list

# Filter by type
grafanactl datasources list --type prometheus
grafanactl datasources list --type loki

# Get JSON for scripting
DS_UID=$(grafanactl datasources list --type prometheus -o json | jq -r '.datasources[0].uid')
```

### Setting Default Datasource

Avoid repeating the datasource UID argument:

```bash
# Set default Prometheus datasource
grafanactl config set contexts.mystack.default-prometheus-datasource <uid>

# Set default Loki datasource
grafanactl config set contexts.mystack.default-loki-datasource <uid>

# Now queries work without specifying a UID
grafanactl datasources prometheus query 'up'
grafanactl datasources loki query '{job="varlogs"}'
```

## Prometheus Query Patterns

### Instant Queries

Query current values:

```bash
# Current uptime for all targets
grafanactl datasources prometheus query <uid> 'up'

# CPU usage by job
grafanactl datasources prometheus query <uid> 'avg by(job) (rate(cpu_usage_seconds[5m]))'

# Memory usage with threshold
grafanactl datasources prometheus query <uid> 'node_memory_MemAvailable_bytes < 1000000000'
```

### Range Queries

Query over time periods:

```bash
# HTTP request rate over last hour
grafanactl datasources prometheus query <uid> 'rate(http_requests_total[5m])' \
  --from now-1h --to now --step 1m

# CPU usage for specific time period
grafanactl datasources prometheus query <uid> 'avg(cpu_usage)' \
  --from 2026-03-01T00:00:00Z --to 2026-03-01T12:00:00Z --step 5m

# Disk usage over last 24 hours
grafanactl datasources prometheus query <uid> 'disk_used_percent' \
  --from now-24h --to now --step 15m
```

### Time Range Formats

grafanactl supports multiple time formats:

```bash
# Relative time (recommended for most cases)
--from now-1h --to now
--from now-24h --to now-1h
--from now-7d --to now

# RFC3339 timestamps
--from 2026-03-01T00:00:00Z --to 2026-03-01T12:00:00Z

# Unix timestamps
--from 1709280000 --to 1709366400
```

### Step Interval

Choose step based on time range:

```bash
# Short ranges: 1-5 second steps
grafanactl datasources prometheus query <uid> 'rate(requests[1m])' \
  --from now-5m --to now --step 1s

# Medium ranges: 1-5 minute steps
grafanactl datasources prometheus query <uid> 'rate(requests[5m])' \
  --from now-6h --to now --step 1m

# Long ranges: 15-60 minute steps
grafanactl datasources prometheus query <uid> 'rate(requests[1h])' \
  --from now-7d --to now --step 1h
```

**Rule of thumb**: Step should be ~1/100th of total range for smooth charts.

### Aggregation Patterns

```bash
# Sum across all instances
grafanactl datasources prometheus query <uid> 'sum(http_requests_total)'

# Average by label
grafanactl datasources prometheus query <uid> 'avg by(job) (cpu_usage)'

# Top 5 by value
grafanactl datasources prometheus query <uid> 'topk(5, http_requests_total)'

# 95th percentile
grafanactl datasources prometheus query <uid> 'histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))'
```

### Combining with Graph

```bash
# Line chart (default) — pass -o graph directly to the query command
grafanactl datasources prometheus query <uid> 'rate(http_requests_total[5m])' \
  --from now-1h --to now --step 1m -o graph

# Instant query as graph
grafanactl datasources prometheus query <uid> 'up' -o graph

# Range query as graph
grafanactl datasources prometheus query <uid> 'cpu_usage' --from now-6h --to now --step 5m -o graph
```

## Loki Query Patterns

### Log Stream Selectors

Basic log filtering:

```bash
# All logs from a job
grafanactl datasources loki query <loki-uid> '{job="varlogs"}'

# Multiple labels (AND)
grafanactl datasources loki query <loki-uid> '{job="varlogs",level="error"}'

# Regex matching
grafanactl datasources loki query <loki-uid> '{job=~"mysql.*",level!="debug"}'

# Exclude specific values
grafanactl datasources loki query <loki-uid> '{namespace="production",pod!~"test.*"}'
```

### Log Stream Operators

```bash
# Contains text
grafanactl datasources loki query <loki-uid> '{job="varlogs"} |= "error"'

# Doesn't contain text
grafanactl datasources loki query <loki-uid> '{job="varlogs"} != "debug"'

# Regex match in log line
grafanactl datasources loki query <loki-uid> '{job="varlogs"} |~ "error|exception"'

# JSON parsing
grafanactl datasources loki query <loki-uid> '{job="varlogs"} | json | level="error"'
```

### Log Range Queries

Query logs over time:

```bash
# Last hour of logs
grafanactl datasources loki query <loki-uid> '{job="varlogs"}' \
  --from now-1h --to now

# Specific time range
grafanactl datasources loki query <loki-uid> '{namespace="prod"}' \
  --from 2026-03-01T00:00:00Z --to 2026-03-01T12:00:00Z
```

### Log Metrics (Rate Queries)

Calculate metrics from logs:

```bash
# Log rate per second
grafanactl datasources loki query <loki-uid> \
  'rate({job="varlogs"}[5m])' \
  --from now-1h --to now --step 1m

# Sum of log rates
grafanactl datasources loki query <loki-uid> \
  'sum(rate({namespace="production"}[5m]))' \
  --from now-6h --to now --step 5m

# Count by level
grafanactl datasources loki query <loki-uid> \
  'sum by(level) (rate({job="varlogs"} | json [5m]))' \
  --from now-1h --to now --step 1m
```

### Combining Loki with Graph

```bash
# Visualize log volume
grafanactl datasources loki query <loki-uid> \
  'sum(rate({job="varlogs"}[5m]))' \
  --from now-6h --to now --step 5m -o graph

# Error rate over time
grafanactl datasources loki query <loki-uid> \
  'sum(rate({job="app"} |= "error" [5m]))' \
  --from now-24h --to now --step 15m -o graph
```

## Prometheus Datasource Operations

### Exploring Metrics

```bash
# List all available labels
grafanactl datasources prometheus labels -d <uid>

# Get values for specific label
grafanactl datasources prometheus labels -d <uid> --label job
grafanactl datasources prometheus labels -d <uid> --label instance

# Get metric metadata
grafanactl datasources prometheus metadata -d <uid>
grafanactl datasources prometheus metadata -d <uid> --metric http_requests_total

# List scrape targets
grafanactl datasources prometheus targets -d <uid>
```

### Discovery Workflow

1. Find interesting labels:
```bash
grafanactl datasources prometheus labels -d <uid>
```

2. Get values for label:
```bash
grafanactl datasources prometheus labels -d <uid> --label job
```

3. Query specific job:
```bash
grafanactl datasources prometheus query <uid> 'up{job="prometheus"}'
```

4. Explore available metrics for that job:
```bash
grafanactl datasources prometheus metadata -d <uid> | grep -i <keyword>
```

## Loki Datasource Operations

### Exploring Log Streams

```bash
# List all available labels
grafanactl datasources loki labels -d <loki-uid>

# Get values for specific label
grafanactl datasources loki labels -d <loki-uid> --label job
grafanactl datasources loki labels -d <loki-uid> --label namespace

# Find series matching selectors
grafanactl datasources loki series -d <loki-uid> -M '{job="varlogs"}'
grafanactl datasources loki series -d <loki-uid> -M '{namespace="production"}' -M '{level="error"}'
```

### Discovery Workflow

1. Find available labels:
```bash
grafanactl datasources loki labels -d <loki-uid>
```

2. Get values for interesting labels:
```bash
grafanactl datasources loki labels -d <loki-uid> --label job
grafanactl datasources loki labels -d <loki-uid> --label namespace
```

3. Find series combinations:
```bash
grafanactl datasources loki series -d <loki-uid> -M '{job="varlogs"}'
```

4. Query specific stream:
```bash
grafanactl datasources loki query <loki-uid> '{job="varlogs",namespace="prod"}'
```

## Output Formats

### Table Format (Default)

For Prometheus queries, shows metric values in a table:

```bash
grafanactl datasources prometheus query <uid> 'up'
# Output:
# METRIC    VALUE  TIMESTAMP
# up{...}   1      2026-03-03T12:00:00Z
```

For Loki queries, shows raw log lines:

```bash
grafanactl datasources loki query <loki-uid> '{job="varlogs"}' --from now-5m --to now
# Output:
# ts=2026-03-06T10:30:00Z level=info msg="request completed" status=200
# ts=2026-03-06T10:30:01Z level=error msg="connection refused"
```

### Wide Format (Loki only)

Shows all labels plus the log line:

```bash
grafanactl datasources loki query <loki-uid> '{job="varlogs"}' --from now-5m --to now -o wide
# Output:
# CLUSTER        DETECTED_LEVEL  JOB       NAMESPACE  POD          LINE
# dev-eu-west-2  info            varlogs   prod       app-abc123   ts=2026-03-06T10:30:00Z...
# dev-eu-west-2  error           varlogs   prod       app-abc123   ts=2026-03-06T10:30:01Z...
```

### JSON Format

Machine-readable for scripting:

```bash
grafanactl datasources prometheus query <uid> 'up' -o json
```

JSON structure:
```json
{
  "status": "success",
  "data": {
    "resultType": "vector",
    "result": [
      {
        "metric": {"__name__": "up", "job": "prometheus"},
        "value": [1709467200, "1"]
      }
    ]
  }
}
```

### YAML Format

```bash
grafanactl datasources prometheus query <uid> 'up' -o yaml
```

### Piping to jq

```bash
# Extract specific fields
grafanactl datasources prometheus query <uid> 'up' -o json | jq '.data.result[].metric.job'

# Filter results
grafanactl datasources prometheus query <uid> 'up' -o json | jq '.data.result[] | select(.value[1] == "1")'

# Count results
grafanactl datasources prometheus query <uid> 'up' -o json | jq '.data.result | length'
```

## Scripting Patterns

### Automated Monitoring

```bash
#!/bin/bash
DS_UID=$(grafanactl datasources list --type prometheus -o json | jq -r '.datasources[0].uid')

# Check if service is up
UP=$(grafanactl datasources prometheus query $DS_UID 'up{job="critical-service"}' -o json | \
     jq -r '.data.result[0].value[1]')

if [ "$UP" != "1" ]; then
  echo "ALERT: critical-service is down!"
  exit 1
fi
```

### Batch Queries

```bash
#!/bin/bash
DS_UID="<your-datasource-uid>"

QUERIES=(
  "up"
  "rate(http_requests_total[5m])"
  "node_memory_MemAvailable_bytes"
)

for query in "${QUERIES[@]}"; do
  echo "Query: $query"
  grafanactl datasources prometheus query $DS_UID "$query" --from now-5m --to now -o graph
  echo "---"
done
```

### Exporting Data

```bash
# Export query results to file
grafanactl datasources prometheus query <uid> 'cpu_usage' --from now-24h --to now --step 1m -o json > cpu-data.json

# Convert to CSV (using jq)
grafanactl datasources prometheus query <uid> 'up' -o json | \
  jq -r '.data.result[] | [.metric.job, .value[0], .value[1]] | @csv' > results.csv
```

## Performance Tips

### Query Optimization

1. **Use specific label filters**: More specific = faster queries
```bash
# Slow
grafanactl datasources prometheus query <uid> 'http_requests_total'

# Fast
grafanactl datasources prometheus query <uid> 'http_requests_total{job="api",status="200"}'
```

2. **Choose appropriate range selectors**:
```bash
# For rate queries, match range to step
grafanactl datasources prometheus query <uid> 'rate(requests[5m])' --step 5m

# Don't use huge ranges for instant queries
grafanactl datasources prometheus query <uid> 'rate(requests[5m])'  # Good
grafanactl datasources prometheus query <uid> 'rate(requests[1h])'  # Usually unnecessary
```

3. **Limit time ranges**:
```bash
# Query only what you need
grafanactl datasources prometheus query <uid> 'up' --from now-1h --to now  # Good
grafanactl datasources prometheus query <uid> 'up' --from now-30d --to now  # Slow
```

### Loki Performance

1. **Use indexed labels for filtering**:
```bash
# Fast (uses indexed labels)
grafanactl datasources loki query <loki-uid> '{job="varlogs",namespace="prod"}'

# Slow (line filter, not indexed)
grafanactl datasources loki query <loki-uid> '{job="varlogs"} |= "namespace:prod"'
```

2. **Limit log queries**:
```bash
# The default limit is 1000 lines
# For production, consider increasing or narrowing time range
grafanactl datasources loki query <loki-uid> '{job="varlogs"}' --from now-5m --to now
```

### Querying at Scale

Loki metric queries (`rate()`, `count_over_time()`, etc.) produce one series per unique label combination. At scale this hits series limits (default 20K). Always aggregate:

```bash
# BAD — one series per pod/namespace/level/... combination
grafanactl datasources loki query <loki-uid> 'count_over_time({job="app"} [5m])'

# GOOD — aggregate down to what you need
grafanactl datasources loki query <loki-uid> 'sum(count_over_time({job="app"} [5m]))'
grafanactl datasources loki query <loki-uid> 'sum by(level) (count_over_time({job="app"} | json [5m]))'
grafanactl datasources loki query <loki-uid> 'topk(10, sum by(pod) (rate({job="app"} [5m])))'
```

Rule of thumb: if your query uses `rate()`, `count_over_time()`, or `bytes_over_time()`, wrap it with `sum()`, `sum by(label)`, or `topk()`.

### Stream Labels vs Extracted Labels

Loki has two kinds of labels — confusing them causes silent failures:

| | Stream labels | Extracted labels |
|---|---|---|
| Set by | Log ingestion config | Parser stages (`| json`, `| logfmt`) |
| Used in | Stream selector `{job="app"}` | Filter expressions after `|` |
| Indexed | Yes (fast) | No (line-by-line scan) |
| Available | Always | Only after parser stage |

Common mistakes:
- Filtering extracted labels in `{}` — fails silently: `{namespace="prod", pod="app-123"}` won't work if `pod` is extracted, not a stream label
- Using `label_format` to rename extracted fields before they're parsed — add the parser stage first
- Assuming a field visible in Grafana Explore is a stream label — check with `grafanactl datasources loki labels -d <uid>` (only shows stream labels)

## Common Patterns

### Health Check

```bash
# Check if services are up
grafanactl datasources prometheus query <uid> 'up{job="critical-service"}' | grep "1"
```

### Error Rate

```bash
# HTTP error rate
grafanactl datasources prometheus query <uid> 'rate(http_requests_total{status=~"5.."}[5m])' \
  --from now-1h --to now --step 1m -o graph
```

### Resource Usage

```bash
# Memory usage by pod
grafanactl datasources prometheus query <uid> 'container_memory_usage_bytes{namespace="production"}' -o graph
```

### Log Analysis

```bash
# Count errors in last hour
grafanactl datasources loki query <loki-uid> \
  'count_over_time({job="app"} |= "error" [1h])'
```

### Comparison Queries

```bash
# Compare current vs 24h ago
grafanactl datasources prometheus query <uid> 'rate(requests[5m])' --from now-1h --to now -o json > now.json
grafanactl datasources prometheus query <uid> 'rate(requests[5m])' --from now-25h --to now-24h -o json > yesterday.json
```
