#!/usr/bin/env bash

echo "(📝) Asserting that source code comply with Go Fmt rules..."
OUT=$(go fmt ./...)
if [[ -n $OUT ]]; then
    echo -e "(❌) The following files violate Go Fmt checks:\n${OUT}"
    exit 1
fi
echo "(✅) Go Fmt check has passed successfully"

exit 0