# GenericHelpers, MathHelpers and small utility types

Covers: the `ECommons.GenericHelpers` static class (core, collection, string, conversion and
WinForms parts), `ECommons.MathHelpers`, `CircularBuffers`, `Collections`, `Numeric`,
`EqualStrings`, `PointerHelpers`, `Constants` and `ErrorCode`. Verified against ECommons `bc38dc7`
(2026-11-08, v3.2.1.26, Dalamud API 15); nothing here was compiled. Paths are relative to the skill
root.

Not covered here although they are parts of the same `GenericHelpers` partial class:
`AddonHelpers.cs` (`TryGetAddonByName`, `IsAddonReady`, `IsScreenReady`, ...) ->
`references/topic-addonmaster.md`; `ExcelHelpers.cs` -> `references/topic-excel.md`; `LogHelpers.cs` (`ToStringFull`,
`e.Log()`, `e.LogWarning()`, ...) -> `references/topic-logging-commands.md`. Stock Dalamud services ->
`dalamud` skill.

## Namespaces and how members are reached

```csharp
using ECommons;                          // GenericHelpers (extension methods), PointerHelpers, Constants, ErrorCode
using static ECommons.GenericHelpers;    // plain statics: Safe, Copy, Paste, IsOccupied, IsKeyPressed, ShellStart, ...
using ECommons.MathHelpers;              // MathHelper, Angle, AngleExtensions, Bitmask, BitHelper, Number, ENullable<T>, ...
using ECommons.CircularBuffers;          // CircularBuffer<T>, CircularArray<T>
using ECommons.Collections;              // OneOrMany<T>, UniqueQueue<T>
using ECommons.Numeric;                  // Int2
using ECommons.Interop;                  // LimitedKeys (argument type of IsKeyPressed)
using ECommons.WindowsFormsReflector;    // Keys (argument type of IsKeyPressed)
```

`GenericHelpers` is declared as `public static unsafe partial class GenericHelpers` in
`namespace ECommons;` across eight files in `references/ecommons/ECommons/GenericHelpers/`. Extension
methods (`this` first parameter) need only `using ECommons;`. Non-extension statics need either the
`GenericHelpers.` prefix or `using static ECommons.GenericHelpers;`. In the consumer checkouts
`using static ECommons.GenericHelpers;` appears in 10 files and AutoRetainer declares it as a
`global using static`, which is why its code calls bare `Safe(...)`, `Copy(...)`, `IsOccupied()`.

No member in this topic needs a `Module` flag and none installs a hook. Members that read game
state through `Svc` (`IsOccupied`, `IsNoConditions`, `IsTarget`, `MathHelper.ToVector3(Vector2)`)
require `ECommonsMain.Init` to have run and should be called on the framework thread.
`ECommonsMain.Dispose()` calls `EqualStrings.Dispose` (sets its registry to `null`); nothing else in
this topic holds state that needs disposing.

In the tables below **ext** marks an extension method, **static** a plain static.

## Exception wrappers

Source: `references/ecommons/ECommons/GenericHelpers/GenericHelpers.cs`.

| Signature | Behaviour |
|---|---|
| `void Safe(System.Action action, bool suppressErrors = false)` | Runs `action` in try/catch. On exception logs `PluginLog.Error($"{e.Message}\n{e.StackTrace ?? ""}")` unless `suppressErrors`. |
| `void Safe(System.Action a, Action<string, object[]> logAction)` | On exception calls `logAction(message + "\n" + stackTrace, Array.Empty<object>())` instead of `PluginLog.Error`. Nothing else is logged. |
| `void Safe(System.Action a, Action<string> fail, bool suppressErrors = false)` | On exception calls `fail(e.Message)` (message only, no stack trace), then logs the original exception with `PluginLog.Error` unless `suppressErrors`. If `fail` itself throws, logs `"Error while trying to process error handler:"` plus that exception and forces the original to be logged too. |
| `bool TryExecute(System.Action action)` | Returns `true` on success; on exception logs with `PluginLog.Error` and returns `false`. |
| `bool TryExecute<T>(Func<T> function, out T result)` | Same; `result` is `default` on failure. |
| `T Run<T>(Func<T> func)`, `void Run(System.Action action)` | Just invoke the delegate. No try/catch. Used to get an expression-bodied block. |

All are **static**. The logged text uses only `Message` and `StackTrace` of the outermost exception;
inner exceptions are not included (use `e.Log()` / `e.ToStringFull()` from `LogHelpers.cs` when you
need them). `ECommonsMain` itself wraps every init/dispose step in `GenericHelpers.Safe`.

```csharp
// AutoRetainer/AutoRetainer.cs
Safe(() => FFXIVInstanceMonitor.ReleaseLock());
```

## Clipboard, shell, files (GenericHelpers.Forms.cs and GenericHelpers.cs)

