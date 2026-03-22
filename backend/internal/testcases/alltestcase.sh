#!/bin/bash

# All Test Cases Runner
# This script runs all backend unit tests in the testcases package

echo "=============================================="
echo "  Running All Backend Unit Tests"
echo "=============================================="
echo ""

# Change to backend directory
cd "$(dirname "$0")/../.."

# Run all tests in the testcases package
echo "Running tests in ./internal/testcases/..."
echo ""

# Run with verbose output and coverage
go test -v -cover ./internal/testcases/...

# Capture exit code
EXIT_CODE=$?

echo ""
echo "=============================================="
if [ $EXIT_CODE -eq 0 ]; then
    echo "  All Tests Passed!"
else
    echo "  Some Tests Failed!"
fi
echo "=============================================="

exit $EXIT_CODE
