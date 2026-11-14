# Reflection, Dalamud internals, Win32 interop, localization, proxy

Covers: `ECommons.Reflection` (`DalamudReflector`, `ReflectionHelper`, `FieldPropertyUnion`),
`ECommons.Interop`, `ECommons.WindowsFormsReflector`, `ECommons.LanguageHelpers` and
`ECommons.Networking`. Verified against ECommons `bc38dc7` (2026-11-08, v3.2.1.26, Dalamud
API 15); nothing here was compiled. Paths are relative to the skill root.

Reflection targets inside Dalamud (type, field and method names) were additionally checked against
the pinned Dalamud checkout (`15.0.3.5`); they are internal and can change with any Dalamud release.

Pointers: stock Dalamud services -> `dalamud` skill; IPC without reflection (`EzIPC`) -> the IPC
topic of this skill; Splatoon connection built on `DalamudReflector` -> `references/topic-splatoonapi.md`.

```csharp
using ECommons.Reflection;                      // DalamudReflector, ReflectionHelper extensions
using ECommons.Reflection.FieldPropertyUnion;   // IFieldPropertyUnion, UnionField, UnionProperty
using ECommons.Interop;                         // WindowFunctions, OpenFileDialog, LimitedKeys
using ECommons.WindowsFormsReflector;           // Keys, Winforms.Clipboard
using ECommons.LanguageHelpers;                 // Localization, .Loc()
using ECommons.Networking;                      // ProxySettings, ApplyProxySettings
```

## Module flag

`references/ecommons/ECommons/ECommonsMain.cs`, `references/ecommons/ECommons/Module.cs`.

`enum Module { All, DalamudReflector, ObjectFunctions, ObjectLife, SplatoonAPI, VfxTracking }`.
`ECommonsMain.Init(IDalamudPluginInterface pluginInterface, object instance,
params Module[] modules)` calls the internal `DalamudReflector.Init()` when `modules` contains any
of `Module.All`, `Module.DalamudReflector` or `Module.SplatoonAPI`. The README writes
`Modules.<Module>`; the enum is named `Module`.

`DalamudReflector.Init()` creates the plugin cache, the change-callback list and the delegate used
by `SetKeyState`. `ECommonsMain.Dispose()` calls `DalamudReflector.Dispose()`, which nulls the
cache and the callback list and removes the `Svc.Framework.Update` monitor.

| Needs the module | Works with plain `ECommonsMain.Init(pi, this)` |
|---|---|
| `TryGetDalamudPlugin(...)` with `ignoreCache: false` (the default): throws `Exception` otherwise | `TryGetDalamudPlugin(...)` with `ignoreCache: true` |
| `RegisterOnInstalledPluginsChangedEvents`: `NullReferenceException` otherwise | `GetService`, `GetPluginManager`, `TryGetLocalPlugin`, `TryGetDalamudStartInfo`, `IsOnStaging` |
| `SetKeyState`: `NullReferenceException` otherwise | `HasRepo`, `AddRepo`, `ReloadPluginMasters`, `SaveDalamudConfig`, `GetPluginMaster`, `AddPlugin`, `RemoveCurrentPlugin` |
| | `GetDtrEntryState`, `SetDtrEntryState`, `DeleteSharedData`, `SetImGuiAssertsState`, `GetPluginName` |

Everything in the right column still needs `Svc.PluginInterface`, so `ECommonsMain.Init` must have
run. YesAlready initialises with no modules and calls `DalamudReflector.GetService`; AutoRetainer
uses `Module.DalamudReflector`; PandorasBox uses `ECommons.Module.All`.

## DalamudReflector

`public static class DalamudReflector`, namespace `ECommons.Reflection`.
Source: `references/ecommons/ECommons/Reflection/DalamudReflector.cs`.

### Other plugins

```csharp
public static bool TryGetDalamudPlugin(string internalName, out object instance,
    bool suppressErrors = false, bool ignoreCache = false)
public static bool TryGetDalamudPlugin(string internalName, out object instance,
    out AssemblyLoadContext context, bool suppressErrors = false, bool ignoreCache = false)
public static bool TryGetLocalPlugin(out object localPlugin, out AssemblyLoadContext context,
    out Type type)
public static bool TryGetLocalPlugin(object instance, out object localPlugin,
    out AssemblyLoadContext context, out Type type)
public static void RegisterOnInstalledPluginsChangedEvents(params Action[] actions)
public static object GetPluginManager()
```

`TryGetDalamudPlugin` behaviour:

- Walks `PluginManager.InstalledPlugins`, matches the `InternalName` property, reads the private
  `instance` field of the `LocalPlugin` (for a `LocalDevPlugin` the field is read from the base
  type). `instance` is the other plugin's `IDalamudPlugin` object, typed `object`.
- `context` is `loader.context` of that plugin and may be null.
- Returns false when the plugin is not installed, or installed but not loaded (its instance is
  null). The second case writes `InternalLog.Warning("Found requested plugin ... but it was null")`
  regardless of `suppressErrors`.
