---
name: add-provider
description: Use when adding a new Grafana Cloud product provider to grafanactl (SLO, OnCall, Synthetic Monitoring, k6, ML, etc.), or when the user says "add provider", "new provider", or "integrate [product]". Also triggers for bead tasks referencing provider implementation.
---

# Add Provider

Add a new Grafana product provider to grafanactl — from API discovery through
verified implementation.

## When to Use

- User wants to add CLI support for a Grafana Cloud product
- User says "add provider", "new provider", "integrate [product]"
- A bead task references provider implementation (e.g., Wave 2/3 providers)

**First**: Check the decision tree in `references/decision-tree.md` to confirm
a provider is the right approach (vs extending the existing resources command).

## Workflow

The workflow has four phases. Each phase has a gate — don't proceed until the
gate condition is met.

| Phase | Gate | Key Actions |
|-------|------|-------------|
| 1. Discover | Design doc produced | AskUser for context → API research → test real calls |
| 2. Design | Decisions answered | Auth, client type, envelope, commands (CRUD + beyond), staging |
| 3. Implement | Compiles, tests pass | provider.go + init() → types/client/adapter → register → tests |
| 4. Verify | All checklists green | Interface + UX + build verification |

### Prerequisites

Before starting, confirm with the user:
- **Product name** — which Grafana product to integrate
- **Access** — do they have a running Grafana instance with the product enabled?
- **Scope** — full provider or single resource type first?

---

## Phase 1: Discover

> **Guide**: `agent-docs/provider-discovery-guide.md` — follow Sections 1.1–1.6

### Step 0: Gather User Context

