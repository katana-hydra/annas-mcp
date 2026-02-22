#!/bin/bash

# Healthcheck script for annas-mcp
# Tests basic functionality of book and article search

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "====================================="
echo "  Anna's Archive MCP Healthcheck"
echo "====================================="
echo ""

# Track test results
TESTS_PASSED=0
TESTS_FAILED=0

# Test 1: Book search for "crypto"
echo -e "${YELLOW}[TEST 1]${NC} Testing book search with term 'crypto'..."
BOOK_RESULT=$(go run ./cmd/annas-mcp book-search "crypto" 2>&1)
BOOK_EXIT_CODE=$?

if [ $BOOK_EXIT_CODE -ne 0 ]; then
    echo -e "${RED}✗ FAILED${NC} - Book search command failed with exit code $BOOK_EXIT_CODE"
    echo "$BOOK_RESULT"
    TESTS_FAILED=$((TESTS_FAILED + 1))
else
    # Check if at least one result was returned
    if echo "$BOOK_RESULT" | grep -q "Book 1:"; then
        echo -e "${GREEN}✓ PASSED${NC} - Book search returned results"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗ FAILED${NC} - Book search returned no results"
        echo "$BOOK_RESULT"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
fi
echo ""

# Test 2: Article search by DOI (Attention is All You Need paper)
echo -e "${YELLOW}[TEST 2]${NC} Testing article search with DOI '10.48550/arXiv.1706.03762'..."
ARTICLE_RESULT=$(go run ./cmd/annas-mcp article-search "10.48550/arXiv.1706.03762" 2>&1)
ARTICLE_EXIT_CODE=$?

if [ $ARTICLE_EXIT_CODE -ne 0 ]; then
    echo -e "${RED}✗ FAILED${NC} - Article search command failed with exit code $ARTICLE_EXIT_CODE"
    echo "$ARTICLE_RESULT"
    TESTS_FAILED=$((TESTS_FAILED + 1))
else
    # Check if paper metadata was returned
    if echo "$ARTICLE_RESULT" | grep -q "DOI:"; then
        echo -e "${GREEN}✓ PASSED${NC} - Article search returned paper details"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗ FAILED${NC} - Article search returned no paper details"
        echo "$ARTICLE_RESULT"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
fi
echo ""

# Summary
echo "====================================="
echo "  Test Summary"
echo "====================================="
echo -e "Total tests: $((TESTS_PASSED + TESTS_FAILED))"
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Failed: $TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All healthchecks passed!${NC}"
    exit 0
else
    echo -e "${RED}Some healthchecks failed!${NC}"
    exit 1
fi
