# Evolving the MCP Server to a Production-Grade Platform

## Purpose

This document describes how the current small MCP server can evolve into a production-grade MCP platform.

The starting point is a single MCP server exposing several domain integrations:

```text
MCP Client
    |
    v
MCP Server
    |
    +--> Todo Service
    +--> Weather Service
    +--> Currency Service
```

The goal is to evolve this into a system capable of supporting:

- many MCP servers
- many teams
- hundreds or thousands of tools
- independent deployment and scaling
- centralized authentication and authorization
- tool discovery
- semantic tool search
- observability
- rate limiting
- reliability controls
- versioning
- governance
- multi-tenancy

The key principle is:

> Do not introduce an MCP gateway or aggregator merely because the system is growing. Introduce each layer when it solves a concrete scaling problem.

---

# 1. Starting Point

The initial architecture is intentionally simple:

```text
                     MCP Client
                         |
                         v
                  +-------------+
                  | MCP Server  |
                  +------+------+
                         |
          +--------------+--------------+
          |              |              |
          v              v              v
        Todos         Weather        Currency
       Service        Service         Service
          |              |              |
          v              v              v
      DummyJSON       Open-Meteo     Frankfurter
```

This is a good architecture when:

- there are few tools
- one team owns the server
- all integrations have similar lifecycle requirements
- one deployment is sufficient
- tool discovery is simple
- the number of MCP clients is small

Do not prematurely add distributed infrastructure.

---

# 2. First Scaling Problem: Too Many Tools

Imagine the server grows from:

```text
8 tools
```

to:

```text
100+ tools
```

The architecture may become:

```text
MCP Server
    |
    +--> Todos
    +--> Weather
    +--> Currency
    +--> CRM
    +--> GitHub
    +--> Slack
    +--> Jira
    +--> Payments
    +--> Analytics
    +--> ...
```

Problems begin to appear:

- huge tool list
- slower tool discovery
- difficult ownership
- unrelated deployments become coupled
- one team's change requires deploying everything
- one provider failure can affect unrelated tools
- scaling is shared
- security boundaries become difficult

At this point the natural evolution is:

> Split the system into multiple MCP servers.

---

# 3. Multiple MCP Servers

Instead of one large server:

```text
                    MCP Client
                        |
                        v
                 MCP Server
                        |
       +----------------+----------------+
       |                |                |
      CRM            GitHub           Weather
```

move to:

```text
                 MCP Client
                     |
       +-------------+-------------+
       |             |             |
       v             v             v
   CRM MCP       GitHub MCP     Weather MCP
   Server          Server         Server
```

Each MCP server owns a bounded domain.

For example:

```text
crm-mcp
    ├── contacts
    ├── accounts
    ├── opportunities
    └── activities

github-mcp
    ├── repositories
    ├── issues
    ├── pull requests
    └── deployments

weather-mcp
    ├── forecast
    └── historical weather
```

---

# 4. Why Split MCP Servers?

Independent MCP servers provide:

### Independent deployment

```text
CRM team
    ↓
deploy crm-mcp

GitHub team
    ↓
deploy github-mcp
```

No need to redeploy unrelated tools.

### Independent scaling

```text
CRM MCP
    replicas: 10

Weather MCP
    replicas: 2
```

### Failure isolation

If GitHub is unavailable:

```text
github-mcp
    ↓
failure
```

the weather MCP server can continue operating.

### Team ownership

Each server can have:

```text
owner
repository
deployment
on-call
security policy
```

---

# 5. MCP Aggregator

Once multiple MCP servers exist, an MCP client may not want to connect to every server individually.

Without aggregation:

```text
                 MCP Client
                /    |     \
               /     |      \
              v      v       v
           CRM     GitHub   Weather
           MCP       MCP      MCP
```

With an aggregator:

```text
                 MCP Client
                     |
                     v
               MCP Aggregator
                /    |    \
               /     |     \
              v      v      v
           CRM     GitHub   Weather
           MCP       MCP      MCP
```

The aggregator presents multiple MCP servers as one logical MCP endpoint.

