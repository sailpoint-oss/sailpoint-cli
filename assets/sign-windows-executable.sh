#!/bin/bash
set -e

EXE="$1"

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

osslsigncode sign -n "SailPoint CLI" -t http://timestamp.digicert.com \
  -certs "$CERT_FILE" -key "$KEY_FILE" \
  -in "$EXE" -out "$EXE"~

mv "$EXE"~ "$EXE"