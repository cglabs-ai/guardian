#!/bin/bash
# Guardian checks for Go projects

set -e

echo "Running Go checks..."

# Run go vet
go vet ./...

# Run staticcheck if available
if command -v staticcheck &> /dev/null; then
    staticcheck ./...
fi

# Check for dangerous patterns (|| true prevents exit on no matches)
if grep -rn "os.Remove\|os.RemoveAll\|exec.Command" --include="*.go" . 2>/dev/null; then
    echo "Warning: Dangerous file operations detected"
fi

echo "Guardian checks complete"