---

# 6. What Does an Aggregator Do?

The aggregator can:

1. connect to multiple MCP servers
2. discover their tools
3. maintain a registry of tools
4. expose aggregated tools to the client
5. route tool calls to the correct MCP server
6. normalize metadata
7. handle connection lifecycle
8. isolate failures
9. optionally apply policies

Conceptually:

```text
MCP Client
    |
    | tools/list
    v
Aggregator
    |
    +--> CRM MCP
    |
    +--> GitHub MCP
    |
    +--> Weather MCP
```

The client does not need to know where individual tools live.

---

# 7. Aggregator Tool Registry

The aggregator can maintain:

```text
Tool Registry

tool_name            server
------------------------------------------------
crm.search_contacts  crm-mcp
crm.get_contact      crm-mcp

github.search_repo   github-mcp
github.get_issue     github-mcp

weather.forecast     weather-mcp
```

The registry is the first step toward centralized discovery.

---

# 8. Tool Namespaces

When multiple servers expose tools, naming collisions become possible.

For example:

```text
get
search
create
delete
```

may exist in many servers.

Use namespacing:

```text
crm.search_contacts
github.search_repositories
weather.forecast
```

This makes ownership and routing explicit.

Possible naming model:

```text
<domain>.<resource>.<action>
```

Examples:

```text
crm.contacts.search
crm.contacts.get
github.repositories.search
github.issues.get
weather.forecast.get
```

The exact naming convention should be standardized platform-wide.

---

# 9. Aggregator vs Gateway

These concepts are related but solve different problems.

## Aggregator

Primarily answers:

> "How do I combine multiple MCP servers into one logical MCP interface?"

Responsibilities:

```text
tool discovery
tool aggregation
routing
MCP server fan-in
```

## Gateway

Primarily answers:

> "How do I control access to and operate MCP traffic at platform scale?"

Responsibilities may include:

```text
authentication
authorization
rate limiting
quotas
tenant isolation
auditing
routing
policy enforcement
observability
caching
retries
circuit breakers
```

---

# 10. Combined Architecture

At production scale, these components may be combined:

```text
                       MCP Client
                           |
                           v
                    +-------------+
                    | MCP Gateway |
                    +------+------+
                           |
                 +---------+---------+
                 |                   |
                 v                   v
          Tool Discovery       Policy Engine
                 |
                 v
          MCP Aggregator
                 |
       +---------+---------+---------+
       |         |         |         |
       v         v         v         v
      CRM      GitHub    Weather   Payments
      MCP       MCP       MCP       MCP
```

However, they do not have to be separate services.

The gateway and aggregator may initially be one process.

---

# 11. When Should a Gateway Be Introduced?

A gateway becomes useful when there are requirements that every MCP server should not implement independently.

For example:

```text
100 MCP servers
    |
    +--> authentication
    +--> authorization
    +--> rate limits
    +--> audit logs
    +--> metrics
```

Without a gateway:

```text
100 servers × 5 cross-cutting concerns
```

creates a lot of duplicated infrastructure.

A gateway centralizes these concerns.

---

# 12. Gateway Responsibilities

A production MCP gateway may handle:

### Authentication

```text
MCP Client
    |
    | token
    v
Gateway
    |
    | authenticated request
    v
MCP Server
```

### Authorization

Example:

```text
user A
    ↓
allowed:
    crm.contacts.search

denied:
    payments.refund
```

### Rate limiting

```text
tenant A
    100 requests/minute

tenant B
    1000 requests/minute
```

### Quotas

For example:

```text
AI Agent
    ↓
10,000 tool calls/day
```

### Audit

Record:

```text
who
what tool
when
which server
result
latency
```

Do not store sensitive payloads unnecessarily.

---

# 13. Tool Discovery

At the MCP protocol level, the fundamental mechanism is tool discovery.

A client can ask the MCP server for its available tools.

Conceptually:

```text
Client
  |
  | tools/list
  v
MCP Server
  |
  v
Tool definitions
```

The official MCP SDK handles the protocol mechanics.

Application code registers tools.

