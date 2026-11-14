OtterGui topic: what differs between the Limiana fork and upstream, the fork's branches, and what
Artisan and AutoRetainer really pin. Verified against `Limiana/OtterGui` `a08fda6` (2026-05-30,
branch `main`), upstream `Ottermandias/OtterGui` `79771ee` (2026-04-29) and the Artisan pin
`32c4454` (2026-09-01). Raw diffs: `references/upstream-to-fork.diff`,
`references/fork-to-artisan-pin.diff`.

Paths such as `references/...` and `references/topic-*.md` are relative to the skill root (the directory that
holds `SKILL.md`), not to this file. Nothing here was compiled; every statement is read from the
diffs and the checkouts.

## Which tree the skill documents

The skill documents **`Limiana/OtterGui` branch `main` at `a08fda6`**, the commit AutoRetainer
pins. Members that exist only in upstream or only in the Artisan pin are listed below and flagged
wherever they are mentioned elsewhere in the skill.

## History in one paragraph

Fork `main` and upstream share the merge base `3ea6164` ("Add overload to dynamis."). After that
point the fork has four commits by Limiana (`4106af6` 7.4, `2fc7cee` 7.4, `76271df` 7.5, `a08fda6`
Update) that only touch `OtterGui.csproj`, `packages.lock.json` and `Custom/IndividualHelpers.cs`.
Upstream has fifteen commits the fork never merged (Dynamis API 1.6, `NotificationActions`,
auto-leaves, clipping and filtered-combo fixes, SDK bumps). So the fork is **behind** upstream in
code and differs mainly by omission.

## Fork `main` versus upstream `79771ee`

Eleven paths differ: ten files (36 insertions, 118 deletions going from upstream to fork) plus
the `OtterGuiInternal` submodule pin. No file is added or removed.

| File | Upstream `79771ee` | Fork `a08fda6` |
|---|---|---|
| `OtterGui.csproj` | `Sdk="Dalamud.NET.Sdk/15.0.0"` | `Sdk="Dalamud.NET.Sdk/14.0.1"`, plus `<TargetFramework>net10.0-windows7.0</TargetFramework>` and `<ContinuousIntegrationBuild>true</ContinuousIntegrationBuild>` |
| `OtterGuiInternal` (submodule pin) | `fb9de85` (2026-02-24) | `49631aa` (2025-08-07) |
| `Services/DynamisIpc.cs` | Dynamis 1.6+, see below | Dynamis 1.3+, see below |
| `Classes/MessageService.cs` | `IMessage.OnNotificationActions(INotificationDrawArgs)`; `Notification` implements it; `AddMessage` subscribes it to `IActiveNotification.DrawActions` | member absent; `AddMessage` discards the notification object |
| `Filesystem/FileSystem.File.cs` | `Load(JObject?, ...)` adds a leaf at `Root` for every object when no saved file exists | that `else` branch is absent: with no saved file the tree stays empty |
| `Widgets/FilteredCombo.cs` | `UpdateFilter` guards `_available.Count <= _lastSelection` | no bounds guard (`_available[_lastSelection]` can throw) |
| `ImGuiClip.cs` | `ClippedTableDraw` wraps each `draw` in `ImUtf8.Group()` | no group |
| `Custom/IndividualHelpers.cs` | `ToName` maps `ObjectKind.Mount => "Mount"` | that arm is deleted (falls through to `kind.ToString()`) |
| `Text/ImUtf8.RotatedText.cs` | | two extra unused locals (`startVertex`, `endVertex`) |
| `Text/ImUtf8.SliderAngle.cs`, `Text/HelperObjects/ImTextExtensions.cs` | | comment and initializer wording only |

`packages.lock.json` is identical in both.

### DynamisIpc members

| Member | Upstream | Fork |
|---|---|---|
| Minimum Dynamis version accepted | 1.6 | 1.3 |
| `InspectObject` | `(nint address, string? name = null)`, IPC `Dynamis.InspectObject.V2` | `(nint address)`, IPC `Dynamis.InspectObject.V1` |
| `InspectRegion` | `(nint, uint size, string typeName, uint typeTemplateId, uint classKindId, string? name = null)`, `.V2` | same without `name`, `.V1` |
| `OpenContextMenu(nint address, Func<string?>? name = null)` | present | **absent** |
| `DrawPointer` | `(nint address, Func<string?>? name = null, string? customText = null, DrawPointerFlags flags = None, ImGuiSelectableFlags selectableFlags = None, Vector2 size = default)`, delegate `Dynamis.GetImGuiDrawPointerDelegate.V3` | `(nint address)` and `(void* address)`, delegate `.V1` |
| `enum DrawPointerFlags` (`None`, `Selected`, `MonoFont`, `Semitransparent`, `RightAligned`) | present | **absent** |
| `DrawTooltipDetails`, `GetClass`, `IsInstanceOf`, `DrawDebugInfo`, `IsSubscribed`, `Features`, `VersionMajor`, `VersionMinor`, `Error` | same | same |

### OtterGuiInternal pin

Upstream's newer pin `fb9de85` changes one thing in `ImGuiInternal.RenderTextClipped`
(`ReadOnlySpan<char>` overloads): the UTF-8 scratch buffer is sized `visibleEnd * 4` instead of
`visibleEnd * 2`. At the fork's pin `49631aa` the buffer is `visibleEnd * 2` bytes, which is too
small for text whose characters need three UTF-8 bytes each (CJK, for instance). Inside the fork
the only caller is `Widgets/ToggleButton.cs` (`ToggleButton.ButtonEx`).

## Branches of the fork

| Branch | Head | Date | Notes |
|---|---|---|---|
| `main` (default) | `a08fda6` | 2026-05-30 | documented here; AutoRetainer pins it |
| `Artisan` | `26b38ec` | 2025-12-18 | `Dalamud.NET.Sdk/13.0.0`; carries the `ItemSelector` / `Table` changes Artisan needs |
| `tw13` | `2adc617` | 2026-07-31 | one commit on top of `main`; this is the push that sets the repo's `pushed_at` |
| `tw` | `c077e4b` | 2026-03-15 | not inspected |
| `old-stuff` | `2da4680` | 2025-06-26 | not inspected |

`tw13` versus `main` (4 files): `OtterGui.csproj` is rewritten to `Sdk="Microsoft.NET.Sdk"` with
`<TargetFramework>net9.0-windows7.0</TargetFramework>` and explicit `$(DalamudLibPath)`
`<Reference>` entries; `IService` loses its `[UsedImplicitly]` attribute; `IndividualHelpers.ToName`
gets `ObjectKind.MountType => "Mount"`; `packages.lock.json` changes. What `tw13` is for is
*(unverified)*; nothing in the repository says.

## What the consumers pin

| Plugin | Commit of plugin | `.gitmodules` URL | Gitlink | Where that commit lives |
|---|---|---|---|---|
| AutoRetainer (`PunishXIV/AutoRetainer` `master`) | `cc020df` (2026-08-28) | `https://github.com/Limiana/OtterGui.git` | `a08fda6` | head of fork `main` |
| Artisan (`PunishXIV/Artisan` `main`) | `c8ee6e0` (2026-11-10) | `https://github.com/Limiana/OtterGui` | `32c4454` | **not on any branch of the fork**; it is the head of branch `woops` in `Taurenkey/OtterGui` |

Artisan's gitlink resolves through the Limiana URL only because GitHub serves objects across a
fork network (`git fetch https://github.com/Limiana/OtterGui 32c4454...` succeeds, and
`git branch -r --contains` on the fork's branches returns nothing). `32c4454` descends from
upstream `79771ee` (merge `031931c`) and from the fork's `Artisan` branch.

AutoRetainer also pins `NightmareXIV/NightmareUI.OtterGuiWrapper` at `371b1f9` (2026-05-30) and
`NightmareXIV/NightmareUI` at `aae510d`.

## Artisan pin `32c4454` versus upstream `79771ee`

Five files (70 insertions, 52 deletions); everything else equals upstream, so the Artisan pin
also has every upstream-only member listed above.

| File | Change in the Artisan pin |
|---|---|
| `OtterGui.csproj` | `Dalamud.NET.Sdk/15.0.0` kept, `<TargetFramework>net10.0-windows7.0</TargetFramework>` added |
| `OtterGuiInternal` | pin moved back to `49631aa` |
| `ItemSelector.cs` | see table below |
| `Table/Table.cs` | `FilteredItems` is `public List<(T, int)>` (fork: `internal readonly`); `Headers` is `public Column<T>[]` (fork: `protected readonly`) |
| `Table/ColumnString.cs` | adds `public virtual event EventHandler<TItem>? OnContextMenuRequest` and `public void InvokeContextMenu(TItem e)` |

`ItemSelector<T>` members:

| Member | Fork `main` and upstream | Artisan pin |
|---|---|---|
| `OnDraw` | `protected virtual bool OnDraw(int idx)` | `protected virtual bool OnDraw(int idx, out bool changes)` |
| `Items` | `protected readonly IList<T>` | `public IList<T>` |
| flags field | `private readonly Flags _flags` | `public Flags ListFlags` |
| `SetCurrent(int idx)` | `private` | `public` |
| `Current` | `{ get; private set; }` | `{ get; set; }` |
| events | none | `ItemAdded`, `ItemDeleted`, `ItemSkipTriggered` (all `EventHandler<bool>`) |
| `AddButtonTooltip()` | none (tooltip is the literal `"Add New"`) | `protected virtual string`, default `"Add new item."` |
| add with empty name | calls `OnAdd("")` | skipped (`!string.IsNullOrEmpty(newName)`) |
| clip row height | `ImGui.GetTextLineHeightWithSpacing()` | same `+ 0.25f` |

In the Artisan pin `ItemSkipTriggered.Invoke(this, changes)` is called without a null check when
`OnDraw` sets `changes` to true, so a subscriber must exist before that happens.

Consequence: code copied from Artisan (`OnDraw(int idx, out bool changes)`, assigning
`this.Headers`, reading `Table.FilteredItems` from outside, `AddButtonTooltip`) does **not**
compile against fork `main` or upstream. Code written for fork `main` that overrides
`OnDraw(int idx)` does not compile against the Artisan pin.

## Choosing a commit for a new plugin

| Goal | Pin |
|---|---|
| Same as AutoRetainer, works with `NightmareUI.OtterGuiWrapper` `371b1f9` | fork `main` `a08fda6` |
| Artisan-style `ItemSelector` events and public `Table` internals | `32c4454` (fetch it by hash; no branch in the fork points at it) |
| Newest library code, SDK 15 csproj, Dynamis 1.6 | upstream `79771ee` |
