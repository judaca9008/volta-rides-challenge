# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run Commands

```bash
go get                 # Install dependencies
go run main.go         # Run the server
go build -o volta-router  # Build the binary
./volta-router         # Run the binary

# Generate test data
go run cmd/generate_data/main.go

# Run tests (when implemented)
go test ./tests -v

# Run specific test
go test ./tests -run TestApprovalRateCalculation -v
```

## Environment Variables

- `PORT` — Server listen port (default: `8080`)
- `ENVIRONMENT` — Deployment environment (`development`, `staging`, `production`)

Example:
```bash
PORT=3000 ENVIRONMENT=production go run main.go
```

---

## Architecture

This is a **smart payment router** built with Go and Echo. It does not own a database in this POC — it uses an in-memory store. In production, it would integrate with PostgreSQL for transaction storage and Redis for caching.

### Layered Structure

```
Controllers → Services → Storage
```

- **`controllers/`** — Echo HTTP handlers. Parse request params/body, call services, return JSON responses.
- **`services/`** — Business logic: approval rate calculation, processor selection, routing decisions.
- **`storage/`** — In-memory store with thread-safe operations (sync.RWMutex). Stores transactions and routing decisions.
- **`models/`** — DTOs with JSON tags. All requests/responses are in `snake_case` following API conventions.

### Routing

Routes are configured in `routers/routers.go`. All routes are grouped under `/volta-router` with `/v1` sub-group. Route path constants live in `common/constants/`.

### Middleware Stack (applied in order)

1. **DataDog APM** — Distributed tracing (`echoDatadog.Middleware`)
2. **Logging** — Request logging with duration and status code
3. **Trace ID** — UUID-based trace propagation via `X-Trace-Id` header
4. **CORS** — Cross-origin support (development)
5. **Recovery** — Panic recovery middleware

---

## Core Business Logic

### Approval Rate Calculation

**Formula:**
```
Approval Rate = (Approved Transactions / Total Transactions) × 100
```

**Time Window:** Last 15 minutes (configurable in `config/config.go`)

**Implementation** (`services/routing_service.go`):
1. Get transactions for processor + country within time window
2. Count approved vs total
3. Return percentage

### Routing Logic

**Algorithm** (`services/routing_service.go → SelectBestProcessor`):
1. Validate country is supported (BR, MX, CO)
2. Get all processors for that country from config
3. Calculate approval rate for each processor
4. Select processor with highest approval rate
5. Classify risk level:
   - **Low**: > 80%
   - **Medium**: 70-80%
   - **High**: < 70%
6. Record decision in store for statistics
7. Return routing response

**Edge Cases:**
- **No data for processor**: Returns 0% approval rate, excluded from routing
- **All processors < 70%**: Routes to "least bad" processor, flags as "high-risk"
- **Unsupported country**: Returns 400 error
- **No processors available**: Returns 503 error

---

## Key Components

### Storage Layer (`storage/store.go`)

**InMemoryStore** with `sync.RWMutex` for thread-safe operations:

- `AddTransaction(tx Transaction)` — Add single transaction
- `AddTransactions(txs []Transaction)` — Batch add
- `GetTransactionsByWindow(processor, country string, window time.Duration)` — Time-window query
- `RecordRoutingDecision(decision RoutingDecision)` — Track routing decision
- `GetRoutingStats(limit int)` — Get last N decisions distribution

**Thread Safety:**
- Reads use `mu.RLock()` (multiple readers can access simultaneously)
- Writes use `mu.Lock()` (exclusive access)

### Service Layer (`services/routing_service.go`)

**RoutingService** orchestrates all business logic:

- `CalculateApprovalRate(processor, country string) float64` — Calculate approval rate for processor
- `SelectBestProcessor(req RoutingRequest) (*RoutingResponse, error)` — Main routing logic
- `GetAllProcessorStats() []ProcessorStats` — Get stats for all 9 processors
- `GetProcessorStats(name string) (*ProcessorStats, error)` — Get stats for specific processor
- `GetRoutingStats() RoutingStats` — Get routing decision distribution

### Controller Layer (`controllers/`)

**RoutingController** (`routing_controller.go`):
- `RouteTransaction` — POST /volta-router/v1/route
- `GetProcessorHealth` — GET /volta-router/v1/processors
- `GetProcessorByName` — GET /volta-router/v1/processors/:name
- `GetRoutingStats` — GET /volta-router/v1/routing/stats