---

# 14. Discovery at One Server

For a small server:

```text
MCP Server
    |
    +--> todos_list
    +--> todos_get
    +--> weather_forecast
    +--> exchange_rate
```

The client receives all tools.

This is fine for:

```text
5–20 tools
```

but becomes problematic when a platform exposes:

```text
5000 tools
```

---

# 15. The Large Tool List Problem

Suppose:

```text
100 MCP servers
×
50 tools
=
5000 tools
```

Sending all 5000 tool definitions to an LLM can create:

- huge context usage
- higher latency
- higher token cost
- tool-selection ambiguity
- lower model accuracy
- larger discovery payloads

Therefore large MCP platforms need a second capability:

> Tool search.

---

# 16. Tool Search

Instead of giving the model every tool:

```text
5000 tools
    ↓
LLM
```

the system can expose a smaller meta-capability:

```text
search_tools
```

The model asks:

```text
"Find tools for searching GitHub pull requests"
```

The system returns:

```text
github.pull_requests.search
github.pull_requests.get
github.pull_requests.list_reviews
```

The model then uses the selected tools.

---

# 17. Tool Search Architecture

```text
                         MCP Client
                             |
                             v
                       MCP Gateway
                             |
                             v
                       Tool Search
                             |
                  +----------+----------+
                  |                     |
                  v                     v
             Keyword Search       Semantic Search
                  |                     |
                  +----------+----------+
                             |
                             v
                       Tool Registry
                             |
                             v
                    Relevant Tool Set
```

---

# 18. Tool Registry

The registry becomes a first-class platform component.

Example record:

```json
{
  "name": "github.pull_requests.search",
  "description": "Search pull requests in a GitHub repository",
  "server": "github-mcp",
  "version": "1.4.0",
  "owner": "developer-platform",
  "tags": [
    "github",
    "pull-request",
    "search"
  ],
  "permissions": [
    "github.read"
  ]
}
```

The registry should contain metadata, not only the raw MCP tool definition.

---

# 19. Tool Metadata

Useful metadata includes:

```text
name
description
server
version
domain
owner
tags
permissions
risk level
availability
latency
cost
deprecated
created_at
updated_at
```

Potentially:

```text
input schema
output schema
examples
```

---

# 20. Keyword Tool Search

The simplest search implementation:

```text
query:
"search github pull requests"
```

Tokenize and match:

```text
github
pull
requests
search
```

against:

```text
tool name
description
tags
domain
```

This is easy to implement and may be sufficient initially.

---

# 21. Semantic Tool Search

At larger scale, use embeddings.

Architecture:

```text
Tool Definition
      |
      v
Embedding Model
      |
      v
Vector
      |
      v
Vector Database
```

When the agent searches:

```text
"find pull requests waiting for review"
```

the query is embedded:

```text
query
  ↓
embedding
  ↓
vector search
  ↓
top K tools
```

---

# 22. Hybrid Search

A production system should usually combine:

```text
keyword search
+
semantic search
+
metadata filters
```

Example:

```text
query:
"search GitHub PRs"

filters:
domain = github
permission = github.read
deprecated = false
```

Then rank the results.

---

# 23. Tool Search Pipeline

A typical pipeline:

```text
User intent
    |
    v
Query normalization
    |
    +--------------------+
    |                    |
    v                    v
Keyword Search      Vector Search
    |                    |
    +---------+----------+
              |
              v
         Candidate Set
              |
              v
           Filters
              |
              v
            Rank
              |
              v
           Top K
              |
              v
        Tool Definitions
```

---

# 24. Ranking

Possible ranking signals:

```text
semantic similarity
keyword relevance
exact name match
domain match
permission compatibility
tool health
latency
usage frequency
success rate
version
```

Example scoring:

```text
score =
    semantic_similarity * 0.50
  + keyword_score       * 0.20
  + domain_match        * 0.15
  + popularity          * 0.05
  + health              * 0.10
```

The exact algorithm should be measured rather than assumed.

---

# 25. Dynamic Tool Exposure

An advanced architecture does not expose all tools at once.

