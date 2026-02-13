#!/usr/bin/env bash
# Bisection script to find which test creates unwanted files/state
# Usage: ./find-polluter.sh <file_or_dir_to_check> <test_pattern>
# Example: ./find-polluter.sh '.git' '**/*_test.go'

set -e

if [ $# -ne 2 ]; then
  echo "Usage: $0 <file_to_check> <test_pattern>"
  echo "Example: $0 '.git' '**/*_test.go'"
  exit 1
fi

POLLUTION_CHECK="$1"
TEST_PATTERN="$2"

echo "Searching for test that creates: $POLLUTION_CHECK"
echo "Test pattern: $TEST_PATTERN"
echo ""

# Get list of test files
TEST_FILES=$(find . -name '*_test.go' -path "$TEST_PATTERN" | sort)
TOTAL=$(echo "$TEST_FILES" | wc -l | tr -d ' ')

echo "Found $TOTAL test files"
echo ""

COUNT=0
for TEST_FILE in $TEST_FILES; do
  COUNT=$((COUNT + 1))

  # Skip if pollution already exists
  if [ -e "$POLLUTION_CHECK" ]; then
    echo "Warning: Pollution already exists before test $COUNT/$TOTAL"
    echo "   Skipping: $TEST_FILE"
    continue
  fi

  # Extract package path from test file
  PKG_DIR=$(dirname "$TEST_FILE")

  echo "[$COUNT/$TOTAL] Testing: $TEST_FILE"

  # Run the test file's package
  go test -count=1 "$PKG_DIR" > /dev/null 2>&1 || true

  # Check if pollution appeared
  if [ -e "$POLLUTION_CHECK" ]; then
    echo ""
    echo "FOUND POLLUTER!"
    echo "   Test file: $TEST_FILE"
    echo "   Package:   $PKG_DIR"
    echo "   Created:   $POLLUTION_CHECK"
    echo ""
    echo "Pollution details:"
    ls -la "$POLLUTION_CHECK"
    echo ""
    echo "To investigate:"
    echo "  go test -v -count=1 $PKG_DIR    # Run package tests"
    echo "  cat $TEST_FILE                    # Review test code"
    exit 1
  fi
done

echo ""
echo "No polluter found - all tests clean!"
exit 0