| Signature | Behaviour |
|---|---|
| `bool Copy(string text, bool silent = false)` | **static**. Null/empty `text` clears the clipboard. Uses `Winforms.Clipboard` (reflection over the runtime's `System.Windows.Forms.Clipboard`). Unless `silent`, shows `Notify.Success("Text copied to clipboard")` / `"Clipboard cleared"`, or `Notify.Error(...)` on failure. Failure also logs a warning. Never throws; returns success. |
| `string? Paste(bool silent = false)` | **static**. Returns clipboard text, or `null` if reading threw (then `Notify.Error` unless `silent`, plus a warning log). Source remark: run on the framework/draw thread when used with ImGui. |
| `void ShellStart(string s)` | **static**. `Process.Start` with `UseShellExecute = true` (URLs, folders, files). Failure shows `Notify.Error($"Could not open {s.Cut(60)}\n{e}")` and is also logged by `Safe`. |
| `void DeleteFileToRecycleBin(string path)` | **static**. `Microsoft.VisualBasic.FileIO.FileSystem.DeleteFile(..., SendToRecycleBin)`; exceptions are logged as warnings. |

```csharp
// AutoRetainer/UI/Windows/SubmarinePointPlanUI.cs
Copy(JsonConvert.SerializeObject(SelectedPlan));
var plan = JsonConvert.DeserializeObject<SubmarinePointPlan>(Paste());   // Paste() may be null
```

## Game state

| Signature | Behaviour |
|---|---|
| `bool IsOccupied()` | **static**. `true` if any flag below is set, or if `Svc.Objects.LocalPlayer?.IsTargetable != true` (so also `true` when there is no local player). |
| `bool IsNoConditions()` | **static**. `true` only if `ConditionFlag.NormalConditions` is set and no condition with index 2..99 is set, ignoring `ParticipatingInCrossWorldPartyOrAlliance`. |
| `bool IsTarget(this IGameObject obj)` | **ext**. Current target is non-null, `obj` non-null and addresses match. |
| `bool AddressEquals(this IGameObject obj, IGameObject other)` | **ext**. `obj?.Address == other?.Address`; two nulls compare equal. |
| `bool TryDecodeSender(SeString sender, out Sender senderStruct)` | **static**. Finds the first `PlayerPayload` and builds `ECommons.ChatMethods.Sender(p.PlayerName, p.World.RowId)`. `false` for null input or no player payload. |

`IsOccupied()` checks exactly these `ConditionFlag` values: `Occupied`, `Occupied30`, `Occupied33`,
`Occupied38`, `Occupied39`, `OccupiedInCutSceneEvent`, `OccupiedInEvent`, `OccupiedInQuestEvent`,
`OccupiedSummoningBell`, `WatchingCutscene`, `WatchingCutscene78`, `BetweenAreas`, `BetweenAreas51`,
`InThatPosition`, `Crafting`, `ExecutingCraftingAction`, `PreparingToCraft`, `Unconscious`,
`MeldingMateria`, `Gathering`, `OperatingSiegeMachine`, `CarryingItem`, `CarryingObject`,
`BeingMoved`, `RidingPillion`, `Mounting`, `Mounting71`, `ParticipatingInCustomMatch`,
`PlayingLordOfVerminion`, `ChocoboRacing`, `PlayingMiniGame`, `Performing`, `Fishing`,
`Transformed`, `UsingHousingFunctions`. `TradeOpen` is commented out in the source. Not checked:
`InCombat`, `Casting`, `Mounted`, `BoundByDuty`, `Jumping`; test those yourself.

`IsScreenReady()` lives in `AddonHelpers.cs`: it returns `false` while any of the addons
`NowLoading`, `FadeMiddle`, `FadeBack` exists and is visible. See `references/topic-addonmaster.md`.

```csharp
// AutoRetainer/AutoRetainer.cs
if(!IsOccupied() && !TaskManager.IsBusy) { ... }
```

## Key state

All **static**, all go through `IsKeyPressed(int key)` which tests bit 15 of a WinAPI key state
(TerraFX `GetKeyState` / `GetAsyncKeyState`). Key code `0` always returns `false`.

| Signature | Notes |
|---|---|
| `bool IsKeyPressed(int key)` | Virtual-key code. |
| `bool IsKeyPressed(LimitedKeys key)` | `ECommons.Interop.LimitedKeys`. |
| `bool IsKeyPressed(Keys key)` | `ECommons.WindowsFormsReflector.Keys` (Forms file). |
| `bool IsKeyPressed(IEnumerable<LimitedKeys> keys)`, `(IEnumerable<int> keys)`, `(IEnumerable<Keys> keys)` | `true` if **any** key is down. |
| `bool IsAnyKeyPressed(IEnumerable<LimitedKeys> keys)` | `[Obsolete("Use IsKeyPressed overload")]`. |
| `bool UseAsyncKeyCheck = false` | Public static field. See gotcha: its effect is inverted relative to its name. |

`Keys.cs` and `LimitedKeys.cs` are excluded from `references/`; they exist only in a checkout
(`ECommons/WindowsFormsReflector/Keys.cs`, `ECommons/Interop/LimitedKeys.cs`).

```csharp
// AutoRetainer/UI/UIUtils.cs: key-capture loop
foreach(var x in Enum.GetValues<LimitedKeys>())
    if(IsKeyPressed(x)) { key = x; }
```

## Equality, null and range tests (GenericHelpers.cs)

| Signature | Behaviour |
|---|---|
| `bool EqualsAny<T>(this T obj, params T[] values)` | **ext**. `values.Any(x => x.Equals(obj))`; a `null` element in `values` throws. |
| `bool EqualsAny<T>(this T obj, IEnumerable<T> values)` | **ext**. Same. |
| `bool IsNullOr<T>(this T source, Predicate<T> testFunction)` | **ext**. `source == null || testFunction(source)`. |
| `bool NotNull<T>(this T obj, [NotNullWhen(true)] out T outobj)` | **ext**. Null test that also hands back the value. |
| `bool TryGetValue<T>(this T? nullable, out T value) where T : struct` | **ext**. Unwraps a `Nullable<T>`. |
| `bool AllNull(params object[] objects)` | **static** (declared in `CollectionHelpers.cs`). |
| `bool ContainsNull<T>(params T[] objects)`, `bool ContainsNull<T>(this IEnumerable<T> objects)` | **static** and **ext**. `master` only: they replace `bool AnyNull(params object[] objects)`, which packages up to 3.2.1.26 have and `master` does not. |
| `bool Invert(this bool b, bool invert)` | **ext**. `invert ? !b : b`. |
| `bool If<T>(this T obj, Func<T, bool> func)` | **ext**. `[Obsolete]`, scheduled for removal, no replacement named. |
| `bool ApproximatelyEquals(this Vector3 one, Vector3 two, float diff = 0.01f)` | **ext**. Per-component `Math.Abs(a - b) < diff`. |
| `bool ApproximatelyEquals(this float one, float two, float diff = 0.01f)` | **ext**. |
| `ref int ValidateRange(this ref int i, int min, int max)` | **ext** on a variable. Clamps **in place** and returns the same ref. |
| `ref float ValidateRange(this ref float i, float min, float max)` | **ext**. Same. |

```csharp
// AutoRetainer/AutoRetainer.cs
if(!retainer.VentureID.EqualsAny(0u, LastVentureID)) { ... }
```

## Strings (StringHelpers.cs, GenericHelpers.cs)

All **ext** on `string` unless noted. Source:
`references/ecommons/ECommons/GenericHelpers/StringHelpers.cs`.

| Signature | Behaviour |
|---|---|
| `bool EqualsIgnoreCase(this string s, string other)` | `StringComparison.OrdinalIgnoreCase`. Throws on null `s`. |
| `bool EqualsIgnoreCaseAny(this string obj, params string[] values)` / `(IEnumerable<string> values)` | Any value equals, ordinal ignore case. |
| `bool ContainsIgnoreCase(this IEnumerable<string> haystack, string needle)` | Collection contains an element equal to `needle` ignoring case. It is **not** a substring test (declared in `CollectionHelpers.cs`). |
| `bool ContainsAny(this string obj, params string[] values)` / `(IEnumerable<string> values)` | Substring test, ordinal. |
| `bool ContainsAny(this string obj, StringComparison comp, params string[] values)` | Substring test with comparison. |
| `bool StartsWithAny(this string source, params string[] values)` | Ordinal. |
| `bool StartsWithAny(this string source, StringComparison stringComparison = Ordinal, params string[] values)` | |
| `bool StartsWithAny(this string source, IEnumerable<string> compareTo, StringComparison stringComparison = Ordinal)` | |
| `bool IsNullOrEmpty(this string? s)` | `string.IsNullOrEmpty`; `[OverloadResolutionPriority(1)]` so it wins over the `IEnumerable<T>` overload. |
| `string? NullWhenEmpty(this string s)` | `null` when `s == string.Empty`. |
| `string? NullWhenFalse(this string s, bool b)` | `b ? s : null`. |
| `string Default(this string s, string defaultValue)` | `defaultValue` when `s` is null or empty. |
| `string Cut(this string s, int num)` | First `num` chars plus `"..."` when longer; unchanged otherwise. |
| `string Repeat(this string s, int num)` | |
| `string[] SplitOnce(this string value, string delimiter)` | One- or two-element array, split at the first ordinal match. |
| `string[] Split(this string s, IEnumerable<string> separators, StringSplitOptions options = None)` | Multi-separator split; at equal position the longest separator wins; honours `RemoveEmptyEntries` and `TrimEntries`. (`GenericHelpers.cs`) |
| `IEnumerable<string> Split(this string str, int chunkSize)` | Fixed-size chunks; a trailing remainder shorter than `chunkSize` is **dropped**. |
| `string ReplaceFirst(this string text, string search, string replace)` | |
| `string ReplaceByChar(this string s, string replaceWhat, string replaceWith, bool replaceWithWhole = false)` | Char-for-char mapping (`replaceWhat[i]` -> `replaceWith[i]`); with `replaceWithWhole` each char of `replaceWhat` becomes the whole `replaceWith`. Length mismatch without the flag logs an error and returns `s` unchanged. |
| `string RemoveOtherChars(this string s, string charsToKeep)` | Keeps only chars present in `charsToKeep`. |
| `string Cleanup(this string s)` | Removes `' '`, `\n`, `\r`, `\t` (StringBuilder). |
| `string RemoveWhitespaces(this string s)` | Same four characters via `Replace`. (`GenericHelpers.cs`) |
| `int? ParseInt(this string number)` | `int.TryParse`, `null` on failure. (`ConversionHelpers.cs`) |

```csharp
// AutoRetainer/AutoRetainer.cs (command handler)
if(arguments.EqualsIgnoreCase("debug")) { ... }
else if(arguments.EqualsIgnoreCaseAny("e", "enable")) { ... }
// AutoRetainer/Scheduler/Tasks/TaskRecursivelyBuyFuel.cs
m.Text.ContainsAny(StringComparison.OrdinalIgnoreCase, "ceruleum", "青燐水バレル", "Erdseim", "céruleum")
```

## SeString / Utf8String

| Signature | Behaviour |
|---|---|
| `string GetText(this SeString seStr, bool onlyFirst = false)` | **ext**. Concatenates `TextPayload.Text` and `AutoTranslatePayload.Text`; an unknown payload encoding to bytes `02 1D 01 03` becomes a space. `onlyFirst` stops after the first text/auto-translate payload. |
| `string GetText(this ReadOnlySeString s, bool onlyFirst = false)` | **ext** on `Lumina.Text.ReadOnly.ReadOnlySeString`; `s.ToDalamudString().GetText(onlyFirst)`. |
| `string GetText(this Utf8String s, bool onlyFirst = false)` | **ext** on the ClientStructs struct (by value). `onlyFirst` is ignored: the body calls `GetText(false)`. |
| `string ExtractText(this SeString / ReadOnlySeString / Utf8String s, bool onlyFirst = false)` | All three `[Obsolete]`: renamed to `GetText` because Dalamud added its own incompatible `ExtractText` for Lumina strings. |
| `SeString Read(this Utf8String str)` | **ext**. `ReadSeString(&str)`. |
| `string Read(this Span<byte> bytes)` | **ext**. UTF-8 decode up to the first `0` byte (or the whole span). |
| `SeString ReadSeString(Utf8String* utf8String)` | **static**. `SeString.Parse(utf8String->AsSpan())`; null pointer gives an empty `SeString`. |
| `bool ContainsPartOf(this string haystack, SeString needle)` | **ext**. `true` if `haystack` contains the text of **any** non-blank `TextPayload` of `needle` (longest first). |
| `bool ContainsPartOf(this SeString haystack, SeString needle)`, `(this string, ReadOnlySeString)`, `(this SeString, ReadOnlySeString)` | **ext**. Convert and forward to the overload above. |
| `SeStringBuilder Add(this SeStringBuilder seStringBuilder, IEnumerable<Payload> payloads)` | **ext**. Appends each payload. |
| `SeString ReadSeString(IntPtr memoryAddress, int maxLength)` | **static**, `[Obsolete("Use MemoryHelper.ReadSeString")]`. Do not call: see gotchas. |
| `byte[] ReadRaw(IntPtr memoryAddress, int length)`, `void ReadRaw(IntPtr memoryAddress, int length, out byte[] value)` | **static**, `[Obsolete("Use MemoryHelper.ReadRaw")]`; forward to `Dalamud.Memory.MemoryHelper`. |
| `ushort GetParsedSeSetingColor(int percent)` | **static**, `[Obsolete]` ("presumably unused, to be removed"). Maps a parse percentile to a UIColor id. |

```csharp
// YesAlready/Features/SelectOk.cs
var text = ((AddonSelectOk*)atk)->PromptText->NodeText.GetText();
// AutoRetainer/Lang.cs (Lumina ReadOnlySeString)
Svc.Data.GetExcelSheet<EObjName>().GetRow(2000401).Singular.GetText()
```

## Collections (CollectionHelpers.cs, GenericHelpers.cs)

Source: `references/ecommons/ECommons/GenericHelpers/CollectionHelpers.cs`. All **ext** unless noted.

Lookup and safe access:

| Signature | Behaviour |
|---|---|
| `V? SafeSelect<K, V>(this IReadOnlyDictionary<K, V> dictionary, K? key)` | `default` if dictionary null, key null or key absent. |
| `V? SafeSelect<K, V>(this IReadOnlyDictionary<K, V> dictionary, K key, V defaultValue)` | `defaultValue` only when the key is absent; a null dictionary or null key still returns `default(V)`. |
| `T SafeSelect<T>(this IReadOnlyList<T> list, int index)` | `default` for null list or index out of range. |
| `T SafeSelect<T>(this T[] array, int index)` | `default` for index out of range; a null array throws. |
| `T? SafeSelectNullable<T>(this IReadOnlyList<T> list, int index) where T : struct`, `(this T[] array, int index)` | Same rules, but `null` instead of `default(T)`, so a missing entry is distinguishable from a zero value. 3.2.1.25 and later. |
| `bool TryGetValue<T>(this IList<T> collection, int index, out T result)`, `(this T[] collection, int index, out T result)` | `false` for a null collection or an index out of range. 3.2.1.23 and later. |
| `T FirstOrAdd<T>(this IList<T> list, Func<T> generator = null)` | Returns `list[0]`, or adds and returns `generator()` (or `Activator.CreateInstance<T>()`) when the list is empty. `master` only. |
| `T GetOrDefault<T>(this IReadOnlyList<T> List, int index)`, `(this T[] Array, int index)` | `[Obsolete]`, use `SafeSelect`. |
| `V GetOrDefault<K, V>(this IDictionary<K, V> dic, K key)` | Not obsolete. `default` when absent; null dictionary throws. |
| `V GetSafe<K, V>(this IDictionary<K, V> dic, K key, V Default = default)` | `[Obsolete("Use SafeSelect")]`. Still used by AutoRetainer. |
| `V GetOrCreate<K, V>(this IDictionary<K, V> dictionary, K key)` | Returns existing value or adds a new one: `""` for `string`, else `Activator.CreateInstance(typeof(V))`, else `default` if that throws. |
| `V GetOrCreate<K, V>(this IDictionary<K, V> dictionary, K key, V defaultValue)` | Adds and returns `defaultValue` when absent. The argument is evaluated on every call. |
| `V GetOrCreate<K, V>(this IDictionary<K, V> dictionary, K key, Func<V> defaultValueGenerator)` | Lazy variant. |
| `int IncrementOrSet<K>(this IDictionary<K, int> dic, K key, int increment = 1)` | Adds to the value or sets it; returns the new value. |
| `IEnumerable<K> FindKeysByValue<K, V>(this IDictionary<K, V> dictionary, V value)` | `value.Equals(x.Value)`; null `value` throws. |
| `T CircularSelect<T>(this IList<T> list, int index)`, `(this T[] list, int index)` | Index wrapped with `MathHelper.Mod`; negative indices count from the end. An empty list or array returns `default` on `master`; packages up to 3.2.1.26 throw `DivideByZeroException` there. |
| `void RemoveAll<K, V>(this IDictionary<K, V> dictionary, Predicate<KeyValuePair<K, V>> predicate)` | (`GenericHelpers.cs`) |

First/last/index:

| Signature | Behaviour |
|---|---|
| `bool TryGetFirst<TSource>(this IEnumerable<TSource> source, out TSource value)` | `false` for null or empty source. |
| `bool TryGetFirst<TSource>(this IEnumerable<TSource> source, Func<TSource, bool> predicate, out TSource value)` | `false` for null source/predicate or no match. |
| `bool TryGetFirst<K, V>(this IDictionary<K, V> dictionary, Func<KeyValuePair<K, V>, bool> predicate, out KeyValuePair<K, V> keyValuePair)` | Uses `First` inside try/catch. |
| `bool TryGetLast<K>(this IEnumerable<K> enumerable, Func<K, bool> predicate, out K value)` | Uses `Last` inside try/catch. |
| `bool TryGetLast<TSource>(this IEnumerable<TSource> source, out TSource value)` | `false` for null or empty source. 3.2.1.23 and later. |
| `T? FirstOrNull<T>(this IEnumerable<T> values) where T : struct`, `(..., Func<T, bool> predicate)` | `null` instead of `default(T)`. |
| `T FirstOr0<T>(this IEnumerable<T> collection, Func<T, bool> predicate)` | First match, else the element at index 0 (throws on empty). |
| `int IndexOf<T>(this IEnumerable<T> values, Predicate<T> predicate)`, `(this IEnumerable<T> values, T value)` | `-1` when not found. |
| `T GetRandom<T>(this IEnumerable<T> enumerable)` | `Random.Shared`; enumerates twice (`Count()` then `ElementAt`); throws on empty. |
| `T MaxSafe<T>(this IEnumerable<T>) where T : IComparable<T>`, `T MinSafe<T>(...)`, `K MaxSafe<T, K>(this IEnumerable<T>, Func<T, K> selector)`, `K MinSafe<T, K>(...)` | Never throw on null/empty, but the running value starts at `default`, see gotchas. (`GenericHelpers.cs`) |
| `uint Sum(this IEnumerable<uint> sequence)` | (`GenericHelpers.cs`) |

Mutation:

| Signature | Behaviour |
|---|---|
| `bool AddIfNotExist<T>(this ICollection<T> collection, T value)` | `false` if already present. (There is no `AddIfNotContains`.) |
| `void AddRange<T>(this ICollection<T> collection, IEnumerable<T> values)` | |
| `void Add<T>(this ICollection<T> collection, params T[] values)` | Multi-add. |
| `void Remove<T>(this ICollection<T> collection, params T[] values)` | Removes the first occurrence of each. |
| `bool Toggle<T>(this HashSet<T> hashSet, T value)`, `bool Toggle<T>(this List<T> list, T value)` | Returns whether the value is present afterwards. The list version removes **all** equal entries. |
| `bool TryDequeue<T>(this IList<T> List, out T result)`, `T Dequeue<T>(this IList<T> List)`, `T DequeueOrDefault<T>(this IList<T> List)` | Treat a list as a queue: remove index 0. `Dequeue` throws `InvalidOperationException` on empty. |
| `bool TryDequeueLast<T>(this IList<T> List, out T result)`, `T DequeueLast<T>(this IList<T> List)` | Same from the end. |
| `T DequeueOrDefault<T>(this Queue<T> Queue)` | |
| `void MoveItemToPosition<T>(IList<T> list, Func<T, bool> sourceItemSelector, int targetedIndex)` | **static**, not an extension. Throws if no item matches (index -1). |

Projection and tests:

| Signature | Behaviour |
|---|---|
| `bool IsNullOrEmpty<T>(this List<T> value)`, `(this T[] value)`, `(this IEnumerable<T> value)` | |
| `bool ContainsAny<T>(this IEnumerable<T> obj, params T[] values)` / `(IEnumerable<T> values)` | |
| `bool ContainsAll<T>(this IEnumerable<T> source, IEnumerable<T> values)` | |
| `bool ContainsNullable<T>(this IEnumerable<T> values, T? value) where T : struct` | `false` for null. |
| `IEnumerable<T?> AsNullable<T>(this IEnumerable<T> values) where T : struct` | |
| `void Each<T>(this IEnumerable<T> collection, Action<T> function)` | |
| `void EachWithIndex<T>(this IEnumerable<T> collection, Action<T, int> function)` | |
| `IEnumerable<(T Value, int Index)> WithIndex<T>(this IEnumerable<T> src)` | |
| `IEnumerable<T> TakeEvery<T>(this IEnumerable<T> values, int num)` | Elements at index 0, `num`, `2*num`, ... `num == 0` throws `ArgumentOutOfRangeException` when enumeration starts (3.2.1.22 and later). |
| `void SortBy<T, TKey>(this List<T> list, Func<T, TKey> keySelector, IComparer<TKey>? comparer = null)`, `SortByDescending`; same pair for `T[]`, `Span<T>` and `IList<T>` | In-place stable sort by key: each key is computed once and ties keep their original order. The `IList<T>` overloads have `[OverloadResolutionPriority(-1)]` and copy through a buffer unless the list is an array or `List<T>`. A null list or `keySelector` throws. 3.2.1.22 and later. |
| `IEnumerable<R> SelectMulti<T, R>(this IEnumerable<T> values, params Func<T, R>[] funcs)` | For each value yields every func result. |
| `IEnumerable<TResult> SelectNested<TSource, TResult>(this IEnumerable<IEnumerable<TSource>> source, Func<TSource, TResult> selector)` | Flatten then project; null inner sequences are skipped. |
| `T[] Together<T>(this T[] array, params T[] additionalValues)` | `Union`, so duplicates are removed. |
| `string Join(this IEnumerable<string> e, string separator)` | `string.Join`. |
| `string Print<T>(this IEnumerable<T> x, string separator = ", ")` | `ToString()` of each element (null -> `""`) joined. |
| `T[] CreateArray<T>(this T element, uint num)` | Array of `num` copies (same reference for classes). (`GenericHelpers.cs`) |
| `uint[] Range(uint inclusiveStart, uint inclusiveEnd)`, `int[] Range(int inclusiveStart, int inclusiveEnd)` | **static**. Both ends inclusive. (`GenericHelpers.cs`) |

```csharp
// AutoRetainer/AutoRetainer.cs
if(C.OfflineData.TryGetFirst(x => x.CID == Svc.ClientState.LocalContentId, out var offlineData)) { ... }
// AutoRetainer/Helpers/Utils.cs
var list = listings.GetOrCreate(x.GrandCompany.RowId, []);
// AutoRetainer/UI/Windows/SubmarinePointPlanUI.cs
SelectedPlan.Points.Toggle(x.RowId);
// AutoRetainer/UI/Windows/SubmarineUnlockPlanUI.cs
$"This plan is used by {my.Select(X => X.Key).Print()}."
```

## Conversions, cloning, ids, misc

Source: `references/ecommons/ECommons/GenericHelpers/ConversionHelpers.cs` and `GenericHelpers.cs`.

| Signature | Behaviour |
|---|---|
| `uint ToUint(this Vector4 color)` | **ext**. `ImGui.ColorConvertFloat4ToU32` (note lower-case `int`). |
| `Vector4 ToVector4(this uint color)` | **ext**. `ImGui.ColorConvertU32ToFloat4` (ImGui ABGR packing). |
| `Vector4 Invert(this Vector4 v)` | **ext**. `1 - X/Y/Z`, alpha kept. |
| `uint ToUInt(this ushort / byte / int value)`, `int ToInt(this byte / ushort / uint value)` | **ext**. Plain casts (unchecked for `int` <-> `uint`). |
| `uint AsUInt32(this float f)`, `float AsFloat(this uint u)` | **ext**. Bit reinterpretation, not numeric conversion. |
| `bool TryParseByteArray(string input, out byte[] output, char separator = ' ')` | **static**. Hex bytes (`"0A FF 10"`). |
| `byte[]? ParseByteArray(string input, char separator = ' ')` | **static**. `null` on failure. |
| `string ToHexString(this IEnumerable<byte> bytes, char separator = ' ')`, `(this IEnumerable<byte> bytes, string separator)` | **ext**. `X2` per byte; an empty/null string separator gives contiguous hex. |
| `T JSONClone<T>(this T obj, JsonSerializerSettings jsonSettings = null)` | **ext**. Newtonsoft serialize then deserialize. |
| `T DSFClone<T>(this T obj)` | **ext**. Same through `EzConfig.DefaultSerializationFactory`. |
| `string GetTemporaryId()` | **static**. 4 hex digit random prefix (fixed per load) plus an incrementing hex counter. Not thread safe. |
| `string GetDisplayTag(this Guid guid)` | **ext**. First GUID group, upper case. |
| `void Regenerate(this ref Guid guid)` | **ext**. Assigns `Guid.NewGuid()` in place. |
| `bool NextBool(this Random random)` | **ext**. |
| `string GetCallStackID(int maxFrames = 3)` | **static**. `"Type.Method <- Type.Method"` of callers; `0` means unlimited. Marked expensive in source. |
| `void SetMinSize(this Window window, float width = 100, float height = 100)`, `(this Window window, Vector2 minSize)` | **ext** on `Dalamud.Interface.Windowing.Window`. Sets `SizeConstraints` with `MaximumSize = new Vector2(float.MaxValue)`. |
| `void SetSizeConstraints(this Window window, Vector2 minSize, Vector2 maxSize)` | **ext**. |
| `string GetTerritoryName(this Number terr)` | **ext**, `[Obsolete("Use ExcelTerritoryHelper.GetName")]`. |
| `bool TryGetWorldByName(string world, out Lumina.Excel.Sheets.World worldId)` | **static**, `[Obsolete]`, use `ExcelWorldHelper.TryGetWorldByName`. |

There are no `Shuffle`, enum `Cast` or time/Unix-timestamp helpers in these files at this commit.

## MathHelper

Source: `references/ecommons/ECommons/MathHelpers/MathHelper.cs`. `public static class MathHelper` in
`ECommons.MathHelpers`. FFXIV world space is X east, Y up, Z south; the horizontal plane is X/Z.
`MathHelper` flattens `Vector3` to `Vector2` as `(X, Z)`, so in every `Vector2` overload
`Vector2.Y` means world Z.

Vector conversion (**ext**):

| Signature | Behaviour |
|---|---|
| `Vector2 ToVector2(this Vector3 vector3)` | `(X, Z)`; height is dropped. |
| `Vector3 ToVector3(this Vector2 vector2, float Y)` | `(v.X, Y, v.Y)`. |
| `Vector3 ToVector3(this Vector2 vector2)` | Height taken from `Svc.Objects.LocalPlayer?.Position.Y ?? 0`. |
| `Vector3 ToVector3(this (float X, float Y, float Z) t)` | |
| `Vector2 ToVector2(this int i)`, `(this float i)` | Both components set to the value. |
| `Vector2 ToVector2(this System.Drawing.Point p)` | |
| `Vector3 SwapYZ(this Vector3 v)` | `(X, Z, Y)`. |
| `Vector4 Add(this Vector4 v, float i)`, `Vector4 AddNoW(this Vector4 v, float i)` | Add to all components / all but W. |
| `float Float(this int i)` | Cast. |

Angles:

| Signature | Behaviour per the code |
|---|---|
| `float DegToRad(this float val)` | **ext**. `PI / 180 * val`, no normalisation. |
| `float RadToDeg(this float f)` | **ext**. `(f * 180 / PI + 360) % 360`: result in [0, 360) for inputs above -2π. |
| `float GetRelativeAngle(Vector2 origin, Vector2 target)` | **Degrees** in [0, 360). Compass bearing from `origin` to `target`: 0 = north (-Z), 90 = east (+X), 180 = south (+Z), 270 = west (-X). |
| `float GetRelativeAngle(Vector3 origin, Vector3 target)` | Same after `ToVector2()`. |
| `CardinalDirection GetCardinalDirection(float angle)` | Degrees in: [45, 135) `East`, [135, 225) `South`, [225, 315) `West`, everything else `North`. |
| `CardinalDirection GetCardinalDirection(Vector3 origin, Vector3 target)`, `(Vector2 origin, Vector2 target)` | Via `GetRelativeAngle`. |
| `float GetAngleBetweenPoints(Vector2 a, Vector2 b)` | **Radians**, `Atan2(b.Y - a.Y, b.X - a.X)`: mathematical convention (0 along +X), not a compass bearing and not a game rotation. |
| `float GetAngleBetweenLines(Vector2 a1, Vector2 a2, Vector2 b1, Vector2 b2)` | **Radians**, unsigned angle between direction `a1->a2` and `b1->b2` via `Acos` of the normalised dot product, range [0, π]. NaN for a zero-length line. |
| `Vector2 GetPointFromAngleAndDistance(Vector2 initialPoint, float angle, float distance)` | `angle` in **radians**; returns `(X + d * Sin(angle), Y + d * Cos(angle))`. Angle 0 points to +Y (world +Z, south) and π/2 to +X, the same convention as a game object's `Rotation`. |
| `Vector3 RotateWorldPoint(Vector3 origin, float angle, Vector3 p)` | `angle` in **radians**. Rotates `p` around `origin` in the X/Z plane, Y untouched: `x' = x*cos - z*sin`, `z' = x*sin + z*cos`. Positive angles turn +X towards +Z. |
| `Vector2 RotateWorldPoint(Vector2 origin, float angle, Vector2 p)` | Same on a `Vector2`. |

`CardinalDirection` (`references/ecommons/ECommons/MathHelpers/CardinalDirection.cs`) is declared in
the order `North, South, West, East` (values 0..3); do not assume clockwise numbering.

Points, lines and paths:

| Signature | Behaviour |
|---|---|
| `Vector2 MovePoint(Vector2 a, Vector2 b, float distance)` | Point at `distance` from `a` towards `b`. NaN if `a == b`. |
| `Vector2 FindClosestPointOnLine(Vector2 point, Vector2 lineA, Vector2 lineB)` | Projection onto the **infinite** line. |
| `bool IsPointPerpendicularToLineSegment(Vector2 point, Vector2 lineA, Vector2 lineB)` | Whether that projection falls within the segment. |
| `bool IsPointOnLine(Vector2 point, Vector2 a, Vector2 b, float tolerance = 0f)` | `|AB - (AP + PB)| <= tolerance`, i.e. on the segment. |
| `float CalculateDistance(IEnumerable<Vector2> vectors)` | Path length. Uses `ElementAt`, so pass a list/array. |
| `float Square(float x)` | |
| `double Mod(double dividend, double divisor)`, `float Mod(float, float)`, `int Mod(int, int)` | Always non-negative remainder. |
| `bool InRange(this T f, T inclusiveStart, T end, bool includeEnd = false)` | **ext** for `double, float, long, int, short, byte, ulong, uint, ushort, sbyte`. Start inclusive, end exclusive unless `includeEnd`. |
| `List<Vector2> CalculateCircularMovement(Vector2 centerPoint, Vector2 initialPoint, Vector2 exitPoint, out List<List<Vector2>> candidates, float precision = 36f, int exitPointTolerance = 1, (float Min, float Max)? clampRadius = null)` | Builds `precision` points on the circle through `initialPoint` around `centerPoint`, returns the shortest arc (list of waypoints) from the point nearest `initialPoint` to the point nearest `exitPoint`. `candidates` holds every tested path. Returns an empty list when `exitPointTolerance > 1` and a straight line is judged sufficient. Throws `Exception` on internal iteration overflow. `clampRadius` pulls waypoints into the given radius band. |
| `List<Vector3> CalculateCircularMovement(Vector3 centerPoint, Vector3 initialPoint, Vector3 exitPoint, out List<List<Vector3>> candidates, float precision = 36f, int exitPointTolerance = 1, (float Min, float Max)? clampRadius = null)` | X/Z wrapper; every output point gets `initialPoint.Y`. |

Clockwise ordering:

| Signature | Behaviour |
|---|---|
| `List<T> EnumerateObjectsClockwise<T>(IEnumerable<T> objects, Func<T, Vector2> getPosition, Vector2 centerPosition, Vector2 startingPosition)` | Sorts by bearing from `centerPosition`, starting at the bearing of `startingPosition`, increasing clockwise (north -> east -> south -> west). |
| `List<T> EnumerateObjectsClockwise<T>(IEnumerable<T> objects, Func<T, Vector2> getPosition, Vector2 centerPosition, float startingAngle)` | `startingAngle` in degrees from north. |
| `List<EnumerationResult<T>> EnumerateObjectsClockwiseEx<T>(...)` | Same two overloads, returning `EnumerationResult<T>`. |

`public readonly record struct EnumerationResult<T>(T Object, float AngleDegrees)`; `AngleDegrees`
is the offset from the starting bearing in [0, 360).

```csharp
// Adapted from AutoRetainer/Helpers/Utils.cs GetAngleTo(Vector2 pos); player = local player object
return (MathHelper.GetRelativeAngle(player.Position.ToVector2(), pos)
    + player.Rotation.RadToDeg()) % 360;

// Order party members clockwise starting from north around an arena centre (not compiled)
var ordered = MathHelper.EnumerateObjectsClockwise(members, m => m.Position.ToVector2(), center.ToVector2(), 0f);
```

## Angle

Source: `references/ecommons/ECommons/MathHelpers/Angle.cs`. `public struct Angle(float radians = 0)`
wrapping radians, in the game-rotation convention (0 = +Z, direction = `(sin, 0, cos)`).

- Field `float Rad`; property `float Deg`.
- Constants: `RadToDeg`, `DegToRad`, `HalfPi`, `DoublePI` (= `MathF.Tau`).
- `static Angle FromDirection(Vector2 dir)` = `Atan2(dir.X, dir.Y)`;
  `static Angle FromDirectionXZ(Vector3 dir)` = `Atan2(dir.X, dir.Z)`.
- `Vector2 ToDirection()` = `(Sin, Cos)`; `Vector3 ToDirectionXZ()` = `(Sin, 0, Cos)`.
- Operators: `==`, `!=`, `+`, `-` (binary and unary), `* float` (both orders), `/ float`, `<`, `>`,
  `<=`, `>=`.
- `Angle Abs()`, `float Sin()`, `float Cos()`, `float Tan()`, `static Angle Asin(float x)`,
  `static Angle Acos(float x)`, `Angle Round(float roundToNearestDeg)`,
  `Angle Normalized()` (into [-π, π]), `bool AlmostEqual(Angle other, float epsRad)`.
- `Equals`, `GetHashCode`; `ToString()` prints whole degrees (`"f0"`).

`public static class AngleExtensions`: `Angle Radians(this float radians)`,
`Angle Degrees(this float degrees)`, `Angle Degrees(this int degrees)`. Note the name clash in
meaning with `MathHelper.RadToDeg(this float)` / `DegToRad(this float)`, which return `float`.

```csharp
var facing = obj.Rotation.Radians();                       // not compiled
var ahead  = obj.Position + facing.ToDirectionXZ() * 5f;   // 5 yalms in front
if(facing.AlmostEqual(180.Degrees(), 0.1f)) { /* faces north (-Z) */ }
```

## Bit helpers, Number, ENullable

`Bitmask` (`references/ecommons/ECommons/MathHelpers/Bitmask.cs`), **static**:
`bool IsBitSet(ulong b, int pos)`, `bool IsBitSet(byte b, int pos)`,
`bool IsBitSet(short b, int pos)`,
`void SetBit(ref ulong b, int pos)`, `void ResetBit(ref ulong b, int pos)`. `pos` is zero-based from
the least significant bit.

`BitHelper` (`references/ecommons/ECommons/MathHelpers/BitHelper.cs`, copied from the .NET Community
Toolkit, branch-free, no argument validation), each for `uint` and `ulong`:

- `bool HasFlag(uint value, int n)` / `(ulong value, int n)`
- `bool HasLookupFlag(uint table, int x, int min = 0)` / `(ulong table, ...)`: bit `x - min` of
  `table`, `false` when out of range.
- `bool HasZeroByte(uint value)` / `(ulong value)`;
  `bool HasByteEqualTo(uint value, byte target)` / `(ulong value, byte target)`
- `void SetFlag(ref uint value, int n, bool flag)`, `uint SetFlag(uint value, int n, bool flag)`
  (and `ulong`)
- `uint ExtractRange(uint value, byte start, byte length)` (and `ulong`)
- `void SetRange(ref uint value, byte start, byte length, uint flags)`,
  `uint SetRange(uint value, byte start, byte length, uint flags)` (and `ulong`)

`Endianness.SwapBytes(this uint x)`: `[Obsolete("Use BinaryPrimitives.ReverseEndianness")]`.

`Number` (`references/ecommons/ECommons/MathHelpers/Number.cs`):
`public readonly unsafe struct Number : IEquatable<Number>`, an 8-byte explicit-layout union.
The struct is `[Obsolete]` (no message, no replacement named) from 3.2.1.23; do not use it in new
code.
Constructors and implicit conversions exist in both directions for `nint, nuint, long, ulong, int,
uint, short, ushort, byte, sbyte`. Operators `==`, `!=`, `+`, `-`, `*`, `/` work on the `long`
view; `ToString()` prints the `long` view. Used as a parameter type so callers can pass any integer
type without casts. Converting out reads the low bytes with no range check, so a negative `int`
stored then read as `uint` reinterprets. Note that the `int`/`short`/`sbyte`/`byte` constructors
write only their own width and leave upper bytes zero, so `new Number(-1)` has `long` value
`4294967295` and does not equal `new Number(-1L)`.

`ENullable<T> where T : struct` (`references/ecommons/ECommons/MathHelpers/ENullable.cs`): a
`Nullable<T>` look-alike that is a normal struct, serialisable by both Newtonsoft
(`ENullableConverter`) and System.Text.Json (`ENullableConverterSystemJson`) through attributes on
the type.

- `ENullable()`, `ENullable(T value)`, `bool HasValue`, `T Value` (throws
  `InvalidOperationException` when empty), `T GetValueOrDefault()`,
  `T GetValueOrDefault(T defaultValue)`.
- Conversions: implicit from `T` and from `T?`, implicit to `T?`, explicit to `T`.
- `ToString()` gives `""` when empty.
- Static class `ENullable`: `int Compare<T>(ENullable<T> n1, ENullable<T> n2)`,
  `bool Equals<T>(ENullable<T> n1, ENullable<T> n2)`, `Type? GetUnderlyingType(Type nullableType)`,
  `ref readonly T GetValueRefOrDefaultRef<T>(ref readonly ENullable<T> nullable)`.
- JSON shape: written as the bare value or `null`. Both readers also accept an object with
  `HasValue` and `Value` properties. The Newtonsoft reader returns an empty `ENullable` on any
  deserialisation error instead of throwing.

## CircularBuffer, CircularArray

`CircularBuffer<T> : IEnumerable<T>`
(`references/ecommons/ECommons/CircularBuffers/CircularBuffer.cs`), fixed capacity, overwrites the
opposite end when full:

| Member | Notes |
|---|---|
| `CircularBuffer(int capacity)`, `CircularBuffer(int capacity, T[] items)` | `ArgumentException` for capacity < 1 or too many items; `ArgumentNullException` for null items. |
| `int Capacity`, `int Size`, `bool IsFull`, `bool IsEmpty` | The count property is `Size`; `Count()` is only LINQ. |
| `T Front()`, `T Back()` | `InvalidOperationException` when empty. |
| `T this[int index] { get; set; }` | 0 = front. `IndexOutOfRangeException` when empty or `index >= Size`. |
| `void PushBack(T item)`, `void PushFront(T item)` | When full, `PushBack` drops the front element and `PushFront` drops the back. |
| `void PopBack()`, `void PopFront()` | Return nothing; read with `Back()`/`Front()` first. |
| `void Clear()`, `T[] ToArray()`, `IList<ArraySegment<T>> ToArraySegments()` | Enumeration runs front to back. |

```csharp
// AutoRetainer/Modules/Multi/MultiMode.cs: rate limiting on the last 5 interactions
internal static CircularBuffer<long> Interactions = new(5);
Interactions.PushBack(Environment.TickCount64);
if(Interactions.Count() == Interactions.Capacity && Interactions.All(x => Environment.TickCount64 - x < 60000)) { ... }
```

`CircularArray<T> : IEnumerable<T>`
(`references/ecommons/ECommons/CircularBuffers/CircularArray.cs`): `CircularArray(int capacity)`
(throws `Exception` when capacity < 2), `void Push(T value)` (fills free slots, then shifts
everything left by one and writes at the end, O(n)), `T this[int index] { get; }` (raw slot, no
filled check). The generic enumerator yields only filled slots; the non-generic one yields the raw
array.

## OneOrMany, UniqueQueue, Int2

`OneOrMany<T>` (`references/ecommons/ECommons/Collections/OneOrMany.cs`), a
`readonly struct` that avoids an array allocation for the single-element case:
constructors `(T single)`, `(T[] many)`, `(List<T> many)`, `(IEnumerable<T> many)`; field
`bool IsPopulated`; `static readonly OneOrMany<T> Empty`; `int Count`; `T this[int index]`;
`Enumerator GetEnumerator()` (pattern-based `foreach`; the struct does not implement
`IEnumerable<T>`, so LINQ does not apply).

`UniqueQueue<T> : IEnumerable<T>` (`references/ecommons/ECommons/Collections/UniqueQueue.cs`), a
queue that ignores items already queued: `int Count`, `void Clear()`, `bool Contains(T item)`,
`void Enqueue(T item)` (no-op for a duplicate), `T Dequeue()`, `bool TryDequeue(out T value)`,
`T Peek()`. An item can be enqueued again after it was dequeued.

`Int2` (`references/ecommons/ECommons/Numeric/Int2.cs`): `public struct Int2` with fields `int X`,
`int Y`, constructors `(int X, int Y)` and `(Vector2 v)` (truncating casts), `operator +`,
`Vector2 ToVector2()`.

## EqualStrings (obsolete)

`references/ecommons/ECommons/StringHelpers/EqualStrings.cs`, namespace `ECommons.StringHelpers`. The
whole class is `[Obsolete("Scheduled to be removed. Please contact the developer if you have an use
for it.")]`; no replacement is named. It keeps a static registry of string sets that should compare
equal:

- `void RegisterEquality(IEnumerable<string> values)`,
  `void RegisterEquality(params string[] values)`
- `bool Equals(string what, string with, StringComparison comparison = StringComparison.Ordinal)`
  and **ext** `ESEquals`
- `bool EqualsAny(string what, IEnumerable<string> with)` / `params`, **ext** `ESEqualsAny`
- `bool EqualsAnyIgnoreCase(string what, IEnumerable<string> with)` / `params`, **ext**
  `ESEqualsAnyIgnoreCase`

`ECommonsMain.Dispose()` nulls the registry. Do not use in new code.

## PointerHelpers, Constants, ErrorCode

`PointerHelpers` (`references/ecommons/ECommons/PointerHelpers.cs`, namespace `ECommons`), **ext**:

- `TResult Reinterpret<TValue, TResult>(this TValue value)` with both type parameters constrained
  to `unmanaged`: reads the bytes of `value` as `TResult`. No size check; `TResult` larger than
  `TValue` reads
  adjacent stack memory.
- `T* As<T>(this IntPtr ptr) where T : unmanaged`: pointer cast.

`Constants` (`references/ecommons/ECommons/Constants.cs`): one member,
`public static readonly XivChatType[] NormalChatTypes`, containing `Party, Alliance, Say, Yell,
Shout, CrossParty, CustomEmote, StandardEmote, TellOutgoing, TellIncoming, FreeCompany,
NoviceNetwork, PvPTeam, Ls1..Ls8, CrossLinkShell1..CrossLinkShell8`.

`ErrorCode` (`references/ecommons/ECommons/ErrorCode.cs`): `public enum ErrorCode { Success,
Player_is_not_logged_in, Invalid_world_specified, Plugin_is_busy }`, intended as a generic IPC
result; members carry `[Obfuscation]` so names survive obfuscation.

## Gotchas

- `UseAsyncKeyCheck` is inverted in the code: when `true`, `IsKeyPressed` calls `GetKeyState`;
  when `false` (default) it calls `GetAsyncKeyState`. The default therefore reports the physical
  key state regardless of which window has focus; AutoRetainer pairs it with a
  `WindowInactive` check.
- The obsolete `GenericHelpers.ReadSeString(IntPtr memoryAddress, int maxLength)` calls itself
  (`=> GenericHelpers.ReadSeString(memoryAddress, maxLength)`), which is unbounded recursion. Use
  `Dalamud.Memory.MemoryHelper.ReadSeString` as the attribute says, or the
  `ReadSeString(Utf8String*)` overload.
- `ENullable<T>` `==` / `!=` call `left.Equals(right)`, which resolves to `Equals(object?)` with a
  boxed `ENullable<T>`; the body then evaluates `value.Equals(other)` against that box (or
  `other == null` when empty). Read from the source, this returns `false` for two equal values and
  for two empty instances. Use the static `ENullable.Equals(n1, n2)` or compare `HasValue` and
  `Value` yourself.
- `OneOrMany<T>(IEnumerable<T> many)` never increments its index variable, so every element
  overwrites the single slot: the result holds only the **last** element with `Count == 1`. Pass an
  array or `List<T>`; note `[OverloadResolutionPriority(1)]` on the array constructor.
- `MaxSafe` / `MinSafe` start from `default(T)`: `MaxSafe` of all-negative numbers returns 0 and
  `MinSafe` of all-positive numbers returns 0. They are only safe for the sign they were meant for.
- `SafeSelect(dictionary, key, defaultValue)` returns `default(V)`, not `defaultValue`, when the
  dictionary or key is null, contrary to its XML doc.
- The dictionary `SafeSelect` takes `IReadOnlyDictionary<K, V>`; `Dictionary<K, V>` qualifies, a
  variable typed `IDictionary<K, V>` does not (use `GetOrDefault`, or `TryGetValue`).
- `EqualsAny`, `EqualsIgnoreCaseAny` and `IndexOf(value)` call `.Equals` on the collection
  elements, and `EqualsIgnoreCase` on the receiver: null elements/receivers throw
  `NullReferenceException`.
- `ContainsIgnoreCase` is a collection-membership test, not a case-insensitive substring test. For
  substrings use `ContainsAny(StringComparison.OrdinalIgnoreCase, ...)`.
- `GetText(this Utf8String, onlyFirst)` ignores `onlyFirst`. It also takes the struct by value, so
  call it on the field (`node->NodeText.GetText()`), not on a pointer.
- `ExtractText` on a Lumina `ReadOnlySeString` now binds to Dalamud's own method, which has
  different behaviour; ECommons code should call `GetText`.
- `Split(this string, int chunkSize)` silently drops a trailing partial chunk.
- `Toggle(List<T>)` removes every equal entry; `Together` deduplicates; `GetRandom` and
  `CalculateDistance` enumerate their input repeatedly.
- `IsOccupied()` returns `true` on the title screen and during login because the local player is
  null; it does not cover combat, casting or being mounted.
- Angle units are mixed: `GetRelativeAngle`, `GetCardinalDirection(float)` and
  `EnumerateObjectsClockwise(..., float startingAngle)` use degrees with 0 = north;
  `RotateWorldPoint`, `GetPointFromAngleAndDistance`, `GetAngleBetweenPoints`,
  `GetAngleBetweenLines` and `Angle` use radians, and among those only
  `GetPointFromAngleAndDistance` and `Angle` follow the game-rotation convention (0 = south).
- `ToVector3(this Vector2)` silently uses the local player's height (0 when there is none); pass
  `Y` explicitly for anything not at the player's elevation.
- `Copy`/`Paste` show a toast through `Notify` unless `silent: true`; pass `silent` in loops or
  background code.
- PunishLib also has a `GenericHelpers` class; importing both namespaces makes unqualified
  `GenericHelpers.X` ambiguous -> `punishlib` skill.