Instead:

```text
Initial MCP session
    |
    +--> search_tools
    +--> get_tool
```

The model discovers relevant tools dynamically.

Example:

```text
search_tools(
    "tools for managing GitHub pull requests"
)
```

returns:

```text
github.pull_requests.search
github.pull_requests.get
github.pull_requests.merge
```

The platform can then expose or describe only those tools.

---

# 26. Tool Search as a Control Plane

Tool discovery should eventually be separated from tool execution.

Think in terms of:

```text
CONTROL PLANE
    |
    +--> tool registry
    +--> metadata
    +--> discovery
    +--> permissions
    +--> versions
    +--> health
```

and:

```text
DATA PLANE
    |
    +--> actual MCP tool calls
    +--> routing
    +--> execution
```

This is an important production architecture distinction.

---

# 27. Control Plane vs Data Plane

## Control plane

Responsible for:

```text
Which tools exist?
Who owns them?
Where are they?
Which version is active?
Who can use them?
Are they healthy?
```

## Data plane

Responsible for:

```text
Execute this tool call.
```

Architecture:

```text
                 Control Plane
          +-------------------------+
          | Registry                |
          | Discovery               |
          | Permissions             |
          | Tool metadata           |
          +------------+------------+
                       |
                       v
                 Gateway / Router
                       |
                       v
                 MCP Servers
```

---

# 28. Tool Registry Synchronization

The aggregator needs to know which tools exist.

Possible approaches:

### Startup discovery

Aggregator connects to each server:

```text
server starts
    ↓
tools/list
    ↓
registry
```

Simple but not enough for dynamic environments.

### Periodic discovery

```text
every N seconds
    ↓
tools/list
    ↓
update registry
```

### Registration

MCP servers register themselves:

```text
MCP Server
    |
    | register
    v
Registry
```

### Event-driven

```text
MCP Server
    |
    | ToolUpdated event
    v
Registry
```

At large scale, event-driven synchronization can reduce unnecessary polling.

---

# 29. Service Discovery

The registry also needs to know where MCP servers live.

Example:

```text
github-mcp
    endpoint:
    mcp://github-mcp.internal
```

In Kubernetes this may map to:

```text
Service
    ↓
github-mcp.default.svc.cluster.local
```

The exact service discovery mechanism is infrastructure-specific.

---

# 30. Health Checks

The gateway should know whether an MCP server is healthy.

Example registry:

```text
github-mcp
    status: healthy

weather-mcp
    status: healthy

payments-mcp
    status: degraded
```

Health signals may include:

```text
connection state
last successful request
error rate
latency
heartbeat
```

---

# 31. Failure Isolation

A critical property:

```text
payments-mcp DOWN
```

must not cause:

```text
CRM MCP
GitHub MCP
Weather MCP
```

to become unavailable.

The aggregator should isolate failures per upstream server.

---

# 32. Circuit Breakers

If an upstream repeatedly fails:

```text
Gateway
    |
    v
Payments MCP
    |
    X
```

the gateway can open a circuit:

```text
CLOSED
  ↓
failures
  ↓
OPEN
  ↓
cooldown
  ↓
HALF-OPEN
  ↓
success
  ↓
CLOSED
```

This prevents sending continuous traffic to a failing service.

---

# 33. Timeouts

Timeouts should exist at multiple levels:

```text
Client timeout
      ↓
Gateway timeout
      ↓
Aggregator timeout
      ↓
MCP server timeout
      ↓
External API timeout
```

The total timeout budget must be coherent.

Avoid:

```text
gateway = 30s
server = 60s
provider = 30s
```

because the downstream request can outlive the upstream request.

---

# 34. Retries

Retries must be carefully controlled.

Do not blindly retry every MCP tool call.

A tool may be non-idempotent:

```text
payments.refund
create_ticket
send_email
```

Retrying may duplicate side effects.

Prefer retries only when:

- operation is known to be safe
- failure is transient
- timeout budget allows it

---

# 35. Idempotency

For write operations, production MCP infrastructure may need:

```text
idempotency key
```

