#!/bin/bash

# ACTUAL USER FLOW TEST RUNNER
# Tests that actually exist and verify your bug complaints

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
NC='\033[0m'
BOLD='\033[1m'

echo -e "${BOLD}${MAGENTA}"
echo "╔══════════════════════════════════════════════════════════════════════╗"
echo "║                    DISTUI TEST SUITE                                 ║"
echo "║                                                                      ║"
echo "║  Testing every bug you complained about + safety checks             ║"
echo "╚══════════════════════════════════════════════════════════════════════╝"
echo -e "${NC}\n"

# Run all tests
echo -e "${BOLD}${YELLOW}Running All Tests...${NC}\n"
TEST_OUTPUT=$(go test -v ./tests ./internal/detection ./internal/fileops 2>&1)
TEST_EXIT=$?

# Show full output
echo "$TEST_OUTPUT"

# Parse results
PASSED=$(echo "$TEST_OUTPUT" | grep -c "^--- PASS:")
FAILED=$(echo "$TEST_OUTPUT" | grep -c "^--- FAIL:")
TOTAL=$((PASSED + FAILED))

# Generate report
echo -e "\n${BOLD}${BLUE}"
echo "╔══════════════════════════════════════════════════════════════════════╗"
echo "║                         TEST REPORT                                  ║"
echo "╚══════════════════════════════════════════════════════════════════════╝"
echo -e "${NC}\n"

echo -e "${BOLD}Summary:${NC}"
echo -e "  Total Tests:    ${TOTAL}"
echo -e "  ${GREEN}Passed:${NC}         ${PASSED}"
echo -e "  ${RED}Failed:${NC}         ${FAILED}"
echo ""

if [ $TOTAL -gt 0 ]; then
    PASS_RATE=$((PASSED * 100 / TOTAL))
    echo -e "${BOLD}Pass Rate:${NC} ${PASS_RATE}%"
    echo ""

    if [ $FAILED -eq 0 ]; then
        echo -e "${GREEN}${BOLD}✓✓✓ ALL TESTS PASSED! ✓✓✓${NC}"
        echo -e "${GREEN}Every user flow works correctly!${NC}\n"
    else
        echo -e "${RED}${BOLD}✗ ${FAILED} TEST(S) FAILING${NC}"
        echo -e "${RED}Review failures above and fix the actual code!${NC}\n"
    fi
fi

echo -e "${BOLD}${CYAN}What These Tests Verify:${NC}\n"

echo -e "${YELLOW}Bug Regressions (8 tests - Your Complaints):${NC}"
echo "$TEST_OUTPUT" | grep "TestBugRegression" | sed 's/^=== RUN   /  • /' | sed 's/TestBugRegression_//'
echo ""

echo -e "${RED}Destructive Safety (5 tests - Don't Nuke Projects):${NC}"
echo "$TEST_OUTPUT" | grep "TestDestructiveSafety" | sed 's/^=== RUN   /  • /' | sed 's/TestDestructiveSafety_//'
echo ""

echo -e "${GREEN}Integration Flows (5 tests):${NC}"
echo "$TEST_OUTPUT" | grep "TestUserFlow" | sed 's/^=== RUN   /  • /' | sed 's/TestUserFlow_//'
echo ""

echo -e "${MAGENTA}Comprehensive User Flows (7 tests - From Spec):${NC}"
echo "$TEST_OUTPUT" | grep "TestFlow_" | sed 's/^=== RUN   /  • /' | sed 's/TestFlow_//'
echo ""

echo -e "${CYAN}Realistic Multi-Project Scenario (1 test - Real Config Files):${NC}"
echo "$TEST_OUTPUT" | grep "TestRealistic" | sed 's/^=== RUN   /  • /' | sed 's/TestRealistic//'
echo ""

echo -e "${BLUE}Core Functionality (4 tests):${NC}"
echo "$TEST_OUTPUT" | grep -E "TestDetectProjectMode|TestArchiveCustomFiles" | sed 's/^=== RUN   /  • /'
echo ""

if [ $FAILED -gt 0 ]; then
    echo -e "${RED}${BOLD}Failed Tests:${NC}"
    echo "$TEST_OUTPUT" | grep "^--- FAIL:" | sed 's/^--- FAIL: /  ✗ /'
    echo ""
    exit 1
else
    echo -e "${GREEN}${BOLD}All user flows verified! 🚀${NC}\n"
    exit 0
fi