Before autonomous research, ask the user what they already know. Use
AskUserQuestion with these questions (adjust based on what's already known):

1. **Source code access** — "Do you have access to the product's source code
   repository? If so, which repo (e.g., `grafana/{product}`)?"
2. **API documentation** — "Can you point me to any API docs or OpenAPI specs?
   (GitHub repo, Grafana docs URL, or 'I don't know')"
3. **Terraform resources** — "Does the Grafana Terraform provider already
   support this product? Any specific resource names you know of?"
4. **Go SDK** — "Is there an existing Go client library for this product's API?
   (e.g., `grafana/{product}-api-go-client`)"
5. **Known quirks** — "Anything unusual about this product's API that I should
   know? (e.g., non-standard auth, async operations, unusual pagination)"

Use the answers to focus research — skip steps where the user has already
provided definitive information, and prioritize areas where they're uncertain.

### Step 1: Map the API Surface (Section 1.1)

Research the product's API surface. Use user-provided pointers from Step 0 first,
then fill gaps autonomously.

Search for the product's OpenAPI spec:
```bash
# Check for OpenAPI spec repos
gh search repos "org:grafana {product}-openapi" --json name,url
# Check plugin API paths
curl -s -H "Authorization: Bearer $TOKEN" "$GRAFANA_URL/api/plugins" | jq '.[] | select(.id | contains("{product}"))'
```

Capture: base path, auth scheme, endpoints, response wrappers, pagination.

### Step 2: Check Existing Tooling (Section 1.2)

```bash
# Check Terraform provider for resource schemas
gh api repos/grafana/terraform-provider-grafana/contents/internal/resources --jq '.[].name'
# Check for Go SDK
gh search repos "org:grafana {product}-go-client OR {product}-api-go" --json name,url
```

Extract: schema fields, types, validation rules, CRUD patterns.

### Step 3: Inspect Source Code (Section 1.3)

```bash
# Find the product's API handlers
gh search code "org:grafana repo:grafana/{product} path:pkg/api" --json path,repository
```

Look for: undocumented endpoints, enum values, validation rules, RBAC requirements.

### Step 4: Identify Auth Model (Section 1.4)

Determine if the product reuses `grafana.token` or needs separate credentials.
This drives `ConfigKeys()` in Phase 3.

### Step 5: Map Resource Relationships (Section 1.5)

Document how resources reference each other and existing grafanactl resources.

### Step 6: Test API Behavior (Section 1.6)

Make real API calls to validate assumptions:
```bash
# List resources
curl -H "Authorization: Bearer $TOKEN" \
  "$GRAFANA_URL/api/plugins/{plugin-id}/resources/v1/{resource}"
```

Verify: response shape, duplicate handling, ID generation, error format.

### Gate: Discovery Complete

Present findings to user. Confirm:
- [ ] API endpoints documented
- [ ] Auth model identified
- [ ] Resource relationships mapped
- [ ] At least one successful API call made

---

## Phase 2: Design

> **Guide**: `agent-docs/provider-discovery-guide.md` Section 2 (Decision Framework)

Answer each decision question. Use the tables in the guide for reference.

### Decision 1: Auth Strategy (Section 2.1)

| Scenario | ConfigKeys | Token Source |
|----------|-----------|--------------|
| Same Grafana SA token | `[]` (empty) | `curCtx.Grafana.Token` |
| Separate product token | `[{Name: "token", Secret: true}]` | Provider config |
| Separate URL + token | `[{Name: "url"}, {Name: "token", Secret: true}]` | Provider config |

### Decision 2: API Client Type (Section 2.2)

| API Type | Client Approach |
|----------|----------------|
| Plugin API (`/api/plugins/...`) | Custom `http.Client` with Bearer token |
| K8s-compatible API (`/apis/...`) | grafanactl's existing dynamic client |
| External service API | Custom `http.Client` with product-specific auth |

**Warning**: Always verify K8s APIs are externally accessible before choosing that path.

### Decision 3: Envelope Mapping (Section 2.3)

Map API objects to grafanactl's K8s envelope:
```yaml
apiVersion: {product}.ext.grafana.app/v1alpha1
kind: {ResourceKind}    # PascalCase singular
metadata:
  name: {unique-id}     # UUID or slug from API
  namespace: default
spec:
  {fields}              # User-editable fields only
```

### Decision 4: Command Surface (Section 2.4)

#### 4a. Standard CRUD (always include)

```
grafanactl {provider}
├── {resource-group}
│   ├── list
│   ├── get <id>
│   ├── push [path...]
│   ├── pull
│   └── delete <id...>
```

#### 4b. Beyond CRUD — Product-Specific Commands

Using findings from Phase 1 (discovered APIs, official docs, source code),
brainstorm what product-specific commands could add value beyond basic CRUD.

Think about:
- **Operational health** — does the product expose status, health, or state
  data? (e.g., SLO's `status` command queries Prometheus for live SLI/budget)
- **Time-series trends** — can you show how something changed over time?
  (e.g., SLO's `timeline` renders error budget burn-down graphs)
- **State history** — does the product track state transitions?
  (e.g., alerting rule evaluation history, incident timeline)
- **Performance data** — are there metrics about execution or efficiency?
  (e.g., k6 load test results, synthetic monitoring check latency)
- **Relationships / graph** — can you visualize how resources connect?
  (e.g., SLO report → SLO definitions mapping)
- **Validation / dry-run** — can the API validate without applying?
  (e.g., rule validation, check preview)

**Reference**: SLO provider's beyond-CRUD commands:
- `status` — hybrid pattern: REST API for resource data + Prometheus instant
  queries for live metrics (SLI percentage, error budget remaining)
- `timeline` — Prometheus range queries + terminal graph rendering (burn-down
  chart over time window)
- `reports status` — aggregates status across multiple SLOs with combined metrics

Present your brainstormed list to the user via AskUserQuestion:
- Show each proposed command with a one-line description
- Include "None — CRUD only for now" as an option
- Ask which commands are valuable for the first implementation
- Commands the user defers can go into later stages (Decision 6)

**Important**: The brainstorming should be grounded in what the discovery phase
actually found — real API endpoints, real data available. Don't propose commands
that would require APIs that don't exist.

**Performance patterns for status/timeline commands:**
- **Aggregate queries** — prefer PromQL with `by (label1, label2)` over
  per-resource queries when showing status for many resources
- **Parallelism** — use `errgroup` for concurrent data fetching (config load,
  resource list, metric queries)
- **Caching** — if a command discovers configuration at runtime (e.g.,
  datasource UID), consider caching it in provider config for subsequent runs

### Decision 5: Package Layout (Section 2.5)

Convention (from SLO reference implementation):

```
internal/providers/{name}/      ← co-located with interface + registry
├── provider.go                 # Provider interface impl + init() + configLoader
├── provider_test.go            # Contract tests
├── {resource}/                 # One subpackage per resource type
│   ├── types.go
│   ├── client.go
│   ├── adapter.go
│   ├── commands.go
│   └── *_test.go
```

Single resource type → flat package. Multiple → subpackage per type.

**Note**: Provider implementations live at `internal/providers/{name}/`,
co-located with the interface and registry. Circular imports are avoided
via Go's self-registration pattern: providers call `providers.Register()`
in `init()`, and `cmd/grafanactl/root/command.go` triggers registration
via blank imports (`_ "github.com/grafana/grafanactl/internal/providers/slo"`).

### Decision 6: Implementation Staging (Section 2.6)

Break into independently shippable stages with a design doc for each:

```
docs/designs/{provider}/
├── {date}-{provider}-plan.md           # Top-level plan (all stages)
├── 1-{resource}-crud/
│   └── {date}-{resource}-crud.md       # Stage 1 design
├── 2-{secondary}-crud/
│   └── {date}-{secondary}-crud.md      # Stage 2 design
└── 3-status/
    └── {date}-status.md                # Stage 3 design
```

Common stage sequence:
1. Core CRUD for primary resource (~1,300 LOC for SLO)
2. Secondary resource types, if any (~500 LOC)
3. Status/monitoring (~350 LOC)
4. Advanced features (graph, timeline, etc.)

**Every stage design doc MUST include a Verification section** with concrete
smoke-test commands using `source .env` or `grafanactl config` values. Do not
defer verification planning to implementation time.

### Gate: Design Complete

Write a top-level plan doc in `docs/designs/{provider}/` capturing all
decisions, file tree, and stage breakdown. Create per-stage docs for each
stage. **Get user approval before implementing.** The SLO plan is the template:
`docs/designs/slo-provider/2026-03-04-slo-provider-plan.md`.

---

## Phase 3: Implement

> **Guide**: `agent-docs/provider-guide.md` — follow Steps 1–7
> **UX Guide**: `agent-docs/design-guide.md` — comply with all [CURRENT] and [ADOPT] items

Implement one stage at a time. Consider splitting work across sessions per stage
to avoid context overflow — each stage's design doc is self-contained enough to
resume in a fresh session.

For each stage:

### Step 1: Provider Interface (`provider-guide.md` Step 1)

Create `internal/providers/{name}/provider.go`. Include `init()` self-registration
and the full `configLoader` — providers cannot import `cmd/grafanactl/config`
(import cycle):

```go
func init() { //nolint:gochecknoinits // Self-registration pattern.
    providers.Register(&{Name}Provider{})
}

type {Name}Provider struct{}
var _ providers.Provider = &{Name}Provider{}

func (p *{Name}Provider) Name() string      { return "{name}" }
func (p *{Name}Provider) ShortDesc() string { return "Manage Grafana {Product} resources." }

// configLoader avoids importing cmd/grafanactl/config (import cycle).
// Copy from internal/providers/slo/provider.go and update as needed.
type configLoader struct {
    configFile string
    ctxName    string
}

func (l *configLoader) bindFlags(flags *pflag.FlagSet) { ... }
func (l *configLoader) LoadRESTConfig(ctx context.Context) (config.NamespacedRESTConfig, error) { ... }
```

**Important**: Copy the full `configLoader` from `internal/providers/slo/provider.go` —
it handles env vars (`GRAFANA_TOKEN`, `GRAFANA_PROVIDER_*`), context switching,
and validation. Don't simplify it; the full implementation is required.

### Steps 2–3: Config Keys + Validate (`provider-guide.md` Steps 2–3)

- **ConfigKeys**: Declare all keys. `Secret: true` for tokens. SLO uses `[]`
  (reuses `grafana.token`) — most plugin API providers can do the same.
  Config key names use **hyphen-case** (`my-url`, `my-token`), not
  underscore_case. Error messages must match the config key format exactly.
- **Validate**: Return actionable errors pointing to `grafanactl config set ...`.

### Step 4: Commands (`provider-guide.md` Step 4)

```go
func (p *{Name}Provider) Commands() []*cobra.Command {
    loader := &configLoader{}
    cmd := &cobra.Command{Use: "{name}", Short: p.ShortDesc()}
    loader.bindFlags(cmd.PersistentFlags())
    cmd.AddCommand({resource}.Commands(loader))
    return []*cobra.Command{cmd}
}
```

**UX requirements** (from `design-guide.md`):
- Register `text` table codec as default for list/get commands
- Use `cmdio.Success/Warning/Error/Info` for status messages
- Include `-o json/yaml` support via `io.Options`
- Include help text with examples (3-5 per command)
- Push is idempotent (create-or-update)
- Data fetching is format-agnostic (Pattern 13)
- PromQL via `promql-builder`, not string formatting (Pattern 14)

**Client decision**: Hand-roll the HTTP client (don't import generated OpenAPI
clients). grafanactl's pattern is direct HTTP calls with `Authorization: Bearer`
headers. Generated clients use awkward types that break the adapter's
`encoding/json` round-trip. ~200 LOC for a typical CRUD client.

### Step 5: Types + Client + Adapter

For each resource type, create:
- `types.go` — Go structs matching API schema. Use camelCase field names
  matching the API (ensures lossless pull → edit → push round-trips)
- `client.go` — HTTP client (List, Get, Create, Update, Delete) with `httptest`
  unit tests
- `adapter.go` — Translate between API objects and K8s `Unstructured`. Test
  with round-trip property tests

Use `internal/providers/slo/definitions/` as the reference for all three files.

### Step 6: Register

Step 1 already added `init()` with `providers.Register()`. The remaining step
is adding the blank import in `cmd/grafanactl/root/command.go`:

```go
import (
    _ "github.com/grafana/grafanactl/internal/providers/{name}" // triggers init()
)
```

This triggers `init()` → `providers.Register()` → `providers.All()` returns it.

### Step 7: Tests (`provider-guide.md` Step 7)

Write contract tests for the provider interface + unit tests for each component:
- Provider interface compliance
- Adapter round-trip (API → K8s → API preserves data)
- Client HTTP behavior (use httptest)
- Command integration (flag parsing, output format)

### Gate: Stage Complete

```bash
make all               # lint + tests + build + docs
grafanactl providers   # New provider listed
grafanactl config view # Secrets redacted correctly
```

---

## Phase 4: Verify

> **Checklist**: `agent-docs/design-guide.md` Section 7 + `agent-docs/provider-guide.md` Checklist

Run through both checklists:

### Interface Compliance
- [ ] All five `Provider` methods implemented
- [ ] `Name()` lowercase, unique, stable
- [ ] All config keys declared in `ConfigKeys()`
- [ ] Secret keys have `Secret: true`
- [ ] `Validate()` returns actionable error with `config set` command
- [ ] Provider in `allProviders()` in `cmd/grafanactl/root/command.go`

### UX Compliance
- [ ] All data-display commands support `-o json/yaml`
- [ ] List/get register `text` table codec as default
- [ ] Error messages include actionable suggestions
- [ ] No `os.Exit()` in command code
- [ ] Status messages use `cmdio` functions
- [ ] `--config` and `--context` inherited via persistent flags
- [ ] Help text: Short (verb, period-terminated), Long, Examples
- [ ] Push is idempotent
- [ ] Data fetching is format-agnostic
- [ ] PromQL uses `promql-builder` (if applicable)

### Build Verification
- [ ] `make all` succeeds (lint + tests + build + docs)
- [ ] `grafanactl providers` lists the new provider
- [ ] `grafanactl config view` redacts secrets
- [ ] Run `/update-agent-docs` to capture new patterns/architecture

---

## Reference Implementations

Two providers serve as reference implementations, demonstrating different
auth models and API types.

### SLO Provider (Wave 1, PR #13)

Uses the same Grafana SA token and a plugin API. Key files:

| Component | Path |
|-----------|------|
| Provider struct + configLoader | `internal/providers/slo/provider.go` |
| Definitions commands | `internal/providers/slo/definitions/commands.go` |
| API client | `internal/providers/slo/definitions/client.go` |
| K8s adapter | `internal/providers/slo/definitions/adapter.go` |
| Status (Prometheus hybrid) | `internal/providers/slo/definitions/status.go` |
| Timeline (range query + graph) | `internal/providers/slo/definitions/timeline.go` |
| Self-registration | `internal/providers/slo/provider.go` (`init()`) |
| Blank import trigger | `cmd/grafanactl/root/command.go` |
| Top-level plan | `docs/designs/slo-provider/2026-03-04-slo-provider-plan.md` |
| Stage 1 design | `docs/designs/slo-provider/1-slo-definitions-crud/` |
| Stage 2 design | `docs/designs/slo-provider/2-reports-crud/` |

### Synth Monitoring Provider (Wave 2)

Uses a separate URL + token and an external service API. Key files:

| Component | Path |
|-----------|------|
| Provider struct + configLoader | `internal/providers/synth/provider.go` |
| Shared config interfaces | `internal/providers/synth/smcfg/loader.go` |
| Checks commands | `internal/providers/synth/checks/commands.go` |
| Status (Prometheus hybrid) | `internal/providers/synth/checks/status.go` |
| Probes commands | `internal/providers/synth/probes/commands.go` |
| Top-level plan | `docs/designs/synth-provider/2026-03-06-synth-provider-plan.md` |
| Stage 1 design | `docs/designs/synth-provider/1-checks-probes-crud/` |
| Stage 2 design | `docs/designs/synth-provider/2-checks-status/` |

### Lessons Learned

**What the SLO commit taught us** (from the actual implementation experience):
- Design docs were produced for each stage before coding began — the plan
  drove implementation, not the other way around
- The `configLoader` needed full env var resolution (`GRAFANA_TOKEN`,
  `GRAFANA_PROVIDER_SLO_*`) — not just flag binding
- Hand-rolling the HTTP client (~200 LOC) was cleaner than importing the
  OpenAPI-generated client (awkward types, poor adapter fit)
- Status commands require a hybrid pattern: REST API for resource data +
  Prometheus instant queries for live metrics
- K8s CRDs for SLO exist internally but are NOT accessible externally —
  verified by real API call, drove the plugin API choice

## Common Pitfalls

| Pitfall | Mitigation |
|---------|------------|
| Incomplete OpenAPI specs | Cross-reference with source code route handlers |
| K8s CRDs not externally accessible | Always verify with real API call before choosing K8s client path |
| readOnly fields in POST/PUT | Adapter must strip server-generated fields on Create/Update |
| Different list response envelopes | Define response types per product (no universal wrapper) |
| configLoader is non-trivial | Copy full implementation from `internal/providers/slo/provider.go`, don't simplify |
| Missing blank import | Add `_ "github.com/grafana/grafanactl/internal/providers/{name}"` in `cmd/grafanactl/root/command.go` |
| Lint failures for init()/global | Add `//nolint:gochecknoinits` and `//nolint:gochecknoglobals` directives |
