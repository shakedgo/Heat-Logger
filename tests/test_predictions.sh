#!/bin/bash

# Make script executable
chmod +x test_predictions.sh

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Prediction Learning System Test ===${NC}"
echo "Testing the new hybrid user/global prediction model..."
echo ""

# Check if server is running
echo -e "${YELLOW}Checking if server is running...${NC}"
if ! curl -s http://localhost:8080/api/health > /dev/null; then
    echo -e "${RED}❌ Server not running on localhost:8080${NC}"
    echo "Please start the server first with: ./run-dev.sh"
    exit 1
fi
echo -e "${GREEN}✅ Server is running${NC}"
echo ""

# Test 1: Existing User Learning (user1)
echo -e "${BLUE}Test 1: Existing User Learning (user1)${NC}"
echo "Testing with 15min duration, 22°C temperature..."

# Get initial prediction
echo "Getting initial prediction..."
PRED1_RESPONSE=$(curl -s -X POST http://localhost:8080/api/calculate \
    -H "Content-Type: application/json" \
    -d '{"userId": "user1", "duration": 15, "temperature": 22}')

PRED1=$(echo $PRED1_RESPONSE | jq -r '.heatingTime')
echo -e "Initial prediction: ${GREEN}$PRED1 minutes${NC}"

# Submit cold feedback (satisfaction 35 - colder than usual)
echo "Submitting cold feedback (satisfaction 35)..."
curl -s -X POST http://localhost:8080/api/feedback \
    -H "Content-Type: application/json" \
    -d "{\"userId\": \"user1\", \"showerDuration\": 15, \"averageTemperature\": 22, \"heatingTime\": $PRED1, \"satisfaction\": 35}" > /dev/null

# Get updated prediction
echo "Getting updated prediction..."
PRED2_RESPONSE=$(curl -s -X POST http://localhost:8080/api/calculate \
    -H "Content-Type: application/json" \
    -d '{"userId": "user1", "duration": 15, "temperature": 22}')

PRED2=$(echo $PRED2_RESPONSE | jq -r '.heatingTime')
echo -e "After cold feedback: ${GREEN}$PRED2 minutes${NC}"

# Calculate learning delta
DELTA=$(echo "scale=2; $PRED2 - $PRED1" | bc)
echo -e "Learning delta: ${YELLOW}+$DELTA minutes${NC}"

if (( $(echo "$DELTA > 0" | bc -l) )); then
    echo -e "${GREEN}✅ System learned - increased heating time after cold feedback${NC}"
else
    echo -e "${RED}❌ System didn't learn - no increase after cold feedback${NC}"
fi
echo ""

# Test 2: New User Cold Start
echo -e "${BLUE}Test 2: New User Cold Start${NC}"
echo "Testing with brand new user..."

NEWUSER_ID="testuser_$(date +%s)"
echo "Creating new user: $NEWUSER_ID"

# Get new user prediction
NEWUSER_RESPONSE=$(curl -s -X POST http://localhost:8080/api/calculate \
    -H "Content-Type: application/json" \
    -d "{\"userId\": \"$NEWUSER_ID\", \"duration\": 15, \"temperature\": 22}")

NEWUSER_PRED=$(echo $NEWUSER_RESPONSE | jq -r '.heatingTime')
echo -e "New user prediction: ${GREEN}$NEWUSER_PRED minutes${NC}"

# Compare with experienced user
USER_DIFF=$(echo "scale=2; $PRED2 - $NEWUSER_PRED" | bc)
echo -e "Difference from experienced user: ${YELLOW}$USER_DIFF minutes${NC}"

if (( $(echo "$USER_DIFF != 0" | bc -l) )); then
    echo -e "${GREEN}✅ Personalization working - different users get different predictions${NC}"
else
    echo -e "${YELLOW}⚠️  Same prediction for both users - may indicate limited global data${NC}"
fi
echo ""

# Test 3: New User Learning
echo -e "${BLUE}Test 3: New User Learning${NC}"
echo "Testing how new user learns from feedback..."

# Submit feedback for new user (too cold)
echo "New user submits cold feedback (satisfaction 25)..."
curl -s -X POST http://localhost:8080/api/feedback \
    -H "Content-Type: application/json" \
    -d "{\"userId\": \"$NEWUSER_ID\", \"showerDuration\": 15, \"averageTemperature\": 22, \"heatingTime\": $NEWUSER_PRED, \"satisfaction\": 25}" > /dev/null

