# Architecture Principles & Engineering Guidelines

## Purpose

This document defines the architectural principles used in the MCP recruitment task.

The goal is not to apply every possible software architecture pattern, but to establish a small set of principles that keep the code:

- easy to understand
- easy to test
- independent from external providers
- easy to extend
- resistant to unnecessary coupling
- idiomatic for Go
- scalable from a small project to a larger system

The main principles are:

1. Clean Architecture / Dependency Inversion
2. Vertical Slicing / Package by Domain
3. Domain-Owned Interfaces
4. Infrastructure Adapters
5. Explicit DTO → Domain Mapping
6. Thin Transport / MCP Layer
7. Dependency Injection
8. Composition Root
9. Context Propagation
10. Explicit Error Boundaries
11. Small Interfaces
12. Testability by Design
13. Avoid Premature Abstraction
14. Bounded External Dependencies

---

# 1. Clean Architecture

## Core idea

Dependencies must point inward toward the application/domain logic.

```text
             Transport
                |
                v
        Application / Domain
                |
                v
           Interfaces
                ^
                |
         Infrastructure
```

The domain/application layer defines what it needs.

Infrastructure provides implementations.

For example:

```text
currency
    |
    | RateRepository
    v
Frankfurter adapter
```

The `currency` package does not know that Frankfurter exists.

It only knows:

```go
type Repository interface {
    GetRate(ctx context.Context, base, quote string) (Rate, error)
}
```

The infrastructure layer implements it.

---

## Why

This protects business/application logic from infrastructure.

If Frankfurter is replaced with another provider:

```text
Frankfurter
     ↓
     X

NewProvider
     ↓
     ✓
```

the currency service should not need to change.

---

## What this does NOT mean

Clean Architecture does not mean:

- dozens of interfaces
- dozens of packages
- a framework
- every function wrapped in an interface
- five layers for every CRUD operation
- complicated dependency injection

The goal is dependency control, not architectural ceremony.

---

# 2. Vertical Slicing / Package by Domain

Instead of organizing the project globally by technical layer:

```text
controllers/
services/
repositories/
models/
```

organize it by business/domain capability:

```text
internal/
├── todo/
├── weather/
└── currency/
```

Each domain contains the code required to implement that capability.

For example:

```text
internal/currency/

model.go
repository.go
service.go
errors.go
service_test.go
```

Infrastructure is then organized around external providers:

```text
internal/infrastructure/

dummyjson/
openmeteo/
frankfurter/
```

---

## Why Vertical Slicing

With global technical layers:

```text
services/
    todo.go
    weather.go
    currency.go

repositories/
    todo.go
    weather.go
    currency.go

handlers/
    todo.go
    weather.go
    currency.go
```

adding one feature requires jumping between many unrelated directories.

With vertical slicing:

```text
currency/
    model.go
    repository.go
    service.go
```

the domain boundary is visible immediately.

---

# 3. Domain Boundaries

Each external integration maps to a meaningful domain:

```text
DummyJSON      → Todo
Open-Meteo     → Weather
Frankfurter    → Currency
```

The domain should represent the application's concept, not the external provider.

Prefer:

```text
currency
```

over:

```text
frankfurter
```

for domain logic.

Frankfurter is an implementation detail.

---

# 4. Domain-Owned Interfaces

Interfaces should generally be declared by the package that consumes them.

Example:

```go
package currency

type Repository interface {
    GetRate(
        ctx context.Context,
        base string,
        quote string,
    ) (Rate, error)
}
```

The currency service needs a repository.

Therefore the currency package defines the interface.

The infrastructure package implements it.

```text
currency
    |
    | defines
    v
Repository interface
    ^
    |
    | implements
    |
Frankfurter adapter
```

---

## Why

This follows Dependency Inversion.

The high-level module says:

> "Here is the capability I require."

The low-level module says:

> "I can provide that capability."

The high-level module does not need to know how.

---

# 5. Interfaces Should Be Small

Prefer:

```go
type Repository interface {
    GetRate(ctx context.Context, base, quote string) (Rate, error)
}
```

over:

```go
type Repository interface {
    GetRate(...)
    GetRates(...)
    ListCurrencies(...)
    Create(...)
    Update(...)
    Delete(...)
    Search(...)
    Export(...)
}
```

A consumer should depend only on what it actually needs.

This is the Interface Segregation Principle.

---

# 6. Infrastructure Adapters

Infrastructure contains provider-specific implementations.

Example:

```text
internal/infrastructure/frankfurter/
    client.go
    dto.go
```

The adapter knows:

- HTTP
- URLs
- query parameters
- JSON
- HTTP status codes
- provider-specific response formats

The domain does not.

---

## Adapter responsibility

The Frankfurter adapter can know:

```text
/v2/rate/EUR/USD
```

The currency service should not.

The currency service should know only:

```go
repo.GetRate(ctx, "EUR", "USD")
```

---

# 7. DTO → Domain Mapping

External API models are not domain models.

Example:

```text
Frankfurter JSON
       |
       v
Frankfurter DTO
       |
       v
mapping
       |
       v
currency.Rate
```

External DTO:

```go
type rateDTO struct {
    Date  string  `json:"date"`
    Base  string  `json:"base"`
    Quote string  `json:"quote"`
    Rate  float64 `json:"rate"`
}
```

Domain:

```go
type Rate struct {
    Date  time.Time
    Base  string
    Quote string
    Rate  float64
}
```

---

## Why map?

Without mapping:

```text
domain → Frankfurter JSON structure
```

the domain becomes coupled to the provider.

With mapping:

```text
provider → adapter → domain
```

the provider becomes replaceable.

---

# 8. Do Not Put JSON Tags in Domain Models

Avoid:

```go
type Rate struct {
    Date string `json:"date"`
}
```

if `Rate` is a domain model.

Prefer:

```go
type Rate struct {
    Date time.Time
}
```

JSON tags belong to transport/infrastructure DTOs.

This keeps the domain independent from serialization.

---

# 9. Thin MCP / Transport Layer

The MCP layer is a transport adapter.

Its job is to:

1. receive MCP input
2. validate/parse transport-level input
3. convert it to application parameters
4. call the service
5. convert the result to MCP output

Conceptually:

```text
MCP request
    |
    v
MCP handler
    |
    v
Service
    |
    v
Domain
```

The handler should NOT:

- construct URLs
- call `http.Client`
- decode upstream JSON
- calculate currency conversion
- contain business rules
- read environment variables
- instantiate dependencies

---

# 10. Business Logic Belongs in Services

For example:

```text
currency_convert
```

should not be:

```text
MCP handler
    ↓
Frankfurter API
    ↓
amount * rate
```

Instead:

```text
MCP handler
    ↓
Currency Service
    ↓
Rate Repository
    ↓
Frankfurter
```

The service performs:

```text
validate amount
validate currencies
normalize currencies
get rate
calculate conversion
return domain result
```

This makes the logic independently testable.

---

# 11. Separate Validation by Responsibility

Not all validation belongs in the same layer.

### Transport-level validation

Examples:

- malformed JSON
- missing required MCP fields
- incorrect primitive types

Handled by MCP/SDK.

### Application/domain validation

Examples:

```text
amount > 0
latitude between -90 and 90
forecast days between 1 and 16
currency code is 3 letters
```

Handled by services.

### Provider validation

Examples:

- provider-specific query requirements
- provider-specific response structure

Handled by infrastructure adapters.

---

# 12. Dependency Injection

Dependencies are passed explicitly.

Prefer:

```go
func NewService(repo Repository) *Service
```

over:

```go
func NewService() *Service {
    return &Service{
        repo: NewFrankfurterClient(),
    }
}
```

The second approach hides dependencies and makes testing harder.

---

# 13. Composition Root

All concrete dependencies should be constructed in one place.

In this project:

```text
cmd/server/main.go
```

is the composition root.

Conceptually:

```text
config
  ↓
provider clients
  ↓
services
  ↓
MCP server
```

Example:

```text
main.go

config
    ↓
FrankfurterClient (creates its own HTTP client via pkg/http)
    ↓
CurrencyService
    ↓
MCPServer
```

The rest of the application receives already-constructed dependencies.

---

# 14. No Global Mutable State

Avoid:

```go
var client *http.Client
```

or:

```go
var currencyService *Service
```

Global mutable state makes:

- tests harder
- concurrency harder to reason about
- dependencies invisible
- lifecycle management harder

Prefer explicit ownership and dependency injection.

---

# 15. Shared Infrastructure vs Shared Domain Logic

It is fine to share technical infrastructure.

For example:

```text
              pkg/http (shared package)
                 |         |         |
                 v         v         v
            DummyJSON  OpenMeteo  Frankfurter
         (own client) (own client) (own client)
```

Each adapter creates its own `http.Client` instance via the shared `pkg/http` package. The package is shared; the instances are not.

But do not create a generic domain client such as:

```go
GenericAPIClient
```

containing:

```text
getTodos()
getWeather()
getRates()
```

Shared infrastructure is useful.

Shared unrelated domain logic is usually a coupling smell.

---

# 16. Context Propagation

Context must flow through the entire request:

```text
MCP request
    ↓
handler
    ↓
service
    ↓
repository
    ↓
HTTP request
```

Use:

```go
func (s *Service) Get(
    ctx context.Context,
    id int,
) (Todo, error)
```

and:

```go
req, err := http.NewRequestWithContext(ctx, ...)
```

Do not replace the incoming context with:

```go
context.Background()
```

inside request processing.

---

# 17. Explicit Error Boundaries

Errors should retain their meaning across layers.

Example:

```text
Frankfurter
    ↓
HTTP 500
    ↓
adapter
    ↓
ErrUpstream
    ↓
service
    ↓
MCP handler
```

Use wrapping:

```go
fmt.Errorf("get exchange rate: %w", err)
```

and inspect with:

```go
errors.Is(...)
errors.As(...)
```

Do not expose raw upstream responses to users.

---

# 18. Domain Errors vs Infrastructure Errors

Infrastructure may know:

```text
HTTP 404
HTTP 429
HTTP 500
timeout
```

The application may care about:

```text
not found
upstream unavailable
invalid input
```

Translate technical errors at appropriate boundaries.

Example:

```text
HTTP 404
   ↓
ErrNotFound
```

while preserving the original error where useful.

---

# 19. Testability by Design

Architecture should make tests natural.

For service tests:

```text
Service
   ↓
Fake Repository
```

No network.

For adapter tests:

```text
Client
   ↓
httptest.Server
```

No real internet.

For MCP integration tests:

```text
MCP Client
   ↓
MCP Server
   ↓
Fake Services
```

This gives deterministic tests at every layer.

---

# 20. Unit Tests Should Test Behavior

Do not test implementation details such as:

```text
private helper X was called exactly once
```

unless that behavior is important.

Prefer:

```text
given invalid currency
when Convert is called
then ErrInvalidInput is returned
```

or:

```text
given rate 1.2 and amount 100
when Convert is called
then converted amount is 120
```

---

# 21. HTTP Tests With `httptest.Server`

External API adapters should be tested against a local HTTP server.

Example:

```text
Test
  ↓
httptest.Server
  ↓
fake API response
```

This verifies:

- HTTP method
- URL
- query parameters
- headers
- response parsing
- error handling

without requiring the actual provider.

---

# 22. MCP Integration Tests

MCP should be tested through the actual MCP SDK rather than by directly invoking handler functions only.

The test should verify:

```text
MCP Client
    |
    | tools/list
    v
MCP Server
```

and then:

```text
MCP Client
    |
    | call tool
    v
MCP Server
    |
    v
Handler
    |
    v
Service
```

This verifies that the actual MCP contract is wired correctly.

---

# 23. MCP Protocol Should Not Be Implemented Manually

Do not write:

```text
POST /json-rpc
tools/list handler
tools/call handler
manual JSON-RPC parser
```

The official MCP SDK owns protocol mechanics.

Application code owns:

```text
tool definitions
tool handlers
services
domain logic
```

This is an important separation.

---

# 24. External APIs Are Implementation Details

The application should think:

```text
Get exchange rate
```

not:

```text
Call Frankfurter /v2/rate/EUR/USD
```

Similarly:

```text
Get weather forecast
```

not:

```text
Call Open-Meteo with query parameters
```

Provider-specific details belong inside adapters.

---

# 25. Replaceability

A useful architectural test is:

> "How much code changes if the external provider changes?"

For example:

```text
Frankfurter
     ↓
   ECB API
```

Ideally:

```text
currency/
    model.go
    repository.go
    service.go
```

does not change.

Only the infrastructure adapter should need to change, assuming the new provider can satisfy the existing repository contract.

---

# 26. Avoid Premature Abstraction

Do not create abstractions because two pieces of code happen to look similar.

Example:

Avoid:

```text
GenericProvider
GenericRepository
GenericService
GenericHTTPAdapter
```

unless there is a real shared behavior.

Prefer three explicit adapters:

```text
DummyJSONClient
OpenMeteoClient
FrankfurterClient
```

This is often clearer and more maintainable.

---

# 27. Abstraction Should Follow Stable Boundaries

Good abstraction:

```text
Currency Service
       |
       v
RateRepository
```

because the application genuinely needs the capability of obtaining rates.

Questionable abstraction:

```text
BaseAPIClient
```

created only because all APIs use HTTP.

The shared HTTP client is infrastructure reuse, not necessarily a domain abstraction.

---

# 28. Single Responsibility

A component should have one primary reason to change.

### MCP handler changes when:

- MCP input/output contract changes

### Service changes when:

- business/application rules change

### Repository interface changes when:

- application capability requirements change

### Infrastructure adapter changes when:

- provider API changes

### Config changes when:

- deployment configuration changes

This makes change propagation predictable.

---

# 29. Dependency Direction Is More Important Than Folder Structure

Folders are not architecture by themselves.

This:

```text
internal/
├── service/
├── repository/
└── handler/
```

can still be well designed.

And this:

```text
internal/
├── todo/
├── weather/
└── currency/
```

can still be badly coupled.

The important property is:

```text
high-level logic
        ↓
interfaces
        ↑
implementations
```

Package structure is used to make those boundaries visible and enforceable.

---

# 30. Keep Domain Models Stable

Domain models should represent concepts used by the application.

Example:

```go
type Conversion struct {
    Amount    float64
    Rate      float64
    Converted float64
}
```

The model should not change because Frankfurter renames a JSON property.

Only the adapter/mapping should change.

---

# 31. Keep Transport Models Separate

MCP input:

```go
type currencyConvertInput struct {
    Amount float64 `json:"amount"`
    From   string  `json:"from"`
    To     string  `json:"to"`
}
```

Application params:

```go
type ConversionParams struct {
    Amount float64
    From   string
    To     string
}
```

Domain result:

```go
type Conversion struct {
    Date      time.Time
    From      string
    To        string
    Amount    float64
    Rate      float64
    Converted float64
}
```

These types may look similar, but they represent different boundaries.

Do not merge them prematurely.

---

# 32. Why Not One Universal Model?

Avoid:

```text
API DTO = domain model = MCP response
```

because then one external API change can affect every layer.

Instead:

```text
External DTO
     ↓
Domain Model
     ↓
MCP Output
```

Each boundary owns its representation.

---

# 33. Bounded Interfaces

Repositories should expose application capabilities, not raw HTTP.

Bad:

```go
DoRequest(ctx, method, url, body)
```

Good:

```go
GetRate(ctx, base, quote)
```

The first leaks infrastructure.

The second expresses business/application intent.

---

# 34. Security / Safety Boundaries

Even public APIs should be treated as untrusted external dependencies.

Validate:

- input ranges
- response structure
- response sizes where appropriate
- timeouts
- upstream status codes

Never assume the upstream response is always valid.

---

# 35. Resource Limits

External API integrations should have bounded inputs.

Examples:

```text
todos limit <= 100
weather forecast days <= 16
currency quotes <= 20
```

This prevents accidentally generating excessively large requests or responses.

---

# 36. Logging Principle

Logs should describe events, not dump data.

Good:

```text
tool=currency_convert
duration_ms=82
status=success
```

Bad:

```text
full upstream JSON response
```

Never log:

- secrets
- tokens
- unnecessary personal data
- full request/response bodies by default

For stdio MCP servers:

```text
stdout = protocol
stderr = logs
```

---

# 37. Production Readiness Without Overengineering

Production-ready does not mean:

```text
Kubernetes
Kafka
Redis
microservices
distributed tracing
service mesh
```

For this project, production readiness means:

```text
timeouts
context
validation
clear boundaries
testability
structured logging
error handling
race safety
```

The architecture should be ready to evolve without pretending the assignment is a distributed production platform.

---

# 38. The Main Architectural Flow

The preferred request flow is:

```text
MCP Client
    |
    | MCP tool call
    v
MCP Handler
    |
    | typed application input
    v
Domain/Application Service
    |
    | interface
    v
Repository
    ^
    |
    | implementation
    |
Infrastructure Adapter
    |
    | HTTP
    v
External API
```

Response:

```text
External API
    |
    v
Provider DTO
    |
    v
Mapper
    |
    v
Domain Model
    |
    v
Service
    |
    v
MCP Handler
    |
    v
MCP Client
```

---

# 39. Mental Model

When adding a new feature, ask these questions in order:

### 1. What is the domain capability?

Example:

```text
Get exchange rate
```

### 2. What does the application need?

Define:

```go
type Repository interface {
    GetRate(...)
}
```

### 3. What business rules exist?

Implement them in the service.

### 4. Which external system provides the capability?

Create an adapter.

### 5. What does the external API look like?

Create provider DTOs.

### 6. How do we map provider data to domain data?

Create explicit mapping.

### 7. How is it exposed to the user/agent?

Add an MCP tool.

### 8. How can every layer be tested independently?

Use:

```text
fake repository
httptest.Server
MCP integration test
```

---

# 40. Final Principles Checklist

Before merging a feature, verify:

- [ ] Domain does not depend on infrastructure
- [ ] Domain does not depend on MCP
- [ ] Interfaces belong to consumers
- [ ] Infrastructure implements interfaces
- [ ] Provider DTOs do not leak into domain
- [ ] DTO → domain mapping exists
- [ ] MCP handlers are thin
- [ ] Business logic lives in services
- [ ] Dependencies are injected
- [ ] Composition happens in `main`
- [ ] Context is propagated
- [ ] External requests have timeouts
- [ ] Inputs are bounded and validated
- [ ] Errors are meaningful and wrapped
- [ ] External integrations have isolated tests
- [ ] Services can be tested without HTTP
- [ ] MCP integration can be tested without real APIs
- [ ] No unnecessary abstractions were introduced
- [ ] No global mutable state was introduced
- [ ] Package boundaries reflect domain boundaries

---

# 41. Summary

The architecture can be reduced to one sentence:

> **Organize code around domain capabilities, keep business logic independent from infrastructure, let consumers own small interfaces, isolate external systems behind adapters, and keep transport layers thin.**

For this MCP server:

```text
                MCP
                 |
                 v
        +----------------+
        | Domain Services|
        +-------+--------+
                |
         small interfaces
                |
                v
        +----------------+
        |   Adapters     |
        +-------+--------+
                |
                v
        External APIs
```

And the most important dependency rule is:

```text
Application / Domain
        |
        v
     Interface
        ^
        |
 Infrastructure
```

not:

```text
Application
    |
    v
Frankfurter/Open-Meteo/DummyJSON
```

This gives the project a clean, testable architecture while keeping the implementation appropriately simple for a recruitment task.
