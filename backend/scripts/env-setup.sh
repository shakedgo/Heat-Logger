#!/bin/bash

# Environment setup script for Heat-Logger

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Setting up environment for Heat-Logger...${NC}"

# Check if .env file exists
if [ ! -f .env ]; then
    echo -e "${YELLOW}Creating .env file from .env.example...${NC}"
    if [ -f .env.example ]; then
        cp .env.example .env
        echo -e "${GREEN}Created .env file${NC}"
    else
        echo -e "${RED}Error: .env.example not found${NC}"
        exit 1
    fi
else
    echo -e "${GREEN}.env file already exists${NC}"
fi

# Function to set environment variable
set_env_var() {
    local key=$1
    local value=$2
    local current_value=$(grep "^${key}=" .env 2>/dev/null | cut -d'=' -f2- || echo "")
    
    if [ -z "$current_value" ]; then
        echo "${key}=${value}" >> .env
        echo -e "${GREEN}Added ${key}=${value}${NC}"
    else
        echo -e "${YELLOW}${key} already set to ${current_value}${NC}"
    fi
}

# Set common environment variables
echo -e "${YELLOW}Setting common environment variables...${NC}"
set_env_var "SERVER_PORT" "8080"
set_env_var "SERVER_HOST" "localhost"
set_env_var "DATABASE_PATH" "./data.db"
set_env_var "DATABASE_DRIVER" "sqlite"
set_env_var "PREDICTOR_VERSION" "v2"
set_env_var "ENVIRONMENT" "development"
set_env_var "GIN_MODE" "debug"

echo -e "${GREEN}Environment setup complete!${NC}"
echo -e "${YELLOW}You can now edit .env file to customize your configuration${NC}"
