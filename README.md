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
      "last_updated": "2024-02-26T15:30:00Z"
    }
  ]
}
```

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
- Risk level classification
- Thread-safe concurrent operations
- DataDog APM integration

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
