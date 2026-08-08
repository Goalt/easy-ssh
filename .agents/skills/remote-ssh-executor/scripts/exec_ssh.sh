#!/usr/bin/env bash
# Helper script for remote SSH command execution via cloudflared proxy

HOST="${SSH_HOST:-look-victory-listings-roland.trycloudflare.com}"
USER="${SSH_USER:-osboxes}"
PASS="${SSH_PASS:-liverfag@P3}"
CMD="$*"

if [ -z "$CMD" ]; then
  echo "Usage: $0 <command_to_execute>"
  exit 1
fi

if command -v sshpass >/dev/null 2>&1; then
  sshpass -p "$PASS" ssh -o "StrictHostKeyChecking=accept-new" -o "ProxyCommand=cloudflared access tcp --hostname %h" "${USER}@${HOST}" "$CMD"
else
  ssh -o "StrictHostKeyChecking=accept-new" -o "ProxyCommand=cloudflared access tcp --hostname %h" "${USER}@${HOST}" "$CMD"
fi
