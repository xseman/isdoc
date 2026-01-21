#!/bin/sh

cd "$(dirname "${0}")" || exit 1

INPUT_FILE="../../testdata/fixtures/sample.isdoc"
OUTPUT_FILE="output.isdoc"

echo "Decoding ISDOC to JSON..."
echo ""

../../bin/isdoc decode "$INPUT_FILE" > decoded.json

if [ $? -eq 0 ]; then
    echo "✓ Successfully decoded to JSON"
    echo ""
    echo "First few lines of decoded JSON:"
    head -n 10 decoded.json
    echo ""
else
    echo "✗ Decoding failed!"
    exit 1
fi

echo "Encoding back to ISDOC XML..."
echo ""

../../bin/isdoc encode decoded.json > "$OUTPUT_FILE"

if [ $? -eq 0 ]; then
    echo "✓ Successfully encoded to ISDOC XML"
    echo ""
    echo "Validating output..."
    ../../bin/isdoc validate "$OUTPUT_FILE"
    
    if [ $? -eq 0 ]; then
        echo ""
        echo "✓ Round-trip successful!"
    else
        echo ""
        echo "✗ Validation failed after round-trip!"
        exit 1
    fi
else
    echo "✗ Encoding failed!"
    exit 1
fi

# Cleanup
rm -f decoded.json "$OUTPUT_FILE"

echo ""
echo "Cleaned up temporary files."