**DataController** (`data_controller.go`):
- `LoadTestData` — POST /volta-router/v1/transactions/load

**HealthCheckController** (`health_controller.go`):
- `HealthCheck` — GET /health

---

## Configuration (`config/config.go`)

### RoutingConfig

- `TimeWindow` — Default: 15 minutes
- `HighRiskThreshold` — Default: 70%
- `MediumRiskThreshold` — Default: 80%
- `CircuitBreakerThreshold` — Default: 60% (for future use)

### ProcessorsByCountry

Static mapping of countries to processors:
```go
"BR": {"RapidPay_BR", "TurboAcquire_BR", "PayFlow_BR"},
"MX": {"RapidPay_MX", "TurboAcquire_MX", "PayFlow_MX"},
"CO": {"RapidPay_CO", "TurboAcquire_CO", "PayFlow_CO"},
```

**Production Note:** This should be database-driven, not hard-coded.

---

## Test Data (`data/generator/generator.go`)

### GenerateTestTransactions(count int)

Creates realistic test transactions with:
- 9 processors across 3 countries
- Timestamps spread over 2-3 hours
- Realistic approval rate patterns:
  - **Strong performers** (90-92%): RapidPay_* processors
  - **Normal** (82-85%): TurboAcquire_* processors
  - **Bad period** (55-90%): PayFlow_* processors (55% for 20-minute window)

**Why Bad Periods?**
Demonstrates router's ability to detect and avoid failing processors.

### SaveTransactionsToFile / LoadTransactionsFromFile

- Save: Writes to `data/test_transactions.json`
- Load: Reads from file, used by `/transactions/load` endpoint

---

## API Endpoints

### POST /volta-router/v1/route

**Request:**
```json
{
  "amount": 100.00,
  "currency": "BRL",
  "country": "BR"
}
```

**Response:**
```json
{
  "processor": "RapidPay_BR",
  "approval_rate": 92.5,
  "risk_level": "low",
  "reason": "Highest approval rate for BR",
  "timestamp": "2024-02-26T15:30:00Z"
}
```

**Validation:**
- `amount` must be > 0
- `currency` must be exactly 3 characters
- `country` must be exactly 2 characters (BR, MX, CO)

### GET /volta-router/v1/processors

Returns all processor stats:
```json
{
  "processors": [
    {
      "name": "RapidPay_BR",
      "country": "BR",
      "approval_rate": 92.5,
      "transaction_count": 145,
      "last_updated": "2024-02-26T15:30:00Z"
    }
  ]
}
```

### GET /volta-router/v1/processors/:name

Returns specific processor stats (same format as above, single object).

### GET /volta-router/v1/routing/stats

Returns routing decision distribution:
```json
{
  "total_decisions": 50,
  "distribution": {
    "RapidPay_BR": 35,
    "PayFlow_BR": 12,
    "TurboAcquire_BR": 3
  },
  "window": "last_50_decisions"
}
```

### POST /volta-router/v1/transactions/load

Loads test data from `data/test_transactions.json`:
```json
{
  "message": "Test data loaded successfully",
  "transactions_loaded": 540
}
```

### GET /health

Simple health check:
```json
{
  "status": "Ok"
}
```

---

## Testing

### Manual Testing Workflow

```bash
# 1. Start server
PORT=8080 go run main.go

# 2. Load test data
curl -X POST http://localhost:8080/volta-router/v1/transactions/load

# 3. Test routing (Brazil)
curl -X POST http://localhost:8080/volta-router/v1/route \
  -H "Content-Type: application/json" \
  -d '{"amount": 100, "currency": "BRL", "country": "BR"}'

# Expected: RapidPay_BR with ~92% approval rate

# 4. Check processor health
curl http://localhost:8080/volta-router/v1/processors

# Expected: 9 processors with approval rates

# 5. Test unsupported country (should fail)
curl -X POST http://localhost:8080/volta-router/v1/route \
  -H "Content-Type: application/json" \
  -d '{"amount": 100, "currency": "USD", "country": "US"}'

# Expected: 400 error with "country US not supported"
```

### Unit Test Structure (when implemented)

