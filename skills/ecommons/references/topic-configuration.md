# Configuration (EzConfig)

Covers: `ECommons.Configuration`: `EzConfig`, `EzCharaConfig<T>`, `ISerializationFactory` /
`DefaultSerializationFactory`, `ExternalWriter`, `IgnoreDefaultValueAttribute`, the obsolete
`IEzConfig`. Verified against ECommons `bc38dc7` (2026-11-08, v3.2.1.26, Dalamud
API 15); nothing here was compiled. Paths are relative to the skill root.

Stock Dalamud configuration (`IPluginConfiguration`, `GetPluginConfig`, `SavePluginConfig`) is
covered by the `dalamud` skill. EzConfig is an alternative to it, not a wrapper around it.

```csharp
using ECommons.Configuration;   // EzConfig, EzCharaConfig<T>, ISerializationFactory, ...
using ECommons;                 // ECommonsMain
```

Source: `references/ecommons/ECommons/Configuration/*.cs` (8 files),
`references/ecommons/ECommons/ECommonsMain.cs`.

## What it is and where files go

`EzConfig` is a static class that holds one "default" configuration object (`EzConfig.Config`) and
offers load/save helpers for any number of additional files. The config type needs only a public
parameterless constructor (`where T : new()`); no interface, no `Version` property.

| Thing | Location |
|---|---|
| EzConfig default file | `<pluginConfigs>/<InternalName>/DefaultConfig.json` |
| Stock Dalamud config (`Svc.PluginInterface.ConfigFile`) | `<pluginConfigs>/<InternalName>.json` |
| Additional files, `appendConfigDirectory: true` | `<pluginConfigs>/<InternalName>/<path>` |
| `EzCharaConfig<T>` files | `<pluginConfigs>/<InternalName>/<prefix><CID:X16>.json` |

The directory is `EzConfig.GetPluginConfigDirectory()`, which is
`Svc.PluginInterface.GetPluginConfigDirectory()` (the same path as
`Svc.PluginInterface.ConfigDirectory.FullName`) unless `PluginConfigDirectoryOverride` is set.
Dalamud creates that directory when asked for it. `DefaultConfig.json` is the value of
`DefaultSerializationFactory.DefaultConfigFileName`; a custom factory can change it.

The two systems use different files. A plugin that calls `EzConfig.Init<T>()` does not read the
stock `<InternalName>.json` unless `EzConfig.Migrate<T>()` is called first.

## EzConfig

`public static class EzConfig`, `references/ecommons/ECommons/Configuration/EzConfig.cs`.

| Member | Notes |
|---|---|
| `static object? Config { get; private set; }` | The default configuration instance. `null` until `Init<T>()` or `Set<T>()`. |
| `static T Init<T>() where T : new()` | `Config = LoadConfiguration<T>(DefaultSerializationFactory.DefaultConfigFileName)`; returns it. |
| `static void Migrate<T>() where T : new()` | One-time import of the stock Dalamud config file. Must run before `Init`. See below. |
| `static void Save()` | If `Config != null`: `SaveConfiguration(Config, DefaultConfigFileName, true)` (pretty-printed), then invokes `OnSave` inside `GenericHelpers.Safe`. No-op when `Config` is null. |
| `static T Get<T>()` | `(T)Config`. Throws `InvalidCastException` on wrong type; for a null `Config` returns null for reference types and throws `NullReferenceException` for value types. |
| `static T Set<T>(T newReference)` | Replaces `Config` with another instance and returns it. Does not save. |
| `static event Action? OnSave` | Raised after each `Save()` (not after `SaveConfiguration` of other files). Cleared by `ECommonsMain.Dispose()`. |
| `static string DefaultConfigurationFileName` | Full path of the default file: `Path.Combine(GetPluginConfigDirectory(), DefaultSerializationFactory.DefaultConfigFileName)`. |
| `static ISerializationFactory DefaultSerializationFactory { get; set; }` | Serializer used when none is passed. Setter throws `InvalidOperationException` once any load, save or `Migrate` has happened. |
| `static string GetPluginConfigDirectory()` | See "Directory override". |
| `static string? PluginConfigDirectoryOverride { get; set; }` | Default `null`. |
| `static bool UseDefaultAsyncSave { get; set; }` | Default `false`. Default for `writeFileAsync` when the argument is null. |
| `static Action<string, string>? SaveFileActionOverride` | Non-functional in this commit, see Gotchas. |

