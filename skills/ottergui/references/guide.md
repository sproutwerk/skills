# OtterGui

OtterGui is a C# class library of ImGui widgets, RAII scopes, UTF-8 text helpers and general
utilities for Dalamud plugins, written by Ottermandias for Penumbra and Glamourer. It has **no
README, no docs and no NuGet package**: consumers add it as a git submodule and the source is the
only documentation. It builds against `Dalamud.Bindings.ImGui` (no `ImGuiNET`).

This skill tracks:

| What | Pin |
|---|---|
| `Limiana/OtterGui` branch `main` (**the tree documented here**) | commit `a08fda6` (2026-05-30); `OtterGui.csproj` `Sdk="Dalamud.NET.Sdk/14.0.1"`, `TargetFramework net10.0-windows7.0` |
| `Ottermandias/OtterGui` branch `main` (upstream, compared against) | commit `79771ee` (2026-04-29); `Sdk="Dalamud.NET.Sdk/15.0.0"` |
| `Ottermandias/OtterGuiInternal` (submodule of both) | fork pins `49631aa` (2025-08-07); upstream pins `fb9de85` (2026-02-24) |
| Artisan's OtterGui gitlink | `32c4454` (2026-09-01), head of `Taurenkey/OtterGui` branch `woops`, **not** on any branch of the Limiana fork |
| `PunishXIV/Artisan` `main` | `c8ee6e0` (2026-11-10), `DalamudApiLevel` 15 |
| `PunishXIV/AutoRetainer` `master` | `cc020df` (2026-08-28), `DalamudApiLevel` 15, pins OtterGui `a08fda6` |
| `NightmareXIV/NightmareUI.OtterGuiWrapper` | `371b1f9` (2026-05-30) |
| `goatcorp/Dalamud` (for the ImRaii comparison) | `43b3411` (2026-11-10), version 15.0.3.9, API 15 |

GitHub reports the fork's last push as 2026-07-31; that push went to branch `tw13`, not `main`.

Pins last checked 2026-11-13: fork `main`, upstream `main`, both OtterGuiInternal pins and both
consumer gitlinks were unchanged. Artisan and Dalamud had new commits; the consumer notes and the
`ImRaii` comparison were re-read at the heads listed above.

**Status.** Upstream is in maintenance mode: its author moved Penumbra and Glamourer to the
successor library Luna, and the last upstream commit is `79771ee` (2026-04-29). The Limiana fork
is *behind* upstream in code (it never merged upstream's last fifteen commits) and differs from it
in ten files plus the submodule pin. Expect fixes to arrive only when a consumer needs them.

**Licensing.** OtterGui is Apache-2.0 (`references/src/LICENSE`). OtterGuiInternal is Boost
Software License 1.0 (`references/src/OtterGuiInternal/LICENSE`). The mirror under `references/`
contains unmodified upstream source with its original notices.

**Verification.** Every type, member and signature named in this skill was read from the pinned
checkouts, which are mirrored verbatim under `references/src/` (253 files, list in
`references/manifest.txt`). Three files are indexed in `references/api-index.txt` but not bundled: `String/Utf8String.cs`, `Text/HelperObjects/SpanTextWriter.cs` and `Extensions/PowerSetExtensions.cs`; read them from the pinned checkout. **Nothing was compiled**: no Dalamud build environment was available,
so the examples are written against the source, not tested. Items marked *(unverified)* could not
be confirmed from the checkouts. Before using a member that is not shown here:

```sh
grep -n 'FilterComboCache' references/api-index.txt    # 2702 public/protected declarations, path:line: decl
grep -n ' class \| struct ' references/public-types.txt # 358 type declarations
```

Use the index to locate a member, then inspect the source. An absent index entry alone does not prove that a member is missing from the fork.


## Guide sections

Paths written as `references/...` are relative to the skill directory. Read only the section needed for the task.

- [sibling skills](guide-sibling-skills.md)
- [topic files](guide-topic-files.md)
- [setup in one screen](guide-setup-in-one-screen.md)
- [namespaces](guide-namespaces.md)
- [what you need on nearly every task](guide-what-you-need-on-nearly-every-task.md)
- [compact catalogue](guide-compact-catalogue.md)
- [how the popular plugins use it](guide-how-the-popular-plugins-use-it.md)
- [ottergui to luna](guide-ottergui-to-luna.md)
- [pitfalls](guide-pitfalls.md)
