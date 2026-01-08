#!/bin/bash
set -e

EXE="$1"

echo "=== Windows Code Signing Script Started ===" >&2
echo "EXE: $EXE" >&2
echo "CERT_FILE: $CERT_FILE" >&2
echo "KEY_FILE: $KEY_FILE" >&2
echo "TEST_MODE: $TEST_MODE" >&2

if [ -z "$CERT_FILE" ]; then
  echo "skipping Windows code-signing; CERT_FILE not set" >&2
  exit 0
fi

if [ ! -f "$CERT_FILE" ]; then
  echo "error Windows code-signing; file '$CERT_FILE' not found" >&2
  exit 1
fi

if [ -z "$KEY_FILE" ]; then
  echo "skipping Windows code-signing; KEY_FILE not set" >&2
  exit 0
fi

if [ ! -f "$KEY_FILE" ]; then
  echo "error Windows code-signing; file '$KEY_FILE' not found" >&2
  exit 1
fi

# Determine if we should use timestamping (skip in test mode)
if [ "$TEST_MODE" = "true" ]; then
  echo "Signing without timestamp (test mode)" >&2
  timeout 120 osslsigncode sign -n "SailPoint CLI" \
    -certs "$CERT_FILE" -key "$KEY_FILE" \
    -in "$EXE" -out "$EXE"~ 2>&1 | tee /dev/stderr
  
  if [ $? -eq 124 ]; then
    echo "ERROR: Signing timed out after 120 seconds" >&2
    exit 1
  fi
  echo "Signing completed successfully" >&2
else
  echo "Signing with timestamp" >&2
  timeout 180 osslsigncode sign -n "SailPoint CLI" -t http://timestamp.digicert.com \
    -certs "$CERT_FILE" -key "$KEY_FILE" \
    -in "$EXE" -out "$EXE"~ 2>&1 | tee /dev/stderr
  
  if [ $? -eq 124 ]; then
    echo "ERROR: Signing with timestamp timed out after 180 seconds" >&2
    exit 1
  fi
  echo "Signing with timestamp completed successfully" >&2
fi

mv "$EXE"~ "$EXE"
echo "=== Windows Code Signing Script Completed ===" >&2