#!/bin/bash

# Volta Rides Smart Payment Router - Demo Script
# This script demonstrates the smart routing system

set -e  # Exit on error

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo ""
echo "╔════════════════════════════════════════════════════════════════╗"
echo "║   Volta Rides - Smart Payment Router Demo                     ║"
echo "║   Proof-of-Concept: Intelligent Processor Routing              ║"
echo "╚════════════════════════════════════════════════════════════════╝"
echo ""

# Step 1: Build
echo -e "${BLUE}Step 1: Building the service...${NC}"
go build -o volta-router
echo -e "${GREEN}✓ Build complete${NC}"
echo ""

# Step 2: Start server in background
echo -e "${BLUE}Step 2: Starting the server...${NC}"
PORT=8080 ./volta-router &
SERVER_PID=$!
sleep 3
echo -e "${GREEN}✓ Server running on port 8080 (PID: $SERVER_PID)${NC}"
echo ""

# Function to cleanup on exit
cleanup() {
    echo ""
    echo -e "${YELLOW}Cleaning up...${NC}"
    kill $SERVER_PID 2>/dev/null || true
    echo -e "${GREEN}✓ Server stopped${NC}"
}
trap cleanup EXIT

# Step 3: Load test data
echo -e "${BLUE}Step 3: Loading test data (540 transactions)...${NC}"
LOAD_RESPONSE=$(curl -s -X POST http://localhost:8080/volta-router/v1/transactions/load)

if command -v jq &> /dev/null; then
    echo "$LOAD_RESPONSE" | jq '.'
else
    echo "$LOAD_RESPONSE"
fi
echo ""

# Step 4: View processor health
echo -e "${BLUE}Step 4: Viewing processor health across all countries...${NC}"
echo "Showing approval rates for all 9 processors:"
echo ""

if command -v jq &> /dev/null; then
    curl -s http://localhost:8080/volta-router/v1/processors | jq -r '.processors[] | "\(.name) (\(.country)): \(.approval_rate)% - \(.transaction_count) transactions"'
else
    curl -s http://localhost:8080/volta-router/v1/processors
fi
echo ""

# Step 5: Make routing decisions for each country
echo -e "${BLUE}Step 5: Making routing decisions for each country...${NC}"
echo ""

echo -e "${YELLOW}→ Routing decision for Brazil (BRL):${NC}"
BRAZIL_RESPONSE=$(curl -s -X POST http://localhost:8080/volta-router/v1/route \
  -H "Content-Type: application/json" \
  -d '{"amount": 100, "currency": "BRL", "country": "BR"}')

if command -v jq &> /dev/null; then
    echo "$BRAZIL_RESPONSE" | jq '.'
else
    echo "$BRAZIL_RESPONSE"
fi
echo ""

echo -e "${YELLOW}→ Routing decision for Mexico (MXN):${NC}"
MEXICO_RESPONSE=$(curl -s -X POST http://localhost:8080/volta-router/v1/route \
  -H "Content-Type: application/json" \
  -d '{"amount": 50, "currency": "MXN", "country": "MX"}')

if command -v jq &> /dev/null; then
    echo "$MEXICO_RESPONSE" | jq '.'
else
    echo "$MEXICO_RESPONSE"
fi
echo ""

echo -e "${YELLOW}→ Routing decision for Colombia (COP):${NC}"
COLOMBIA_RESPONSE=$(curl -s -X POST http://localhost:8080/volta-router/v1/route \
  -H "Content-Type: application/json" \
  -d '{"amount": 75, "currency": "COP", "country": "CO"}')

if command -v jq &> /dev/null; then
    echo "$COLOMBIA_RESPONSE" | jq '.'
else
    echo "$COLOMBIA_RESPONSE"
fi
echo ""

# Step 6: Check routing statistics
echo -e "${BLUE}Step 6: Checking routing statistics...${NC}"
STATS_RESPONSE=$(curl -s http://localhost:8080/volta-router/v1/routing/stats)

if command -v jq &> /dev/null; then
    echo "$STATS_RESPONSE" | jq '.'
else
    echo "$STATS_RESPONSE"
fi
echo ""

# Step 7: Test edge case (unsupported country)
echo -e "${BLUE}Step 7: Testing edge case (unsupported country)...${NC}"
echo -e "${YELLOW}→ Attempting to route to United States (should fail):${NC}"
ERROR_RESPONSE=$(curl -s -X POST http://localhost:8080/volta-router/v1/route \
  -H "Content-Type: application/json" \
  -d '{"amount": 100, "currency": "USD", "country": "US"}')

if command -v jq &> /dev/null; then
    echo "$ERROR_RESPONSE" | jq '.'
else
    echo "$ERROR_RESPONSE"
fi
echo ""

# Step 8: Check specific processor
echo -e "${BLUE}Step 8: Checking specific processor (RapidPay_BR)...${NC}"
PROCESSOR_RESPONSE=$(curl -s http://localhost:8080/volta-router/v1/processors/RapidPay_BR)

if command -v jq &> /dev/null; then
    echo "$PROCESSOR_RESPONSE" | jq '.'
else
    echo "$PROCESSOR_RESPONSE"
fi
echo ""

# Demo complete
echo "╔════════════════════════════════════════════════════════════════╗"
echo "║                      Demo Complete! ✓                          ║"
echo "╚════════════════════════════════════════════════════════════════╝"
echo ""
echo -e "${GREEN}Key Observations:${NC}"
echo "1. RapidPay processors have the highest approval rates (90-92%)"
echo "2. PayFlow processors show degraded performance (55-57%) during bad periods"
echo "3. Router automatically selects best-performing processor per country"
echo "4. Risk levels are correctly classified based on approval rates"
echo "5. Unsupported countries return appropriate error messages"
echo ""
echo -e "${BLUE}For more information:${NC}"
echo "  - API Documentation: README.md"
echo "  - Architecture Details: ARCHITECTURE.md"
echo "  - Development Guide: CLAUDE.md"
echo ""