Example:

```text
request_id = abc123
```

If the gateway retries:

```text
abc123
```

the downstream service can recognize the duplicate request.

---

# 36. Authentication

At platform scale:

```text
MCP Client
    |
    | identity/token
    v
MCP Gateway
```

The gateway authenticates the client.

Possible identity information:

```text
user
tenant
application
agent
service
```

Do not assume every MCP server needs to independently implement the same authentication flow.

---

# 37. Authorization

Authorization should operate at tool level.

Example policy:

```text
role: analyst

allow:
    analytics.read
    crm.contacts.search

deny:
    payments.refund
    crm.contacts.delete
```

This can be represented in the tool registry.

---

# 38. Tool Risk Classification

Not every tool has the same risk.

Example:

```text
LOW
    weather.forecast

MEDIUM
    crm.contacts.update

HIGH
    payments.refund
    production.deploy
```

The gateway can enforce stronger policies for high-risk tools.

For example:

```text
HIGH risk
    ↓
require additional authorization
```

---

# 39. Multi-Tenancy

A production platform may serve many tenants.

Architecture:

```text
Tenant A
   |
   v
Gateway
   |
   +--> allowed tools

Tenant B
   |
   v
Gateway
   |
   +--> different allowed tools
```

The registry can contain tenant-specific policy.

---

# 40. Rate Limiting

Rate limiting can happen at:

```text
tenant
user
application
MCP server
tool
```

Example:

```text
tenant: acme
    10,000 calls/minute

tool: payments.refund
    10 calls/minute
```

Use distributed rate limiting if gateway instances are horizontally scaled.

---

# 41. Horizontal Scaling

The gateway itself should be stateless where possible.

```text
                  Load Balancer
                       |
           +-----------+-----------+
           |           |           |
           v           v           v
       Gateway 1   Gateway 2   Gateway 3
           |           |           |
           +-----------+-----------+
                       |
                  MCP Servers
```

Shared state should live in external systems:

```text
Tool Registry
Policy Store
Rate Limit Store
Vector DB
```

---

# 42. Gateway vs Aggregator Deployment

There are multiple valid deployment models.

## Model A: One service

```text
MCP Gateway
    |
    +--> aggregation
    +--> discovery
    +--> auth
    +--> routing
```

Best for early platform maturity.

## Model B: Separate services

```text
Gateway
   |
   v
Aggregator
   |
   v
MCP Servers
```

Useful when responsibilities or scaling characteristics diverge.

## Model C: Gateway + control plane

```text
               Control Plane
          +---------------------+
          | Registry            |
          | Search              |
          | Policy              |
          +----------+----------+
                     |
                     v
                MCP Gateway
                     |
                     v
                MCP Servers
```

This is a strong production architecture.

---

# 43. Recommended Evolution Path

Do not jump directly from:

```text
1 MCP server
```

to:

```text
100 microservices
vector database
service mesh
```

Use incremental stages.

---

# 44. Stage 1 — Single MCP Server

```text
MCP Client
    |
    v
MCP Server
    |
    +--> Todo
    +--> Weather
    +--> Currency
```

Characteristics:

```text
simple
single deployment
few tools
```

Use the architecture from the recruitment task.

---

# 45. Stage 2 — Domain-Based MCP Servers

Split when domains become large:

```text
MCP Client
   |
   +--> Todo MCP
   +--> Weather MCP
   +--> Currency MCP
```

Each server can scale independently.

---

# 46. Stage 3 — MCP Aggregator

Introduce an aggregator when clients should not manage many MCP connections.

```text
                 MCP Client
                     |
                     v
                Aggregator
              /     |      \
             v      v       v
           Todo   Weather Currency
           MCP      MCP      MCP
```

Aggregator responsibilities:

```text
tools/list aggregation
tool routing
server lifecycle
```

---

# 47. Stage 4 — Central Tool Registry

When the number of tools grows:

```text
                   Registry
                  /        \
                 /          \
           Aggregator      Search
                |
                v
           MCP Servers
```

The registry becomes the source of truth for tool metadata.

---

