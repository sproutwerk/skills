# PunishLib

PunishLib is a 19-file C# class library (1,166 lines) published by Puni.sh at
`github.com/PunishXIV/PunishLib`. It gives Puni.sh plugins a common branded About tab, a Puni.sh
API-key dialog, a few ImGui widgets and a copy of an older ECommons configuration helper. It is
not on NuGet; plugins embed it as a git submodule and reference the project.

This skill tracks:

| What | Pin |
|---|---|
| `PunishXIV/PunishLib` branch `master` | commit `96c0cef` (2026-05-30), csproj `Version 2.0.0.1`, `TargetFramework net10.0-windows7.0`, `PackageReference ECommons 3.2.0.11` |
| `NightmareXIV/ECommons` (to resolve what PunishLib calls) | commit `bc38dc7` (2026-11-08), csproj `BaseVersion 3.2.1.26` |
| Dalamud (to resolve Dalamud symbols) | `goatcorp/Dalamud` `985a56f`, API 15 |
| Consumers read for the examples | AutoHook `467cc6a`, Artisan `c8ee6e0`, Saucy `427f7c6`, AutoRetainer `cc020df`, PandorasBox `3a35634`, YesAlready `c531fb8` |

Reviewed 2026-11-13: `master` HEAD is still the 2026-05-30 commit, and the repository has no
tags. The newer push GitHub reports (2026-07-31) went to the `tw13` branch; the
other branches are `API9` and `tw`. Only `master` is covered.

Everything named below was read from those checkouts. Nothing was compiled or run in game;
statements about runtime behaviour are derived from reading the source, and the few that could
not be settled that way are marked *(unverified)*. The whole library is mirrored verbatim under
`references/PunishLib/` (23 files, list in `references/source-files.txt`). To look something up:
`grep -n 'Name' references/all-public-members.txt` (87 lines, `path:line: declaration`) and
`references/all-internal-members.txt` (22 lines). Never invent a member; if it is not in the
reference, it does not exist in this pin.

The index matches the `public` keyword textually, so it also lists things a plugin cannot call:
members of the internal classes `API` and `ImGuiEx`, a private nested class (`AboutTab.cs:245`)
and the commented-out `Motd.PrintOnce`.

Out of scope, one pointer each:

- ECommons (`ECommonsMain`, `Svc`, `ECommons.ImGuiMethods.ImGuiEx`, `Notify`,
  `ECommons.Configuration.EzConfig`, `ThreadLoadImageHandler`) → sibling skill `ecommons`.
- Stock Dalamud services, plugin lifecycle, DataShare, manifests → sibling skill `dalamud`.
- Stock Dalamud ImGui (`Dalamud.Bindings.ImGui`, `ImRaii`, `ImGuiHelpers`, `ImGuiComponents`,
  windows, fonts, textures) → sibling skill `dalamudimgui`.
- FFXIVClientStructs → sibling skill `clientstructs`. PunishLib references the assembly in its
  csproj but no source file uses it.


## Guide sections

Paths written as `references/...` are relative to the skill directory. Read only the section needed for the task.

- [adding punishlib to a plugin](guide-adding-punishlib-to-a-plugin.md)
- [initialization and disposal](guide-initialization-and-disposal.md)
- [public api by folder](guide-public-api-by-folder.md)
- [branding specific versus generally useful](guide-branding-specific-versus-generally-useful.md)
- [worked examples](guide-worked-examples.md)
- [pitfalls](guide-pitfalls.md)
