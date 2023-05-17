#!/usr/bin/env bash

echo "(📝) Asserting that source code comply with Go Vet rules..."
OUT=$(go vet ./...)
if [[ -n $OUT ]]; then
    echo -e "(❌) The following files violate Go Vet checks:\n${OUT}"
    exit 1
fi
echo "(✅) Go Vet check has passed successfully"

exit 0