# 48. Stage 5 — Tool Search

When the tool catalog becomes too large:

```text
MCP Client
    |
    v
search_tools
    |
    v
Tool Search
    |
    v
Registry
```

Start with:

```text
keyword + metadata filtering
```

Then add:

```text
semantic search
```

when there is enough tool volume to justify it.

---

# 49. Stage 6 — MCP Gateway

When cross-cutting requirements appear:

```text
authentication
authorization
rate limits
audit
quotas
routing
```

introduce the gateway.

```text
MCP Client
    |
    v
MCP Gateway
    |
    v
Aggregator
    |
    v
MCP Servers
```

---

# 50. Stage 7 — Production Control Plane

At larger scale:

```text
                         Control Plane
        +------------------------------------------+
        | Tool Registry                            |
        | Tool Search                              |
        | Policy / Authorization                   |
        | Server Registry                           |
        | Version Management                        |
        | Health / Metadata                         |
        +-------------------+----------------------+
                            |
                            v
                       MCP Gateway
                            |
                +-----------+-----------+
                |           |           |
                v           v           v
             CRM MCP     GitHub MCP   Payments MCP
```

This separates management from execution.

---

# 51. Stage 8 — Multi-Region / Large Scale

Only if required:

```text
                         Global DNS
                             |
               +-------------+-------------+
               |                           |
               v                           v
          EU Gateway                  US Gateway
               |                           |
        +------+-----+               +-----+------+
        |            |               |            |
      MCP          MCP             MCP          MCP
    servers      servers         servers      servers
```

The registry/control plane may become globally distributed or region-aware.

This is infrastructure-specific and should not be introduced before the business requires it.

---

# 52. Production Reference Architecture

A mature platform could look like:

```text
                              MCP Clients
                                  |
                                  v
                           Load Balancer
                                  |
                                  v
                        +-------------------+
                        |    MCP Gateway    |
                        |-------------------|
                        | Auth              |
                        | Authorization     |
                        | Rate Limiting     |
                        | Routing           |
                        | Audit             |
                        | Observability     |
                        +---------+---------+
                                  |
                                  v
                         +----------------+
                         | MCP Aggregator |
                         +-------+--------+
                                 |
                 +---------------+----------------+
                 |               |                |
                 v               v                v
             Tool Registry   Tool Search      Policy Store
                 |
                 |
                 v
        +--------+---------+---------+---------+
        |        |         |         |         |
        v        v         v         v         v
      CRM      GitHub    Weather   Payments  Analytics
      MCP       MCP       MCP        MCP       MCP
```

---

# 53. Control Plane Architecture

A dedicated control plane may contain:

```text
                 Control Plane
                       |
       +---------------+----------------+
       |               |                |
       v               v                v
 Tool Registry     Tool Search       Policy
       |               |                |
       v               v                v
 PostgreSQL        Vector DB       Policy Store
```

Possible technologies are implementation choices.

For example:

```text
PostgreSQL
OpenSearch
pgvector
Redis
```

should not be selected solely because they are popular.

Choose based on actual requirements.

---

# 54. Tool Registry Data Model

A registry entry might conceptually contain:

```text
Tool
-------------------------
id
name
description
server_id
version
schema
domain
owner
tags
permissions
risk_level
status
deprecated
created_at
updated_at
```

Server:

```text
MCP Server
-------------------------
id
name
endpoint
version
owner
environment
region
status
last_seen
```

---

# 55. Versioning

Tool definitions evolve.

Example:

```text
github.issue.create.v1
github.issue.create.v2
```

or:

```text
tool:
    github.issue.create

version:
    2
```

The gateway/registry should support:

```text
active version
deprecated version
compatible clients
migration period
```

Avoid silently changing a tool contract in incompatible ways.

---

# 56. Tool Deprecation

A production registry should support:

```text
ACTIVE
DEPRECATED
DISABLED
```

Example:

```text
github.issue.create
    status: deprecated
    replacement:
        github.issues.create
```

Search should prefer active tools.

---

# 57. Observability

At production scale, measure:

