# NightmareUI and NightmareUI.OtterGuiWrapper

NightmareUI is NightmareXIV's ImGui component library for Dalamud plugins: a fluent settings-page
builder (`NuiBuilder`), a button tab bar (`NuiTools.ButtonTabs`), a world picker, a name censor and
a serializable UI-state object. It is 14 `.cs` files on top of ECommons' `ImGuiEx`.
`NightmareUI.OtterGuiWrapper` is a second repo (9 `.cs` files) that joins it to OtterGui's
`FileSystem<T>`. Both READMEs are one line: "This library is currently for internal use only!".
Consumers take both as git submodules.

This skill tracks:

| What | Pin |
|---|---|
| `NightmareXIV/NightmareUI` `main` | commit `aae510d` (2026-06-18), csproj `Version 1.0.0.0`, `Microsoft.NET.Sdk`, `net10.0-windows7.0`, `PackageReference ECommons 3.2.0.10` |
| `NightmareXIV/NightmareUI.OtterGuiWrapper` `main` | commit `371b1f9` (2026-05-30), same SDK/TFM, `PackageReference ECommons 3.0.0.2` |
| `Limiana/OtterGui` (the fork the wrapper is built against) | commit `a08fda6` (2026-05-30), `Dalamud.NET.Sdk/14.0.1`, Apache-2.0; read only, not mirrored |
| `NightmareXIV/ECommons` `master` | commit `bc38dc7` (2026-11-08), `BaseVersion 3.2.1.26`; read only to resolve symbols |
| `goatcorp/Dalamud` `master` | commit `43b3411` (2026-11-10), `DalamudVersion 15.0.3.9`, API 15; read only |
| Consumers | `PunishXIV/AutoRetainer` `cc020df` (2026-08-28), `PunishXIV/Splatoon` `f396ef6` (2026-11-13) |

Submodule pins in the consumers: AutoRetainer pins NightmareUI `aae510d`, the wrapper `371b1f9`,
OtterGui `a08fda6`, ECommons `e6be8f0`; Splatoon pins NightmareUI `aae510d`, ECommons `26d30b9`.
Both consumers pin exactly the commits mirrored here and declare `DalamudApiLevel 15`. GitHub
reports `pushed_at 2026-07-31` for both NightmareUI repos while `main` is older; each repo also
has branches `tw` and `tw13`, which are not covered here.

Pins last checked 2026-11-13: the two mirrored repos, the OtterGui fork and AutoRetainer had not moved;
ECommons, Dalamud and Splatoon were re-read at their new heads.

Everything named below was read from those checkouts. **Nothing was compiled or run in game.**
Statements about runtime behaviour are readings of the source; items that reading could not settle
are marked *(unverified)*. Verbatim sources: `references/NightmareUI/...` and
`references/NightmareUI.OtterGuiWrapper/...` (upstream paths preserved, list in
`references/source-files.txt`). To find a member: `grep -n Name references/all-public-members.txt`
(`path:line: declaration`, public and protected), or `references/all-internal-members.txt`. Never
invent a member; if it is not in the reference, say so.

Out of scope, one pointer each:

- ECommons and its `ImGuiEx` layer (`EzConfig`, `Svc`, `PluginLog` ...) → skill `ecommons`.
- OtterGui itself (`FileSystem<T>`, `ISortMode<T>`, the full selector, `ImRaii`) → skill `ottergui`.
- Stock Dalamud ImGui (`Dalamud.Bindings.ImGui`, `Window`, `ImRaii`, fonts) → skill `dalamudimgui`.
- Dalamud services and plugin lifecycle → skill `dalamud`.
- PunishLib (About tab, Puni.sh branding) → skill `punishlib`.


## Guide sections

Paths written as `references/...` are relative to the skill directory. Read only the section needed for the task.

- [topic files](guide-topic-files.md)
- [1 adding it to a plugin](guide-1-adding-it-to-a-plugin.md)
- [2 nightmareui public surface](guide-2-nightmareui-public-surface.md)
- [3 nightmareuiotterguiwrapper](guide-3-nightmareuiotterguiwrapper.md)
- [4 worked examples](guide-4-worked-examples.md)
- [5 pitfalls](guide-5-pitfalls.md)
