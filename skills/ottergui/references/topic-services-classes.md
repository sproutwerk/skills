OtterGui topic: the non-UI half of the library: `Services/` (ServiceManager, IService markers,
HookManager, FastHook, ImGuiCacheService, DynamisIpc), `Classes/`, `Log/`, `Tasks/`,
`Compression/`, `String/`, `Extensions/`, and the `OtterGuiInternal` submodule. Verified against
`Limiana/OtterGui` `a08fda6` (2026-05-30, branch `main`) and `Ottermandias/OtterGuiInternal`
`49631aa` (2025-08-07, the commit the fork pins). Against upstream `79771ee` only
`Services/DynamisIpc.cs` and `Classes/MessageService.cs` differ in these folders; see
`references/topic-fork-delta.md`.

Paths such as `references/...` and `references/topic-*.md` are relative to the skill root (the directory that
holds `SKILL.md`), not to this file. Signatures are copied from `references/api-index.txt`; nothing
was compiled.

Neither Artisan nor AutoRetainer uses `Services/`; they are built on ECommons (`Svc`, `EzConfig`).
This half matters when reading Penumbra-lineage code or when a plugin wants OtterGui's DI.

## Log (`Log/Logger.cs`, namespace `OtterGui.Log`)

```csharp
public class Logger
{
    public Logger();                                   // name taken from Assembly.GetCallingAssembly()
    public string PluginName { get; }
    public Serilog.ILogger MainLogger { get; }
    public enum LogLevel { Excessive, Verbose, Debug, Information, Warning, Error, Fatal }  // Serilog values

    public void Fatal(string text);        public void Fatal(FatalInterpolatedStringHandler builder);
    public void Error(string text);        public void Error(ErrorInterpolatedStringHandler builder);
    public void Warning(string text);      public void Warning(WarningInterpolatedStringHandler builder);
    public void Information(string text);  public void Information(InformationInterpolatedStringHandler builder);
    public void Debug(string text);        public void Debug(DebugInterpolatedStringHandler builder);
    public void Verbose(string text);      public void Verbose(VerboseInterpolatedStringHandler builder);
    public void Verbose(string format, params object?[] args);
    public void Message(LogLevel level, string text);
    public void Message(LogLevel level, LogLevelInterpolatedStringHandler builder);
    public void Excessive(...);                        // [Conditional("EXCESSIVE_LOGGING")], three overloads
}
```

`Log.Debug($"Loaded {count} items in {ms} ms")` binds to the handler overload; the handler's
constructor checks whether the level is enabled and skips formatting when it is not. Output goes
to `Serilog.Log.ForContext("Dalamud.PluginName", name)`, the process-wide Serilog logger.
`LogLevel.Excessive` has the same value as `Verbose`. `LazyString` (same file) wraps a `Func<string>` whose `ToString()` evaluates it, for
expensive arguments.

Almost every other service takes a `Logger` in its constructor, so this is the first object to
create.

## Services (`Services/`, namespace `OtterGui.Services`)

### ServiceManager

A thin wrapper over `Microsoft.Extensions.DependencyInjection` with singletons only.

```csharp
public class ServiceManager : IDisposable
{
    public ServiceManager(Logger logger);              // registers the logger and itself
    public readonly StartTimeTracker Timers;           // construction time per service
    public ServiceProvider? Provider { get; }

    public ServiceManager AddSingleton<T>();
    public ServiceManager AddSingleton<T>(Func<IServiceProvider, T> factory) where T : class;
    public ServiceManager AddExistingService<T>(T service) where T : class;
    public ServiceManager AddDalamudService<T>(IDalamudPluginInterface pi) where T : class;
    public void AddIServices(Assembly assembly);       // every exported, concrete type implementing IService
    public ServiceProvider CreateProvider();           // ValidateOnBuild = true, ValidateScopes = false
    public void EnsureRequiredServices();              // instantiates every IRequiredService
    public T GetService<T>() where T : class;
    public IEnumerable<T> GetServicesImplementing<T>();
    public void Dispose();                             // disposes the provider and thereby the services
}
```