- `suppressErrors` only silences the two `PluginLog.Error` lines written when the lookup throws.
- With `ignoreCache: false`: a null cache throws `Exception("PluginCache is null. Have you
  initialised the DalamudReflector module on ECommons initialisation?")`. This throw is outside the
  `try`, so `suppressErrors` does not stop it. The first cached call also starts plugin monitoring
  (adds a `Svc.Framework.Update` handler and logs at Information level).
- A successful lookup is stored in the cache when the cache exists, even with `ignoreCache: true`;
  `ignoreCache` only skips reading it.
- The cache is cleared whenever the installed plugin set changes (see below), on the next tick. In
  between, a cached entry can refer to a plugin instance that was just unloaded.

`TryGetLocalPlugin` returns the `LocalPlugin` object of the calling plugin. The overload that takes
`instance` ignores that argument and always compares against `ECommonsMain.Instance`. Exceptions are
logged with `e.Log()` and the method returns false.

`RegisterOnInstalledPluginsChangedEvents` appends actions; there is no unregister method. Monitoring
compares `Svc.PluginInterface.InstalledPlugins` (`InternalName`, `Version`, `IsLoaded`) with the
previous snapshot on every framework update. The first comparison runs against an empty snapshot, so
the actions fire once shortly after monitoring starts. Actions are run through a `TickScheduler`
after the cache is cleared, and the change is logged at Verbose level.

Consumer patterns (all three pass `ignoreCache: true`):

```csharp
// PandorasBox/IPC/TeleporterIPC.cs: presence check, silent
if (DalamudReflector.TryGetDalamudPlugin("TeleporterPlugin", out var plugin, true, true))
    return true;

// AutoRetainer/Helpers/YesAlready.cs: read and write a static of another plugin
if(DalamudReflector.TryGetDalamudPlugin("Yes Already", out var pl, false, true))
{
    pl.GetStaticFoP("YesAlready.Service", "Configuration").SetFoP("Enabled", false);
}

// AutoRetainer/Helpers/Utils.cs: call a static method in another plugin's assembly
if(DalamudReflector.TryGetDalamudPlugin("NotificationMaster", out var instance, true, true))
{
    instance.GetType().Assembly.GetType("NotificationMaster.TrayIconManager", true)
        .GetMethod("ShowToast").Invoke(null, new object[] { s, P.Name });
}
```

### Dalamud services and configuration

```csharp
public static object GetService(string serviceFullName)
public static bool TryGetDalamudStartInfo(out DalamudStartInfo dalamudStartInfo,
    IDalamudPluginInterface pluginInterface = null)
public static bool IsOnStaging()
public static string GetPluginName()
public static bool HasRepo(string repoURL)
public static void AddRepo(string repoURL, bool enabled)
public static void ReloadPluginMasters()
public static void SaveDalamudConfig()
public static bool GetDtrEntryState(string name)
public static void SetDtrEntryState(string name, bool enable)
public static void DeleteSharedData(string name)
public static void SetImGuiAssertsState(bool? enableAsserts = false,
    bool? enableVerboseAssertLogging = false, bool? enableAssertsAtStartup = false)
public static void SetKeyState(VirtualKey key, int state)
```

| Member | Behaviour |
|---|---|
| `GetService` | `Dalamud.Service<T>.Get()` for the type named by `serviceFullName` in the Dalamud assembly. Throws if the type does not exist (`GetType(name, true)`). Example name: `"Dalamud.Configuration.Internal.DalamudConfiguration"`. |
| `GetPluginManager` | Same, fixed to `Dalamud.Plugin.Internal.PluginManager`. |
| `TryGetDalamudStartInfo` | Reads `StartInfo` of the `Dalamud.Dalamud` service. Type is `Dalamud.Common.DalamudStartInfo`. On failure logs an error and returns false with `default`. |
| `IsOnStaging` | Reads the file at `DalamudStartInfo.ConfigurationPath` as JSON and returns true when `DalamudBetaKind` is non-empty and not `"release"`. Not wrapped in try/catch; AutoRetainer wraps the call. |
| `GetPluginName` | `Svc.PluginInterface?.InternalName ?? "Not initialized"`. |
| `HasRepo` | Scans `DalamudConfiguration.ThirdRepoList`; URLs compared trimmed, without trailing `/`, case-insensitive. |
| `AddRepo` | Returns silently if the repo exists. Otherwise creates `Dalamud.Configuration.ThirdPartyRepoSettings`, sets `Url` and `IsEnabled`, adds it to the list. Does not save and does not reload; call `SaveDalamudConfig()` and `ReloadPluginMasters()` yourself. |
| `ReloadPluginMasters` | Invokes `PluginManager.SetPluginReposFromConfigAsync(true)` and does not await the returned task. |
| `SaveDalamudConfig` | Invokes `DalamudConfiguration.QueueSave()`. |
| `GetDtrEntryState` | True when `name` is absent from `DalamudConfiguration.DtrIgnore`. Returns false on exception (logged). |
| `SetDtrEntryState` | Adds or removes `name` in `DtrIgnore` and calls `QueueSave` only when the state changes. |
| `DeleteSharedData` | Removes `name` from the private `caches` dictionary of `Dalamud.Plugin.Ipc.Internal.DataShare` and from `EzSharedData.Cache`. |
| `SetImGuiAssertsState` | Null means "leave unchanged". Note the defaults are `false`, so a parameterless call turns all three off. Writes `ImGuiAssertsEnabledAtStartup` (config, saved) and `ShowAsserts` / `EnableVerboseAssertLogging` (`InterfaceManager`). Exceptions are logged. |
| `SetKeyState` | Writes through `KeyState.GetRefValue(int)` (private, bound in `Init`). Needs the module. |