```text
tool calls
latency
success rate
error rate
timeouts
upstream failures
token/context impact
```

Useful dimensions:

```text
tool
server
tenant
client
region
status
```

Do not use unbounded high-cardinality values in metrics.

---

# 58. Distributed Tracing

For complex request paths:

```text
MCP Client
   |
   v
Gateway
   |
   v
Aggregator
   |
   v
MCP Server
   |
   v
External API
```

distributed tracing can show:

```text
gateway: 20ms
aggregator: 5ms
github-mcp: 120ms
github API: 110ms
```

This makes latency problems easier to diagnose.

---

# 59. Caching

Caching can be useful for read-heavy tools.

Examples:

```text
currencies_list
weather metadata
GitHub repository metadata
```

But do not blindly cache tool calls.

For each tool ask:

```text
Is it safe?
How stale can data be?
Does the tool have side effects?
Does the user expect fresh data?
```

Never cache non-idempotent side effects as if they were normal reads.

---

# 60. Tool Execution Policies

The gateway can eventually enforce policies before execution:

```text
Tool call
    |
    v
Policy
    |
    +--> allowed → execute
    |
    +--> denied  → reject
```

Policies can consider:

```text
user
tenant
tool
risk
environment
time
region
data classification
```

---

# 61. Human Approval for High-Risk Tools

For sensitive tools:

```text
payments.refund
production.deploy
database.delete
```

the platform may require:

```text
Agent
  ↓
Gateway
  ↓
Approval required
  ↓
Human approval
  ↓
Tool execution
```

This should be an explicit platform capability rather than hidden inside individual MCP servers.

---

# 62. MCP Server Contract

Even at large scale, individual MCP servers should remain relatively simple.

A domain MCP server should still follow:

```text
MCP
 ↓
Service
 ↓
Repository
 ↓
Adapter
```

The gateway should not move business logic out of the server.

The gateway handles platform concerns.

---

# 63. What Should NOT Move to the Gateway?

Avoid putting domain logic in the gateway.

Bad:

```text
Gateway
    |
    +--> calculate invoice
    +--> determine refund amount
    +--> transform CRM business object
```

Good:

```text
Gateway
    |
    +--> authenticate
    +--> authorize
    +--> route
    +--> rate limit
    |
    v
MCP Server
    |
    +--> business logic
```

The gateway is infrastructure/platform logic.

---

# 64. Team Ownership Model

A mature platform can establish:

```text
Platform Team
    |
    +--> Gateway
    +--> Aggregator
    +--> Registry
    +--> Search
    +--> Policies

Domain Teams
    |
    +--> CRM MCP
    +--> GitHub MCP
    +--> Payments MCP
    +--> Analytics MCP
```

This is one of the biggest benefits of the architecture.

---

# 65. Repository Structure at Organization Scale

Each MCP server can remain independently owned:

```text
repositories/

crm-mcp/
github-mcp/
payments-mcp/
weather-mcp/
analytics-mcp/
```

The platform infrastructure can be separate:

```text
mcp-gateway/
mcp-registry/
mcp-search/
```

A monorepo is also possible:

```text
platform/
    gateway/
    registry/
    search/

servers/
    crm/
    github/
    payments/
```

The decision between monorepo and polyrepo is organizational rather than an MCP architectural requirement.

---

# 66. Scaling Individual Servers

Because MCP servers are stateless where possible:

```text
CRM MCP
    |
    +--> replica 1
    +--> replica 2
    +--> replica 3
```

The gateway routes requests to healthy replicas.

If a server has local state, that state should be explicitly managed.

Prefer:

```text
MCP server
    ↓
external database/cache
```

over relying on local process memory for durable state.

---

# 67. Aggregator Connection Management

The aggregator may maintain connections to many MCP servers.

It should handle:

```text
connect
disconnect
reconnect
health
timeouts
backoff
```

Do not allow one unavailable MCP server to block startup or discovery of all other servers.

Prefer:

```text
CRM connected
GitHub connected
Weather unavailable
Payments connected
```

rather than:

```text
startup failed
```

