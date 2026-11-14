#!/usr/bin/env bash
# Re-copy the curated file set and regenerate every index.
# usage: tools/refresh.sh <Dalamud checkout> [<dalamud-docs checkout>]
set -euo pipefail
SRC=$(realpath "${1:?path to a goatcorp/Dalamud checkout}")
DOCS=${2:-}
HERE=$(cd "$(dirname "$0")/.." && pwd)
REF="$HERE/../../skills/dalamudimgui/references"
BIN=$(mktemp -d)
GEN="$BIN/gen-index"; ENUMS="$BIN/gen-enums"
go -C "$HERE" build -o "$GEN" ./tools/gen-index
go -C "$HERE" build -o "$ENUMS" ./tools/gen-enums

# 1. verbatim copies, same file set as references/manifest.txt
rm -rf "$REF/src"
missing=0
while IFS= read -r f; do
  [ -z "$f" ] && continue
  if [ -f "$SRC/$f" ]; then
    mkdir -p "$REF/src/$(dirname "$f")"
    cp "$SRC/$f" "$REF/src/$f"
  else
    echo "MISSING upstream (moved/renamed?): $f" >&2
    missing=$((missing + 1))
  fi
done < "$REF/manifest.txt"

# 2. indexes (only files that are actually compiled, see the csproj Compile Remove rules)
mkdir -p "$REF/index"
B="$SRC/imgui/Dalamud.Bindings.ImGui"
export INDEX_BASE="$SRC/imgui"
"$GEN" "$REF/index/imgui-methods.txt"  'ImGui'  "$B/Custom" "$B/ImGui.cs"
"$GEN" "$REF/index/imguip-methods.txt" 'ImGuiP' "$B/Custom" "$B/ImGuiP.cs"
"$GEN" "$REF/index/imgui-struct-methods.txt" '.*' --exclude '^(ImGui|ImGuiP|ImGuiNative|ImGuiPNative)$' \
  "$B/Custom" "$B/ImU8String.cs" "$B/ImTextureID.cs" "$B/ImVector.cs"
# ImPlot: raw-pointer overloads dropped (the generator emits every pointer/ref/span permutation, ~20k lines)
INDEX_SKIP_POINTERS=1 "$GEN" "$REF/index/implot-methods.txt" 'ImPlot' "$SRC/imgui/Dalamud.Bindings.ImPlot"
"$GEN" "$REF/index/imguizmo-methods.txt" 'ImGuizmo' "$SRC/imgui/Dalamud.Bindings.ImGuizmo"
INDEX_BASE="$SRC" "$GEN" "$REF/index/interface-methods.txt" '.*' \
  $(grep -E '^Dalamud/' "$REF/manifest.txt" | sed "s|^|$SRC/|")

# enums: one line per enum with its members
"$ENUMS" "$SRC/imgui" "$REF/index/enums.txt"

# FontAwesome icon names
grep -oE '^\s+[A-Za-z0-9_]+ = 0x[0-9A-Fa-f]+' "$SRC/Dalamud/Interface/FontAwesome/FontAwesomeIcon.cs" \
  | sed -E 's/^\s+//' > "$REF/index/fontawesome-icons.txt"
grep -m1 'Font-Awesome Version' "$SRC/Dalamud/Interface/FontAwesome/FontAwesomeIcon.cs" \
  | sed 's|^// ||' > "$REF/index/fontawesome-version.txt"

# file index of Dalamud/Interface
( cd "$SRC/Dalamud/Interface" && find . -type f | sed 's|^\./||' | sort ) > "$REF/index/interface-files.txt"

# 3. upstream migration docs
if [ -n "$DOCS" ]; then
  # v16 became a directory upstream; keep its layout so index.md can link to sestring.md
  rm -rf "$REF/upstream-docs"
  mkdir -p "$REF/upstream-docs/v16"
  for v in v13 v14 v15 v16/index v16/sestring; do cp "$DOCS/docs/versions/$v.md" "$REF/upstream-docs/$v.md"; done
fi

# 4. provenance
{
  echo "Dalamud: $(git -C "$SRC" log -1 --format='%H %ci')"
  echo "DalamudVersion: $(grep -oP '(?<=<DalamudVersion>)[^<]+' "$SRC/Dalamud/Dalamud.csproj")"
  echo "submodules:"
  git -C "$SRC" submodule status | sed 's/^/  /'
  [ -n "$DOCS" ] && echo "dalamud-docs: $(git -C "$DOCS" log -1 --format='%H %ci')"
} > "$REF/PROVENANCE.txt"

echo "done; missing=$missing"
wc -l "$REF"/index/*.txt