In pinned Dalamud `DtrIgnore` is `List<string>?` and `ImGuiAssertsEnabledAtStartup` is `bool?`; a
null value makes the corresponding ECommons call throw inside its own try/catch and log.

YesAlready (`UI/MainWindow.cs`) does the DTR toggle by hand with the same calls that
`GetDtrEntryState`/`SetDtrEntryState` wrap:

```csharp
var config = DalamudReflector.GetService("Dalamud.Configuration.Internal.DalamudConfiguration");
var dtrList = config.GetFoP<List<string>>("DtrIgnore");
// ... dtrList.Add / dtrList.Remove ...
config.Call("QueueSave", []);
```

AutoRetainer (`AutoRetainer.cs`):

```csharp
ECommonsMain.Init(pi, this, Module.DalamudReflector);
var gameVersion = DalamudReflector.TryGetDalamudStartInfo(out var ver)
    ? ver.GameVersion.ToString() : "unknown";
```

### Installing and removing plugins

```csharp
public static async Task<List<object>?> GetPluginMaster(string masterURL)
public static async Task<bool> AddPlugin(string masterURL, string pluginInternalName)
public static void RemoveCurrentPlugin()
```

- `GetPluginMaster` builds an internal `PluginRepository(happyHttpClient, masterURL, true)`, awaits
  `ReloadAsync`, and returns the `PluginMaster` entries (`RemotePluginManifest` objects typed
  `object`). Returns null and logs an error on failure.
- `AddPlugin` finds the manifest by `InternalName`, adds the repo if missing (enabled), calls
  `ReloadPluginMasters()`, then `PluginManager.InstallPluginAsync(manifest, false,
  PluginLoadReason.Installer)`. Returns true only when the resulting plugin reports `IsLoaded`.
  Every failure path logs an error and returns false.
- `RemoveCurrentPlugin` schedules, through `Svc.Framework.RunOnTick`, `LocalPlugin.UnloadAsync(
  PluginLoaderDisposalMode.WaitBeforeDispose)` followed by `PluginManager.RemovePlugin` for the
  calling plugin. It unloads and deletes your own plugin.

### CachedPluginEntry

`internal class CachedPluginEntry` with internal fields `object Plugin` and
`AssemblyLoadContext Context`. Not usable from plugin code.
Source: `references/ecommons/ECommons/Reflection/CachedPluginEntry.cs`.

## ReflectionHelper

`public static partial class ReflectionHelper`, namespace `ECommons.Reflection`. Sources:
`references/ecommons/ECommons/Reflection/ReflectionHelper/ReflectionHelper.cs`, `Call.cs`,
`CallStatic.cs`, `Utils.cs` in the same folder.

```csharp
public const BindingFlags AllFlags =
    BindingFlags.Public | BindingFlags.NonPublic | BindingFlags.Instance | BindingFlags.Static;
public const BindingFlags StaticFlags =
    BindingFlags.Public | BindingFlags.NonPublic | BindingFlags.Static;
public const BindingFlags InstanceFlags =
    BindingFlags.Public | BindingFlags.NonPublic | BindingFlags.Instance;
```

### Field or property access ("FoP")

```csharp
public static object GetFoP(this object obj, string name)
public static T GetFoP<T>(this object obj, string name)
public static void SetFoP(this object obj, string name, object value)
public static object GetStaticFoP(this object obj, string type, string name)
public static object GetStaticFoP(this Type obj, string name)
public static T GetStaticFoP<T>(this object obj, string type, string name)
public static void SetStaticFoP(this object obj, string type, string name, object value)
```

| Member | Lookup | Not found |
|---|---|---|
| `GetFoP` | Field first, then property, `AllFlags`, on the runtime type and then each base type up the chain. | Returns null. Also returns null when `obj` or `name` is null. |
| `GetFoP<T>` | Cast of `GetFoP`. | Null cast to a value type `T` throws `NullReferenceException`. |
| `SetFoP` | Field first, then property, `AllFlags`, on the runtime type only (no base-type walk, so private members of a base class are not found). | `NullReferenceException`. |
| `GetStaticFoP(this object, type, name)` | `type` is a full type name resolved in the assembly of `obj.GetType()`. Field first, then property, `StaticFlags`. | Unknown type throws `NullReferenceException`; unknown member returns null. A static field holding null falls through to a property of the same name. |
| `GetStaticFoP(this Type, name)` | Same on the given `Type`. There is no generic form of this overload. | Returns null. |
| `SetStaticFoP` | As `GetStaticFoP(this object, ...)`. There is no `this Type` overload. | `NullReferenceException`. |

The four static variants carry the source comment "Subject for future refactoring and changes!".

### Calling methods