```csharp
// Plugin constructor (pattern; not compiled)
var log      = new Logger();
var services = new ServiceManager(log)
    .AddDalamudService<IFramework>(pluginInterface)
    .AddDalamudService<IKeyState>(pluginInterface)
    .AddDalamudService<IGameInteropProvider>(pluginInterface)
    .AddExistingService(this);
services.AddIServices(typeof(OtterGui.Services.HookManager).Assembly);   // OtterGui's own IServices
services.AddIServices(typeof(Plugin).Assembly);                          // yours
services.CreateProvider();
services.EnsureRequiredServices();
```

- `AddDalamudService<T>` obtains the service through `pi.Inject` on a private wrapper with a
  `[PluginService]` property, and also registers the plugin interface itself.
- Auto-constructed services use the public constructor with the **most parameters**; every
  parameter must be resolvable. `ValidateOnBuild` reports missing registrations when
  `CreateProvider` runs.
- `AddIServices` only sees `ExportedTypes`, so `internal` services are not picked up.
- `AddIServices(OtterGui assembly)` registers `HookManager`, `ImGuiCacheService`, `DynamisIpc`,
  `FileCompactor`, `FrameworkManager`, `KeyboardManager` and every other public `IService` in the
  library; each then needs its constructor dependencies registered. `IUiBuilder` is the
  `UiBuilder` property of `IDalamudPluginInterface`: `AddExistingService(pluginInterface.UiBuilder)`.

### IService markers (`Services/IService.cs`)

| Interface | Members | Meaning |
|---|---|---|
| `IService` | none | picked up by `AddIServices`; `[UsedImplicitly]` for inheritors |
| `IRequiredService : IService` | none | constructed eagerly by `EnsureRequiredServices()` |
| `IAsyncService : IService` | `Task Awaiter`, `bool Finished` | service with background initialisation |
| `IAwaitedService : IRequiredService, IAsyncService` | | |
| `IHookService : IAwaitedService` | `nint Address`, `Enable()`, `Disable()` | |
| `IDataContainer : IService` | `long Time`, `long Memory`, `string Name`, `int TotalCount` | |
| `IAsyncDataContainer : IDataContainer, IAsyncService` | | |
| `IInvokeService : IService` | `void Invoke()` | |
| `ILoadService : IService` | `void Load()` | |
| `IUiService : IService`, `IApiService : IService` | none | tags |

Only `IService` and `IRequiredService` are interpreted by `ServiceManager`; the others are
contracts for your own code.

### HookManager and FastHook

```csharp
public sealed class HookManager(IGameInteropProvider provider) : IDisposable, IService
{
    public readonly IGameInteropProvider Provider;
    public Task<Hook<T>>  CreateHook<T>(string name, nint address, T detour, bool enable = false) where T : Delegate;
    public Task<Hook<T>>  CreateHook<T>(string name, string signature, T detour, bool enable = false) where T : Delegate;
    public Task<Hook<T>?> TryReplaceHook<T>(string name, T detour) where T : Delegate;
    public Task<bool>     DisposeHook(string name);
    public IEnumerable<(string Name, nint Address, long Time, Type Delegate)> Diagnostics { get; }
}

public abstract class FastHook<T> : IHookService where T : Delegate
{
    protected Task<Hook<T>> Task { get; init; }
    public Task Awaiter { get; }   public bool Finished { get; }   public nint Address { get; }
    public void Enable();   public void Disable();   public void Toggle(bool value);
}
```

Hooks are created on the thread pool, one after another (each call is chained onto the previous
task). Names must be unique: a duplicate name throws inside the task. `CreateHook` with an
address `<= 0` throws synchronously. `Dispose` waits for the running task and disposes every
hook. `Hook<T>` is `Dalamud.Hooking.Hook<T>`; hooking rules are in the `dalamud` skill.

```csharp
public sealed unsafe class MyHook : FastHook<MyHook.Delegate>      // pattern; not compiled
{
    public delegate void Delegate(nint a1);

    public MyHook(HookManager hooks)
        => Task = hooks.CreateHook<Delegate>("My Hook", "<your signature>", Detour, true);

    private void Detour(nint a1)
        => Task.Result.Original(a1);
}
```

### ImGuiCacheService

Per-frame-expiring caches for UI data that is expensive to compute.

