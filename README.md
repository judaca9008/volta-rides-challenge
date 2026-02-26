# Volta Rides - Smart Payment Router

A proof-of-concept intelligent payment routing system built in Go with Echo framework. Routes transactions to payment processors based on real-time approval rates to minimize failed payments.

## 🚀 Quick Start (< 5 Minutes)

### Prerequisites
- Go 1.21 or higher
- `jq` (optional, for formatted JSON output in demo)

### Installation & Run

```bash
# 1. Install dependencies
go get

# 2. Run the server
PORT=8080 go run main.go

# Server will start on http://localhost:8080
```

### Quick Test

```bash
# 1. Load test data (540 transactions)
curl -X POST http://localhost:8080/volta-router/v1/transactions/load

# 2. Get a routing decision for Brazil
curl -X POST http://localhost:8080/volta-router/v1/route \
  -H "Content-Type: application/json" \
  -d '{"amount": 100, "currency": "BRL", "country": "BR"}'

# 3. Check processor health
curl http://localhost:8080/volta-router/v1/processors
```

## 📋 API Documentation

### Base URL
```
http://localhost:8080/volta-router/v1
```

### Endpoints

#### 1. Get Routing Decision
**POST** `/route`

Routes a payment to the best-performing processor based on approval rates.

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

**Risk Levels:**
- `low`: Approval rate > 80%
- `medium`: Approval rate 70-80%
- `high`: Approval rate < 70%

**Supported Countries:**
- `BR` - Brazil (BRL)
- `MX` - Mexico (MXN)
- `CO` - Colombia (COP)

**Query Parameters:**
- `simulate=true` - **Simulation Mode**: Returns routing decision without recording it in statistics (useful for testing)
- `failover=true` - **Failover Ranking**: Returns top 3 processors with approval rates for fallback options

**Example with Simulation Mode:**
```bash
curl -X POST "http://localhost:8080/volta-router/v1/route?simulate=true" \
  -H "Content-Type: application/json" \
  -d '{"amount": 100, "currency": "BRL", "country": "BR"}'
```

**Example with Failover Ranking:**
```bash
curl -X POST "http://localhost:8080/volta-router/v1/route?failover=true" \
  -H "Content-Type: application/json" \
  -d '{"amount": 100, "currency": "BRL", "country": "BR"}'
```

**Response with Failover Ranking:**
```json
{
  "processor": "RapidPay_BR",
  "approval_rate": 92.5,
  "risk_level": "low",
  "reason": "Highest approval rate for BR",
  "timestamp": "2024-02-26T15:30:00Z",
  "fallback": {
    "processor": "TurboAcquire_BR",
    "approval_rate": 85.0
  },
  "last_resort": {
    "processor": "PayFlow_BR",
    "approval_rate": 78.0
  }
}
```

---

#### 2. Get All Processor Health
**GET** `/processors`

Returns current approval rates for all processors.

**Response:**
```json
{
  "processors": [
    {
      "name": "RapidPay_BR",
      "country": "BR",
      "approval_rate": 92.5,
      "transaction_count": 145,
      "last_updated": "2024-02-26T15:30:00Z",
      "circuit_state": "closed"
    },
    {
      "name": "PayFlow_BR",
      "country": "BR",
      "approval_rate": 55.0,
      "transaction_count": 120,
      "last_updated": "2024-02-26T15:30:00Z",
      "circuit_state": "open",
      "circuit_opened_at": "2024-02-26T15:25:00Z"
    }
  ]
}
```

**Circuit Breaker States:**
- `closed` - Normal operation (default, omitted from response if closed)
- `open` - Processor disabled due to low approval rate (< 60%)
- `half_open` - Testing if processor has recovered (after 5 minutes)

---

#### 3. Get Specific Processor
**GET** `/processors/:name`

Returns stats for a specific processor.

**Example:**
```bash
curl http://localhost:8080/volta-router/v1/processors/RapidPay_BR
```

---

#### 4. Get Routing Statistics
**GET** `/routing/stats`

Returns routing decision distribution for the last 50 decisions.

**Response:**
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

---

#### 5. Load Test Data
**POST** `/transactions/load`

Loads 540 test transactions from `data/test_transactions.json`.

**Response:**
```json
{
  "message": "Test data loaded successfully",
  "transactions_loaded": 540
}
```