```go
// tests/routing_test.go
func TestCalculateApprovalRate(t *testing.T) {
    // Test approval rate calculation accuracy
}

func TestSelectBestProcessor(t *testing.T) {
    // Test processor selection logic
}

func TestRiskClassification(t *testing.T) {
    // Test risk level classification
}

func TestConcurrentAccess(t *testing.T) {
    // Test thread-safe storage operations
}
```

---

## Dependencies

### Core
- `github.com/labstack/echo/v4` — Web framework (Yuno standard)
- `github.com/google/uuid` — UUID generation for trace IDs and transaction IDs
- `github.com/go-playground/validator/v10` — Request validation
- `gopkg.in/DataDog/dd-trace-go.v1` — DataDog APM integration

### Why These Choices?
- **Echo**: Fast, minimal, used across Yuno's Go services
- **DataDog**: Standard APM for Yuno
- **Validator**: Declarative validation with struct tags

---

## Production Considerations

### What's Simplified for POC

1. **Storage**: In-memory (should be PostgreSQL + Redis)
2. **Middleware**: Simplified logging (should use Yuno's structured logger)
3. **Configuration**: Hard-coded processor mappings (should be database-driven)
4. **Authentication**: None (should require API keys/JWT)
5. **Rate Limiting**: None (should limit per client)

### Migration Path to Production

**Phase 1: Database Integration**
1. Replace `InMemoryStore` with PostgreSQL
2. Add Redis for approval rate caching
3. Keep service interface unchanged

**Phase 2: Observability**
1. Add Prometheus metrics
2. Set up Grafana dashboards
3. Configure alerts for approval rate drops

**Phase 3: Advanced Features**
1. Circuit breaker implementation
2. Cost-aware routing
3. Failover ranking

---

## Common Tasks

### Add a New Country

1. Add processor mappings in `config/config.go`:
```go
ProcessorsByCountry = map[string][]string{
    "AR": {"RapidPay_AR", "TurboAcquire_AR", "PayFlow_AR"},
}
```

2. Generate test data for new processors in `data/generator/generator.go`

3. No code changes needed in services or controllers (country-agnostic logic)

### Change Time Window

Edit `config/config.go`:
```go
TimeWindow: 20 * time.Minute,  // Changed from 15 to 20 minutes
```

### Add New Risk Threshold

Edit `config/config.go`:
```go
CriticalRiskThreshold: 50.0,  // New threshold for critical risk
```

Then update `services/routing_service.go → classifyRiskLevel()` to use new threshold.

---

## Debugging Tips

### High Approval Rates Not Matching Expected

**Check:**
1. Time window configuration (are transactions within window?)
2. Transaction timestamps (use `curl /volta-router/v1/processors` to see transaction counts)
3. Test data generation (verify patterns in `data/generator/generator.go`)

### Routing Always Returns Same Processor

**Likely Cause:** All transactions outside time window

**Fix:** Regenerate test data with recent timestamps:
```bash
go run cmd/generate_data/main.go
curl -X POST http://localhost:8080/volta-router/v1/transactions/load
```

### DataDog Traces Not Showing

**Check:**
1. DataDog agent is running locally
2. `DD_AGENT_HOST` environment variable is set (if needed)
3. Service name matches in DataDog UI: `volta-router`

---

## Project Structure Summary

```
volta-rides-challenge/
├── main.go                   # Entry point, DataDog tracer init
├── cmd/
│   ├── httpServer/          # Server initialization
│   └── generate_data/       # Test data generator CLI
├── controllers/             # HTTP handlers
├── services/                # Business logic
├── storage/                 # In-memory store
├── models/                  # DTOs and data structures
├── routers/                 # Route configuration
│   └── middleware/         # Logging, trace ID
├── config/                  # Configuration
├── common/constants/        # Route paths, service name
├── data/
│   ├── generator/          # Test data generation logic
│   └── test_transactions.json  # Generated test data
└── tests/                   # Unit tests (when implemented)
```

---

## Quick Reference

### Start Server
```bash
PORT=8080 go run main.go
```

### Load Test Data
```bash
curl -X POST http://localhost:8080/volta-router/v1/transactions/load
```

### Route Transaction
```bash
curl -X POST http://localhost:8080/volta-router/v1/route \
  -H "Content-Type: application/json" \
  -d '{"amount": 100, "currency": "BRL", "country": "BR"}'
```

### Check Health
```bash
curl http://localhost:8080/volta-router/v1/processors
```

---

🤖 Generated with [Claude Code](https://claude.com/claude-code)
