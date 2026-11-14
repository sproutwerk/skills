# ECommons

ECommons is a static-heavy helper library for Dalamud plugins: service locator, ImGui extension
layer, UI automation, attribute-driven IPC and hooks, configuration, throttling, game-state
shortcuts and several hundred extension methods. Repo `github.com/NightmareXIV/ECommons`, branch
`master`, MIT (Copyright (c) 2023 NightmareXIV). There is no documentation site and the README is
short and partly outdated, so the source is the API truth.

This skill tracks:

| What | Pin |
|---|---|
| `NightmareXIV/ECommons` `master` | commit `bc38dc7` (`bc38dc71340fa3b266de58e3d6d876a73313b61e`, 2026-11-08), `ECommons.csproj` `BaseVersion 3.2.1.26`, `net10.0-windows7.0` |
| NuGet `ECommons` | 3.2.1.26, published 2026-10-14; its nuspec names the earlier commit `26d30b9`, so the package lacks a few `master` members ([list](topic-setup.md#package-and-master-differences)) |
| Dalamud | API 15 (`goatcorp/Dalamud` `43b3411`, `DalamudVersion` 15.0.3.9, read on 2026-11-13) |
| Consumers read for usage | AutoRetainer `cc020df`, YesAlready `c531fb8`, PandorasBox `3a35634` |

Everything named here was read from that checkout unless marked *(unverified)*. **Nothing was
compiled**: examples are assembled from verified signatures. Verbatim source is under
`references/ecommons/<upstream path>` (408 files, list in `references/source-files.txt`). To find a
member that is not covered, grep `references/all-public-members.txt` (4996 lines,
`path:line: declaration`) or `references/all-public-types.txt` (646 type declarations), then read
the file. Never invent a member: ECommons renames things between versions, so a remembered name
may not exist at this pin.

Out of scope, one pointer each:

- Stock Dalamud services, plugin lifecycle, csproj and manifest, `IAddonLifecycle`, call gates →
  skill `dalamud`.
- Stock Dalamud ImGui: bindings, `Window` / `WindowSystem`, `ImRaii`, fonts, textures → skill
  `dalamudimgui`.
- FFXIVClientStructs structs, `AtkUnitBase`, `AtkValue`, agents → skill `clientstructs`.
- PunishLib (About tab, Puni.sh shared config) → skill `punishlib`.
- NightmareUI (`NuiBuilder`, the AutoRetainer-style settings pages) → skill `nightmareui`.
- Writing Splatoon scripts → skill `splatoonscripts`; Splatoon preset and Element semantics →
  skill `splatoonpresets`.


## Guide sections

Paths written as `references/...` are relative to the skill directory. Read only the section needed for the task.

- [topic files](guide-topic-files.md)
- [adding ecommons to a plugin](guide-adding-ecommons-to-a-plugin.md)
- [bootstrap](guide-bootstrap.md)
- [namespaces](guide-namespaces.md)
- [svc](guide-svc.md)
- [needed on nearly every task](guide-needed-on-nearly-every-task.md)
- [catalog](guide-catalog.md)
- [worked examples](guide-worked-examples.md)
- [pitfalls](guide-pitfalls.md)

## Changes since 3.2.1.21

The previous pin of this skill was `15f40a0` (3.2.1.21, 2026-09-28). A plugin on that version or
older lacks everything below. Items marked `master` are also absent from the 3.2.1.26 package
(`26d30b9`); see [package and `master` differences](topic-setup.md#package-and-master-differences).

- TaskManager stacks nest. `Stack` changed from `List<TaskManagerTask>` to `List<StackDetails>`,
  `IsStackActive` is computed, and the manual `BeginStack` / `EnqueueStack()` / `InsertStack()` /
  `DiscardStack()` calls are `[Obsolete]` in favour of the `Action` overloads, which now also
  take a stack name. See [stack mode](topic-taskmanager.md#stack-mode).
- `ClickHelper.InvokeReceiveEvent` calls the vtable entry through an unmanaged function pointer,
  which fixes an `InvalidCastException` with Dalamud's addon hooks. See
  [automation](topic-automation.md#clickhelper).
- `AddonMaster.SelectYesno.TextLegacy` is `[Obsolete]` (use `Text`); `ButtonsVisible` tests
  `AtkValueType.ConstString` where it tested `String8`. `AddonMaster.Bank` gained `InputNumeric`.
  `AddonMaster.Gathering.GatheredItem` gained `IsHidden`, `IsRare` and `GatherAmount` (`master`).
- Seven `ReaderXBM*` `AtkReader` classes and `ECommons.StringHelpers.DigitParsers` are new. See
  [AddonMaster](topic-addonmaster.md).
- `GenericHelpers` gained `TryGetValue` for lists and arrays, `SortBy` / `SortByDescending`,
  `SafeSelectNullable`, `TryGetLast` and, on `master`, `FirstOrAdd` and `ContainsNull`.
  `AnyNull` was removed on `master`. `TakeEvery(0)` now throws. See
  [generic helpers](topic-generic-helpers.md).
- `ImGuiEx.EnumCombo` overloads take an optional `T[]? valuesList` and moved to `EnumCombo.cs`.
  `ImGuiEx.Selectable` takes an optional `Vector2? alignment`. New: non-nullable
  `InputFloat(float width, ...)` and `ColorEdit4(ImU8String label, scoped ref uint col, ...)`.
- `ExcelWorldHelper.GetName` has `includeId` overloads. `PluginLog.Trace` and
  `PluginLog.TraceEnabled` are new. `MathHelpers.Number` is `[Obsolete]`.
- `EffectHeader.ActionType` is a computed property over the new `ActionTypeByte` field at offset
  31. `TaskManagerTask.Name` has a public setter.

These are source comparisons. Nothing was built or run in game.
