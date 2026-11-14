# Coming from OtterGui: OtterGui -> Luna mapping

Covers the OtterGui -> Luna / ImSharp equivalents table and the OtterGui types with no Luna
counterpart. Verified against Luna `52eb0fe` and ImSharp `4d09327` (branch `gc1.88`, ImGui 1.88).
Paths are relative to the skill root. Grep `references/api-index.txt` before using a member that is
not shown here.

OtterGui is Apache-2.0, built on `Dalamud.Bindings.ImGui`, with its low-level
half in the `OtterGuiInternal` submodule. Luna is a rewrite, not a rename: the
right column is the nearest equivalent, and signatures differ.

| OtterGui | Luna / ImSharp |
|---|---|
| `OtterGui.Text.ImUtf8.Text(...)`, `ImUtf8.Button`, `ImUtf8.Table(id, columns, flags, …)` | `Im.Text`, `Im.Button`, `Im.Table.Begin(id, columns, flags, …)` |
| `Utf8StringHandler<LabelStringHandlerBuffer>` etc. (`OtterGui/Text/HelperObjects`) | Same type names in namespace `ImSharp`; aliases `Utf8LabelHandler`, `Utf8TextHandler`, `Utf8HintHandler` |
| `ImRaii.PushColor(ImGuiCol, uint/Vector4, condition)` | `ImGuiColor.X.Push(color[, condition])` / `Im.Color.Push(ImGuiColor, …)` |
| `ImRaii.PushStyle(ImGuiStyleVar, float/Vector2, condition)` | `ImStyleSingle.X.Push(float)`, `ImStyleDouble.X.Push(Vector2)` |
| `ImRaii.PushFrameBorder(thickness, color, condition)` | `ImStyleBorder.Frame.Push(color, thickness, condition)` |
| `ImRaii.PushId(...)`, `PushFont(...)`, `PushIndent(...)` | `Im.Id.Push(...)`, `Im.Font.Push(...)`, `Im.Indent(...)` |
| `ImRaii.DefaultColors()`, `DefaultStyle()` | `Im.Style.PushDefault()` (returns `ColorStyleDisposable`) |
| `ImRaii.IEndObject` results (`ImUtf8.Child`, `TabBar`, `Popup`, …) | `Im.*Disposable` ref structs from `Im.Child.Begin`, `Im.TabBar.Begin`, `Im.Popup.Begin`, … |
| `ImUtf8.IconButton(FontAwesomeIcon, tooltip, size, disabled, …)` | `ImEx.Icon.Button(AwesomeIcon, tooltip, …)` |
| `ImUtf8.HoverTooltip(...)`, `ImGuiUtil.HoverTooltip` | `Im.Tooltip.OnHover(...)` |
| `ImUtf8.LabeledHelpMarker`, `ImGuiUtil.LabeledHelpMarker` | `LunaStyle.DrawHelpMarkerLabel` / `DrawAlignedHelpMarkerLabel` |
| `ImUtf8.CopyOnClickSelectable`, `TextFramed`, `TextRightAligned`, `TextCentered`, `TextShadowed`, `ResizableChild`, `Spinner`, `RotatedText` | Same names on `ImEx` |
| `ImUtf8.InputTextDeactivated`, `InputScalarDeactivated`, `InputMultiLineDeactivated` files | `ImEx.InputOnDeactivation.Text/Scalar/MultiLine` |
| `ImGuiClip.ClippedDraw(...)`, `ImUtf8.ListClipper` | `Im.ListClipper` (`Iterate`, static `Draw`) |
| `ImGuiTable.DrawTable`, `DrawTabbedTable` | `ImEx.Table.Draw`, `ImEx.Table.DrawTabbed` |
| `OtterGui.Table.Table<T>`, `Column<T>`, `ColumnString<T>`, `ColumnNumber<T>`, `ColumnFlags<T, TItem>`, `YesNoColumn<T>` | `ImSharp.Table.DefaultTable<T>` / `TableBase<,>`, `BasicColumn<T>`, `TextColumn<T>`, `NumberColumn<TNumber, T>`, `FlagColumn<TEnum, T>`, `YesNoColumn<T>` |
| `OtterGui.Widgets.FilterComboBase<T>`, `FilterComboCache<T>` | `FilterComboBase<TCacheItem>`, `SimpleFilterCombo<T>`, `GenericCombo<T>` |
| `FilterComboColors`, `Changelog`, `Tutorial`, `TagButtons` | Same names under `Luna/Widgets` |
| `MultiStateCheckbox<T>`, `TriStateCheckbox` (`OtterGui/Text/Widget`) | Same class names in `ImSharp`; call through `ImEx.TriStateCheckbox`, `TwoStateCheckbox`, `XCheckbox`, `IconCheckbox` |
| `Widget.*` (framed group, hex viewer, color picker, key selectors in `OtterGui/Widgets`) | `ImEx.FramedGroup`, `ImEx.HexViewer`, `ImEx.ColorPickerButton`, `Luna.KeySelector.*` |
| `OtterGui.Classes.EventWrapper<T1, …, TPriority>(name)` | `Luna.EventBase<TArguments, TPriority>(name, log)`: one arguments struct passed by `in` |
| `OtterGui.Log.Logger` | `LunaLogger` / `MainLogger` |
| `OtterGui.Services.ServiceManager`, `IService` markers, `HookManager`, `FastHook`, `DynamisIpc` | Same names in namespace `Luna` |
| `ImGuiCacheService` | `CacheManager` + `BasicCache` |
| `SaveServiceBase<T>` | `BaseSaveService<T>` |
| `MessageService`, `FrameworkManager`, `KeyboardManager`, `ModifiableHotkey`, `ModifierHotkey`, `DoubleModifier`, `UserRegex`, `OptionalBool`, `IndexSet`, `SingleArray`, `SubList`, `TransformList`, `ZipList`, `ReadWriteDictionary`, `ConcurrentSet`, `BoolSetter`, `InMethodChecker`, `StartTimeTracker`, `DayChangeTracker`, `Backup`, `FileCompactor`, `SingleTaskQueue`, `IAction` | Same names in Luna (folders differ) |
| `MultiDictionary<,>` | `ListDictionary<,>` / `SetDictionary<,>` |
| `FileSystem<T>` with `Folder` / `Leaf`, `FileSystemSelector<T, TState>` | `BaseFileSystem` with `IFileSystemFolder` / `IFileSystemData<T>`, `FileSystemDrawer<TNodeCache>` |
| `OtterGui.String.StringU8` | `ImSharp.StringU8`, plus `StringPair` |
| `ImGuiCol.WindowBg`, `ImGuiCol.FrameBg`, … | `ImGuiColor.WindowBackground`, `ImGuiColor.FrameBackground`, … |

No Luna counterpart found for OtterGui's `LowerString`, `LazyList`,
`TemporaryList`, `TextureCache`, `PerformanceTracker`, `ItemSelector<T>`,
`ToggleButton`, `ChatTypeSelector`, `ImGuiPie`, `ClippedSelectableCombo<T>`
(no type of that name in `references/api-index.txt`).
