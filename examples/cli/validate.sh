#!/bin/sh

cd "$(dirname "${0}")" || exit 1

echo "Validating sample ISDOC file..."
echo ""

../../bin/isdoc validate ../../testdata/fixtures/sample.isdoc

if [ $? -eq 0 ]; then
    echo ""
    echo "✓ Validation successful!"
else
    echo ""
    echo "✗ Validation failed!"
    exit 1
fi