for the entire platform.

---

# 68. Backpressure

At large scale:

```text
10,000 agents
    ↓
gateway
    ↓
MCP server
```

can overload a downstream service.

The gateway may need:

```text
concurrency limits
queue limits
rate limits
timeouts
load shedding
```

Example:

```text
payments-mcp
max concurrent requests = 100
```

Additional requests are rejected or queued according to policy.

---

# 69. Bulkheads

Use isolation between critical domains.

Conceptually:

```text
Gateway
   |
   +--> CRM pool
   |
   +--> GitHub pool
   |
   +--> Payments pool
```

If GitHub receives huge traffic, it should not consume all gateway resources and starve Payments.

---

# 70. Security Boundary

The gateway can become the primary security boundary:

```text
Untrusted MCP Clients
          |
          v
     MCP Gateway
          |
          v
Trusted MCP Network
```

However, defense in depth is still important.

Individual MCP servers should not blindly trust the network.

---

# 71. Recommended Evolution for This Project

For this specific recruitment project, the appropriate future roadmap is:

```text
NOW

Single MCP Server
    |
    +--> Todos
    +--> Weather
    +--> Currency
```

Then:

```text
STEP 1

Separate MCP servers by domain
```

Then:

```text
STEP 2

MCP Aggregator
    |
    +--> Domain MCPs
```

Then:

```text
STEP 3

Central Tool Registry
```

Then:

```text
STEP 4

Tool Search
    |
    +--> keyword search
    +--> metadata filtering
```

Then:

```text
STEP 5

Semantic / hybrid tool search
```

Then:

```text
STEP 6

MCP Gateway
    |
    +--> auth
    +--> authorization
    +--> rate limits
    +--> routing
    +--> audit
```

Then:

```text
STEP 7

Control Plane
    |
    +--> Registry
    +--> Search
    +--> Policies
    +--> Health
    +--> Versioning
```

---

# 72. What We Should Keep From the Original Architecture

Even after introducing all this infrastructure, the internal MCP server architecture should remain:

```text
MCP Handler
    ↓
Domain/Application Service
    ↓
Repository Interface
    ↑
Infrastructure Adapter
    ↓
External API
```

Do not throw away the original architecture.

The gateway is an additional platform layer, not a replacement for domain architecture.

---

# 73. Final Architecture Principle

The system evolves by adding infrastructure around stable domain boundaries:

```text
             ┌──────────────────────────────┐
             │        Control Plane         │
             │                              │
             │ Registry | Search | Policies │
             └──────────────┬───────────────┘
                            |
                            v
                    ┌───────────────┐
                    │ MCP Gateway   │
                    │               │
                    │ Auth          │
                    │ Rate limits   │
                    │ Routing       │
                    │ Audit         │
                    └───────┬───────┘
                            |
                            v
                    ┌───────────────┐
                    │ MCP Aggregator│
                    └───────┬───────┘
                            |
          +-----------------+------------------+
          |                 |                  |
          v                 v                  v
       CRM MCP          GitHub MCP        Payments MCP
          |                 |                  |
       Domain            Domain             Domain
       Logic             Logic              Logic
          |                 |                  |
       Adapters           Adapters           Adapters
          |                 |                  |
      External          External           External
       APIs              APIs               APIs
```

The important separation is:

```text
Control Plane
    ↓
manages and discovers tools

Gateway
    ↓
controls access and traffic

Aggregator
    ↓
combines MCP servers

MCP Server
    ↓
implements domain capabilities

Infrastructure Adapter
    ↓
integrates external systems
```

---

# 74. Final Mental Model

A useful way to remember the architecture is:

### MCP Server

> "I provide these capabilities."

### Aggregator

> "I combine capabilities from many MCP servers."

### Tool Registry

> "I know what capabilities exist and where they live."

### Tool Search

> "I find the most relevant capabilities for a given intent."

### Gateway

> "I control who can use capabilities and how traffic reaches them."

### Control Plane

> "I manage the platform that makes all of this possible."

The architecture should evolve in that order only when the corresponding scaling problem actually appears.