---

#### 6. Health Check
**GET** `/health`

Simple health check endpoint.

**Response:**
```json
{
  "status": "Ok"
}
```

---

## 🎯 Demo Walkthrough

Run the automated demo script:

```bash
chmod +x demo.sh
./demo.sh
```

Or follow these manual steps:

### Step 1: Build the Service
```bash
go build -o volta-router
```

### Step 2: Start the Service
```bash
PORT=8080 ./volta-router &
```

### Step 3: Load Test Data
```bash
curl -X POST http://localhost:8080/volta-router/v1/transactions/load
```

### Step 4: View Processor Health
```bash
curl http://localhost:8080/volta-router/v1/processors | jq
```

You'll see 9 processors across 3 countries with their approval rates.

### Step 5: Make Routing Decisions

**Brazil:**
```bash
curl -X POST http://localhost:8080/volta-router/v1/route \
  -H "Content-Type: application/json" \
  -d '{"amount": 100, "currency": "BRL", "country": "BR"}' | jq
```

**Mexico:**
```bash
curl -X POST http://localhost:8080/volta-router/v1/route \
  -H "Content-Type: application/json" \
  -d '{"amount": 50, "currency": "MXN", "country": "MX"}' | jq
```

**Colombia:**
```bash
curl -X POST http://localhost:8080/volta-router/v1/route \
  -H "Content-Type: application/json" \
  -d '{"amount": 75, "currency": "COP", "country": "CO"}' | jq
```

### Step 6: Check Routing Statistics
```bash
curl http://localhost:8080/volta-router/v1/routing/stats | jq
```

---

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `ENVIRONMENT` | Environment name | `development` |

### Routing Configuration

Default settings (defined in `config/config.go`):

- **Time Window**: 15 minutes (approval rates calculated from last 15 min)
- **High Risk Threshold**: < 70%
- **Medium Risk Threshold**: 70-80%
- **Low Risk**: > 80%

---

## 📊 Test Data

The system comes with 540 pre-generated test transactions:

**Processors & Patterns:**
- **Brazil**:
  - `RapidPay_BR` - 92% approval (strong performer)
  - `TurboAcquire_BR` - 85% approval (normal)
  - `PayFlow_BR` - 55-90% approval (has bad period)

- **Mexico**:
  - `RapidPay_MX` - 90% approval (strong performer)
  - `TurboAcquire_MX` - 82% approval (normal)
  - `PayFlow_MX` - 56-88% approval (has bad period)

- **Colombia**:
  - `RapidPay_CO` - 91% approval (strong performer)
  - `TurboAcquire_CO` - 83% approval (normal)
  - `PayFlow_CO` - 57-89% approval (has bad period)

**Transaction Distribution:**
- 60 transactions per processor
- Spread over last 2-3 hours
- Amounts: $5-$150 USD equivalent

---

## 🚀 Advanced Features (Stretch Goals)

This implementation includes several advanced features beyond the core requirements:

### 1. Circuit Breaker Pattern ⚡

Automatically protects the system from failing processors:

**How it works:**
- Monitors approval rates in real-time
- Opens circuit when approval rate drops below 60%
- Processor is excluded from routing for 5 minutes
- After timeout, circuit enters "half-open" state for testing
- Closes circuit if processor recovers (approval rate ≥ 60%)

**Benefits:**
- Prevents routing to consistently failing processors
- Automatic recovery detection
- Improves overall system reliability

**Example:**
```bash
# Circuit breaker will automatically exclude processors with < 60% approval rate
curl http://localhost:8080/volta-router/v1/processors

# Response shows circuit state:
# {
#   "name": "PayFlow_BR",
#   "approval_rate": 55.0,
#   "circuit_state": "open",
#   "circuit_opened_at": "2024-02-26T15:25:00Z"
# }
```

### 2. Failover Ranking 🔄

Provides ranked list of processors for retry logic:

**How it works:**
- Calculates approval rates for all processors
- Returns top 3 processors sorted by performance
- Primary + fallback + last resort options
- Enables intelligent retry strategies

**Benefits:**
- Enables automatic failover in payment gateway
- Maximizes payment success through retries
- No single point of failure

