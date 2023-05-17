#!/usr/bin/env bash

MAIN=$1; shift
OUT_EXEC=$1

echo "(🚧) Building $MAIN"
OUT=$(go build -a -o $OUT_EXEC $MAIN)
if [[ -n $OUT ]]; then
    echo -e "(❌) Error occurred while building remote work processor executable:\n${OUT}"
    exit 1
fi
echo "(✅) Remote work processor has been built successfully"

exit 0