```csharp
public class ImGuiCacheService : IDisposable, IUiService
{
    public ImGuiCacheService(IUiBuilder uiBuilder, Logger log);    // subscribes to uiBuilder.Draw
    public int KeepAliveFrames { get; set; }                       // default 5, must be >= 1
    public CacheId GetNewId();
    public T? Cache<T>(CacheId id, Func<(T Cache, string Name)> generator);
    public bool TryGetCache<T>(CacheId id, out T? cache);          // does not extend the lifetime
    public bool InvalidateCache(CacheId id);
}
public readonly struct CacheId(int id) : IEquatable<CacheId>, IComparable<CacheId>
```

Get one `CacheId` per cache (store it in a field), call `Cache(id, generator)` every frame you
draw. The generator runs only when no cache exists; a cache that is not queried for
`KeepAliveFrames` frames is dropped, and disposed if it implements `IDisposable`. Known quirk,
read from the setter: `KeepAliveFrames = value` stores `value - oldValue` instead of `value`, so
changing it from the default 5 to 10 yields 5. Leave it at the default or verify the result.

### DynamisIpc

Client for the IPC of the Dynamis plugin (`Exter-N/Dynamis`, named in a comment of upstream's
file; subscribes to `Dynamis.*` call gates). Constructor
`DynamisIpc(IDalamudPluginInterface pi, Logger log)`. Fork members: `IsSubscribed`, `Features`,
`VersionMajor`, `VersionMinor`, `Error`, `InspectObject(nint address)`, `InspectRegion(nint
address, uint size, string typeName, uint typeTemplateId, uint classKindId)`,
`DrawTooltipDetails(nint address)`, `GetClass(nint address)`, `IsInstanceOf(nint address, string?
className, Type? classType)`, `DrawPointer(nint address)`, `DrawPointer(void* address)`,
`DrawDebugInfo()`. When Dynamis is absent `DrawPointer` falls back to a copy-on-click selectable.
Upstream has a wider API (`references/topic-fork-delta.md`).

## Classes (`Classes/`, namespace `OtterGui.Classes`)

### Application services

