# `grafanactl query` refactor

# 1. `query` should be a subcommand of `datasources`

`grafanactl query` is a top-level command but semantically belongs under
`grafanactl datasources query`, consistent with how other resource-scoped
operations are organised.

# 2. `query` should have subcommands for known datasource kinds

* `grafanactl datasources query prometheus $DATASOURCE_UID '$EXPR' [--from=] [--to=] [--window=]` - query
    metrics, instant or range-query depending on from / to (window as an
    alternative options for [now-window,now])
* `grafanactl datasources query loki $DATASOURCE_UID '$EXPR' `- same opts as metrics
* `grafanactl datasources query tempo $DATASOURCE_UID '$EXPR' `- same opts as metrics (note, tempo datasources not currently implemented)
* `grafanactl datasources query pyroscope $DATASOURCE_UID '$EXPR' `- same opts as metrics
* `grafanactl datasources query generic $DATASOURCE_UID '$EXPR' `- generic option for datasources with unknown (or any) type

The idea is that:

a) known datasources from `datasources prometheus / loki / pyroscope` map to
known query types, making it easier to construct correct syntax for queries
(promql vs logql, etc)
b) we can pre-filter only specific datasource uids in e.g. `query loki` so that
users don't accidentally try to query logs from a prometheus datasource
c) generic is an escape hatch for community / other datasources
d) eventually we can add more kinds like sql and so on

