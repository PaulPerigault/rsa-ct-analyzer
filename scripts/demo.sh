#!/bin/sh
# Offline end-to-end demo: simulated CT log -> crawl -> scan. No network, no root.
set -eu

BIN="${BIN:-./bin/rsa-ct-analyzer}"
ADDR="127.0.0.1:18080"
WORK="$(mktemp -d)"

"$BIN" mock-log --addr "$ADDR" --precert-every 4 testdata/certs 2>"$WORK/mock.log" &
LOG_PID=$!
trap 'kill "$LOG_PID" 2>/dev/null || true; rm -rf "$WORK"' EXIT INT TERM

echo "== crawl: simulated CT log at http://$ADDR"
"$BIN" crawl --log-url "http://$ADDR" --start 0 --count 100 --rate-limit 0 \
  --retries 10 -o "$WORK/crawled.pem"

echo
echo "== scan: crawled certificates"
status=0
"$BIN" scan --fail-on critical "$WORK/crawled.pem" || status=$?
echo
echo "scan exit status: $status (2 = a critical finding, expected on the sample data)"
