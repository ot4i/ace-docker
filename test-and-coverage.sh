#!/bin/bash -e

# Array of packages at 100% test coverage - Once a package has been fully covered it will be added to this list
COVERED_PACKAGES=('webadmin')
TEST_OUTPUT=$(go test $(go list ./... | grep -v /test/) -covermode=atomic -coverprofile cover.out)
COVERAGE_REGRESSION=false

for i in ${COVERED_PACKAGES[@]}; do
    COV=$(echo "$TEST_OUTPUT" | grep "$i" | awk '{ print $5 }')
    if [[ "$COV" != "100.0%" ]]; then
        echo "$i is not at 100% test coverage."
        COVERAGE_REGRESSION=true
    fi
done

if [[ $COVERAGE_REGRESSION == true ]]; then
    echo "Please address the coverage regression."
    exit 1
fi


# This is the current expected code coverage
CURRENT_COVERAGE=58.1
LATEST=$(go tool cover -func cover.out | grep total | awk '{print substr($3, 1, length($3)-1)}')
echo "Latest coverage: $LATEST"
echo "Expected Coverage: $CURRENT_COVERAGE"
result=$(echo "$LATEST >= $CURRENT_COVERAGE" | bc -l)
if [ $result -gt 0 ]; then
    echo "PASSED - Coverage Check"
else
    echo "Failed to meet required coverage value: $CURRENT_COVERAGE"
    exit 1

fi
if [ "$LATEST" != "$CURRENT_COVERAGE" ]; then
    echo "\nFAILED - You must update the CURRENT_COVERAGE in travis to match the new benchmark: $LATEST"
    exit 1
fi