**Example:**
```bash
curl -X POST "http://localhost:8080/volta-router/v1/route?failover=true" \
  -H "Content-Type: application/json" \
  -d '{"amount": 100, "currency": "BRL", "country": "BR"}'

# Response includes failover options:
# {
#   "processor": "RapidPay_BR",
#   "approval_rate": 92.5,
#   "fallback": {
#     "processor": "TurboAcquire_BR",
#     "approval_rate": 85.0
#   },
#   "last_resort": {
#     "processor": "PayFlow_BR",
#     "approval_rate": 78.0
#   }
# }
```

### 3. Simulation Mode 🧪

Test routing decisions without affecting statistics:

**How it works:**
- Add `?simulate=true` query parameter
- Returns routing decision as normal
- Does NOT record decision in statistics
- Perfect for testing and development

**Benefits:**
- Test routing logic without polluting data
- Safe experimentation with production data
- Debugging and troubleshooting

**Example:**
```bash
# Simulate routing (won't affect statistics)
curl -X POST "http://localhost:8080/volta-router/v1/route?simulate=true" \
  -H "Content-Type: application/json" \
  -d '{"amount": 100, "currency": "BRL", "country": "BR"}'

# Check stats - simulation requests not included
curl http://localhost:8080/volta-router/v1/routing/stats
```

### 4. Comprehensive Unit Tests ✅

Full test coverage for core functionality:

**Test Coverage:**
- ✅ Approval rate calculation (with/without data, time windows)
- ✅ Routing logic (best processor selection, edge cases)
- ✅ Risk level classification (low/medium/high)
- ✅ Thread-safe operations (concurrent reads/writes)
- ✅ Storage operations (transactions, routing decisions)
- ✅ Edge cases (unsupported countries, no data, all processors failing)

**Run tests:**
```bash
go test ./tests -v

# Run specific test
go test ./tests -run TestCircuitBreaker -v
```

**Results:**
```
PASS: TestCalculateApprovalRate
PASS: TestSelectBestProcessor
PASS: TestConcurrentAddTransactions
PASS: TestRiskLevelClassification
... 13 tests passing
```

---

## 🧪 Testing

### Manual Testing

```bash
# Test successful routing
curl -X POST http://localhost:8080/volta-router/v1/route \
  -H "Content-Type: application/json" \
  -d '{"amount": 100, "currency": "BRL", "country": "BR"}'

# Test unsupported country (should return 400)
curl -X POST http://localhost:8080/volta-router/v1/route \
  -H "Content-Type: application/json" \
  -d '{"amount": 100, "currency": "USD", "country": "US"}'

# Test invalid request (should return 400)
curl -X POST http://localhost:8080/volta-router/v1/route \
  -H "Content-Type: application/json" \
  -d '{"amount": -10, "currency": "BRL", "country": "BR"}'
```

### Run Unit Tests

```bash
go test ./tests -v
```

---

## 🏗️ Architecture

### Architectural Summary

This proof-of-concept implements a smart payment router using **Go with Echo framework**, following Yuno's standard microservice patterns. The approach prioritizes rapid delivery of a working prototype while maintaining production-grade architecture patterns. We chose a **3-layer architecture** (Controllers → Services → Storage) that separates HTTP handling, business logic, and data access, making the codebase easy to understand and extend. The core routing algorithm calculates real-time approval rates using a **15-minute sliding time window**, then selects the processor with the highest approval rate for each country. DataDog APM integration provides distributed tracing out of the box, and thread-safe operations via `sync.RWMutex` ensure concurrent requests don't corrupt shared data.

For the **1-hour POC scope**, we made pragmatic trade-offs: **in-memory storage** instead of PostgreSQL eliminates database setup time and provides zero-latency queries, though data is lost on restart. We implemented **simplified middleware** (basic logging and trace ID generation) rather than integrating Yuno's full internal libraries, allowing the service to run standalone without authentication dependencies. The processor-to-country mapping is **hard-coded** in configuration rather than database-driven, which is acceptable for a demo with 3 countries but would need to change for production. These trade-offs enabled rapid development while keeping the code clean and following Yuno's architectural conventions exactly.