```csharp
// ReflectionHelper.cs
public static object Call(this object obj, string name, object[] @params,
    bool matchExactArgumentTypes = false)
public static T Call<T>(this object obj, string name, object[] @params,
    bool matchExactArgumentTypes = false)

// Call.cs: generic-capable forms
public static object Call(this object instance, string methodName,
    IEnumerable<string> methodTypeArguments, object[] parameters)
public static object Call(this object obj, IEnumerable<Assembly>? assemblies, string methodName,
    IEnumerable<string>? methodTypeArguments, object[] parameters)
public static T Call<T>(this object obj, IEnumerable<Assembly> assemblies, string methodName,
    IEnumerable<string> methodTypeArguments, object[] parameters)

// CallStatic.cs
public static object CallStatic(this object obj, string type, string name, object[] values)
public static T CallStatic<T>(this object obj, string type, string name, object[] values)
public static object CallStatic(IEnumerable<Assembly> assemblies, string typeName,
    IEnumerable<string> typeArguments, string methodName,
    IEnumerable<string> methodTypeArguments, object[] parameters)
```

| Form | Method resolution | Not found |
|---|---|---|
| `Call(name, @params, false)` | `GetMethod(name, AllFlags)`: by name only. | Overloaded name throws `AmbiguousMatchException`; missing method throws `NullReferenceException`. |
| `Call(name, @params, true)` | `GetMethod(name, AllFlags, types)` where types are `@params[i].GetType()`. A null argument throws; an argument of a derived type does not match a base-typed parameter exactly unless the binder accepts it. | `NullReferenceException`. |
| `Call(methodName, methodTypeArguments, parameters)` | Exact runtime argument types. Type argument names are full type names searched in the instance's own assembly only. | Returns null, no exception. Also null when `MakeGenericMethod` fails. |
| `Call(assemblies, methodName, methodTypeArguments, parameters)` | Same, type argument names searched in `assemblies`. | Returns null. |
| `CallStatic(this object, type, name, values)` | Type resolved in the assembly of `obj`; exact runtime argument types. | `NullReferenceException`. |
| `CallStatic(assemblies, ...)` | Not an extension method. `typeArguments` and `methodTypeArguments` are type names (null for non-generic). | Returns null. |

There is no generic `Call<T>` for the three-argument `IEnumerable<string>` form.

Overload selection with collection expressions: `obj.Call("QueueSave", [])` and
`obj.Call("UserDisableTweak", [x], true)` bind to the `object[]` form; a call with two collection
arguments, `obj.Call("Method", ["Full.Type.Name"], [arg])`, binds to the generic-capable form.

```csharp
// AutoRetainer/Helpers/Utils.cs (inside a commented-out block in the pinned commit):
// static generic method HaselCommon.Service.Get<HaselTweaks.Windows.PluginWindow>()
var configWindow = ReflectionHelper.CallStatic(context.Assemblies, "HaselCommon.Service", [],
    "Get", ["HaselTweaks.Windows.PluginWindow"], []);
configWindow.GetFoP("TweakManager").Call("UserDisableTweak", [x], true);
```

### Type and method search, delegates

```csharp
public static MethodInfo FindStaticMethodInAssemblies(IEnumerable<Assembly> assemblies,
    string typeName, Type[] typeArguments, string methodName, Type[] methodTypeArguments,
    Type[] parameterTypes)
public static List<Type> FindTypesInAssemblies(IEnumerable<Assembly> assemblies,
    IEnumerable<string> typeNames)
public static List<Type> FindTypesInAssemblies(IEnumerable<Assembly> assemblies,
    IEnumerable<(string TypeName, Type[] TypeArguments)> typeNames)
public static Type[] GetTypes(this IEnumerable<object> objects)
public static Delegate CreateDelegate(MethodInfo methodInfo, object target)
public static Type GetTypeFromRuntimeAssembly(string assemblyName, string type)
public static MethodInfo? GetAccessibleSetMethod(PropertyInfo property)
public static DelegateInfo AnalyzeDelegateField(IFieldPropertyUnion field)
public static void AssignDelegateToField(IFieldPropertyUnion field, object target,
    Delegate delegateInstance)
```

- `FindTypesInAssemblies` returns only the types it found, in request order, first assembly wins.
  A missing type is skipped, so compare the result count with the request count.
- `FindStaticMethodInAssemblies` returns null when nothing matches.
- `CreateDelegate` builds an `Action<...>` or `Func<...>` delegate type from the method signature.
  For instance methods it binds by `methodInfo.Name` on `target`. `EzIPC` uses it internally.
- `GetTypeFromRuntimeAssembly` does `Assembly.Load(assemblyName).GetType(type)`; on exception it
  logs and returns null.
- `GetAccessibleSetMethod` returns the setter including non-public ones, and falls back to the
  setter declared on `property.DeclaringType`.
- `ReflectionHelper.DelegateInfo` (nested class): `bool IsDelegate`, `Type[] ParameterTypes`,
  `Type ReturnType` (null for void), `Type DelegateType`.
- `AssignDelegateToField` throws `ArgumentNullException` for a null field, a null delegate or a
  null target on a non-static member, and `ArgumentException` when the member is not a delegate
  type. A delegate of a different type is re-created as the member's delegate type.

## FieldPropertyUnion

Namespace `ECommons.Reflection.FieldPropertyUnion`. Sources:
`references/ecommons/ECommons/Reflection/FieldPropertyUnion/IFieldPropertyUnion.cs`,
`UnionField.cs`, `UnionProperty.cs`.

