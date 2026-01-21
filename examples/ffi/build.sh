#!/usr/bin/env bash
set -e
cd "$(dirname "${0}")" || exit

echo "Building FFI shared library..."
cd ../..
make build-ffi

echo ""
echo "Running examples..."
echo ""

for lang in python php swift; do
    echo "[$lang]"
    cd "examples/ffi/$lang"
    bash run.sh || echo "  Failed"
    echo ""
    cd - > /dev/null
done
