#!/bin/bash

# Make script executable
chmod +x seed_diverse_data.sh

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Seeding Diverse Global Data ===${NC}"
echo "Adding varied user patterns to improve personalization..."
echo ""

# Check if server is running
if ! curl -s http://localhost:8080/api/health > /dev/null; then
    echo -e "${RED}❌ Server not running on localhost:8080${NC}"
    echo "Please start the server first"
    exit 1
fi
echo -e "${GREEN}✅ Server is running${NC}"
echo ""

# Define diverse user profiles with different heating preferences
declare -A USER_PROFILES=(
    ["quick_shower_user"]="Short showers, high efficiency"
    ["long_shower_user"]="Long relaxing showers"
    ["cold_climate_user"]="Lives in cold climate"
    ["hot_climate_user"]="Lives in hot climate" 
    ["eco_conscious_user"]="Minimizes energy usage"
    ["comfort_user"]="Prioritizes comfort over efficiency"
    ["morning_rusher"]="Quick morning showers"
    ["evening_relaxer"]="Long evening showers"
)

echo -e "${YELLOW}Creating diverse user profiles...${NC}"

# Quick shower user - prefers short, efficient showers
echo "Adding data for quick_shower_user..."
QUICK_DATA=(
    '{"duration": 3, "temperature": 25, "heatingTime": 2.5, "satisfaction": 50}'
    '{"duration": 4, "temperature": 22, "heatingTime": 3.0, "satisfaction": 48}'
    '{"duration": 5, "temperature": 20, "heatingTime": 4.0, "satisfaction": 52}'
    '{"duration": 3, "temperature": 28, "heatingTime": 2.0, "satisfaction": 50}'
    '{"duration": 4, "temperature": 24, "heatingTime": 2.8, "satisfaction": 49}'
)

for data in "${QUICK_DATA[@]}"; do
    duration=$(echo $data | jq -r '.duration')
    temperature=$(echo $data | jq -r '.temperature')
    heatingTime=$(echo $data | jq -r '.heatingTime')
    satisfaction=$(echo $data | jq -r '.satisfaction')
    
    curl -s -X POST http://localhost:8080/api/feedback \
        -H "Content-Type: application/json" \
        -d "{\"userId\": \"quick_shower_user\", \"showerDuration\": $duration, \"averageTemperature\": $temperature, \"heatingTime\": $heatingTime, \"satisfaction\": $satisfaction}" > /dev/null
done

# Long shower user - enjoys extended showers
echo "Adding data for long_shower_user..."
LONG_DATA=(
    '{"duration": 25, "temperature": 22, "heatingTime": 18.0, "satisfaction": 50}'
    '{"duration": 30, "temperature": 20, "heatingTime": 22.0, "satisfaction": 48}'
    '{"duration": 28, "temperature": 24, "heatingTime": 19.5, "satisfaction": 52}'
    '{"duration": 35, "temperature": 18, "heatingTime": 25.0, "satisfaction": 50}'
    '{"duration": 27, "temperature": 23, "heatingTime": 18.5, "satisfaction": 49}'
)

for data in "${LONG_DATA[@]}"; do
    duration=$(echo $data | jq -r '.duration')
    temperature=$(echo $data | jq -r '.temperature')
    heatingTime=$(echo $data | jq -r '.heatingTime')
    satisfaction=$(echo $data | jq -r '.satisfaction')
    
    curl -s -X POST http://localhost:8080/api/feedback \
        -H "Content-Type: application/json" \
        -d "{\"userId\": \"long_shower_user\", \"showerDuration\": $duration, \"averageTemperature\": $temperature, \"heatingTime\": $heatingTime, \"satisfaction\": $satisfaction}" > /dev/null
done

# Cold climate user - needs more heating in cold weather
echo "Adding data for cold_climate_user..."
COLD_DATA=(
    '{"duration": 15, "temperature": 5, "heatingTime": 20.0, "satisfaction": 50}'
    '{"duration": 12, "temperature": 2, "heatingTime": 18.5, "satisfaction": 48}'
    '{"duration": 18, "temperature": 8, "heatingTime": 22.0, "satisfaction": 52}'
    '{"duration": 10, "temperature": 0, "heatingTime": 16.0, "satisfaction": 50}'
    '{"duration": 20, "temperature": 3, "heatingTime": 24.0, "satisfaction": 49}'
)

for data in "${COLD_DATA[@]}"; do
    duration=$(echo $data | jq -r '.duration')
    temperature=$(echo $data | jq -r '.temperature')
    heatingTime=$(echo $data | jq -r '.heatingTime')
    satisfaction=$(echo $data | jq -r '.satisfaction')
    
    curl -s -X POST http://localhost:8080/api/feedback \
        -H "Content-Type: application/json" \
        -d "{\"userId\": \"cold_climate_user\", \"showerDuration\": $duration, \"averageTemperature\": $temperature, \"heatingTime\": $heatingTime, \"satisfaction\": $satisfaction}" > /dev/null