| Type | Constructor | Purpose and main members |
|---|---|---|
| `FrameworkManager : IDisposable, IService` | `(IFramework framework, Logger log)` | queue work onto the framework thread. `RegisterOnTick(string tag, Action action)`: one action per frame; `RegisterImportant(string tag, Action action)`: all on the next frame; `RegisterDelayed(string tag, Action action, TimeSpan delay)`: after the delay, re-registering a tag keeps the earlier time. Remaining actions run on dispose. Debug lists `Important`, `OnTick`, `Delayed`; field `Framework` |
| `SaveServiceBase<T>` | `protected (Logger log, FrameworkManager framework, T fileNames, Task? awaiter = null)` | file saving for `ISavable<T>` objects. `Save(SaveType type, in ISavable<T> group)`, `QueueSave`, `DelaySave(value)` (10 s in Release, 2 s in Debug), `DelaySave(value, TimeSpan delay)`, `ImmediateSave`, `ImmediateSaveSync`, `ImmediateDelete`; field `FileNames`. Writes to `<name>.tmp` and moves it over the target when the file already exists; saves are serialised on one task chain |
| `ISavable<in T>` | | `string ToFilename(T fileNames)`, `void Save(StreamWriter writer)`, default `LogName(string fileName)`, `TypeName` |
| `enum SaveType` | | `Queue`, `Delay`, `Immediate`, `ImmediateSync`, `None` |
| `MessageService` | `(Logger log, IUiBuilder builder, IChatGui chat, INotificationManager notificationManager)` | one entry point for log + notification + chat + an in-plugin message list. `AddMessage(IMessage message, bool doPrint = true, bool doNotify = true, bool doLog = true, bool doChat = false)` returns the timestamp key; `AddTaggedMessage(string tag, IMessage message)` suppresses repeats of a tag for five minutes; `Draw()` renders the list; `CleanTaggedMessages(bool force)`. Implements a read-only dictionary `DateTime -> IMessage`. Note it is **not** an `IService` |
| `MessageService.IMessage` | | `NotificationType`, `Message`, `NotificationMessage`, `NotificationTitle`, `NotificationDuration`, `LogMessage`, `ChatMessage`, `LogLevel`, `PrintMessage`, `PrintTooltip` (upstream adds `OnNotificationActions`) |
| `Notification : MessageService.IMessage` | `(string content, NotificationType type, uint duration = 5000)`, `(Exception ex, string content1, string content2, NotificationType type, uint duration = 5000)` | the standard message |
| `MessageServiceExtensions` | | `NotificationMessage(this MessageService service, string content, NotificationType type = None, bool doPrint = true)`, and overloads taking `Exception ex, string content1[, string content2]` |
| `SeStringBuilderExtensions` | | `AddText(text, int color, bool brackets = false)`, `AddGreen`, `AddYellow`, `AddRed`, `AddBlue`, `AddWhite`, `AddPurple`, `AddCommand(string command, string description)`, `AddInitialPurple(string word, bool withComma = true)`; colour constants `Green = 504`, `Yellow = 31`, `Red = 534`, `Blue = 517`, `White = 1`, `Purple = 541` |
| `KeyboardManager : IService, IDisposable` | `(IFramework framework, IKeyState keyState)` | `RegisterKey(ModifiableHotkey key)` and overloads `(VirtualKey)`, `(VirtualKey, ModifierHotkey)`, `(VirtualKey, ModifierHotkey, ModifierHotkey)`; on the next framework update the registered keys whose modifiers were held are cleared from `IKeyState`, so the game does not see them. `DrawDebug()` |
| `TextureCache` | `(IDataManager dataManager, ITextureProvider textureProvider)` | `LoadIcon(uint iconId)`, `LoadFile(string file)`, `TryLoadIcon`, `TryLoadFile`, `FileExists`, `IconExists` |
| `PerformanceTracker<T> : IDisposable where T : unmanaged, Enum` | `(IFramework framework)` | per-frame timing by enum key. `Enable()`, `Disable()`, `Measure(T timingType)` (returns a disposable), `Start`, `Stop`, `Draw(string label, string textBox, Func<T, string> toNames)`; `Enabled`, `TotalFrames` |
| `StartTimeTracker` | | one-shot startup timings by name. `Measure(string name)` (disposable), `Measure(string name, Action action)`, `Measure<TRet>(string name, Func<TRet> func)`, `Start`, `Stop`, `Draw(string label)` |
| `Backup` (static) | | `CreateAutomaticBackup(Logger logger, DirectoryInfo dir, IReadOnlyCollection<FileInfo> files)`: timestamped backup, skipped when identical to the newest, oldest deleted beyond `MaxNumBackups = 10`; `CreatePermanentBackup(logger, dir, files, string name)`; `TryGetFile<T>(DirectoryInfo dir, string fileName, out T? parsedFile, out string message, Func<string, T?>? parse = null)` |
| `DelayedReferenceLoader<TObj, TData>` (abstract) | `(MessageService messager)` | resolve cross references after loading; implement `TryGetObject`, `SetObject`; call `SetAllObjects()` |
| `DayChangeTracker` (static) | | `event Action<int, int, int>? DayChanged` (day, month, year) raised by a timer at local midnight |

### Events

`EventWrapper<..., TPriority>` is a prioritised, exception-isolating multicast event, declared by
deriving a named class.

```csharp
public sealed class SelectionChanged() : EventWrapper<Item?, Item?, SelectionChanged.Priority>(nameof(SelectionChanged))
{
    public enum Priority { Ui = 0, Cache = 10 }
}

// subscriber:  _event.Subscribe(OnSelection, SelectionChanged.Priority.Ui);
// publisher:   _event.Invoke(oldItem, newItem);
// teardown:    _event.Unsubscribe(OnSelection);
```

- Base `EventWrapperBase<TPriority>(string name) : IDisposable, IService` with `Name`,
  `HasSubscribers`; variants `EventWrapper<TPriority>` and `EventWrapper<T1, ..., TPriority>` for
  one to six arguments.
- By-reference and pointer variants: `EventWrapperRef2`, `EventWrapperRef3`, `EventWrapperRef34`,
  `EventWrapperPtr` (1 to 3 pointers), `EventWrapperPtr1`, `EventWrapperPtr234`,
  `EventWrapperPtr12Ref34`, with matching delegate types `ActionRef`, `ActionRef2`, `ActionRef3`,
  `ActionRef34`, `ActionPtr`, `ActionPtr1`, `ActionPtr234`, `ActionPtr12Ref34`.
