#!/bin/bash

# Get current year and term
year=$(date +%Y)
month=$(date +%m)

# Determine current term
if [ $month -ge 1 ] && [ $month -le 4 ]; then
    term="spring"
elif [ $month -ge 5 ] && [ $month -le 8 ]; then
    term="summer"
else
    term="fall"
fi

# Determine next term
if [ "$term" = "spring" ]; then
    nextTerm="summer"
    nextYear=$year
elif [ "$term" = "summer" ]; then
    nextTerm="fall"
    nextYear=$year
else
    nextTerm="spring"
    nextYear=$((year + 1))
fi

# Build all binaries
echo "Building binaries..."
go build -o bin/fetch-sections scripts/fetchSections/main.go
go build -o bin/sync-offerings scripts/syncOfferings/main.go
go build -o bin/sync-instructors scripts/syncInstructors/main.go

# Run the commands
echo "Running data sync commands..."
./bin/fetch-sections $year $term
./bin/fetch-sections $nextYear $nextTerm
./bin/sync-offerings
./bin/sync-instructors

# Add any changes to git
git add internal/store/json/ 