done

# Hot climate user - needs less heating in warm weather
echo "Adding data for hot_climate_user..."
HOT_DATA=(
    '{"duration": 15, "temperature": 35, "heatingTime": 3.0, "satisfaction": 50}'
    '{"duration": 12, "temperature": 38, "heatingTime": 2.5, "satisfaction": 48}'
    '{"duration": 18, "temperature": 32, "heatingTime": 4.0, "satisfaction": 52}'
    '{"duration": 10, "temperature": 40, "heatingTime": 2.0, "satisfaction": 50}'
    '{"duration": 20, "temperature": 30, "heatingTime": 5.0, "satisfaction": 49}'
)

for data in "${HOT_DATA[@]}"; do
    duration=$(echo $data | jq -r '.duration')
    temperature=$(echo $data | jq -r '.temperature')
    heatingTime=$(echo $data | jq -r '.heatingTime')
    satisfaction=$(echo $data | jq -r '.satisfaction')
    
    curl -s -X POST http://localhost:8080/api/feedback \
        -H "Content-Type: application/json" \
        -d "{\"userId\": \"hot_climate_user\", \"showerDuration\": $duration, \"averageTemperature\": $temperature, \"heatingTime\": $heatingTime, \"satisfaction\": $satisfaction}" > /dev/null
done

# Eco conscious user - minimizes energy usage
echo "Adding data for eco_conscious_user..."
ECO_DATA=(
    '{"duration": 8, "temperature": 22, "heatingTime": 5.0, "satisfaction": 50}'
    '{"duration": 6, "temperature": 25, "heatingTime": 4.0, "satisfaction": 48}'
    '{"duration": 10, "temperature": 20, "heatingTime": 6.0, "satisfaction": 52}'
    '{"duration": 7, "temperature": 24, "heatingTime": 4.5, "satisfaction": 50}'
    '{"duration": 9, "temperature": 21, "heatingTime": 5.5, "satisfaction": 49}'
)

for data in "${ECO_DATA[@]}"; do
    duration=$(echo $data | jq -r '.duration')
    temperature=$(echo $data | jq -r '.temperature')
    heatingTime=$(echo $data | jq -r '.heatingTime')
    satisfaction=$(echo $data | jq -r '.satisfaction')
    
    curl -s -X POST http://localhost:8080/api/feedback \
        -H "Content-Type: application/json" \
        -d "{\"userId\": \"eco_conscious_user\", \"showerDuration\": $duration, \"averageTemperature\": $temperature, \"heatingTime\": $heatingTime, \"satisfaction\": $satisfaction}" > /dev/null
done

# Comfort user - prioritizes comfort, uses more heating
echo "Adding data for comfort_user..."
COMFORT_DATA=(
    '{"duration": 15, "temperature": 22, "heatingTime": 15.0, "satisfaction": 50}'
    '{"duration": 12, "temperature": 25, "heatingTime": 12.0, "satisfaction": 48}'
    '{"duration": 18, "temperature": 20, "heatingTime": 18.0, "satisfaction": 52}'
    '{"duration": 10, "temperature": 24, "heatingTime": 10.0, "satisfaction": 50}'
    '{"duration": 20, "temperature": 21, "heatingTime": 20.0, "satisfaction": 49}'
)

for data in "${COMFORT_DATA[@]}"; do
    duration=$(echo $data | jq -r '.duration')
    temperature=$(echo $data | jq -r '.temperature')
    heatingTime=$(echo $data | jq -r '.heatingTime')
    satisfaction=$(echo $data | jq -r '.satisfaction')
    
    curl -s -X POST http://localhost:8080/api/feedback \
        -H "Content-Type: application/json" \
        -d "{\"userId\": \"comfort_user\", \"showerDuration\": $duration, \"averageTemperature\": $temperature, \"heatingTime\": $heatingTime, \"satisfaction\": $satisfaction}" > /dev/null
done

echo ""
echo -e "${GREEN}✅ Diverse global data seeded successfully!${NC}"
echo ""
echo -e "${BLUE}Added user profiles:${NC}"
for user in "${!USER_PROFILES[@]}"; do
    echo -e "  • ${YELLOW}$user${NC}: ${USER_PROFILES[$user]}"
done

echo ""
echo -e "${BLUE}Data distribution:${NC}"
echo "  • Quick showers (3-5 min): 5 records"
echo "  • Long showers (25-35 min): 5 records"
echo "  • Cold climate (0-8°C): 5 records"
echo "  • Hot climate (30-40°C): 5 records"
echo "  • Eco-friendly (short duration, efficient): 5 records"
echo "  • Comfort-focused (longer heating times): 5 records"
echo ""
echo -e "${GREEN}Total: 30 diverse records added to improve personalization!${NC}"