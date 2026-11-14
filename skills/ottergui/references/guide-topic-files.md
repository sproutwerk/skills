## Topic files

Deep material lives in `references/topic-*.md` (first-party, same pins). Open the file when its trigger applies;
do not answer those questions from the summaries in this file alone.

| File | Read this when ... |
|---|---|
| `references/topic-fork-delta.md` | code from Artisan, Penumbra, Glamourer or upstream does not compile against the OtterGui commit in use; choosing which commit or branch to pin; you need to know whether a member exists in the fork, upstream or the Artisan pin; anything about `DynamisIpc`, `OnNotificationActions`, the fork's branches |
| `references/topic-setup.md` | adding the submodule and `ProjectReference`; a build fails on `OtterGuiInternal`, the SDK version, the target framework or `packages.lock.json`; preparing for or recovering from a Dalamud API bump |
| `references/topic-imutf8.md` | writing or reading any `ImUtf8.*` call; choosing between `u8`, `string` and interpolated overloads; an `ImUtf8FormatException`; you need an `EndObjects` type's members or the signature of an `Extended` helper (`IconButton`, `HoverTooltip`, `InputTextOnDeactivated`, `ResizableChild`, ...) |
| `references/topic-raii.md` | using `OtterGui.Raii.ImRaii`; `ImRaii` is ambiguous (CS0104); deciding between OtterGui's and Dalamud's `ImRaii`; `DefaultColors` / `DefaultStyle` / `Enabled` do not revert what you expected |
| `references/topic-table.md` | building or debugging a `Table<T>`: column types, filter and sort behaviour, a complete example, `ImGuiTable.DrawTable` |
| `references/topic-filesystem.md` | anything with `FileSystem<T>`, `Folder`, `Leaf`, `ISortMode<T>`, saving/loading the tree, or `FileSystemSelector` (context menus, buttons, filters, custom row drawing); the `Filesystem` / `FileSystem` namespace confusion |
| `references/topic-widgets.md` | you need the constructor or draw signature of any widget: `FilterComboBase` / `FilterComboCache` / `FilterComboColors`, `ClippedSelectableCombo`, `Tutorial`, `Changelog`, `TagButtons`, `TabBar` / `ITab`, `MultiStateCheckbox` family, `IEditor` family, `Widget.*` helpers, `ToggleButton` |
| `references/topic-helpers.md` | writing an `ItemSelector<T>` subclass; clipping a long list with `ImGuiClip`; any `ImGuiUtil.*` or `Functions.*` call; `ImGuiPie`; the `Custom/` buttons |
| `references/topic-services-classes.md` | using `ServiceManager`, `IService`, `HookManager` / `FastHook`, `ImGuiCacheService`, `Logger`, `EventWrapper`, `SaveServiceBase`, `MessageService`, `FrameworkManager`, hotkey types, `LowerString`, containers, `Tasks/`, `Compression/`, `String/`, `Extensions/`, or writing a custom widget with `OtterGuiInternal` |
| `references/topic-usage-in-the-wild.md` | deciding which part of OtterGui fits a puni.sh-style plugin; reading or imitating Artisan's selectors and tables or AutoRetainer's settings tree; a file-by-file list of what the consumers use |