- Subscribers are kept sorted by ascending priority and `Invoke` walks the list **from the end**,
  so the subscriber with the highest priority value runs first.
- An exception in one subscriber is caught and logged through the static
  `EventWrapperBase.Logger`; set it once with `EventWrapperBase.ChangeLogger(logger)`, otherwise
  exceptions are swallowed silently.
- Subscribing a delegate that is already subscribed moves it to the new priority.

### Hotkeys

| Type | Members |
|---|---|
| `ModifierHotkey` (readonly struct) | `NoKey`, `Shift`, `Control`, `Alt`, `ValidModifiers`, `ValidKeys`, field `Modifier`, implicit conversions from and to `VirtualKey`, `IsActive()` |
| `DoubleModifier` (struct) | `NoKey`, `Modifier1`, `Modifier2`, constructors `()`, `(ModifierHotkey)`, `(ModifierHotkey, ModifierHotkey)`, `SetModifier1`, `SetModifier2`, `ForcedModifier(DoubleModifier defaultValue)`, `IsActive()` |
| `ModifiableHotkey` (struct) | `Hotkey`, `Modifier1`, `Modifier2`, constructors `(VirtualKey hotkey, VirtualKey[]? validKeys = null)`, `(hotkey, modifier1, validKeys)`, `(hotkey, modifier1, modifier2, validKeys)`, `SetHotkey(VirtualKey key, IReadOnlyList<VirtualKey>? validKeys = null)`, `SetModifier1`, `SetModifier2`, `IsPressed()` |
| `VirtualKeyExtensions` | `IsPressed(this VirtualKey key, IKeyState state)` |

UI for these: `Widget.KeySelector`, `ModifierSelector`, `DoubleModifierSelector`,
`ModifiableKeySelector` (`references/topic-widgets.md`). `DoubleModifier` is what
`FileSystemSelector.DeleteSelectionButton` takes as its confirm modifier.

### Strings and filters

| Type | Members |
|---|---|
| `LowerString` (readonly struct) | a string stored with its lower-case form: `Text`, `Lower`, `Empty`, `Length`, `IsEmpty`, `Contains`, `StartsWith`, `EndsWith`, `IsContained(string other)`, implicit conversions from and to `string`, static `InputWithHint(string label, string hint, ref LowerString s, int maxLength = 128, ImGuiInputTextFlags flags = None)`; has a Newtonsoft converter |
| `UserRegex` (struct) | user-entered regex with fallback: `Text`, `Regex`, `Empty`, `DefaultOptions`, `UpdateText(string text, RegexOptions options = DefaultOptions)`, `Match(string text)`, static `DrawRegexInput(string label, ref UserRegex regex, string hint = "", Action? tooltip = null, float width = 0, uint failureColor = 0, RegexOptions options = DefaultOptions)` |
| `FilterUtility<T>` (abstract) | `Draw(float width)`, `IsEmpty`, abstract `ApplyFilter(in T value)`, `FilterChange(string input)`, `Tooltip` |
| `OptionalBool`, `QuadBool` (readonly structs) | tri-state and four-state booleans. `OptionalBool.True / False / Null`, `HasValue`, `Value`; `QuadBool.True / False / NullTrue / NullFalse / Null`, `Value`, `Enabled`, `ForcedValue`, `Split`, `SetEnabled`, `SetValue`, `TryParse`, `ToJObject`, `FromJObject` |

### Containers and small utilities