```csharp
public static T LoadConfiguration<T>(string path, bool appendConfigDirectory = true,
    ISerializationFactory? serializationFactory = null) where T : new()

public static void SaveConfiguration(this object Configuration, string path,
    bool prettyPrint = false, bool appendConfigDirectory = true,
    ISerializationFactory? serializationFactory = null, bool? writeFileAsync = null,
    bool? useReliableFileStorage = null)
```

`LoadConfiguration<T>`:

- `appendConfigDirectory: true` prepends `GetPluginConfigDirectory()` via `Path.Combine`. Because
  of `Path.Combine` semantics, passing an absolute path with the default `true` still yields that
  absolute path (AutoRetainer relies on this to read another plugin's file).
- Returns `new T()` when the file does not exist, and when the deserializer returns `null`
  (for example a file containing `null`). It never writes the file.
- There is no `try`/`catch`: malformed JSON or an IO error throws out of `LoadConfiguration` and
  therefore out of `Init<T>()`. `.new` leftovers are not read back automatically.
- Synchronous, on the calling thread.

`SaveConfiguration`:

- It is an extension method on `object`, so `myObj.SaveConfiguration("file.json")` and
  `EzConfig.SaveConfiguration(myObj, "file.json")` are the same call.
- Serialization always happens on the calling thread. A `null` result from the factory throws
  `NullReferenceException`; serializer exceptions propagate to the caller.
- The file write happens on the calling thread unless `writeFileAsync` (or, when it is null,
  `UseDefaultAsyncSave`) is `true`, in which case only the write runs in `Task.Run`.
- The write takes `lock(Configuration)` and catches every exception, logging it with
  `PluginLog.Error` (`e.Log()`). A failed write is therefore silent to the caller.
- `prettyPrint` defaults to `false` here; `EzConfig.Save()` passes `true`.
- `useReliableFileStorage` is accepted and never read in this commit.

### Migrate

```csharp
EzConfig.Migrate<Config>();            // optional, before Init
var config = EzConfig.Init<Config>();
```

- Throws `NullReferenceException("Migrate must be called before initialization")` if `Config` is
  already set.
- Runs only when the EzConfig default file does not exist and the stock file
  (`Svc.PluginInterface.ConfigFile`, or `<parent>/<PluginConfigDirectoryOverride>.json` when the
  override is set) exists. Otherwise logs `Migrating conditions are not met, skipping...` at
  Information level.
- When it runs: logs a warning, loads the stock file as `T` with the default factory, saves it as
  the EzConfig default file, resets `Config` to `null`, and renames the stock file to
  `<name>.json.old`.
- The stock file is parsed with EzConfig's serializer settings, not Dalamud's
  (`TypeNameHandling` is not enabled), so `$type` metadata written by Dalamud is not used to pick
  types.
- Counts as "a configuration has been loaded": set `DefaultSerializationFactory` before it.

### Directory override

`PluginConfigDirectoryOverride = "Name"` makes `GetPluginConfigDirectory()` return
`<pluginConfigs>/Name` (a sibling of the plugin's own directory; it is created on each call). Every
EzConfig and `EzCharaConfig` path that appends the config directory follows it. Set it before the
first load.

## What ECommonsMain.Dispose() does for configuration

From `references/ecommons/ECommons/ECommonsMain.cs`, in this order, each inside
`GenericHelpers.Safe`:

1. `if(EzConfig.Config != null) EzConfig.Save();` so the default configuration is written once on
   plugin unload, synchronously unless `UseDefaultAsyncSave` is `true`. `OnSave` fires.
2. `EzConfig.Dispose()` (internal), which only sets `OnSave = null`.
3. Later: `ExternalWriter.Dispose()` (internal): marks disposed, `CompleteAdding()` and disposes
   the request queue.

Not done by `Dispose()`:

- `EzConfig.Config` is not reset to null, and the `WasCalled` latch that locks
  `DefaultSerializationFactory` is not reset.
- Files loaded with `LoadConfiguration<T>` are not saved. Save them yourself.
- `EzCharaConfig<T>` instances are not saved directly. Their `EzLogout` hook is removed through
  `EzEvent.DisposeAll`; whether pending character data gets written on unload is up to you (call
  `SaveAll()` in your own `Dispose` before `ECommonsMain.Dispose()`).

Separately, `ECommonsMain.Dispose()` calls
`Svc.PluginInterface.SavePluginConfig(EzConfigGui.Config)`
when an `IPluginConfiguration` was passed to `EzConfigGui.Init(...)`. That is the stock Dalamud
path and is unrelated to `EzConfig.Config`.

## Serialization

### ISerializationFactory

`references/ecommons/ECommons/Configuration/ISerializationFactory.cs`

```csharp
public interface ISerializationFactory
{
    string DefaultConfigFileName { get; }
    T? Deserialize<T>(string inputData);
    string? Serialize(object config, bool prettyPrint);
    string? Serialize(object config);
    T? Deserialize<T>(byte[] inputData);
    byte[]? SerializeAsBin(object config);
    bool IsBinary { get; }

    void WriteFile(string fullPath, string data);
    void WriteFile(string fullPath, byte[] data);
    string ReadFileAsText(string fullPath);
    byte[] ReadFileAsBin(string fullPath);
    bool FileExists(string fullPath);
}
```

`EzConfig` chooses the text or binary trio by `IsBinary`: text uses `Serialize(config,
prettyPrint)`,
`WriteFile(string, string)`, `ReadFileAsText`, `Deserialize<T>(string)`; binary uses
`SerializeAsBin`, `WriteFile(string, byte[])`, `ReadFileAsBin`, `Deserialize<T>(byte[])`
(`prettyPrint` is not passed in binary mode). File existence always goes through `FileExists`.

### DefaultSerializationFactory

`references/ecommons/ECommons/Configuration/DefaultSerializationFactory.cs`. Newtonsoft.Json
(`JsonConvert`). All members are `virtual` except `Serialize(object config)`, which forwards to
`Serialize(config, false)`.

| Member | Default behaviour |
|---|---|
| `DefaultConfigFileName` | `"DefaultConfig.json"` |
| `IsBinary` | `false` |
| `Deserialize<T>(string)` | `new JsonSerializerSettings { ObjectCreationHandling = ObjectCreationHandling.Replace }` |
| `Deserialize<T>(byte[])` | UTF-8 decode, then the string overload |
| `Serialize(object, bool prettyPrint)` | `Formatting = Indented` or `None`; `DefaultValueHandling = Ignore` if the config type has `[IgnoreDefaultValue]`, else `Include` |
| `SerializeAsBin(object)` | UTF-8 bytes of `Serialize(config)` (not indented) |
| `ReadFileAsText` / `ReadFileAsBin` / `FileExists` | `File.ReadAllText(path, UTF8)` / `File.ReadAllBytes` / `File.Exists` |
| `WriteFile` (both) | Anti-corruption write, below |

Consequences of the settings:

- No `TypeNameHandling`: no `$type` in output, so a field's declared type can be changed as long
  as the JSON still fits. Polymorphic members (a `List<Base>` holding derived instances) do not
  round-trip unless you supply your own settings or converters.
- `ObjectCreationHandling.Replace`: collections initialised in field initialisers or the
  constructor are replaced by the deserialized ones rather than appended to. With stock Dalamud
  settings, default collection entries get duplicated on each load; here they do not.
- Newtonsoft defaults otherwise apply: public fields and properties are serialized, `[JsonIgnore]`
  and `[JsonProperty]` work, private members need `[JsonProperty]`.

Per-type settings override: if the config type has a static field or property named exactly
`JsonSerializerSettings` (public or non-public; looked up with `GetFieldPropertyUnion`) whose
value is a Newtonsoft `JsonSerializerSettings`, that object is used as-is for both directions and a
Verbose line is logged. In that case `prettyPrint`, `[IgnoreDefaultValue]` and the `Replace`
handling are all ignored; put `Formatting`, `DefaultValueHandling` and `ObjectCreationHandling` in
your own settings if you want them. For deserialization the lookup is on `typeof(T)`, for
serialization on `config.GetType()`.

```csharp
public class Config
{
    // picked up by DefaultSerializationFactory by name
    private static JsonSerializerSettings JsonSerializerSettings = new()
    {
        Formatting = Formatting.Indented,
        ObjectCreationHandling = ObjectCreationHandling.Replace,
        TypeNameHandling = TypeNameHandling.Auto,
    };
    public List<string> Names = [];
}
```

### Anti-corruption write

`DefaultSerializationFactory.WriteFile(fullPath, data)`:

1. If `<fullPath>.new` already exists (a previous write did not finish), it is renamed to
   `<fullPath>.new.<unix ms>`, two warnings are logged and `Notify.Warning("Detected unsuccessfully
   saved configuration file.")` is shown.
2. Data is written to `<fullPath>.new` (`File.WriteAllText(..., Encoding.UTF8)`, which emits a
   BOM, or `File.WriteAllBytes`).
3. `File.Move("<fullPath>.new", fullPath, true)` replaces the target.

There is no `.bak` copy of the previous good file and no rotation; the only extra files are the
orphaned `.new.<timestamp>` ones from step 1, which are never cleaned up or read back. The write
does not go through Dalamud's `ReliableFileStorage`.

### Plugging in another serializer

Either derive from `DefaultSerializationFactory` and override what differs (keeps the
anti-corruption write), or implement `ISerializationFactory` from scratch. Install it globally
before any EzConfig load/save/migrate, or pass it per call.

```csharp
public class MyFactory : DefaultSerializationFactory
{
    public override string DefaultConfigFileName => "Config.json";

    public override string Serialize(object config, bool prettyPrint)
        => JsonConvert.SerializeObject(config, prettyPrint ? Formatting.Indented : Formatting.None);

    public override T? Deserialize<T>(string inputData) where T : default
        => JsonConvert.DeserializeObject<T>(inputData);
}

// plugin constructor, after ECommonsMain.Init, before Migrate/Init:
EzConfig.DefaultSerializationFactory = new MyFactory();
var config = EzConfig.Init<Config>();

// per call:
var other = EzConfig.LoadConfiguration<Other>("other.bin", serializationFactory: myBinaryFactory);
other.SaveConfiguration("other.bin", serializationFactory: myBinaryFactory);
```

The `where T : default` clause on the override is what C# requires for an unconstrained `T?`
override; adjust if the compiler disagrees *(unverified, not compiled)*.

The default factory is also usable as a plain JSON helper. AutoRetainer uses it for clipboard
import/export:

```csharp
Copy(EzConfig.DefaultSerializationFactory.Serialize(selectedPlan, false));
var plan = EzConfig.DefaultSerializationFactory.Deserialize<EntrustPlan>(Paste())
    ?? throw new NullReferenceException();
```

Calling `Serialize`/`Deserialize` directly on the factory does not trip the `WasCalled` latch;
only `LoadConfiguration`, `SaveConfiguration` and `Migrate` do.

## IgnoreDefaultValueAttribute

`[AttributeUsage(AttributeTargets.Class | AttributeTargets.Struct)] public class
IgnoreDefaultValueAttribute : Attribute`. Put it on the configuration type itself. Checked with
`IsDefined(..., inherit: false)` on the runtime type of the root object only, so it must be on the
exact class being saved, and it then applies `DefaultValueHandling.Ignore` to the whole object
graph. "Default" means the CLR default (`0`, `false`, `null`) unless `[DefaultValue]` says
otherwise, not the field initialiser: a `bool Enabled = true` that the user sets to `false` is
omitted from the file and comes back as `true`. Use it for large append-only data files, not for
settings with non-zero defaults.

## IEzConfig

```csharp
[Obsolete("You are no longer required to implement IEzConfig")]
public interface IEzConfig { }
```

Empty marker interface, obsolete; the replacement is to implement nothing. It has no relation to
Dalamud's `IPluginConfiguration`: it does not extend it and has no `Version`. No EzConfig method
has a constraint on either interface. A class may additionally implement `IPluginConfiguration`
if it is also passed to stock Dalamud APIs or to `EzConfigGui.Init(..., config)`, but EzConfig
does not need it. AutoRetainer's `StatisticsFile : IEzConfig` is a leftover.

## EzConfigValueStorage

`internal static class` with one field, `internal static ISerializationFactory
DefaultSerializationFactory = new DefaultSerializationFactory();`. Backing store for
`EzConfig.DefaultSerializationFactory`; not usable from plugin code.

## EzCharaConfig<T>

`public class EzCharaConfig<T> where T : new()`,
`references/ecommons/ECommons/Configuration/EzCharaConfig.cs`. Per-character configuration files
keyed by content ID, with an in-memory cache.

```csharp
public EzCharaConfig(IEnumerable<Option>? options = null, string prefix = "EzConfig")
public enum Option { AllowDefaultConfig, UnloadOnLogout, NotIndented }   // nested: EzCharaConfig<T>.Option
```

| Member | Notes |
|---|---|
| `T Get()` | `Get(Player.CID)` |
| `T GetDefault()` | `Get(0)` |
| `T Get(ulong CID)` | Cached instance, or `EzConfig.LoadConfiguration<T>("<prefix><CID:X16>.json")` added to the cache. CID `0` throws `InvalidOperationException("Player is not currently logged in")` unless `Option.AllowDefaultConfig`. |
| `bool TryGet(ulong CID, out T? value)` | `false` only for CID `0` without `AllowDefaultConfig`; otherwise `Get(CID)`. |
| `void Save(ulong cid, bool unload = false)` | Saves only if that CID is in the cache. Indented unless `Option.NotIndented`. `unload` removes it from the cache. |
| `void SaveAll(bool unload = false)` | `Save` for every cached CID. |
| `string CurrentCharaConfigFile` | Full path for `Player.CID`. |
| `string DefaultCharaConfigFile` | Full path of the CID-0 file. |

- File name is `$"{Prefix}{CID:X16}.json"` in `EzConfig.GetPluginConfigDirectory()`, for example
  `EzConfig00412345678ABCDE.json`. Use a distinct `prefix` per `EzCharaConfig` instance.
- The constructor creates `new EzLogout(() => SaveAll(Options.Contains(Option.UnloadOnLogout)))`,
  so all cached character configs are saved on `IClientState.Logout`. `EzLogout` is removed by
  `ECommonsMain.Dispose()`; the class has no `Dispose` of its own.
- Uses the default serialization factory and `UseDefaultAsyncSave`.
- `Player.CID` is `0` when not logged in, so `Get()` throws on the title screen unless
  `AllowDefaultConfig` is set (then it returns the shared CID-0 config).

```csharp
public class CharaSettings { public bool Enabled = true; public List<uint> Items = []; }

EzCharaConfig<CharaSettings> CharaConfig =
    new([EzCharaConfig<CharaSettings>.Option.UnloadOnLogout], prefix: "Chara");

if(CharaConfig.TryGet(Player.CID, out var settings)) settings.Enabled = false;
CharaConfig.Save(Player.CID);

// plugin Dispose, before ECommonsMain.Dispose():
CharaConfig.SaveAll();
```

No consumer in the three pinned plugin checkouts uses `EzCharaConfig`.

## ExternalWriter

`public static class ExternalWriter`, `references/ecommons/ECommons/Configuration/ExternalWriter.cs`.
Writes files from a separate `dotnet.exe` child process so a game crash cannot truncate them.
`EzConfig` does not use it; it is opt-in and standalone.

```csharp
public static void PlaceWriteOrder(FileSaveStruct order)

public class FileSaveStruct   // nested: ExternalWriter.FileSaveStruct
{
    public string? Name; public byte[]? Data; public byte[]? NameHash; public byte[]? DataHash; // properties
    public FileSaveStruct(string name, string data)      // data encoded as UTF-8
    public FileSaveStruct(string name, byte[] dataBytes)
    public string Serialize()                            // System.Text.Json
}
```

- `Name` is the full target path. Hashes are SHA-1 of the UTF-8 name and of the data.
- First successful order starts a background `Thread` that, per batch: copies
  `ECommons.FileWriter.dll`, `ECommons.FileWriter.deps.json` and
  `ECommons.FileWriter.runtimeconfig.json` from the plugin's assembly directory into
  `Svc.PluginInterface.ConfigDirectory` if missing or different, starts
  `%AppData%\XIVLauncher\runtime\dotnet.exe <configDir>\ECommons.FileWriter.dll <pipe handle>`,
  and sends each order as one JSON line over an anonymous pipe. Child stdout is logged at
  Information, stderr at Error.
- If the queue rejects an order, a warning is logged and the order is retried via
  `Svc.Framework.RunOnTick`.
- After `ECommonsMain.Dispose()` the queue is disposed (the field is not nulled), so further
  `PlaceWriteOrder` calls are not supported.
- The `ECommons.FileWriter` project is not part of the ECommons repository at this commit (the
  string occurs only in `ExternalWriter.cs`). The three files must be supplied next to the plugin
  assembly by the plugin author; otherwise the copy throws, the exception is logged at Debug level
  only, and nothing is written. The `dotnet.exe` path is Windows/XIVLauncher specific.
- Only known use in the pinned consumers is an AutoRetainer debug button:

```csharp
ExternalWriter.PlaceWriteOrder(new(
    System.IO.Path.Combine(Svc.PluginInterface.ConfigDirectory.FullName, "WriterTest.json"),
    EzConfig.DefaultSerializationFactory.Serialize(C, true)));
```

Treat `ExternalWriter` as experimental; prefer `EzConfig.Save()`.

## Usage from real plugins

Of the three pinned consumers only AutoRetainer uses EzConfig. YesAlready and PandorasBox both use
stock Dalamud configuration (`Svc.PluginInterface.GetPluginConfig() as Configuration ?? new
Configuration()` and `SavePluginConfig(this)`, class implements `IPluginConfiguration`) alongside
ECommons; YesAlready passes nothing to EzConfig at all. Using ECommons does not require EzConfig.

AutoRetainer (`AutoRetainer/AutoRetainer.cs`), condensed:

```csharp
public AutoRetainer(IDalamudPluginInterface pi)
{
    ECommonsMain.Init(pi, this, Module.DalamudReflector);
    // ...
}

public void Load()
{
    EzConfig.Migrate<Config>();
    config = EzConfig.Init<Config>();
    // ...
}

internal void SetConfig(Config c) => config = EzConfig.Set<Config>(c);   // after importing a config
```

Its `Config` class is a plain `[Serializable] internal unsafe class Config` with public fields and
no interface. It calls `EzConfig.Save()` when windows close (`Window.OnClose`) and after specific
mutations, not every frame.

Additional files (AutoRetainer statistics, `Modules/Statistics/StatisticsFileWrapper.cs`):

```csharp
internal string FileName => $"{CID:X16}_{RetainerName}.statistic.json";
File = EzConfig.LoadConfiguration<StatisticsFile>(FileName);   // relative to plugin config dir
File.SaveConfiguration(FileName);                              // extension method on object
```

Timestamped backup before overwriting (AutoRetainer `ExpertTab.cs`):

```csharp
EzConfig.SaveConfiguration(C, $"Backup_{DateTimeOffset.Now.ToUnixTimeMilliseconds()}.json");
```

Peeking at the default file without making it the active config:

```csharp
var c = EzConfig.LoadConfiguration<Config>(EzConfig.DefaultSerializationFactory.DefaultConfigFileName);
```

Minimal new plugin:

```csharp
public sealed class Plugin : IDalamudPlugin
{
    internal static Config C = null!;

    public Plugin(IDalamudPluginInterface pi)
    {
        ECommonsMain.Init(pi, this);
        C = EzConfig.Init<Config>();
    }

    public void Dispose()
    {
        ECommonsMain.Dispose();   // saves EzConfig.Config
    }
}

public class Config
{
    public bool Enabled = true;
    public List<string> Names = [];
}
```

## Gotchas

- `EzConfig.Init<T>()` requires `Svc` to be initialised, so call `ECommonsMain.Init` first.
- `ECommonsMain.Dispose()` always saves a non-null `EzConfig.Config`. If the in-memory object is in
  a state you do not want persisted (failed import, read-only second instance), there is no public
  way to null `Config`; `Set<T>(default)` on a reference type does set it to null, which skips the
  save.
- The unload save is synchronous on the thread that disposes the plugin, unless
  `UseDefaultAsyncSave` is `true`; with async writes, an unload-time `Task.Run` write may race the
  plugin being unloaded.
- `SaveFileActionOverride` cannot be used: its setter never assigns the backing field (it only
  throws if the field is already non-null, which cannot happen) and nothing reads the property.
  Assigning it is a silent no-op. To redirect writes, override `WriteFile` in a factory.
- `useReliableFileStorage` on `SaveConfiguration` is ignored.
- Write failures are logged, not thrown. Serialization failures and all load failures are thrown.
  A corrupt `DefaultConfig.json` makes `Init<T>()` throw in the plugin constructor unless you catch
  it.
- Setting `EzConfig.DefaultSerializationFactory` after any load/save/migrate throws
  `InvalidOperationException`. The latch is a static that `ECommonsMain.Dispose()` does not reset.
- `SaveConfiguration` locks on the configuration object for the write only; serialization is not
  locked. Mutating collections on another thread while saving can throw from Newtonsoft
  ("Collection was modified"). Save from the framework thread if the config is mutated there.
- Saving is not throttled. Calling `EzConfig.Save()` every frame from a draw function serializes
  and writes the whole file each frame.
- `[IgnoreDefaultValue]` drops every value equal to the CLR default; see that section before
  using it on a settings class.
- A static `JsonSerializerSettings` member on the config type silently replaces all default
  settings, including `ObjectCreationHandling.Replace` and indentation.
- `EzCharaConfig<T>.DefaultCharaConfigFile` is built from `$"{Prefix}{0:X16}.json"`, that is
  `<prefix>0000000000000000.json`, the same name `Get(0)` / `GetDefault()` loads.
- `EzCharaConfig<T>.Save(cid)` does nothing if `Get(cid)` was never called for that CID in this
  session (or it was unloaded).
- `Migrate<T>()` renames the stock config to `.json.old`. After that, stock
  `GetPluginConfig()` returns null for the plugin; do not mix both systems on one class without
  intent.