```csharp
// in ReflectionHelper (Utils.cs)
public static IFieldPropertyUnion[] GetFieldPropertyUnions(this Type type,
    BindingFlags bindingFlags = BindingFlags.Default, bool propertyFirst = false)
public static IFieldPropertyUnion? GetFieldPropertyUnion(this Type type, string name,
    BindingFlags bindingFlags = BindingFlags.Default)
```

`IFieldPropertyUnion` members: `Type UnionType`, `bool IsSpecialName`,
`object? GetRawConstantValue()`,
`object? GetValue(object? obj)`, `void SetValue(object? obj, object? value)`,
`void SetValue(object? obj, object? value, BindingFlags invokeAttr, Binder? binder,
CultureInfo? culture)`, `string Name`, `MemberTypes MemberType`, `Type? DeclaringType`,
`Type? ReflectedType`, `System.Reflection.Module Module`, `bool IsDefined(Type attributeType,
bool inherit)`, `object[] GetCustomAttributes(bool inherit)`,
`object[] GetCustomAttributes(Type attributeType, bool inherit)`,
`T? GetCustomAttribute<T>() where T : Attribute`,
`IEnumerable<T> GetCustomAttributes<T>() where T : Attribute`,
`IEnumerable<CustomAttributeData> CustomAttributes`, `bool IsCollectible`, `bool IsStatic`.

- `UnionField(FieldInfo fieldInfo)` exposes `public readonly FieldInfo FieldInfo`;
  `UnionProperty(PropertyInfo propertyInfo)` exposes `public readonly PropertyInfo PropertyInfo`.
- `UnionProperty.SetValue` goes through `GetAccessibleSetMethod` and throws
  `InvalidOperationException` when no setter exists. `UnionProperty.IsStatic` throws
  `InvalidOperationException` when the property has neither getter nor setter.
- Compiler-generated backing fields are not filtered out, so an auto-property appears twice with
  non-public flags: once as the property and once as its `<Name>k__BackingField` field.
- `GetFieldPropertyUnion` tries the field first, then the property, and returns null when neither
  exists.

```csharp
// ECommons/Singletons/SingletonServiceManager.cs and AutoRetainer EntrustManager.cs
foreach(var member in type.GetFieldPropertyUnions(ReflectionHelper.AllFlags))
{
    var value = member.GetValue(instanceOrNull);
}
```

## Interop

### WindowFunctions

`public static unsafe class WindowFunctions`, namespace `ECommons.Interop`.
Source: `references/ecommons/ECommons/Interop/WindowFunctions.cs`. `HWND` is
`TerraFX.Interop.Windows.HWND`.

```csharp
public const int SW_MINIMIZE = 6;
public const int SW_FORCEMINIMIZE = 11;
public const int SW_HIDE = 0;
public const int SW_SHOW = 5;
public const int SW_SHOWNA = 8;
public static char* FFXIVClassNamePtr { get; }
public static bool TryFindGameWindow(out HWND hwnd)
public static bool ApplicationIsActivated()
public static bool SendKeypress(int keycode)
```

- `TryFindGameWindow` enumerates top-level windows of class `FFXIVGAME` with `FindWindowEx` and
  returns the one owned by the current process id. Safe with several game clients running.
- `ApplicationIsActivated` is true when the foreground window belongs to the current process.
- `SendKeypress` sends `WM_KEYDOWN` then `WM_KEYUP` with `SendMessage` to the game window; returns
  false when the window was not found. The `LimitedKeys` overloads and modifier handling live in
  `ECommons.Automation.WindowsKeypress` (automation topic).
- `FFXIVClassNamePtr` allocates unmanaged memory once and registers the free with `Purgatory`,
  which `ECommonsMain.Dispose()` purges.
- `ECommonsMain.MainWindowHandle` (`nint*`) is a separate, signature-scanned pointer.

### NativeFunctions

`public static unsafe class NativeFunctions`.
Source: `references/ecommons/ECommons/Interop/NativeFunctions.cs`.

Usable: `ushort LOWORD(uint value)`, `ushort HIWORD(uint value)`,
`uint MAKEWPARAM(ushort low, ushort high)`.

Everything else is `[Obsolete(..., true)]` (compile error on use) with a stub body returning
`default`. Replacement in every case is the member of the same name on
`TerraFX.Interop.Windows.Windows`: `SendMessage`, `FindWindowEx`, `GetWindowThreadProcessId`,
`GetForegroundWindow`, `GetKeyState`, `GetAsyncKeyState`, `ShowWindow`, `ClientToScreen`,
`SetCursorPos`, `SetWindowText`, `SendInput`.

### OpenFileDialog, OpenFileName

Sources: `references/ecommons/ECommons/Interop/OpenFileDialog.cs`,
`references/ecommons/ECommons/Interop/OpenFileName.cs`.

```csharp
public static bool IsSelecting()
public static void SelectFile(Action<OpenFileName> successCallback, Action cancelCallback = null,
    string initialDir = null, string title = "Select a file...",
    IEnumerable<(string Description, IEnumerable<string> Extensions)> fileTypes = null,
    bool save = false)
public static void SelectFile(OpenFileName ofn, Action<OpenFileName> successCallback,
    Action cancelCallback = null, string initialDir = null, string title = "Select a file...",
    IEnumerable<(string Description, IEnumerable<string> Extensions)> fileTypes = null,
    bool save = false)
```

- Native Win32 dialog (`Comdlg32.dll` `GetOpenFileName`, or `GetSaveFileName` when `save` is true),
  owned by the game window when found.
- Runs on a new `Thread`. Both callbacks are invoked on that thread, not on the framework thread;
  marshal back before touching game state.
- One dialog at a time per plugin: a second call while one is open shows
  `Notify.Error("Failed to open file dialog")` and does nothing. `IsSelecting()` reports it.
- `Extensions` are given without dot or wildcard (`"csv"`); ECommons prefixes `*.` and joins
  with `;`.
- The selected path is `OpenFileName.file` (buffer of 1024 chars); `fileTitle` is the file name
  (256 chars). In save mode no extension is appended for you.
- `SelectFile` overwrites `dlgOwner`, `structSize`, `filter`, `file`, `maxFile`, `fileTitle`,
  `maxFileTitle`, `initialDir` and `title` of the passed `OpenFileName`. Other public fields
  (`flags`, `defExt`, `filterIndex`, `flagsEx`, ...) are kept, which is the reason for the overload.
- Exceptions inside the thread are logged and shown with `Notify.Error`.
- When `fileTypes` is given, the filter string ends with an "All files" entry that is preceded by
  two NUL characters. By the Win32 filter format a double NUL ends the list, so that entry is
  probably never shown *(unverified, inferred from the string built in source)*.

`OpenFileName` is a `[StructLayout(LayoutKind.Sequential, CharSet = CharSet.Auto)]` class mirroring
`OPENFILENAME` with public fields: `structSize`, `dlgOwner`, `instance`, `filter`, `customFilter`,
`maxCustFilter`, `filterIndex`, `file`, `maxFile`, `fileTitle`, `maxFileTitle`, `initialDir`,
`title`, `flags`, `fileOffset`, `fileExtension`, `defExt`, `custData`, `hook`, `templateName`,
`reservedPtr`, `reservedInt`, `flagsEx`.

```csharp
// AutoRetainer/UI/NeoUI/DeployablesTab.cs
OpenFileDialog.SelectFile(x =>
{
    var name = x.file;
    if(!name.EndsWith(".csv", StringComparison.OrdinalIgnoreCase))
    {
        name = $"{name}.csv";
    }
    Utils.WriteCsv(name, headers, data);
}, title: "Save as...", fileTypes: [("Comma-separated values", ["csv"])], save:true);
```

### LimitedKeys

`public enum LimitedKeys`, namespace `ECommons.Interop`, file `ECommons/Interop/LimitedKeys.cs`.
Not in the reference copy; open it in an ECommons checkout. It carries
`[Obfuscation(Exclude = true, ApplyToMembers = true)]`.

It is the Windows virtual-key list with one name per value and readable names, without the modifier
flag members. Members whose names differ from `Keys`:

| `LimitedKeys` | value | `Keys` name |
|---|---|---|
| `LeftMouseButton`, `RightMouseButton`, `MiddleMouseButton` | 1, 2, 4 | `LButton`, `RButton`, `MButton` |
| `XMouseButton1`, `XMouseButton2` | 5, 6 | `XButton1`, `XButton2` |
| `Digit_0` ... `Digit_9` | 48-57 | `D0` ... `D9` |
| `LeftWin`, `RightWin` | 91, 92 | `LWin`, `RWin` |
| `LeftShiftKey`, `RightShiftKey` | 160, 161 | `LShiftKey`, `RShiftKey` |
| `LeftControlKey`, `RightControlKey` | 162, 163 | `LControlKey`, `RControlKey` |
| `LeftAltKey`, `RightAltKey` | 164, 165 | `LMenu`, `RMenu` |

Shared names include `None` (0), `Back`, `Tab`, `Enter`, `Pause`, `CapsLock`, `Escape`, `Space`,
`PageUp`, `PageDown`, `End`, `Home`, `Left`, `Up`, `Right`, `Down`, `PrintScreen`, `Insert`,
`Delete`, `A`-`Z`, `NumPad0`-`NumPad9`, `Multiply`, `Add`, `Separator`, `Subtract`, `Decimal`,
`Divide`, `F1`-`F24`, `NumLock`, `Scroll`, the `Browser*`, `Volume*`, `Media*` and `Launch*` keys
and the `Oem*` keys. `LimitedKeys` has no `ShiftKey` (16), `ControlKey` (17), `Menu` (18), `Return`,
`Capital`, `NoName`, `KeyCode`, `Modifiers`, `Shift`, `Control` or `Alt`.

Consumers take it from `GenericHelpers` and `WindowsKeypress`:

```csharp
// ECommons/GenericHelpers/GenericHelpers.cs
public static bool IsKeyPressed(LimitedKeys key)
public static bool IsKeyPressed(IEnumerable<LimitedKeys> keys)   // any of them
public static bool IsKeyPressed(int key)
[Obsolete("Use IsKeyPressed overload")]
public static bool IsAnyKeyPressed(IEnumerable<LimitedKeys> keys)

// AutoRetainer/PluginData/Config.cs: stored in configuration as a keybind
public LimitedKeys Suppress = LimitedKeys.LeftControlKey;
// AutoRetainer/Internal/ContextMenuManager.cs
if(IsKeyPressed([LimitedKeys.LeftControlKey, LimitedKeys.RightControlKey])
    && IsKeyPressed([LimitedKeys.RightShiftKey, LimitedKeys.LeftShiftKey]))
// PandorasBox/Features/Other/ATMSolver.cs
WindowsKeypress.SendKeypress(ECommons.Interop.LimitedKeys.Escape);
```

## WindowsFormsReflector

README "v3 changes": ECommons 3.0.0.0 and higher no longer reference Windows Forms in any way. The
`RELEASEFORMS` and `DEBUGFORMS` build configurations were removed. Code that used
`System.Windows.Forms.Keys` with ECommons must use `ECommons.WindowsFormsReflector.Keys`. Internal
copy/paste methods call Windows Forms through reflection.

### Keys

`public enum Keys`, namespace `ECommons.WindowsFormsReflector`, file
`ECommons/WindowsFormsReflector/Keys.cs`. Not in the reference copy; open it in an ECommons
checkout. It is a member-for-member copy of `System.Windows.Forms.Keys`, including aliases
(`Return`/`Enter` = 13, `Capital`/`CapsLock` = 20, `Prior`/`PageUp` = 33, `Next`/`PageDown` = 34,
`Snapshot`/`PrintScreen` = 44, `Oem1`...`Oem7`, `Oem102`) and the flag members
`Modifiers = -65536`, `KeyCode = 65535`, `Shift = 65536`, `Control = 131072`, `Alt = 262144`.

`GenericHelpers.IsKeyPressed(Keys key)` and `GenericHelpers.IsKeyPressed(IEnumerable<Keys> keys)`
(`references/ecommons/ECommons/GenericHelpers/GenericHelpers.Forms.cs`) cast the value to `int` and
pass it to the Win32 key-state call, so the modifier flag members are not valid there. Prefer
`LimitedKeys` for new keybind settings.

### Winforms.Clipboard

`public static partial class Winforms` with nested `public static class Clipboard`. Sources:
`references/ecommons/ECommons/WindowsFormsReflector/Winforms.cs`,
`references/ecommons/ECommons/WindowsFormsReflector/Winforms.Clipboard.cs`.

```csharp
public static Type Instance { get; }      // System.Windows.Forms.Clipboard, resolved on each access
public static void Clear()
public static void SetText(string text)
public static string GetText()
```

`Instance` calls `ReflectionHelper.GetTypeFromRuntimeAssembly("System.Windows.Forms",
"System.Windows.Forms.Clipboard")`. If the assembly cannot be loaded it is null and the three
methods throw `NullReferenceException`. `Winforms` has no other members. Plugin code normally uses
the wrappers, which catch and report errors:

```csharp
public static bool Copy(string text, bool silent = false)   // GenericHelpers; empty text clears
public static string? Paste(bool silent = false)            // GenericHelpers; null on failure
```

## LanguageHelpers

Sources: `references/ecommons/ECommons/LanguageHelpers/Localization.cs`,
`references/ecommons/ECommons/LanguageHelpers/LocalizationExtensions.cs`.

```csharp
public static class Localization
{
    public static string PararmeterSymbol = "??";      // spelling as in source
    public static string Separator = "==";
    public static string CurrentLanguage { get; internal set; }
    public static bool Logging = false;
    public static void Init(string Language = null)
    public static List<string> GetAvaliableLanguages(bool rescan = false)   // spelling as in source
    public static string GameLanguageString { get; }
    public static void Save(string lang)
    public static string GetLocFileLocation(string lang)
    public static string Loc(string s)
    public static string Loc(string s, params object[] values)
}
public static class LocalizationExtensions
{
    public static string Loc(this string s)
    public static string Loc(this string s, params object[] values)
}
```

- The English source string is the key. `"text".Loc()` returns the translation when one exists and
  is non-empty, otherwise the input.
- `Loc(values)` then replaces each `??` in order with `value.ToString()` (first occurrence per
  value). A null value throws.
- Translation file: `Language<lang>.ini` in the plugin assembly directory
  (`Svc.PluginInterface.AssemblyLocation.DirectoryName`). UTF-8, one entry per line,
  `<english>==<translation>`. A line break inside a string is written as the two characters `\n`.
  A line that does not split into exactly two parts (empty lines, text containing `==`) is skipped
  with a warning. Duplicate keys log a warning, last one wins.
- `Init(null)` or a missing file clears the table and leaves `CurrentLanguage` null. `Init` is not
  called by `ECommonsMain.Init`; without it every `.Loc()` returns its input.
- `GetAvaliableLanguages` returns `"English"` plus one entry per `Language*.ini` file, cached until
  `rescan` is true.
- `GameLanguageString` maps `Svc.Data.Language` to `"Japanese"`, `"French"`, `"German"`,
  `"Chinese"` (client language value 4) or `"English"`.
