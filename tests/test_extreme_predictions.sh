#!/bin/bash

# Make script executable
chmod +x test_extreme_predictions.sh

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Extreme Prediction Learning Test ===${NC}"
echo "Testing with extreme satisfaction scores and diverse conditions..."
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

# Test 1: Extreme Cold Feedback (Satisfaction 10)
echo -e "${BLUE}Test 1: Extreme Cold Feedback${NC}"
echo "Testing with very low satisfaction (10) to see large learning delta..."

# Get initial prediction for user1
PRED1_RESPONSE=$(curl -s -X POST http://localhost:8080/api/calculate \
    -H "Content-Type: application/json" \
    -d '{"userId": "user1", "duration": 15, "temperature": 22}')

PRED1=$(echo $PRED1_RESPONSE | jq -r '.heatingTime')
echo -e "Initial prediction: ${GREEN}$PRED1 minutes${NC}"

# Submit extreme cold feedback (satisfaction 10)
echo "Submitting EXTREME cold feedback (satisfaction 10)..."
curl -s -X POST http://localhost:8080/api/feedback \
    -H "Content-Type: application/json" \
    -d "{\"userId\": \"user1\", \"showerDuration\": 15, \"averageTemperature\": 22, \"heatingTime\": $PRED1, \"satisfaction\": 10}" > /dev/null

# Get updated prediction
PRED2_RESPONSE=$(curl -s -X POST http://localhost:8080/api/calculate \
    -H "Content-Type: application/json" \
    -d '{"userId": "user1", "duration": 15, "temperature": 22}')

PRED2=$(echo $PRED2_RESPONSE | jq -r '.heatingTime')
echo -e "After extreme cold feedback: ${GREEN}$PRED2 minutes${NC}"

DELTA1=$(echo "scale=2; $PRED2 - $PRED1" | bc)
echo -e "Extreme cold learning delta: ${YELLOW}+$DELTA1 minutes${NC}"
echo ""

# Test 2: Extreme Hot Feedback (Satisfaction 90)
echo -e "${BLUE}Test 2: Extreme Hot Feedback${NC}"
echo "Testing with very high satisfaction (90) to see negative learning delta..."

# Submit extreme hot feedback (satisfaction 90)
echo "Submitting EXTREME hot feedback (satisfaction 90)..."
curl -s -X POST http://localhost:8080/api/feedback \
    -H "Content-Type: application/json" \
    -d "{\"userId\": \"user1\", \"showerDuration\": 15, \"averageTemperature\": 22, \"heatingTime\": $PRED2, \"satisfaction\": 90}" > /dev/null

# Get updated prediction
PRED3_RESPONSE=$(curl -s -X POST http://localhost:8080/api/calculate \
    -H "Content-Type: application/json" \
    -d '{"userId": "user1", "duration": 15, "temperature": 22}')

PRED3=$(echo $PRED3_RESPONSE | jq -r '.heatingTime')
echo -e "After extreme hot feedback: ${GREEN}$PRED3 minutes${NC}"

DELTA2=$(echo "scale=2; $PRED3 - $PRED2" | bc)
echo -e "Extreme hot learning delta: ${YELLOW}$DELTA2 minutes${NC}"
echo ""

# Test 3: Test Different User with Diverse Conditions
echo -e "${BLUE}Test 3: Diverse Global Data Test${NC}"
echo "Adding diverse global data and testing personalization..."

# Create diverse test users with different patterns
USERS=("power_user" "eco_user" "cold_climate_user" "hot_climate_user")
CONDITIONS=(
    '{"duration": 5, "temperature": 30, "heatingTime": 3.0, "satisfaction": 50}'   # Short hot shower
    '{"duration": 30, "temperature": 10, "heatingTime": 20.0, "satisfaction": 50}' # Long cold shower  
    '{"duration": 20, "temperature": 5, "heatingTime": 25.0, "satisfaction": 50}'  # Cold climate
    '{"duration": 8, "temperature": 35, "heatingTime": 2.0, "satisfaction": 50}'   # Hot climate
)

echo "Seeding diverse global data..."
for i in "${!USERS[@]}"; do
    user="${USERS[$i]}"
    condition="${CONDITIONS[$i]}"
    
    # Extract values from condition
    duration=$(echo $condition | jq -r '.duration')
    temperature=$(echo $condition | jq -r '.temperature')
    heatingTime=$(echo $condition | jq -r '.heatingTime')
    satisfaction=$(echo $condition | jq -r '.satisfaction')
    
    echo "Adding data for $user: ${duration}min, ${temperature}°C, ${heatingTime}min heating"
    
    # Add feedback for this user
    curl -s -X POST http://localhost:8080/api/feedback \
        -H "Content-Type: application/json" \
        -d "{\"userId\": \"$user\", \"showerDuration\": $duration, \"averageTemperature\": $temperature, \"heatingTime\": $heatingTime, \"satisfaction\": $satisfaction}" > /dev/null
done
echo ""

# Test 4: Compare Predictions with Diverse Data
echo -e "${BLUE}Test 4: Personalization with Diverse Data${NC}"
echo "Testing how different users get different predictions now..."

# Test same conditions for different user types
TEST_CONDITIONS=(
    '{"duration": 15, "temperature": 22, "desc": "Standard conditions"}'
    '{"duration": 10, "temperature": 30, "desc": "Short hot shower"}'
    '{"duration": 25, "temperature": 10, "desc": "Long cold shower"}'
)

for condition in "${TEST_CONDITIONS[@]}"; do
    duration=$(echo $condition | jq -r '.duration')
    temperature=$(echo $condition | jq -r '.temperature')
    desc=$(echo $condition | jq -r '.desc')
    
    echo -e "\n${YELLOW}Testing: $desc (${duration}min, ${temperature}°C)${NC}"
    
    # Get predictions for different users
    for user in "user1" "power_user" "eco_user" "new_test_user"; do
        response=$(curl -s -X POST http://localhost:8080/api/calculate \
            -H "Content-Type: application/json" \
            -d "{\"userId\": \"$user\", \"duration\": $duration, \"temperature\": $temperature}")
        
        prediction=$(echo $response | jq -r '.heatingTime')
        echo "  $user: ${prediction} minutes"
    done
done
echo ""

# Test 5: Progressive Learning with Extreme Scores
echo -e "${BLUE}Test 5: Progressive Learning with Extreme Scores${NC}"
echo "Testing how system learns progressively with extreme feedback..."

NEW_USER="learning_test_user"
echo "Testing progressive learning for: $NEW_USER"

# Initial prediction
INITIAL_RESPONSE=$(curl -s -X POST http://localhost:8080/api/calculate \
    -H "Content-Type: application/json" \
    -d "{\"userId\": \"$NEW_USER\", \"duration\": 12, \"temperature\": 20}")

INITIAL_PRED=$(echo $INITIAL_RESPONSE | jq -r '.heatingTime')
echo -e "Initial prediction: ${GREEN}$INITIAL_PRED minutes${NC}"

# Submit series of extreme cold feedback
EXTREME_SCORES=(15 20 10 25)
CURRENT_PRED=$INITIAL_PRED

for i in "${!EXTREME_SCORES[@]}"; do
    score="${EXTREME_SCORES[$i]}"
    echo "Round $((i+1)): Submitting satisfaction $score..."
    
    # Submit feedback
    curl -s -X POST http://localhost:8080/api/feedback \
        -H "Content-Type: application/json" \
        -d "{\"userId\": \"$NEW_USER\", \"showerDuration\": 12, \"averageTemperature\": 20, \"heatingTime\": $CURRENT_PRED, \"satisfaction\": $score}" > /dev/null
    
    # Get new prediction
    NEW_RESPONSE=$(curl -s -X POST http://localhost:8080/api/calculate \
        -H "Content-Type: application/json" \
        -d "{\"userId\": \"$NEW_USER\", \"duration\": 12, \"temperature\": 20}")
    
    NEW_PRED=$(echo $NEW_RESPONSE | jq -r '.heatingTime')
    DELTA=$(echo "scale=2; $NEW_PRED - $CURRENT_PRED" | bc)
    
    echo -e "  After feedback: ${GREEN}$NEW_PRED minutes${NC} (Δ: ${YELLOW}+$DELTA${NC})"
    CURRENT_PRED=$NEW_PRED
done

TOTAL_LEARNING=$(echo "scale=2; $CURRENT_PRED - $INITIAL_PRED" | bc)
echo -e "Total learning after extreme feedback: ${YELLOW}+$TOTAL_LEARNING minutes${NC}"
echo ""

# Summary
echo -e "${BLUE}=== Enhanced Test Summary ===${NC}"
echo -e "Extreme cold learning (satisfaction 10): ${YELLOW}+$DELTA1 minutes${NC}"
echo -e "Extreme hot learning (satisfaction 90): ${YELLOW}$DELTA2 minutes${NC}"
echo -e "Progressive learning total: ${YELLOW}+$TOTAL_LEARNING minutes${NC}"
echo ""

# Recommendations
echo -e "${BLUE}=== Results Analysis ===${NC}"
if (( $(echo "$DELTA1 > 0.5" | bc -l) )); then
    echo -e "${GREEN}✅ Extreme cold feedback produces significant learning${NC}"
else
    echo -e "${YELLOW}⚠️  Extreme cold feedback learning could be more aggressive${NC}"
fi

if (( $(echo "$DELTA2 < -0.2" | bc -l) )); then
    echo -e "${GREEN}✅ Extreme hot feedback reduces heating time effectively${NC}"
else
    echo -e "${YELLOW}⚠️  Extreme hot feedback could be more responsive${NC}"
fi

if (( $(echo "$TOTAL_LEARNING > 1.0" | bc -l) )); then
    echo -e "${GREEN}✅ Progressive learning with extreme scores is working well${NC}"
else
    echo -e "${YELLOW}⚠️  Progressive learning could accumulate more aggressively${NC}"
fi

echo ""
echo -e "${GREEN}Enhanced testing completed! The system now has diverse global data and has been tested with extreme satisfaction scores.${NC}"