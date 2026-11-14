# Source provenance and snapshot use

The curated FFXIVClientStructs source, generator files, upstream docs and indexes
track [80cb4a7b7e11a6702e2eb8839afd301a918c638e](https://github.com/aers/FFXIVClientStructs/tree/80cb4a7b7e11a6702e2eb8839afd301a918c638e)
(2026-11-10), which was upstream `main` on the review date. Its `ida/data.yml`
executable version is `2026.11.03.0000.0000`; its assembly version is still
`7.57.0.$(GitCommits)`, although the upstream commit that moved `data.yml` to
this build is titled "7.58". Upstream publishes no tags or releases.
Dalamud [43b3411](https://github.com/goatcorp/Dalamud/tree/43b34118bcf7213e575ec39517d033191b1acf17)
(version 15.0.3.9) pins that commit. This identifies the reviewed Dalamud
source tree, not every installed release or testing channel.

The upstream MIT notice is preserved in
[licenses/FFXIVClientStructs-MIT.txt](licenses/FFXIVClientStructs-MIT.txt).
`ecommons/Callback.cs` is identical to `ECommons/Automation/Callback.cs` at
ECommons [bc38dc7](https://github.com/NightmareXIV/ECommons/tree/bc38dc71340fa3b266de58e3d6d876a73313b61e)
(2026-11-08); upstream last changed that file in `d435e12`. The ECommons MIT
notice is in [licenses/ECommons-MIT.txt](licenses/ECommons-MIT.txt).

## Review on 2026-11-13

The [comparison from the previous snapshot](https://github.com/aers/FFXIVClientStructs/compare/dd57986b8287e7c6ed9c27198f8c76cb04a81dfe...80cb4a7b7e11a6702e2eb8839afd301a918c638e)
changes the game build and every `data.yml` address. Six curated files changed:

- `AtkResNode.GetAsAtkComponentTextInput` has a new signature.
- `AtkTextNode.SetText(ReadOnlySpan<byte>)` and
  `AtkComponentButton.SetText(ReadOnlySpan<byte>)` now pass a pointer to a zero
  byte when the span is empty.
- `StringArrayType` gains `InventoryBuddy = 87`.
- `FFXIVClientStructs.csproj` sets `UseTemplatedCode` and `GitRepositoryUrl`.
- `pr-change-check.yaml` now runs only for pull requests against `main`.

The other 96 curated files, including the generator, attributes and runtime
resolver, are unchanged. Outside the curated set, upstream added inventory
string and number array types under `Client/UI/Arrays`, look-at and IK
animation structs, DLSS structs, and the `SQEX.Sd` audio driver types; the
type index lists them.

All 102 copied FFXIVClientStructs files were compared byte for byte with a
checkout of the selected commit. Upstream's `.gitattributes` checks text files
out with CRLF line endings, and the copies keep them. The type index now lists
1686 C# files; the `data.yml` excerpt retains the same selection of globals,
functions and the Framework, InventoryManager and AtkUnitBase entries from a
file with 2637 classes. The guide's version strings, example addresses, counts
and Dalamud crossing notes were rechecked against these sources. Nothing was
compiled. No game, memory reads, callbacks or signature resolver checks were
run. Source review alone does not validate runtime addresses.

See [refresh criteria](guide-conventions-and-version-hygiene.md#when-to-refresh-this-skill)
for when a newer upstream commit calls for another refresh.

Copied upstream documentation retains repository-relative links to omitted
files. Resolve Ghidra/IDA links against the
[pinned upstream tree](https://github.com/aers/FFXIVClientStructs/tree/80cb4a7b7e11a6702e2eb8839afd301a918c638e).
The full `ida/data.yml` and tools are not bundled.
