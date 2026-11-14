#!/usr/bin/env bash
# Usage: tools/refresh.sh <Luna checkout> <ImSharp checkout> <Luna.Targa checkout>
# Re-copies exactly the files listed in skills/luna/references/manifest-{luna,imsharp,targa}.txt from
# the given checkouts into skills/luna/references/{Luna,ImSharp,Luna.Targa}/ and regenerates the
# indices. Files that vanished upstream are reported, not silently dropped.
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"; skill="$root/../../skills/luna"
luna="$(cd "$1" && pwd)"; imsharp="$(cd "$2" && pwd)"; targa="$(cd "$3" && pwd)"
copy_set() { # <src root> <manifest> <dest>
    rm -rf "$3"; mkdir -p "$3"
    while IFS= read -r f; do
        [ -z "$f" ] && continue
        if [ -f "$1/$f" ]; then mkdir -p "$3/$(dirname "$f")"; cp "$1/$f" "$3/$f"
        else echo "MISSING upstream: $f" >&2; fi
    done < "$2"
}
list_cs() { # <root> <label>
    ( cd "$1" && find . -name '*.cs' -not -path './.git/*' -not -path '*/bin/*' -not -path '*/obj/*' | sed "s|^\./|$2/|" )
}
copy_set "$luna"    "$skill/references/manifest-luna.txt"    "$skill/references/Luna"
copy_set "$imsharp" "$skill/references/manifest-imsharp.txt" "$skill/references/ImSharp"
copy_set "$targa"   "$skill/references/manifest-targa.txt"   "$skill/references/Luna.Targa"
for d in "$skill/LICENSES" "$skill/references/LICENSES"; do
    cp "$luna/LICENSE"    "$d/AGPL-3.0-only.txt"
    cp "$imsharp/LICENSE" "$d/MIT-ImSharp.txt"
    cp "$targa/LICENSE"   "$d/MIT-Luna.Targa.txt"
done
{
    "$root/tools/build-index.sh" "$luna" Luna
    "$root/tools/build-index.sh" "$imsharp" ImSharp
    "$root/tools/build-index.sh" "$targa" Luna.Targa
} > "$skill/references/api-index.txt"
{
    list_cs "$luna" Luna
    list_cs "$imsharp" ImSharp
    list_cs "$targa" Luna.Targa
} | LC_ALL=C sort > "$skill/references/all-files.txt"
pin() { git -C "$luna" ls-tree HEAD "$1" | awk '{print $3}'; }
{
    echo "Luna       $(git -C "$luna" log -1 --format='%H %cI')"
    echo "ImSharp    $(git -C "$imsharp" log -1 --format='%H %cI') (Luna pins: $(pin ImSharp))"
    echo "Luna.Targa $(git -C "$targa" log -1 --format='%H %cI') (Luna pins: $(pin Luna.Targa))"
} > "$skill/references/VERSIONS.txt"
cat "$skill/references/VERSIONS.txt"
echo "files under references/: $(find "$skill/references" -type f | wc -l)"