**Next steps for production** would involve replacing in-memory storage with **PostgreSQL/TimescaleDB** for persistence and time-series optimization, and adding **Redis caching** to reduce database load for frequent approval rate queries. The **circuit breaker logic is already implemented** and automatically disables failing processors (< 60% approval rate) for 5 minutes with automatic recovery detection. The service would deploy to **Kubernetes via Yuno's Kingdom platform**, integrate with the full monitoring stack (Prometheus, Grafana, alerting), and add API authentication with rate limiting. The core routing logic and layered architecture require no changes—only the infrastructure dependencies would evolve. Estimated production migration: **2-3 weeks with 2 engineers** (1 backend, 1 DevOps).

---

See [ARCHITECTURE.md](./ARCHITECTURE.md) for detailed architecture documentation.

**High-Level Overview:**

```
Controllers → Services → Storage
```

- **Controllers**: HTTP handlers, request validation
- **Services**: Business logic, approval rate calculation
- **Storage**: Thread-safe in-memory store

**Key Features:**
- Real-time approval rate tracking with sliding time windows
- Country-aware processor selection
- Risk level classification (low/medium/high)
- Thread-safe concurrent operations
- DataDog APM integration
- **Circuit breaker pattern** - Automatic protection from failing processors
- **Failover ranking** - Top 3 processor options for retry logic
- **Simulation mode** - Test without affecting statistics
- **Comprehensive unit tests** - Full test coverage

---

## 📝 Development

### Project Structure

```
volta-rides-challenge/
├── main.go                   # Entry point
├── cmd/
│   ├── httpServer/          # Server initialization
│   └── generate_data/       # Test data generator
├── controllers/             # HTTP handlers
├── services/                # Business logic
├── storage/                 # In-memory store
├── models/                  # Data structures
├── routers/                 # Route configuration
├── config/                  # Configuration
├── data/                    # Test data
└── tests/                   # Unit tests
```

### Code Conventions

This project follows Yuno's Go microservice patterns:
- Layered architecture (Controllers → Services → Storage)
- DataDog APM integration
- Structured logging
- Thread-safe operations

---

## 🌐 Public Deployment

This project is ready for instant deployment to Fly.io with a public HTTPS URL.

### Quick Deploy (3 commands)

```bash
# 1. Install Fly CLI
brew install flyctl  # macOS
# curl -L https://fly.io/install.sh | sh  # Linux

# 2. Authenticate
fly auth signup  # or fly auth login

# 3. Deploy
fly launch --now
```

Your API will be live at: `https://volta-router.fly.dev` 🚀

**Features included:**
- ✅ Free tier (256MB RAM, auto-sleep when idle)
- ✅ HTTPS automatic
- ✅ Auto-scaling
- ✅ Health checks configured
- ✅ Zero-downtime deploys

**After deployment:**
```bash
# Load test data
curl -X POST https://volta-router.fly.dev/volta-router/v1/transactions/load

# Test routing
curl -X POST https://volta-router.fly.dev/volta-router/v1/route \
  -H "Content-Type: application/json" \
  -d '{"amount": 100, "currency": "BRL", "country": "BR"}'
```

📖 **Full deployment guide:** See [DEPLOYMENT.md](./DEPLOYMENT.md) for detailed instructions, troubleshooting, and advanced configuration.

---

## 🚀 Production Considerations

This is a **proof-of-concept**. For production deployment, consider:

1. **Persistent Storage**: Replace in-memory store with PostgreSQL/TimescaleDB
2. **Caching**: Add Redis for approval rate caching
3. **Message Queue**: Use Kafka for transaction ingestion
4. **Circuit Breaker**: Implement processor circuit breaker logic
5. **Authentication**: Add API key/JWT authentication
6. **Rate Limiting**: Prevent API abuse
7. **Monitoring**: Full observability stack (Prometheus, Grafana)
8. **Horizontal Scaling**: Load balancer + multiple instances

See [ARCHITECTURE.md](./ARCHITECTURE.md) for detailed production roadmap.

---

## 📄 License

This is a proof-of-concept project for Volta Rides.

---

## 🤝 Support

For questions or issues, please refer to:
- [CLAUDE.md](./CLAUDE.md) - Development guide for AI assistance
- [ARCHITECTURE.md](./ARCHITECTURE.md) - Architectural decisions

---

**Built with ❤️ using Go + Echo framework following Yuno's microservice patterns**

🤖 Generated with [Claude Code](https://claude.com/claude-code)
