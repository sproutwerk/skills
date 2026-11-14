# License and source notices

Original `SKILL.md`, `references/guide.md`, `references/guide-*.md`, the Markdown files under `references/topic-*.md`, and `README.md` are Copyright (c) 2026 sproutwerk, licensed under MIT. The full text is in `LICENSE` beside this file.

Everything under `references/ecommons/` is an unmodified copy of upstream ECommons source and keeps its own copyright notices. Most of it is covered by the ECommons license; some files carry a different notice in their header or credit another project:

| Files | Source | License | License text |
| --- | --- | --- | --- |
| `references/ecommons/` (except the rows below and the files in the next section) | [NightmareXIV/ECommons](https://github.com/NightmareXIV/ECommons) at `bc38dc71340fa3b266de58e3d6d876a73313b61e` | MIT | `references/ecommons/LICENSE.md` |
| `references/ecommons/ECommons/Hooks/ActionEffectTypes/ActionEffectType.cs`, `references/ecommons/ECommons/Hooks/ActionEffectTypes/AttackType.cs` | Header: "authored by lmcintyre and distributed under GNU GPL v3 license" (<https://github.com/lmcintyre/>); copied here as part of ECommons | GPL-3.0-only | `LICENSES/GPL-3.0-only.txt` |
| `references/ecommons/ECommons/MathHelpers/BitHelper.cs` | Header: licensed by the .NET Foundation ([dotnet/runtime](https://github.com/dotnet/runtime)); copied here as part of ECommons | MIT | `LICENSES/MIT-runtime.txt` |
| Parts of `references/ecommons/ECommons/GameHelpers/VfxManager.cs` marked "From VFXEditor" | [0ceal0t/Dalamud-VFXEditor](https://github.com/0ceal0t/Dalamud-VFXEditor) at `c320f08c981f3bf7353157c3d2faebcb00cba511`; copied here as part of ECommons | MIT | `LICENSES/MIT-Dalamud-VFXEditor.txt` |
| Part of `references/ecommons/ECommons/ImGuiMethods/ImGuiEx/ExcelCombos.cs` marked "from koenari" | [Koenari/HimbeertoniRaidTool](https://github.com/Koenari/HimbeertoniRaidTool) at `b28313e6d62de940acc073f203e3032e846bfb13`; copied here as part of ECommons | MIT | `LICENSES/MIT-HimbeertoniRaidTool.txt` |
| `references/all-files.txt`, `references/all-public-members.txt`, `references/all-public-types.txt`, `references/source-files.txt` | Generated from the ECommons source at the same commit; they list its file paths and public declarations | Same terms as the files they index | See the rows above |

The GPL text is the SPDX copy, because the header names no repository. The three MIT texts are the upstream projects' own license files; the .NET one was taken from `dotnet/runtime` at `4cc1f5a863b20a3f06b7d878f256b68c4259cece`, since the mirrored file records no revision.

## Not covered by these licenses

Upstream ECommons distributes the following files under its MIT license, but their own headers say something else or nothing. Their redistribution terms are not settled. The files are bundled unchanged.

| File | What the file says | What the cited source declares |
| --- | --- | --- |
| `references/ecommons/ECommons/Interop/OpenFileName.cs` | "Copyright Microsoft Corporation All rights reserved"; no license grant | No source is cited |
| `references/ecommons/ECommons/ExcelServices/Sheets/ExtendedBaseParam.cs` | "This code is authored by Caraxi", with a link to SimpleTweaksPlugin; no license stated | [Caraxi/SimpleTweaksPlugin](https://github.com/Caraxi/SimpleTweaksPlugin) declares AGPL-3.0 |
| `references/ecommons/ECommons/Automation/Chat.cs` | "from" ChatTwo `ChatBox.cs` at `5b58513a5226a2d3b40aca96be4f51c779fab363`; no license stated | [Infiziert90/ChatTwo](https://github.com/Infiziert90/ChatTwo) declares EUPL-1.2 at that commit |

ECommons is the upstream project's name. This independently maintained reference skill is not an official ECommons release or endorsement.
