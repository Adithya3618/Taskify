#!/bin/bash

# Run All Backend Tests
# Usage: ./run_all_tests.sh

set -e

echo "========================================="
echo "  Taskify Backend - Running All Tests"
echo "========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get the directory where the script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(dirname "$SCRIPT_DIR")"

cd "$BACKEND_DIR"

echo "Backend directory: $BACKEND_DIR"
echo ""

# Check if go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

echo -e "${YELLOW}Running all unit tests in ./internal/testcases/...${NC}"
echo ""

# Run all tests with verbose output and coverage
go test -v -cover ./internal/testcases/...

EXIT_CODE=$?

echo ""
echo "========================================="

if [ $EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
else
    echo -e "${RED}Some tests failed!${NC}"
fi

echo "========================================="

exit $EXIT_CODE