| Type | What it is |
|---|---|
| `ICachingList<out T> : IReadOnlyList<T>` | `ClearList()`, `IsInitialized` |
| `LazyList<T>(Func<IReadOnlyList<T>> generator)` | list generated on first access, dropped by `ClearList()` |
| `TemporaryList<T>(IEnumerable<T> items)` | materialises the enumerable on first access, dropped by `ClearList()` |
| `MultiDictionary<TKey, TValue>` | key to list of values: `TryAdd`, `Remove`, `RemoveValue`, `TryGetValue(in TKey key, out IReadOnlyList<TValue> values)`, `Grouped`, `KeyCount`, `ValueCount` |
| `UniquePairMultiDictionary<TKey, TValue>` | key to set of values; same surface with `IReadOnlySet<TValue>` and `int TryAdd(in TKey key, IEnumerable<TValue> values)` |
| `MultiDictionaryExtensions` | `ToMultiDictionary(...)`, `ToUniquePairMultiDictionary(...)` |
| `ReadWriteDictionary<TKey, TValue>` | `IDictionary` that derives from `ReaderWriterLockSlim` |
| `ConcurrentSet<T>` | `ConcurrentDictionary<T, NullValue>` with `TryAdd(T value)`, `TryRemove(T value)` |
| `IndexSet` | bit set of indices: `Add`, `Remove`, `AddRange`, `RemoveRange`, `Indices(...)`, `Ranges(...)`, `Count`, `IsEmpty`, `IsFull`, constructor `(int capacity, bool initiallyFull)` |
| `SingleArray<T>` | zero, one or many values without an array for the single case; `Append`, `Prepend`, `Remove` |
| `SubList<T>`, `ZipList<T1, T2>`, `TransformList<TIn, TOut>` | read-only list views |
| `BoolSetter` (ref struct) | `using var _ = new BoolSetter(ref flag, setTo: true);` sets a bool and resets it on dispose |
| `InMethodChecker` | re-entrancy guard: `EnterMethod()` returns a `RaiiSetter`; `InMethod`, `RecursionDepth` |
| `DisposableContainer` | disposes several `IDisposable?` at once; `Empty` |

## Tasks (`Tasks/`, namespace `OtterGui.Tasks`)

| Type | Members |
|---|---|
| `IAction : IEquatable<IAction>` | `void Execute(CancellationToken token)` |
| `enum ActionState` | `NotQueued`, `NotStarted`, `Cancelled`, `Running`, `Failed`, `Succeeded` |
| `SingleTaskQueue` | `Task Enqueue(IAction action, CancellationToken token = default)`: runs actions one after another |
| `TrackedTask`, `AwaitedTask`, `AsyncTask` (static) | each only has `Task Run(Action action)`, which calls `Task.Run(action)`. The three names exist to mark intent at the call site |

## Compression (`Compression/`, namespace `OtterGui.Compression`)

