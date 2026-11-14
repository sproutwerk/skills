#!/usr/bin/env bash
# Usage: build-index.sh <root-dir> <label>
# Prints a grep-able index of public/protected declarations for every .cs file
# under <root-dir> (skipping bin/obj/.git), as "<label>/<relative path>:<line>: <declaration>".
set -euo pipefail
root="$1"; label="$2"
cd "$root"
find . -name '*.cs' -not -path './.git/*' -not -path '*/bin/*' -not -path '*/obj/*' | sed 's|^\./||' | LC_ALL=C sort |
while IFS= read -r f; do
    grep -nE '^[[:space:]]*(\[[^]]*\][[:space:]]*)*(public|protected)[[:space:]]' "$f" 2>/dev/null |
        sed -E 's/^([0-9]+):[[:space:]]+/\1: /; s/[[:space:]]+$//' |
        sed "s|^|$label/$f:|" || true
done
