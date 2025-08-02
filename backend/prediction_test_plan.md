# Prediction System Test Plan

## Overview
This test plan demonstrates how the new hybrid user/global prediction system learns from feedback and adapts over time. We'll use systematic test cases to show the learning behavior.

## Historical Context
Based on your existing data for user1:
- **15min @ 22°C**: Multiple records with 10.3-10.4 heating time, satisfaction 45-46 (slightly cold)
- **20min @ 22°C**: Records with 11.8-14.2 heating time, satisfaction 15-30 (very cold - system learning)

## Test Cases

### Test Case 1: Learning from Similar Conditions (15min @ 22°C)
**Goal**: Show how the system learns from user's specific feedback for familiar conditions

**Steps**:
1. **Initial Prediction**: Call prediction API with `{"userId": "user1", "duration": 15, "temperature": 22}`
2. **Expected Result**: Should predict around 10.5-11.0 minutes (slightly higher than 10.4 due to satisfaction 45-46)
3. **Submit Feedback**: Submit satisfaction 40 (colder than before)
4. **Re-test Prediction**: Same inputs, should now predict higher (~11.2-11.5 minutes)
5. **Submit Better Feedback**: Submit satisfaction 50 (perfect)
6. **Final Prediction**: Should stabilize around the heating time that got satisfaction 50

### Test Case 2: New User Experience
**Goal**: Show how new users get global predictions

**Steps**:
1. **New User Prediction**: Call with `{"userId": "newuser", "duration": 15, "temperature": 22}`
2. **Expected Result**: Should use global data + default algorithm (likely 8-9 minutes)
3. **Submit Feedback**: Submit satisfaction 30 (too cold)
4. **Second Prediction**: Should be higher, blending user data with global
5. **Gradual Learning**: After 3-5 feedback entries, predictions should be more user-specific

### Test Case 3: Edge Case - Very Different Conditions
**Goal**: Test behavior with conditions outside user's history

**Steps**:
1. **Unusual Conditions**: Call with `{"userId": "user1", "duration": 5, "temperature": 10}`
2. **Expected Result**: Should fall back to global data + defaults (no similar user history)
3. **Submit Feedback**: Submit satisfaction rating
4. **Build History**: Repeat 2-3 times to build user-specific data for these conditions

### Test Case 4: Relative Adjustment Testing
**Goal**: Demonstrate the new relative adjustment feature

**Steps**:
1. **Short Heating Time**: Predict for conditions that typically need 5 minutes
2. **Submit Low Satisfaction**: Submit satisfaction 20 (very cold)
3. **Expected Adjustment**: Should increase by up to 25% (1.25 minutes max)
4. **Long Heating Time**: Predict for conditions that typically need 20 minutes  
5. **Submit Low Satisfaction**: Submit satisfaction 20 (very cold)
6. **Expected Adjustment**: Should increase by up to 25% (5 minutes max)

## Detailed Test Scenarios

### Scenario A: Progressive Learning
```bash
# Step 1: Initial state
curl -X POST http://localhost:8080/api/calculate \
  -H "Content-Type: application/json" \
  -d '{"userId": "user1", "duration": 15, "temperature": 22}'

# Expected: ~10.8 minutes (learning from satisfaction 45-46)

# Step 2: Submit cold feedback
curl -X POST http://localhost:8080/api/feedback \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user1",
    "showerDuration": 15,
    "averageTemperature": 22,
    "heatingTime": 10.8,
    "satisfaction": 35
  }'

# Step 3: Check learning
curl -X POST http://localhost:8080/api/calculate \
  -H "Content-Type: application/json" \
  -d '{"userId": "user1", "duration": 15, "temperature": 22}'

# Expected: Higher than 10.8 (maybe 11.5-12.0)
```

### Scenario B: New User Cold Start
```bash
# Step 1: Brand new user
curl -X POST http://localhost:8080/api/calculate \
  -H "Content-Type: application/json" \
  -d '{"userId": "testuser_new", "duration": 15, "temperature": 22}'

# Expected: Default algorithm result (~8-9 minutes)

# Step 2: Submit feedback showing it was too cold
curl -X POST http://localhost:8080/api/feedback \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "testuser_new",
    "showerDuration": 15,
    "averageTemperature": 22,
    "heatingTime": 8.5,
    "satisfaction": 25
  }'

# Step 3: Check adaptation
curl -X POST http://localhost:8080/api/calculate \
  -H "Content-Type: application/json" \
  -d '{"userId": "testuser_new", "duration": 15, "temperature": 22}'

# Expected: Blend of user learning + global data (maybe 9.5-10.5)
```