File system compression of individual files on Windows ("Available compression methods for file
system compression" in the source). `FileCompactor(Logger logger) :
IDisposable, IService` with `CanCompact` (false under Wine: `!Dalamud.Utility.Util.IsWine()`),
`Enabled`, `MassCompactRunning`, `CurrentFile`, `CurrentIndex`, `TotalFiles`,
`StartMassCompact(IEnumerable<FileInfo> files, CompressionAlgorithm algorithm, bool
logFileNotFound)`, `CancelMassCompact()`, `GetFileSizeOnDisk(string filePath)`,
`WriteAllBytes(string filePath, byte[] decompressedFile)`, `WriteAllBytesAsync(...)`.
`enum CompressionAlgorithm { None = -2, Lznt1 = -1, Xpress4K = 0, Lzx = 1, Xpress8K = 2,
XPress16K = 3 }`; `Interop.cs` holds the P/Invoke
layer (`IsCompactedFile` and `CompactFile` default to `CompressionAlgorithm.Xpress8K`).

## String (`String/`, namespace `OtterGui.String`)

`StringU8` (file `Utf8String.cs`): an immutable UTF-8 string struct implementing
`IReadOnlyList<byte>`, `IEquatable<StringU8>`, `IComparable<StringU8>`,
`IComparisonOperators<StringU8, StringU8, bool>`, `ISpanFormattable` and `IUtf8SpanFormattable`,
with a nested `Utf8InterpolatedStringHandler`. `Interop.cs` has
`VirtualQuery(nint address, out MemoryBasicInformation info)` and the matching structs/enums.
Neither Artisan nor AutoRetainer imports this namespace.

## Extensions (`Extensions/`, namespace `OtterGui.Extensions`)

| Class | Members |
|---|---|
| `ArrayExtensions` | `WithIndex<T>(this IEnumerable<T>)` returning `(T Value, int Index)`, `WithoutIndex`, `WithoutValue`, `IndexOf<T>(this IEnumerable<T>, Predicate<T>)`, `IndexOf<T>(this IEnumerable<T>, T needle)`, `FindFirst<T>(..., out T? result)` (enumerable and span forms), `SelectWhere<TIn, TOut>(this IEnumerable<TIn>, Func<TIn, (bool, TOut?)> filterMap)`, `WriteHexBytes`, `WriteHexByteDiff` |
| `StableInsertionSortExtension` | `StableSort<T, TKey>(this IList<T> list, Func<T, TKey> selector)`, `StableSort<T>(this IList<T> list, Comparison<T> comparer)` |
| `VectorExtensions` | `Ceiling`, `Floor`, `Round` on `Vector2` |
| `FileExtensions` | `EnumerateNonHiddenFiles(this DirectoryInfo)`, `EnumerateNonHiddenFilesFiltered(this DirectoryInfo, Func<FileInfo, bool>? filter)`, `IsHidden(this FileSystemInfo)` |
| `GuidExtensions` | `ShortGuid`, `OptimizedString`, `FromOptimizedString(ReadOnlySpan<char> / ReadOnlySpan<byte> text, out Guid guid)` |
| `PowerSetExtensions` | `MoveWithPowerSet`, `RemoveWithPowerSet`, `AddNewWithPowerSet` |

If another library in the plugin defines extension methods with the same names (`IndexOf`,
`WithIndex`, `Move`), the compiler reports an ambiguous call; qualify with the class name.

## OtterGuiInternal (submodule, compiled into the OtterGui assembly)

Repository `Ottermandias/OtterGuiInternal`, license **Boost Software License 1.0**
(`references/src/OtterGuiInternal/LICENSE`), 13 `.cs` files including `GlobalUsings.cs`. It wraps ImGui internals that the
public bindings do not expose at the top level.

| File | Namespace | Contents |
|---|---|---|
| `ImGuiInternal.cs` | `OtterGuiInternal` | `static unsafe class ImGuiInternal`: `GetCurrentWindow()`, `ItemSize(ImRect / Vector2, float textBaseLineY = -1f)`, `ItemAdd(ImRect boundingBox, ImGuiId id, ImGuiItemFlags flags = 0)` (+ overload with `navBoundingBox`), `ItemSetKeyOwner(ImGuiKey key, ImGuiInputFlags flags)`, `ItemSetUsingMouseWheel()`, `ButtonBehavior(ImRect boundingBox, ImGuiId id, out bool hovered, out bool held, ImGuiButtonFlags flags = 0)`, `RenderNavHighlight`, `RenderFrame` (two forms), `CalcItemSize(Vector2 min, float defaultWidth, float defaultHeight)`, eight `RenderTextClipped` overloads, `Rotate(Vector2 vec, float cos, float sin)` |
| `Utility/StringHelpers.cs` | `OtterGuiInternal.Utility` | `AddText`, `ComputeId`, `ComputeSize`, `ComputeSizeAndId`, `SplitString`, `SplitStringWithNull`, `SplitStringWithoutNull`; `const int MaxStackAlloc = 2047` |
| `Utility/SymbolHelpers.cs` | same | `RenderCross`, `RenderCheckmark`, `RenderDot`, `RenderDash` `(ImDrawListPtr drawList, Vector2 position, uint color, float size)` |
| `Utility/ColorHelpers.cs` | same | `GetFrameBg(bool hovered, bool held)`, `GetButtonColor(bool hovered, bool held)` |
| `Enums/ImGuiId.cs`, `Direction.cs`, `ChildFlags.cs` | `OtterGuiInternal.Enums` | `enum ImGuiId : uint` (a typed ImGui ID), `Direction`, `ChildFlags` |
| `Enums/ImGuiInputFlags.cs` | `OtterGui.OtterGuiInternal.Enums` (sic) | `enum ImGuiInputFlags : ulong` |
| `Structs/*.cs`, `Wrappers/ImGuiWindowTempPtr.cs` | `OtterGuiInternal.Structs`, `.Wrappers` | `ImVec2`, `ImVector<T>`, `ImGuiWindowTemp`, `ImGuiWindowTempPtr` |

You need it only when writing a custom widget the way `MultiStateCheckbox<T>` does: compute size
and ID with `StringHelpers.ComputeSizeAndId`, reserve space with `ItemSize`, register with
`ItemAdd`, handle input with `ButtonBehavior`, draw with the draw list.
