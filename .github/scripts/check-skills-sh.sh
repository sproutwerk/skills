#!/usr/bin/env bash
# Copyright (c) 2026 sproutwerk. SPDX-License-Identifier: MIT
set -euo pipefail

repo="$(cd "${1:?Supply the repository path}" && pwd)"
project="$(mktemp -d)"
trap 'rm -rf "$project"' EXIT
export DISABLE_TELEMETRY=1 DO_NOT_TRACK=1

# Every skill directory must be grouped in skills.sh.json, and nothing else.
diff <(find "$repo/skills" -mindepth 2 -maxdepth 2 -name SKILL.md -printf '%h\n' | xargs -n1 basename | sort) \
     <(jq -r '.groupings[].skills[]' "$repo/skills.sh.json" | sort)

# A failed discovery or an empty installation must not look like a pass.
npx --yes skills@1.7.2 add "$repo" --list
cd "$project"
npx --yes skills@1.7.2 add "$repo" --agent codex --skill '*' --copy --yes
count=0
for entrypoint in "$repo"/skills/*/SKILL.md; do
  skill="$(basename "$(dirname "$entrypoint")")"
  grep -qx "name: $skill" "$entrypoint" || { printf 'Frontmatter name does not match directory: %s\n' "$skill" >&2; exit 1; }
  installed="$project/.agents/skills/$skill/SKILL.md"
  if [[ ! -f "$installed" ]]; then
    printf 'Missing installed skill: %s\n' "$skill" >&2
    exit 1
  fi
  cmp "$entrypoint" "$installed"
  count=$((count + 1))
done
[[ "$count" -gt 0 ]]
printf 'Installed and verified %d skills\n' "$count"