### Scenario C: Weight Progression Test
**Goal**: Show how user weight increases with more data

```bash
# Test with user having 1, 5, and 10+ similar records
# User weight should be: 0.1, 0.5, and 1.0 respectively

# Check prediction for user with few records
curl -X POST http://localhost:8080/api/calculate \
  -H "Content-Type: application/json" \
  -d '{"userId": "user_with_2_records", "duration": 15, "temperature": 22}'

# vs user with many records (user1 has 8+ records)
curl -X POST http://localhost:8080/api/calculate \
  -H "Content-Type: application/json" \
  -d '{"userId": "user1", "duration": 15, "temperature": 22}'
```

## Expected Learning Patterns

### Pattern 1: Satisfaction-Based Adjustment
- **Satisfaction < 50**: Heating time should increase
- **Satisfaction > 50**: Heating time should decrease  
- **Satisfaction = 50**: Minimal change

### Pattern 2: User Weight Evolution
- **0-2 records**: Weight ≈ 0.0-0.2 (mostly global predictions)
- **3-7 records**: Weight ≈ 0.3-0.7 (hybrid predictions)
- **8+ records**: Weight ≈ 0.8-1.0 (mostly user-specific)

### Pattern 3: Relative Adjustments
- **Low heating times** (5 min): Max adjustment ±1.25 min
- **Medium heating times** (10 min): Max adjustment ±2.5 min  
- **High heating times** (20 min): Max adjustment ±5.0 min

## Validation Criteria

### ✅ Success Indicators:
1. **Learning**: Predictions change meaningfully after feedback
2. **Personalization**: Different users get different predictions for same inputs
3. **Convergence**: Repeated good feedback (satisfaction 50) stabilizes predictions
4. **Cold Start**: New users get reasonable predictions from global data
5. **Relative Scaling**: Adjustments scale proportionally with heating time

### ❌ Failure Indicators:
1. **No Learning**: Predictions don't change after feedback
2. **Extreme Changes**: Predictions jump by >50% from single feedback
3. **Poor Cold Start**: New users get unreasonable predictions (e.g., 1 min or 60 min)
4. **Fixed Adjustments**: All satisfaction feedback causes same absolute change

## Automated Test Script

Create a test script to run these scenarios automatically:

```bash
#!/bin/bash
# prediction_learning_test.sh

echo "=== Testing Prediction Learning System ==="

# Test 1: User1 learning
echo "Test 1: Existing user learning..."
PRED1=$(curl -s -X POST http://localhost:8080/api/calculate -H "Content-Type: application/json" -d '{"userId": "user1", "duration": 15, "temperature": 22}' | jq -r '.heatingTime')
echo "Initial prediction: $PRED1 minutes"

# Submit cold feedback
curl -s -X POST http://localhost:8080/api/feedback -H "Content-Type: application/json" -d "{\"userId\": \"user1\", \"showerDuration\": 15, \"averageTemperature\": 22, \"heatingTime\": $PRED1, \"satisfaction\": 35}"

PRED2=$(curl -s -X POST http://localhost:8080/api/calculate -H "Content-Type: application/json" -d '{"userId": "user1", "duration": 15, "temperature": 22}' | jq -r '.heatingTime')
echo "After cold feedback: $PRED2 minutes"
echo "Learning delta: $(echo "$PRED2 - $PRED1" | bc) minutes"

# Test 2: New user
echo -e "\nTest 2: New user cold start..."
NEWUSER_PRED=$(curl -s -X POST http://localhost:8080/api/calculate -H "Content-Type: application/json" -d '{"userId": "newuser_test", "duration": 15, "temperature": 22}' | jq -r '.heatingTime')
echo "New user prediction: $NEWUSER_PRED minutes"
echo "Difference from experienced user: $(echo "$PRED2 - $NEWUSER_PRED" | bc) minutes"
```

This comprehensive test plan will demonstrate all the key improvements in your prediction system!