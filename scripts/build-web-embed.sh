#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
WEB_DIR="$ROOT_DIR/web"
EMBED_DIR="$ROOT_DIR/pkg/embeds/web"

if ! command -v yarn >/dev/null 2>&1; then
  echo "yarn is required but not found" >&2
  exit 1
fi

cd "$WEB_DIR"
yarn install --frozen-lockfile
yarn build

if [ ! -d "$WEB_DIR/dist" ]; then
  echo "web build output '$WEB_DIR/dist' not found" >&2
  exit 1
fi

rm -rf "$EMBED_DIR"
mkdir -p "$EMBED_DIR"
cp -R "$WEB_DIR/dist/." "$EMBED_DIR/"

if [ ! -f "$EMBED_DIR/index.html" ]; then
  echo "embedded index.html missing after copy" >&2
  exit 1
fi

echo "embedded web assets updated in $EMBED_DIR"
