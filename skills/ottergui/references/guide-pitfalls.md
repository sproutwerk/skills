## Pitfalls

**Commit and fork**

- **Artisan's OtterGui is not the fork's `main`.** Its gitlink `32c4454` lives in
  `Taurenkey/OtterGui`. `OnDraw(int idx, out bool changes)`, `AddButtonTooltip()`, the
  `ItemAdded` / `ItemDeleted` / `ItemSkipTriggered` events, public `Items` / `ListFlags` /
  `SetCurrent(int)`, writable `Table<T>.Headers`, public `FilteredItems` and
  `ColumnString.OnContextMenuRequest` exist only there (and on the fork's `Artisan` branch).
- **Upstream-only members** (absent in the fork): `DynamisIpc.OpenContextMenu`,
  `DynamisIpc.DrawPointerFlags`, the wide `DrawPointer` / `InspectObject` / `InspectRegion`
  overloads, `MessageService.IMessage.OnNotificationActions`. Penumbra and Glamourer code may
  use them.
- **The fork lacks three upstream fixes**: `FilterComboBase.UpdateFilter` can index out of range
  after mouse-wheel scrolling followed by opening the combo; `FileSystem<T>.Load` leaves the tree
  empty when no saved file exists; `ImGuiClip.ClippedTableDraw` does not group each row.
- **`RenderTextClipped` buffer**: at the fork's OtterGuiInternal pin the `ReadOnlySpan<char>`
  overloads allocate `2 * length` bytes for UTF-8; three-byte characters (CJK) overflow it.
  Inside the library only `ToggleButton.ButtonEx` calls it.
- GitHub serves commits across a fork network, so a submodule URL does not tell you which fork a
  pinned commit belongs to. Check with `git branch -r --contains <hash>`.

**Build**

- Clone and update submodules **recursively**; OtterGuiInternal is compiled from source files
  inside the OtterGui project directory.
- The fork's csproj names `Dalamud.NET.Sdk/14.0.1` while its consumers target API 15. Whether
  that combination needs a change on your machine is *(unverified)*; upstream uses `15.0.0`.
- `NightmareUI.OtterGuiWrapper` references `..\..\OtterGui\OtterGui.csproj`: OtterGui must sit at
  the repository root next to the wrapper's directory.
- A bump of the Dalamud API level can break OtterGui through renamed cloned enum members
  (`ObjectKind.MountType` became `ObjectKind.Mount`) and through the ImGui binding internals.

**Names**

- `OtterGui.Raii.ImRaii` and `Dalamud.Interface.Utility.Raii.ImRaii` clash (CS0104) when both
  namespaces are imported. Alias one: `using ImRaii = OtterGui.Raii.ImRaii;`.
- `OtterGui.Filesystem` (tree, lower-case `s`) versus `OtterGui.FileSystem.Selector` (selector,
  upper-case `S`).
- `Changelog` and `ChangeLogDisplayType` exist in both `OtterGui.Widgets` and
  `OtterGui.Text.Widget`. `TristateCheckbox` (`sbyte`) and `TriStateCheckbox` (`bool?`) are
  different classes.
- `ImUtf8.ColorPicker` has two unrelated families: `(label, ref color, flags)` and the
  setter-style `(label, tooltip, Vector3 input, Action<Vector3> setter, letter)`.

**Behaviour**

- **Filter polarity differs.** `ItemSelector<T>.Filtered(idx)` and
  `FileSystemSelector.ApplyFilters(path)` return `true` to **hide**. `Column<T>.FilterFunc(item)`
  and `FilterComboBase<T>.IsVisible(...)` return `true` to **show**.
- `ItemSelector<T>` hooks default to `throw new NotImplementedException()`. `OnDraw` and
  `Filtered` are always called; the others when their flag is set.
- `ItemSelector<T>` and `Table<T>` never modify your collection. Your hooks do, and for
  `Table<T>` you must set `FilterDirty = true` afterwards (`protected`; expose it).
- `ColumnSelect<T, TItem>` draws a filter combo but implements neither `FilterFunc` nor
  `DrawColumn`. `ColumnFlags<T, TItem>` stores no filter value; override `FilterValue`,
  `SetValue` and `FilterFunc`.
- `ColumnString` compiles the filter text as a regular expression first. Text such as `(` or
  `[` is an invalid pattern and silently falls back to substring search; text such as `a.c` is a
  valid pattern and matches more than the literal.
- Fixed row height everywhere: `ImGuiClip`, `ItemSelector<T>` (text line height with spacing),
  `Table<T>.Draw(itemHeight)`, `FilterComboBase<T>.Draw(..., itemHeight)`. Taller rows break
  scrolling.
- `ImGuiClip.ClippedDraw(IEnumerable<T>, ...)` returns the remainder **or** the bitwise
  complement of the next start index. Do not treat a negative result as an error.
- `Tutorial` step ids are list positions. Removing a registered step shifts every later id;
  use `Deprecated()` as placeholder. Call `Open` directly after the widget it explains.
- `TabBar` / `ITab.Label` is passed to ImGui through a pointer and must be null-terminated:
  return a `u8` literal.
- `FileSystem<T>` requires `T : class`, names may not contain `/` (replaced by `\`), and
  `Folder.Children` is internal. Its public methods throw `System.Exception` with a message.
- Modifying a `FileSystem<T>` from inside the selector's draw callbacks (context menu entries,
  buttons) must go through `EnqueueFsAction`.
- `FileSystemSelector` takes an `IKeyState` but keyboard navigation is commented out in the
  source.
- `OtterGui.Raii.ImRaii.DefaultColors()` / `DefaultStyle()` / `DefaultFont()` / `Enabled()` only
  see pushes made through OtterGui's `ImRaii`, not Dalamud's and not raw ImGui calls.
- `ImRaii.Style.Push` checks float-versus-`Vector2` only in DEBUG builds.
- `OtterGui.Raii.ImRaii` scopes return an interface over a struct and allocate per call;
  `ImUtf8` scopes are `ref struct`s and cannot be stored or captured.
- `ImUtf8` static buffers: one per buffer type, no locking. Do not call `ImUtf8` off the draw
  thread, and do not hold the span of a formatted string across another formatting call.
- `new Logger()` derives its name from the **calling assembly**.
- `EventWrapper` swallows subscriber exceptions and logs them only if
  `EventWrapperBase.ChangeLogger(logger)` was called. The highest priority value runs first.
- `ImGuiCacheService.KeepAliveFrames`' setter stores the difference to the old value, not the
  value.
- `Logger.Excessive` is `[Conditional("EXCESSIVE_LOGGING")]`: calls vanish unless the calling
  project defines that symbol.
- `ServiceManager.AddIServices` registers only exported (public) types and picks the constructor
  with the most parameters.
- `Custom/CustomGui` link buttons open hard-coded URLs of the upstream author's Discord, guide
  and donation pages. Only `DrawLinkButton` is generic.
- `FileCompactor.CanCompact` is `false` under Wine.
