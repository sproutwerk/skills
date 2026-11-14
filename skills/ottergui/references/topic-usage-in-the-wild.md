OtterGui topic: which parts of OtterGui Artisan and AutoRetainer actually use, file by file, and
what that implies for writing similar plugins. Verified against `PunishXIV/Artisan` `c8ee6e0`
(2026-11-10, branch `main`), `PunishXIV/AutoRetainer` `cc020df` (2026-08-28, branch `master`),
`NightmareXIV/NightmareUI.OtterGuiWrapper` `371b1f9` (2026-05-30) and `Limiana/OtterGui`
`a08fda6` (2026-05-30).

Paths such as `references/...` and `references/topic-*.md` are relative to the skill root (the directory that
holds `SKILL.md`), not to this file. Consumer paths are relative to the consumer repository. The
consumer sources are not mirrored in `references/`; clone them to read more.

## Scan context

A scan of the 90 plugins listed on puni.sh on 2026-09-29 found OtterGui in four: Artisan
(submodule URL `Limiana/OtterGui`), AutoRetainer (`Limiana/OtterGui` plus
`NightmareXIV/NightmareUI.OtterGuiWrapper`), Ice's Cosmic Exploration
(`LeontopodiumNivale14/OtterGui`) and AetherLove (`Ottermandias/OtterGui`). Only Artisan and
AutoRetainer were cloned and read here. The other two plugins and the `LeontopodiumNivale14` fork
are *(unverified)*.

## Summary

| Part of OtterGui | Artisan | AutoRetainer (direct) | NightmareUI.OtterGuiWrapper |
|---|---|---|---|
| `ItemSelector<T>`, `ItemDetailsWindow` | yes (3 selectors) | | |
| `Table<T>`, `ColumnString`, `ColumnFlags` | yes (2 tables) | | |
| `OtterGui.Raii.ImRaii` | yes (2 files) | | yes (5 files) |
| `ImGuiUtil` | yes (`Center`, `HoverIcon`, `HoverTooltip`) | | yes |
| `OtterGui.Filesystem` list extensions (`Move`) | yes | | |
| `FileSystem<T>` | | | yes (both file systems derive from it) |
| `FileSystemSelector<T, TStateStorage>` | | | yes (`GenericFileSystem`), plus its own simplified copy |
| `OtterGui.Classes` | `TextureCache` | | yes (2 files) |
| `OtterGui.Log.Logger` | | | yes (1 file) |
| `OtterGui.Text` (`ImUtf8`) | | | yes (1 file) |
| `OtterGui.Extensions` | imported in 1 file | | |
| `Services/`, `Tasks/`, `Compression/`, `String/`, `Widgets/` | not imported | not imported | not imported |

AutoRetainer's own code contains no active `using OtterGui...` line. The only occurrence is a
commented-out `//using OtterGui.Text.EndObjects;` in `AutoRetainer/Helpers/Utils.cs`. It reaches
OtterGui solely through the wrapper, although `AutoRetainer.csproj` references
`..\OtterGui\OtterGui.csproj` directly as well.

Practical reading: in the puni.sh ecosystem OtterGui is a **widget donor** for list selectors,
filterable tables and the folder tree. Everything else in those plugins (services, config, hooks,
most ImGui helpers) comes from ECommons and stock Dalamud. Do not introduce
`ServiceManager`, `SaveServiceBase` or `MessageService` into such a plugin just because OtterGui
is already referenced.

## Artisan

`using` lines across `Artisan/Artisan/` (count of files): `OtterGui` 5, `OtterGui.Filesystem` 3,
`OtterGui.Table` 2, `OtterGui.Raii` 2, `OtterGui.Classes` 2, `OtterGui.Extensions` 1.

| File | What it uses |
|---|---|
| `UI/ListEditor.cs` | `RecipeSelector : ItemSelector<ListItem>` (flags `Add \| Delete \| Move`), `ListFolders : ItemSelector<NewCraftingList>`, `ItemDetailsWindow`, `ImRaii.Popup`, `PushColor`, `PushId`, `PushStyle` from `OtterGui.Raii` |
| `UI/ExpertProfilesUI.cs` | `ExpertProfileList : ItemSelector<ExpertProfile>` (flags `Add \| Delete \| Move \| Filter \| Duplicate`); `ImRaii` here is Dalamud's (`using Dalamud.Interface.Utility.Raii`) |
| `UI/ExpertProfileEditor.cs` | imports `OtterGui`, `OtterGui.Filesystem` |
| `UI/Tables/IngredientTable.cs` | `IngredientTable : Table<Ingredient>`; nine `ColumnString<Ingredient>` subclasses and `ItemFilterColumn : ColumnFlags<ItemFilter, Ingredient>`; `ImGuiUtil.Center`, `ImGuiUtil.HoverIcon`, `ImGuiUtil.HoverTooltip`; `OtterGui.Raii` |
| `UI/Tables/RaphaelCacheTable.cs` | `RaphaelCacheTable : Table<RaphaelOptions>`, `ClickableColumn : ColumnString<RaphaelOptions>` |
| `UI/RaphaelCacheUI.cs` | calls `Table.Draw(ImGui.GetTextLineHeightWithSpacing() - 4f)`, reads `Table.FilteredItems` |
| `Artisan.cs`, `Extensions.cs` | `OtterGui.Classes.TextureCache` (`internal TextureCache Icons;`, and a C# extension block on `TextureCache`) |
| `UI/RecipeWindowUI.cs` | imports `OtterGui.Extensions`; `ImRaii.PushFont` there is Dalamud's |
| `RawInformation/ExtendedIngredient.cs` | imports `OtterGui` |

Artisan depends on members that only exist in its pinned commit `32c4454` (and on the fork's
`Artisan` branch), not on fork `main`:

| Artisan code | Member needed |
|---|---|
| `protected override bool OnDraw(int idx, out bool changes)` | changed `OnDraw` signature |
| `protected override string AddButtonTooltip()` | new virtual |
| `RecipeSelector.ItemAdded += RefreshTable;` and `ItemDeleted`, `ItemSkipTriggered` | events |
| `this.ListFlags = Flags.Move;` | public `ListFlags` |
| `RecipeSelector.Items = SelectedList.Recipes.Distinct().ToList();` | public, writable `Items` |
| `RecipeSelector.SetCurrent(newIdx);`, `RecipeSelector.Current = ...` | public `SetCurrent(int)`, public setter |
| `this.Headers = [.. headers];` | writable `Table<T>.Headers` |
| `Table.FilteredItems.Count`, `.Clear()` from another class | public `FilteredItems` |
| `_nameColumn.OnContextMenuRequest += OpenContextMenu;`, `InvokeContextMenu(item)` | `ColumnString` additions |

See `references/topic-fork-delta.md` for the exact declarations. When a user shows Artisan code and asks
why it does not compile in their plugin, check which OtterGui commit their submodule points at
first.

Patterns worth copying from Artisan (they work on every tree):

- One selector class per list, flags chosen per list, `Filtered` returning `false` when no filter
  is wanted (`RecipeSelector`).
- Table columns as nested classes of the table, labels set with object initializers
  (`new() { Label = "Progress" }`), widths recomputed in `PreDraw()` only when
  `ImGuiHelpers.GlobalScale` changed, dividing by the scale and adding `Table.ArrowWidth`.
- Extra table flags in the constructor:
  `Flags |= ImGuiTableFlags.Hideable | ImGuiTableFlags.Reorderable | ImGuiTableFlags.Resizable;`
- Mixing: Artisan uses ECommons' `ImGuiEx` and stock `ImGui` calls inside `OnDraw` /
  `DrawColumn`. The OtterGui widgets only own the frame around your drawing code.

## AutoRetainer and NightmareUI.OtterGuiWrapper

AutoRetainer's settings window (`UI/NeoUI/!Definitions/NeoWindow.cs`) holds
`internal ConfigFileSystem FileSystem;`, builds it with `new(() => Tabs)`, calls
`FileSystem.Reload()` and draws with `FileSystem.Draw(null)`. Its pages derive from
`NeoUIEntry : ConfigFileSystemEntry`. All three types are from
`NightmareUI.OtterGuiWrapper.FileSystems.Configuration`.

Inside the wrapper (`NightmareUI.OtterGuiWrapper/FileSystems/`, 9 files, 1475 lines):

| File | OtterGui usage |
|---|---|
| `Configuration/ConfigFileSystem.cs` | `sealed class ConfigFileSystem : FileSystem<ConfigFileSystemEntry>`; nested selector derives from the wrapper's simplified selector |
| `Configuration/SimplifiedSelector/FileSystemSelector*.cs` (4 files) | a reduced copy of OtterGui's `FileSystemSelector<T, TStateStorage>` in namespace `NightmareUI.OtterGuiWrapper.FileSystems.Configuration.SimplifiedSelector` |
| `Generic/GenericFileSystem.cs` | `sealed class GenericFileSystem<TData> : FileSystem<TData> where TData : class, IFileSystemStorage, new()`; nested `FileSystemSelector : FileSystemSelector<TData, FileSystemSelector.State>` from `OtterGui.FileSystem.Selector`; persistence through the protected `Load` / `SaveToFile` |

`using` lines in the wrapper: `OtterGui.Filesystem` 7, `OtterGui.Raii` 5, `OtterGui` 4,
`OtterGui.Classes` 2, `OtterGui.FileSystem.Selector` 1, `OtterGui.Log` 1, `OtterGui.Text` 1.

The wrapper's csproj references `..\..\OtterGui\OtterGui.csproj`, i.e. it expects the OtterGui
submodule to sit **beside** the wrapper's own directory in the consuming repository. A plugin
that adds the wrapper must therefore add OtterGui at `<repo>/OtterGui` too. The wrapper's API is
documented in the `nightmareui` skill.

## Decision guide

| You want | Use |
|---|---|
| A list of user-defined entries with add / delete / duplicate / reorder | `ItemSelector<T>` (`references/topic-helpers.md`) |
| A sortable, per-column filterable table over a collection | `Table<T>` (`references/topic-table.md`) |
| A settings window with a tree of pages like AutoRetainer's | `ConfigFileSystem` from NightmareUI.OtterGuiWrapper (`nightmareui` skill) |
| Users arranging their own objects in folders | `FileSystem<T>` + `FileSystemSelector` (`references/topic-filesystem.md`), or the wrapper's `GenericFileSystem<TData>` |
| A searchable combo over hundreds of entries | `FilterComboCache<T>` (`references/topic-widgets.md`) |
| Scoped style/colour pushes only | Dalamud's own `ImRaii` (`dalamudimgui` skill); OtterGui is not needed |