- `Logging = true` makes every untranslated string that passes through `.Loc()` get recorded with
  an empty translation; `Save(lang)` then writes the whole table to the language file, which
  produces a template for translators. `Save` overwrites the file.
- Not thread-safe (plain `Dictionary`), and `.Loc()` with `Logging` on mutates it.

```csharp
Localization.Init(Localization.GameLanguageString);
ImGui.Text("Processed ?? of ?? items".Loc(done, total));
```

No usage of `.Loc()` exists in the three consumer checkouts; the example is assembled from the
signatures. Splatoon itself uses it (see the `Element.cs` copy in the `splatoonpresets` skill).

## Networking

Sources: `references/ecommons/ECommons/Networking/ProxySettings.cs`,
`references/ecommons/ECommons/Networking/HttpClientProxyHelper.cs`.

```csharp
[Serializable]
public class ProxySettings
{
    [JsonProperty("UseProxy")] public bool UseProxy = false;
    [JsonProperty("ProxyAddress")] public string ProxyAddress = "";
    [JsonProperty("UseProxyAuthentication")] public bool UseProxyAuthentication = false;
    [JsonProperty("BypassLocal")] public bool BypassLocal = true;
    [JsonProperty("ProxyLogin")] public string ProxyLogin = "";
    [JsonProperty("ProxyPassword")] public string ProxyPassword = "";
    public void ImGuiDraw()                      // inside a "Proxy Settings" collapsing header
    public void ImGuiDrawNoCollapsingHeader()    // bare table
}
public static class HttpClientProxyHelper
{
    public static HttpClient ApplyProxySettings(this HttpClient client, ProxySettings settings)
}
```

Intended use, from the source doc comment: keep a `ProxySettings` instance in the plugin
configuration, draw it with one of the two draw methods, and call `ApplyProxySettings` on the
`HttpClient` immediately after constructing it.

```csharp
public ProxySettings Proxy = new();                       // in the configuration class
var client = new HttpClient().ApplyProxySettings(C.Proxy);
C.Proxy.ImGuiDraw();                                      // in the settings UI
```

- `ApplyProxySettings` reads the private `_handler` field of `HttpClient` and casts it to
  `HttpClientHandler`, then sets `Proxy` to a `WebProxy` (or to `default` when `UseProxy` is
  false). It works for `new HttpClient()` and for a client built on an `HttpClientHandler`. With
  any other handler type the cast fails.
- Every failure (wrong handler type, invalid `ProxyAddress` URI, client already used for a request)
  is caught and logged with `e.Log()`; the client is returned unchanged and no error reaches the
  caller.
- Settings are applied once. Changing `ProxySettings` later does not affect an existing client.
- `ProxyPassword` is stored as plain text in whatever file holds the configuration.
- No usage exists in the three consumer checkouts; the example is assembled from the signatures.

## Gotchas

- `TryGetDalamudPlugin` with default arguments throws (not returns false) when the
  `DalamudReflector` module was not requested. Either request the module or pass
  `ignoreCache: true`, as all three consumers do.
- `internalName` is the plugin's Dalamud `InternalName`, matched exactly and case-sensitively.
- The returned plugin instance lives in another `AssemblyLoadContext`. Its types are not your types
  even when the source is the same; pass only BCL or Dalamud types across, and keep using
  reflection (`GetFoP`, `Call`) on the results. Do not cache the instance across plugin reloads;
  look it up again or clear your copy in a `RegisterOnInstalledPluginsChangedEvents` action.
- For supported cross-plugin calls prefer IPC; reflection on another plugin breaks whenever that
  plugin renames a member.
- `GetFoP` returns null both for "member holds null" and "member does not exist". `GetFoP<int>` on
  a missing member throws `NullReferenceException`, which is easy to misread as a null target.
- `SetFoP` does not walk base types while `GetFoP` does.
- `Call` with the default `matchExactArgumentTypes: false` fails on overloaded method names.
- `GetFieldPropertyUnions()` and `GetFieldPropertyUnion(name)` default to `BindingFlags.Default`,
  which is 0. .NET member lookup needs an instance/static flag and a public/non-public flag to
  return anything, so a call without flags yields an empty array or null. Always pass flags, for
  example `ReflectionHelper.AllFlags`. AutoRetainer's `ExpertTab.cs` calls it without flags.
- `TryGetLocalPlugin(object instance, ...)` ignores `instance`.
- `AddRepo` neither saves nor reloads.
- `SetImGuiAssertsState()` with no arguments disables asserts; pass null for the values to keep.
- `NativeFunctions` P/Invoke wrappers are compile errors now; use `TerraFX.Interop.Windows.Windows`.
- `OpenFileDialog` callbacks run on a background thread.
- `GenericHelpers.UseAsyncKeyCheck` is inverted relative to its name in the pinned source: when
  true, `IsKeyPressed` calls `GetKeyState`; when false (default) it calls `GetAsyncKeyState`.
- `ECommons.WindowsFormsReflector.Keys` and `ECommons.Interop.LimitedKeys` share many member names.
  Importing both namespaces is fine, but the enums are not interchangeable without a cast, and
  several names differ (table above).
- `Localization.Init` must be called by the plugin; member names `PararmeterSymbol` and
  `GetAvaliableLanguages` are misspelled in source and must be typed that way.
