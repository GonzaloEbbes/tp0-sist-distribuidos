#!/usr/bin/env bash

HOST="server"
PORT="12345"
MESSAGE="Hola!"

RESPONSE="$(printf '%s\n' "$MESSAGE" | nc "$HOST" "$PORT" 2>/dev/null || true)"
RESPONSE="${RESPONSE%$'\n'}"

if [[ "$RESPONSE" == "$MESSAGE" ]]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi
