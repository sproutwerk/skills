# Luna folders

Covers the per-folder catalog of `Luna/`: UI (LunaStyle, Layout, Widgets, Text, Colors, Keyboard,
Conditions), the file system tree, and services and infrastructure. Verified against Luna `52eb0fe`
and ImSharp `4d09327` (branch `gc1.88`, ImGui 1.88). Paths are relative to the skill root. Grep
`references/api-index.txt` before using a member that is not shown here.

## UI

| Folder / file | Contents |
|---|---|
| `LunaStyle.cs` | Shared icon set as `AwesomeIcon` fields (`SaveIcon`, `DeleteIcon`, `HelpMarker`, `AddObjectIcon`, `FolderIcon`, `RefreshIcon`, `EditIcon`, `ConfigIcon`, `LockedIcon`, `ImportIcon`, `ToClipboardIcon`, `FromClipboardIcon`, `ErrorIcon`, `WarningIcon`, `InfoIcon`, …); `Vector4` properties forwarded from Dalamud's `ImGuiColors`: `InfoForeground`, `SuccessForeground`, `WarningForeground`, `ErrorForeground`, `AttentionForeground` and the `…Background` set; `DrawSeparator()`, `DrawGuid(Guid)`; `LunaStyle.Modifier` (`ModifierHelper`: `Destructive`, `Misclick`, `Add`, `Group`, `Bulk`, each a `CachedModifier` usable as `bool`, in `$"Hold {…}"`, and with `.Tooltip(action)`). |
| `LunaStyle.HelpMarker.cs` | `DrawHelpMarker(tooltip, color, treatAsHovered)`, `DrawAlignedHelpMarker(…)`, `DrawHelpMarker(label, tooltip, color)`, `DrawHelpMarkerLabel`, `DrawAlignedHelpMarkerLabel`, `DrawRightAlignedHelpMarker`, and no-text `DrawHelpMarker(color)` returning hover state. "Aligned" = aligned to frame padding. |
| `LunaStyle.YesNoColumn.cs` | `LunaStyle.YesNoColumn<T>` table column. |
| `Layout/` | `Window`, `OverlayWindow` (no sounds, click-through, pinning or blur), `IndexedWindow` + `WindowFactory<TWindow>` (numbered pop-out windows), `WindowSystem`, `TitleBarButton`; `TabBar<T>` + `ITab` / `ITab<T>` (`Label`, `DrawContent()`, `IsVisible`, `IsEnabled`, `Flags`, `Identifier`), with `CurrentTab`, `NextTab` and event `TabSelected`; `TwoPanelLayout` (`LeftPanel`/`RightPanel` : `IPanel`, `…Header` : `IHeader`, `…Footer` : `IFooter`, `Draw(in TwoPanelWidth)`, `TwoPanelWidth(float Width, ScalingMode Mode)`); `BaseButton` / `BaseButton<T>` (`Label`, `OnClick`, `Enabled`, `DrawButton(size)`, `DrawMenuItem()`, `DrawTabBarButton`; on `BaseButton<T>`, override `Button(size, in data)` or `MenuItem(in data)` to change the widget), `BaseIconButton<TIcon>`, `ButtonList` / `ButtonList<T>` (`AddButton(button, priority)`), `ButtonFooter`, `SplitButtonHeader`, `SubMenuButton`, `MenuSeparator`; `Editors/` (`IEditor<T>.Draw(Span<T> values, bool disabled)`, `DragEditor<T>`, `InputEditor<T>`, `SliderEditor<T>`, `EnumEditor<T>`, `ColorEditor`, `Editors.DefaultFloat`, `Editors.DefaultInt`). |
| `Widgets/` | `Changelog` (window; `NextVersion(title).RegisterEntry(…).RegisterHighlight(…).RegisterImportant(…)`), `Tutorial` (`Register(name, text)`, `Open(id, current, setter)`), `TagButtons.Draw(…)`, `SetButtons.DrawCombo/DrawCheckables…`, `KeySelector.Combo/Modifier/DoubleModifier/ModifiableKeySelector`, `InputPopup.OpenName/Open`, `ObjectEditPopup`, `HeaderLine`, `CaretButton`, `FilterComboColors`, `SupportButton.Discord/ReniGuide/Link/KoFiPatreon`. |
| `Text/` | `AwesomeIcon`, `UserRegex` (`Match(text)`, static `DrawRegexInput(label, ref regex, hint, …)`). |
| `Colors/` | User-editable color schemes: `ColorDictionary<TColorId, TColorData>` where `TColorData : IColorData<TColorId>` supplies `ColorData<TColorId>` defaults; `ColorDataUnion` (constant, or a reference to an `ImGuiColor`, `ImNodesColor`, `DalamudColor`, `LunaColor` or another own color); `ColorCache<,>` resolved values; `ColorSettingsDrawer.Draw(…)`; combos per color kind. |
| `Keyboard/` | `ModifierHotkey` (`NoKey`, `Shift`, `Control`, `Alt`, `IsActive()`), `DoubleModifier`, `ModifiableHotkey` (`IsPressed()`), `KeyboardManager.RegisterKey(…)`, `IKeyBindProvider`, `KeyExtensions`. |
| `ErrorWindow.cs` | Load-failure window; described at the end of `references/topic-services.md`. |
| `Conditions/` | Boolean condition trees `ICondition<TContext>` with `AndCondition`, `OrCondition`, `NotCondition`, `TrueCondition`, `FalseCondition`, `ConditionParser`, and `Drawing/ConditionDrawerCache<TContext>`, a node-graph editor on ImNodes. |

## File system tree

`Luna/Filesystem/` is a virtual folder tree over plugin objects (mods,
designs).
`BaseFileSystem(string name, LunaLogger log, bool allowsMultiSelection, IComparer<ReadOnlySpan<char>>? comparer = null)`
exposes `Root`, `Selection` (`FileSystemSelection`), events `Changed`
(`FileSystemChanged`) and `DataNodeChanged`, and mutators `CreateFolder`,
`FindOrCreateFolder`, `FindOrCreateAllFolders(path)`, `CreateDataNode<T>(parent, name, data)`,
`CreateDuplicateDataNode<T>`, `CreateSeparator`, `Rename`, `RenameAndMove`,
`Move`, `Merge`, `Delete`, `Find(fullPath, out node)`, `ChangeLockState`,
`ChangeExpandedState`, `ExpandAllAncestors`, `ChangeFolderSortMode`, `Clear`.
Nodes are `IFileSystemNode` → `IFileSystemFolder`, `IFileSystemData` /
`IFileSystemData<T>`, `IFileSystemSeparator`; stored values implement
`IFileSystemValue<TSelf>` and hold a `DataPath`. Sort modes are `ISortMode`
structs (`FoldersFirst`, `Lexicographical`, `FoldersLast`, `InternalOrder` and
inverses). Persistence: `FileSystemSaver<TSaveService, TProvider>`.

UI (`Filesystem/Ui/`): derive
`FileSystemDrawer<TNodeCache>(MessageService messager, BaseFileSystem fileSystem, IFileSystemFilter<TNodeCache>? filter)`
and implement `CreateCache()` returning a `FileSystemCache<TNodeCache>`. The
drawer is an `IPanel` (drop it in `TwoPanelLayout.LeftPanel`) with `Header`
(`FilterHeader<TNodeCache>`), `Footer` (`ButtonFooter`), and context menus as
button lists: `MainContext`, `FolderContext`, `SeparatorContext`,
`DataContext`. Dragging a node near the top or bottom edge scrolls the list;
the read-only `DragDropAutoScrollMargin` (one text line with spacing) is the margin. Ready buttons in `Filesystem/Ui/Buttons/`
(`CreateFolderButton`, `RenameFolderInput`, `DissolveFolderButton`,
`LockFolderButton`, `MoveSelectionButton`, `SortModeSelector`, …).

## Services and infrastructure

| Folder | Contents |
|---|---|
| `Services/` | `ServiceManager`, `ImSharpDalamudContext`, `FrameworkManager` (`RegisterOnTick(tag, action)`, `RegisterImportant`, `RegisterDelayed(tag, action, delay)`; registering the same tag again replaces the pending action), `DragDropManager` (external file drops: `AddSource`, `AddTarget`), `DynamisIpc`, `PredefinedTagManager<,>`, `DelayedReferenceLoader<,,>`, `SupportInfoProvider`, `DayChangeTracker` (instance class behind `IDayChangeTracker`; event `DayChanged(day, month, year)`; dispose it). Lifetimes and the marker interfaces are in `references/topic-services.md`. |
| `Events/` | `EventBase<TPriority>` and `EventBase<TArguments, TPriority>(string name, ILogger log, IComparer<TPriority>? comparer = null)`: `Subscribe(subscriber, priority)`, `Unsubscribe`, `Invoke(in arguments)`, `HasSubscribers`. Subscribers are `InAction<TArguments>` (`void (in T)`) and run in priority order. They are stored in a copy-on-write array, so `Invoke` takes no lock, and an exception from one subscriber is logged while the rest still run. Both classes are `IScopedService`. Convention in Penumbra: `class ModSelection : EventBase<ModSelection.Arguments, ModSelection.Priority>` with a nested `Arguments` record struct and `Priority` enum. |
| `Messaging/` | `MessageService(LunaLogger log, IChatGui chat, INotificationManager notificationManager)`: `AddMessage(IMessage message, doStore, doNotify, doLog, doChat)`, `AddTaggedMessage(tag, message)`, `DrawNotificationLog()`, `RemoveMessages`; extension `NotificationMessage(content, type, …)` and `NotificationMessage(Exception ex, …)`; `Notification : IMessage`; `AmassingNotification<T>` to fold many events into one toast. |
| `Ipc/` | Typed wrappers over Dalamud call gates: `FuncProvider<…>`, `ActionProvider<…>`, `EventProvider<…>`, `FuncSubscriber<…>`, `ActionSubscriber<…>`, `EventSubscriber<…>` (arities 0..8 produced by `[GenerateArities]`), `PluginSubscriber` (version handshake, `Available`, `Reattach`, `InitializationFailure`), `CallerPlugin`, and the `IIdDataShareAdapter`-based object sharing (`IpcObjectManager`, `BasicWrapper<TSelf, TEnum>` with `Version`, `IsDisposed` and `ThrowAdapterVersionMismatch(requiredMajor, minimumMinor)`, `AdapterVersionMismatchException`, `IAdapterFactory`). |
| `Files/` | `BaseSaveService<T>(LunaLogger log, FrameworkManager framework, T fileNames, Task? awaiter = null)` where `T : BaseFilePathProvider`: `Save(SaveType, in value)`, `QueueSave`, `DelaySave`, `ImmediateSave`, `ImmediateSaveSync`, `ImmediateDelete`, `Flush`. `SaveType`: `Queue`, `Delay`, `Immediate`, `ImmediateSync`, `None`. Savables implement `ISavable<T>` (`ToFilePath(T)`, `Save(Stream)`). Writes are atomic with backups (`BackupMode`, `BackupLimit`). `ConfigurationFile<TProvider>` (abstract `CurrentVersion`, `AddData(Utf8JsonWriter)`, `LoadData(in JsonElement)`, `ToFilePath`). `Backup.CreateAutomaticBackup/CreatePermanentBackup`, `BaseBackupService<>`, `JsonRecoveryStream` + `JsonRecoveryFlags` (repairs damaged JSON), `IJsonParsable<TSelf>`. |
| `Json/` | `JsonFunctions` (System.Text.Json): shared `SerializerOptions`, `WriterOptions`, `ReaderOptions`; `Utf8JsonWriter` extensions `WriteIfNot`, `WriteNonEmptyString`, `WriteEnumIfNot`, `TemporaryObject(name)`; reader/`JsonElement` extensions; `Utf8JsonObjectReader`; `JsonObjectConversion` between Newtonsoft `JObject` and `JsonDocument`. |
| `Logging/` | `LunaLogger` (abstract; implements `Microsoft.Extensions.Logging.ILogger` and `Serilog.ILogger`) with `Fatal/Error/Warning/Information/Debug/Verbose/Excessive`, each taking a `string` or a level-specific interpolated-string handler that skips formatting when the level is off. `MainLogger(string? pluginName = null)`, `MainLogger<T>`, `FileLogger(filePath, level)`. |
| `Tasks/` | `SingleTaskQueue.Enqueue(IAction, token)`, `IAction.Execute(token)`. |
| `Containers/` | `IndexList<T>`, `ListDictionary<,>`, `SetDictionary<,>`, `ReadWriteDictionary<,>`, `ConcurrentSet<T>`, `SingleArray<T>`, `SortedListAdapter<T>`, `SubList<T>`, `TransformList<,>`, `ZipList<,>`, `CollectionAdapter<T>`, `DisposableContainer`, `ReadOnlyWrapper<TKey1, TKey2, TValue>` (read-only view of a dictionary of dictionaries). |
| `Spans/` | `SpanBinaryReader`, `SpanTextWriter` (ref structs). |
| `Extensions/` | LINQ-like helpers (`IndexOf` and `FindFirst` for enumerables, arrays and lists, with context-passing overloads; `SelectWhere`), `Array.Repeat` and `array.Without(index)`, `Version.Equals(major, minor, …)`, list `Move`/`AddOrReplace`/`RemoveDuplicates`, span hex dumps and trimming, `Guid.ShortGuid()`, file helpers, `SeStringBuilder` colors, `DalamudExtensions.CreateFileDialog(WindowFlags)`. |
| `Utility/` | `OptionalBool`, `QuadBool`, `IndexSet`, `BoolSetter`, `BitFunctions`, `FormattingFunctions`, `CycleChecker`, `ValueCollection<T>`, `NullValue`, `ModifierHelper`, `WindowsFunctions`, `GenericJsonConverterAttribute`. |
| `Diagnostics/` | `StartTimeTracker` (`Measure(name)` scope; `StartTimeTracker.Draw(id, tracker)`), `InMethodChecker`, `Base64Tester`. |
| `DirectX/` | D3D11 via TerraFX: `Shader<T>` kinds, `Buffer`, `ConstantBuffer<T>`, `Sampler`, `FullScreenQuad`, `CustomRenderManager` + `ICustomRenderable`, image effects (`IEffect`, `EffectGraph`, `LunaEffects`, `ShaderFilterEffect`, `ComputeFilterEffect`, `Image`, `RwImage`), `TextureArraySlicer`. Not mirrored in `references/`; indexed only. |
| `GameInterop/` | `HookManager` (`CreateHook<T>(name, address|signature, detour, enable)` → `Task<Hook<T>?>`, `TryReplaceHook<T>`, `DisposeHook(name)`, `Hooks` as `IEnumerable<IHookData>`, event `HookStateChanged`, `LogExceptions(log)`), `FastHook<T>` base, `IHookData`, `HookManagerDrawer` (an `IUiService` with `Draw()` that lists the hooks), `SafeResourceHandle`, `SafeTextureHandle`. |
| `Compression/` | `FileCompactor` (Windows transparent compression), `CompressionFunctions`, `CompressionAlgorithm`. |
| `Resources/` | `ResourceProvider` (internal) plus shader sources; compiled `.dxbc` are embedded resources, built by `Build-Shaders.ps1`. |

Members that differ on older pins are listed under
[recent API changes](guide.md#recent-api-changes).