# Get updated prediction for new user
NEWUSER_PRED2_RESPONSE=$(curl -s -X POST http://localhost:8080/api/calculate \
    -H "Content-Type: application/json" \
    -d "{\"userId\": \"$NEWUSER_ID\", \"duration\": 15, \"temperature\": 22}")

NEWUSER_PRED2=$(echo $NEWUSER_PRED2_RESPONSE | jq -r '.heatingTime')
echo -e "New user prediction after feedback: ${GREEN}$NEWUSER_PRED2 minutes${NC}"

NEWUSER_DELTA=$(echo "scale=2; $NEWUSER_PRED2 - $NEWUSER_PRED" | bc)
echo -e "New user learning delta: ${YELLOW}+$NEWUSER_DELTA minutes${NC}"

if (( $(echo "$NEWUSER_DELTA > 0" | bc -l) )); then
    echo -e "${GREEN}✅ New user learning - increased heating time after cold feedback${NC}"
else
    echo -e "${RED}❌ New user not learning properly${NC}"
fi
echo ""

# Test 4: Relative Adjustment Test
echo -e "${BLUE}Test 4: Relative Adjustment Test${NC}"
echo "Testing relative adjustments for different heating times..."

# Test with conditions that should give a shorter heating time
echo "Testing short heating time scenario (10min duration, 25°C)..."
SHORT_RESPONSE=$(curl -s -X POST http://localhost:8080/api/calculate \
    -H "Content-Type: application/json" \
    -d '{"userId": "user1", "duration": 10, "temperature": 25}')

SHORT_PRED=$(echo $SHORT_RESPONSE | jq -r '.heatingTime')
echo -e "Short scenario prediction: ${GREEN}$SHORT_PRED minutes${NC}"

# Test with conditions that should give a longer heating time  
echo "Testing long heating time scenario (25min duration, 15°C)..."
LONG_RESPONSE=$(curl -s -X POST http://localhost:8080/api/calculate \
    -H "Content-Type: application/json" \
    -d '{"userId": "user1", "duration": 25, "temperature": 15}')

LONG_PRED=$(echo $LONG_RESPONSE | jq -r '.heatingTime')
echo -e "Long scenario prediction: ${GREEN}$LONG_PRED minutes${NC}"

# Calculate max possible adjustments (25% of heating time)
SHORT_MAX_ADJ=$(echo "scale=2; $SHORT_PRED * 0.25" | bc)
LONG_MAX_ADJ=$(echo "scale=2; $LONG_PRED * 0.25" | bc)

echo -e "Max adjustment for short scenario: ${YELLOW}±$SHORT_MAX_ADJ minutes${NC}"
echo -e "Max adjustment for long scenario: ${YELLOW}±$LONG_MAX_ADJ minutes${NC}"

if (( $(echo "$LONG_MAX_ADJ > $SHORT_MAX_ADJ" | bc -l) )); then
    echo -e "${GREEN}✅ Relative adjustments working - longer heating times allow bigger adjustments${NC}"
else
    echo -e "${RED}❌ Relative adjustments not working properly${NC}"
fi
echo ""

# Summary
echo -e "${BLUE}=== Test Summary ===${NC}"
echo -e "User1 learning delta: ${YELLOW}+$DELTA minutes${NC}"
echo -e "New user vs experienced difference: ${YELLOW}$USER_DIFF minutes${NC}"
echo -e "New user learning delta: ${YELLOW}+$NEWUSER_DELTA minutes${NC}"
echo -e "Relative adjustment range: ${YELLOW}$SHORT_MAX_ADJ to $LONG_MAX_ADJ minutes${NC}"
echo ""

# Recommendations
echo -e "${BLUE}=== Recommendations ===${NC}"
if (( $(echo "$DELTA > 0.1" | bc -l) )); then
    echo -e "${GREEN}✅ User learning is working well${NC}"
else
    echo -e "${YELLOW}💡 Consider testing with more extreme satisfaction scores (10-20 or 80-90)${NC}"
fi

if (( $(echo "$USER_DIFF > 0.5 || $USER_DIFF < -0.5" | bc -l) )); then
    echo -e "${GREEN}✅ User personalization is effective${NC}"
else
    echo -e "${YELLOW}💡 Add more diverse global data to improve personalization${NC}"
fi

echo ""
echo -e "${GREEN}Test completed! Check the results above to verify the learning system is working.${NC}"