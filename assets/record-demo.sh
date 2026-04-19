#!/bin/bash
# Script to record the demo GIF
# Chains push + pull so the link is captured automatically
# Usage: bash assets/record-demo.sh

set -e

CLI="./bin/kip"
ENV_FILE="test.env"

echo "$ cat $ENV_FILE"
cat "$ENV_FILE"
echo ""
sleep 2

echo "$ kip push $ENV_FILE"
OUTPUT=$($CLI push "$ENV_FILE")
echo "$OUTPUT"
echo ""
sleep 3

# Extract the link from output
LINK=$(echo "$OUTPUT" | grep -oP 'http\S+')

echo "# --- On your teammate's machine ---"
echo ""
sleep 1

echo "$ kip pull $LINK"
$CLI pull "$LINK"
sleep 3
