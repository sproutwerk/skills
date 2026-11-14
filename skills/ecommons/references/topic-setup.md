# Setup: distribution, csproj, Init and Dispose

Covers: how ECommons is distributed (NuGet and git submodule), how real plugins reference it, how
versions couple to the Dalamud API level, and exactly what `ECommonsMain.Init` and
`ECommonsMain.Dispose` do. Verified against ECommons `bc38dc7` (2026-11-08, v3.2.1.26, Dalamud
API 15); nothing here was compiled. Paths are relative to the skill root.

Sources: `references/ecommons/ECommons/ECommonsMain.cs`, `references/ecommons/ECommons/Module.cs`,
`references/ecommons/ECommons/ECommons.csproj`, `references/ecommons/ECommons/DalamudServices/Svc.cs`,
`references/ecommons/README.md`. Stock csproj / `Dalamud.NET.Sdk` questions → skill `dalamud`.

## Distribution

Two channels, both built from the same `master` branch of `NightmareXIV/ECommons`.

### NuGet

| Package | Latest (2026-11-13) | What it is |
|---|---|---|
| `ECommons` | 3.2.1.26, published 2026-10-14T23:21Z | the library; 140 versions on nuget.org |
| `ECommons.IPC` | 1.0.0.24 | separate repo `NightmareXIV/ECommons.IPC`; depends on `ECommons >= 3.2.0.11`. Typed IPC clients for NightmareXIV plugins *(contents unverified, not cloned)* |
| `FFXIVOcelot.ECommons` | 2.0.1 | third-party package by another author, not the library. Do not use it by mistake |

Facts about the `ECommons` package, from the nuspec and the nupkg listing of 3.2.1.26:

- Contents are `lib/net10.0-windows7.0/ECommons.dll` and `ECommons.xml` (XML docs). Nothing else.
- The dependency group is empty. Dalamud, FFXIVClientStructs, Lumina and the ImGui bindings are
  referenced by the library as `Private=False` hint-path references into the Dalamud dev folder, so
  the DLL binds to whatever Dalamud the game loads at runtime.
- The nuspec `repository` element carries the commit it was built from. For 3.2.1.26 that is
  `26d30b9ec14c3d3c18e335a9e4c7a5421dd134db` (2026-10-15). This skill is pinned to the later
  `master` commit `bc38dc7`, which still declares `BaseVersion 3.2.1.26`. See
  [package and `master` differences](#package-and-master-differences).
- Publishing is automatic: `.github/workflows/CompileAndPublish.yml` runs on every push to
  `master` that touches `ECommons/ECommons.csproj` (that is, on every `BaseVersion` bump) and on
  manual dispatch. It checks out `NightmareXIV/NightmareXIVNugetUploader` and runs it. There are no
  git tags or GitHub releases to look at; the version history is the NuGet version list.
- Version string: `<BaseVersion>` plus an optional `-<PackageKind>` suffix in the csproj. The only
  suffixed versions ever published are `3.0.0.9-api13` to `3.0.0.18-api13`.

Version lines and the Dalamud API they were built for. The target framework and publish dates are
from nuget.org; the API level is inferred from those dates and the Dalamud release history, upstream
does not state it anywhere:

| ECommons | Target framework | Published | Dalamud API (inferred) |
|---|---|---|---|
| 2.x, 3.0.0 to 3.0.0.8 | net9.0-windows7.0 | up to 2025-08-04 | 12 |
| 3.0.0.9-api13 to 3.0.0.18-api13 | net9.0-windows7.0 | 2025-08 | 13 (prerelease) |
| 3.0.1 to 3.0.1.32 | net9.0-windows7.0 | 2025-08-07 to 2025-12-05 | 13 |
| 3.1.0 to 3.1.0.34 | net10.0-windows7.0 | 2025-12 to 2026-04-21 | 14 |
| 3.2.0 to 3.2.1.26 | net10.0-windows7.0 | 2026-04-28 onwards | 15 |

There is no compatibility shim between lines. A plugin built for API 15 needs a 3.2.x package; an
older package fails at load or at the first call into a changed Dalamud or ClientStructs member.
Within a line, versions are not guaranteed source compatible either: members get renamed and
reworked without a major bump (the `Player` rework and the TaskManager namespace move are the
examples the README calls out).

#### Package and `master` differences

A version number does not identify one source tree. Commits pushed after a `BaseVersion` bump
stay unpublished until the next bump, so a submodule on `master` can be ahead of the package with
the same version. At this pin, `master` (`bc38dc7`) differs from the 3.2.1.26 package (`26d30b9`)
in three files:

- `GenericHelpers.AnyNull(params object[])` is gone. `master` has `ContainsNull<T>(params T[])`
  and `ContainsNull<T>(this IEnumerable<T>)` instead. The package has `AnyNull` and no
  `ContainsNull`.
- `FirstOrAdd<T>(this IList<T>, Func<T> generator = null)` exists only on `master`, and
  `CircularSelect` returns `default` for an empty list or array there instead of throwing.
- `AddonMaster.Gathering.GatheredItem` has `IsHidden`, `IsRare` and `GatherAmount` only on
  `master`.
- The `Struct()`, `Character()`, `GameObject()` and `IGameObject()` pointer extensions in
  `CharacterFunctions` use `chr?.Address` on `master`, so a null receiver yields a null pointer
  instead of a `NullReferenceException`.

The bundled mirror and indexes describe `master`. For a plugin on the NuGet package, treat those
members as absent.

### Git submodule

```sh
git submodule add https://github.com/NightmareXIV/ECommons ECommons
```

```xml
<ProjectReference Include="..\ECommons\ECommons\ECommons.csproj" />
```

The project to reference is `ECommons/ECommons/ECommons.csproj` (repo root holds only the solution,
README and licence). Things to know about building it as a project:

- It is a plain `Microsoft.NET.Sdk` project, not `Dalamud.NET.Sdk`. It locates Dalamud through its
  own `DalamudLibPath`: `$(appdata)\xivlauncher\Addon\Hooks\dev\` on Windows; on Linux
  `$(DALAMUD_HOME)/`, defaulting to `$(HOME)/.xlcore/dalamud/Hooks/dev/`. If your plugin resolves
  Dalamud from somewhere else (CI download, custom path), set `DALAMUD_HOME` for the whole build or
  ECommons fails with missing `Dalamud`, `FFXIVClientStructs` and `Lumina` references.
- `TargetFramework` is `net10.0-windows7.0`, `Platforms` is `x64`, `LangVersion` is `preview` (the
  source uses C# 14 extension blocks, for example in `TaskManagerConfigurationExtensions.cs`),
  `Nullable` is `enable` but most files opt out with `#nullable disable`.
- `GeneratePackageOnBuild` is `True`, so every build also writes a `.nupkg` into the ECommons
  output folder. Harmless, but it surprises people.
- `RestorePackagesWithLockFile` is `true` and a `packages.lock.json` is committed.
- It references `Microsoft.CodeAnalysis.BannedApiAnalyzers` with `BannedSymbols.txt` and sets
  `WarningsAsErrors` to `RS0030`. This applies to ECommons' own code only (the analyzer is a
  private asset). It matters when you patch the submodule: `ImRaii`, most `IPlayerState` members,
  several `IBattleChara` cast properties, `ReadOnlySeString.ExtractText` and `World.Region` are
  banned inside ECommons and fail the build. See `references/ecommons/ECommons/BannedSymbols.txt`.
- The submodule pins a commit. Bump it together with Dalamud API bumps; an old pin against a new
  Dalamud is the usual cause of a wall of compile errors inside `ECommons/`.

### What consumers actually do

From the clones used for this skill (2026-11-13):

| Plugin | Reference | Evidence |
|---|---|---|
| AutoRetainer `cc020df` | submodule + `ProjectReference` | `.gitmodules` lists `ECommons` (NightmareXIV); csproj has `<ProjectReference Include="..\ECommons\ECommons\ECommons.csproj">`. It also lists `<PackageReference Include="ECommons" Version="*">` with `ExcludeAssets` runtime/build/... and `PrivateAssets=all`, which contributes nothing at runtime |
| YesAlready `c531fb8` | NuGet | `<PackageReference Include="ECommons" Version="3.2.0.19" />`, project SDK `Dalamud.NET.Sdk/15.0.0`; no ECommons submodule |
| PandorasBox `3a35634` | NuGet | `<PackageReference Include="ECommons" Version="3.2.1.5" />`, project SDK `Dalamud.NET.Sdk/15.0.0`; `.gitmodules` still lists an `ECommons` submodule that the csproj no longer references |

So the presence of an ECommons entry in `.gitmodules` does not prove the plugin builds from it;
read the csproj. PunishLib itself depends on the NuGet package (`"ECommons": "[3.2.0.11, )"` in
PandorasBox's lock file under the PunishLib project) → skill `punishlib`.

Forks exist and are used as submodules by some plugins (`ffxivcode/ECommons`,
`Taurenkey/ECommons`). Their API can differ from this pin *(unverified, not cloned)*.

### Choosing

- NuGet: simplest, reproducible, works with `Dalamud.NET.Sdk` without path setup. You cannot patch
  the library and you wait for a release when a game patch breaks a signature.
- Submodule: you can step into and patch the source, and you can pin any commit. You own the
  `DalamudLibPath` problem and the longer build.
- Never both in one plugin. Two copies of `ECommons.dll` types in one load context give duplicate
  type errors or, worse, two sets of static state.

## Namespaces to import

`GenericHelpers` and `ECommonsMain` live in the root namespace `ECommons`; extension methods from
`GenericHelpers` become visible with `using ECommons;`. Common imports:

```csharp
using ECommons;                    // ECommonsMain, Module, GenericHelpers (+ extensions), EzCmd
using ECommons.DalamudServices;    // Svc
using ECommons.Logging;            // PluginLog, DuoLog, InternalLog
using ECommons.ImGuiMethods;       // ImGuiEx, Notify, EColor, ThreadLoadImageHandler ...
using ECommons.Configuration;      // EzConfig
using ECommons.SimpleGui;          // EzConfigGui
using ECommons.Throttlers;         // EzThrottler, FrameThrottler
using ECommons.Automation;         // Chat, Callback
using ECommons.Automation.NeoTaskManager;   // TaskManager (current)
using ECommons.UIHelpers.AddonMasterImplementations;   // AddonMaster
using ECommons.GameHelpers;        // Player
```

Many consumers put these in a `GlobalUsings.cs`. `Module` clashes with `System.Reflection.Module`
when `System.Reflection` is imported; PandorasBox writes `ECommons.Module.All` for that reason.

## ECommonsMain.Init

```csharp
public unsafe static void Init(IDalamudPluginInterface pluginInterface, object instance,
                               params Module[] modules)
```

Call it first in the plugin constructor, before touching `Svc` or any other ECommons class. In
order, it:

1. Throws `InvalidOperationException` unless `instance` implements `IDalamudPlugin` or
   `IAsyncDalamudPlugin`. Stores it in `ECommonsMain.Instance`.
2. Calls `Svc.Init(pluginInterface)`, which is `pi.Create<Svc>()`: Dalamud injects every
   `[PluginService]` static property of `Svc`. If that throws, ECommons starts a thread that shows
   a Win32 message box ("Error initializing ECommons services ...") and rethrows, so the plugin
   fails to load.
3. Generates `ECommonsMain.InstanceUniqueId` (random `uint`, unique among loaded ECommons plugins
   through a Dalamud DataShare dictionary named `ECommons.UniqueIdConcurrentDictionary`).
4. Resolves `ECommonsMain.MainWindowHandle` (`nint*`) by signature; failure is only a warning.
5. Logs the "This is ECommons v... Hello!" line unless `ECommonsMain.ReducedLogging` is true.
6. Sets `Svc.Log.MinimumLogLevel = LogEventLevel.Verbose`. Your plugin then emits verbose and
   debug lines regardless of the level Dalamud would have applied.
7. Runs `CmdManager.Init` (registers `[Cmd]` methods found on the plugin instance).
8. Initialises the requested modules, each wrapped in `GenericHelpers.Safe` so a failure is logged
   and does not abort `Init`.

`ECommonsMain.ReducedLogging` is a public static field; set it before `Init` to silence the
banner and the module log lines.

Thread: `Init` itself has no thread requirement in the source. YesAlready and AutoRetainer call it
directly in the constructor; PandorasBox wraps its whole initialisation, `Init` included, in
`framework.RunOnFrameworkThread(...)` because the code that follows touches game state.

### Modules

```csharp
public enum Module { All, DalamudReflector, ObjectFunctions, ObjectLife, SplatoonAPI, VfxTracking }
```

| Flag | What `Init` does | You need it for |
|---|---|---|
| (none) | services, `[Cmd]` registration | everything not listed below. Most of the library needs no module |
| `DalamudReflector` | `DalamudReflector.Init()` | reflection into Dalamud's plugin manager: `DalamudReflector.TryGetDalamudPlugin` and friends → `references/topic-reflection-interop.md` |
| `ObjectFunctions` | `ObjectFunctions.Init()`: resolves the nameplate-colour function by signature up front | nothing strictly: `IsHostile()` resolves the same delegate lazily on first use. The flag only front-loads the scan |
| `ObjectLife` | `ObjectLife.Init()`: hooks the game object constructor on the next framework tick | `ObjectLife` lifetime queries → `references/topic-game-hooks.md` |
| `SplatoonAPI` | `DalamudReflector.Init()` then `Splatoon.Init()` | drawing through an installed Splatoon → `references/topic-splatoonapi.md` |
| `VfxTracking` | `VfxManager.Init()` | `VfxManager` tracked VFX lists → `references/topic-game-hooks.md` |
| `All` | all of the above | convenience; installs hooks you may not need |

`SplatoonAPI` implies `DalamudReflector`. Hooks under `ECommons.Hooks` (`ActionEffect`,
`DirectorUpdate`, `MapEffect`, ...) are not modules; they install themselves on first use.

## ECommonsMain.Dispose

```csharp
public static void Dispose()
```

Call it last in the plugin's `Dispose`, after your own teardown (your code may still use `Svc`,
`EzConfig` or a TaskManager while shutting down). It sets `ECommonsMain.Disposed = true` and then,
each step wrapped in `GenericHelpers.Safe`:

| Step | Effect |
|---|---|
| `SingletonServiceManager.DisposeAll` | disposes services created by `SingletonServiceManager.Initialize` |
| `PluginLoader.Dispose`, `CmdManager.Dispose` | loader state; `[Cmd]` handlers |
| `EzConfig.Save` (only if `EzConfig.Config != null`), `EzConfig.Dispose` | final save of the EzConfig configuration |
| EzConfigGui teardown (only if `EzConfigGui.WindowSystem != null`) | unsubscribes `OpenMainUi` / `OpenConfigUi`, saves `EzConfigGui.Config` through `SavePluginConfig` and shows a "Configuration saved" notification if a config was passed to `EzConfigGui.Init`, removes all windows |
| `EzCmd.RegisteredCommands` | `Svc.Commands.RemoveHandler` for every command added through `EzCmd.Add` |
| `Splatoon.Reset` / `Splatoon.Shutdown` | removes this plugin's Splatoon elements |
| `ProperOnLogin`, `DirectorUpdate`, `ActionEffect`, `MapEffect`, `SendAction`, `ActorVfx`, `StaticVfx`, `GameObjectCtor`, `VfxManager`, `ObjectLife`, `DalamudReflector`, `AutoCutsceneSkipper`, `Callback`, `TradeDetectionManager`, `RenderDisableManager`, `ContextMenuPrefixRemover` | unhooks and unsubscribes each |
| `LegacyTaskManager.TaskManager.DisposeAll`, `NeoTaskManager.TaskManager.DisposeAll` | detaches every TaskManager instance from `Framework.Update` |
| `ThreadLoadImageHandler.ClearAll`, its `httpClient` | cached textures and the HTTP client |
| `EzThrottler.Throttler = null`, `FrameThrottler.Throttler = null` | static throttlers are gone: any later `EzThrottler.Throttle` call throws `NullReferenceException` |
| `EzEvent.DisposeAll`, `EzHookCommon.DisposeAll`, `EzSharedData.Dispose`, `EzIPC.Dispose`, `EzDtr.DisposeAll`, `ExternalWriter.Dispose` | every `EzFrameworkUpdate`/`EzLogout`/..., every `EzHook`, shared data, IPC registrations, DTR entries |
| `Purgatory.Purge` | runs actions registered with `Purgatory.Add(Action)` and nulls the static class-typed members of types registered with `Purgatory.Add(Type)` (this is also how `EzPatch` and `Svc.GameNetwork` are released) |
| `RelinquishData` for the unique-id dictionary | DataShare reference count |

Finally `ECommonsMain.Instance = null`.

One mismatch to be aware of: `EzConfigGui.Init` subscribes `WindowSystem.Draw` to `UiBuilder.Draw`,
while `Dispose` unsubscribes `EzConfigGui.Draw` (the draw action passed to `Init`), which is a
different delegate. The window system is emptied (`RemoveAllWindows`) and set to null, and Dalamud
drops the plugin's `UiBuilder` on unload, so this has no visible effect in normal use.

### If Dispose is skipped

Everything in the table stays alive after your plugin unloads, inside a plugin whose assembly
Dalamud is trying to unload:

- Hooks (`EzHook`, `ActionEffect`, `Callback.InstallHook`, `ObjectLife`, ...) remain enabled with
  detours pointing into your assembly. This is the crash-on-unload or crash-on-reload case.
- Every TaskManager keeps its `Framework.Update` subscription and keeps running queued tasks.
- `EzFrameworkUpdate`, `EzLogout`, `EzTerritoryChanged`, `ProperOnLogin` subscriptions keep firing.
- Commands added with `EzCmd.Add` or `[Cmd]` stay registered; the reload then fails to register
  them again or the old handler runs.
- IPC providers registered through `EzIPC` stay registered, so other plugins keep calling dead
  code.
- The last EzConfig changes are not written.
- The assembly cannot be collected, so a dev-plugin reload leaks the old copy.

### After Dispose

`ECommonsMain.Disposed` is true and static state is torn down. Background work (tasks, timers,
threads) that outlives `Dispose` must check `ECommonsMain.Disposed` before touching ECommons, in
particular the throttlers (null) and `Svc`.

## Minimal plugin

```csharp
using Dalamud.Plugin;
using ECommons;
using ECommons.DalamudServices;
using ECommons.Logging;

namespace MyPlugin;

public sealed class MyPlugin : IDalamudPlugin
{
    public MyPlugin(IDalamudPluginInterface pluginInterface)
    {
        ECommonsMain.Init(pluginInterface, this);          // add Module flags only when needed
        PluginLog.Information($"Loaded on territory {Svc.ClientState.TerritoryType}");
    }

    public void Dispose()
    {
        // own teardown first
        ECommonsMain.Dispose();                            // last
    }
}
```

More complete bootstraps (config, window, commands, TaskManager, IPC) → `references/topic-examples.md`.

## README statements that still matter

- `Player` errors after an update: the old class lives on as
  `ECommons.GameHelpers.LegacyPlayer.Player`; the README suggests
  `global using Player = ECommons.GameHelpers.LegacyPlayer.Player;` as a stopgap →
  `references/topic-game.md`.
- TaskManager namespace changed on 2024-04-15: the old one is
  `ECommons.Automation.LegacyTaskManager`, the current one `ECommons.Automation.NeoTaskManager` →
  `references/topic-taskmanager.md`.
- Since 3.0.0.0 ECommons does not reference Windows Forms; the `RELEASEFORMS` / `DEBUGFORMS`
  configurations are gone; use `ECommons.WindowsFormsReflector.Keys` instead of
  `System.Windows.Forms.Keys` → `references/topic-reflection-interop.md`.
- The README's module list (`All`, `SplatoonAPI`, `DalamudReflector`, `ObjectLife`,
  `ObjectFunctions`) omits `VfxTracking` and writes the enum as `Modules.<Module>`; the type is
  `Module`.

## Gotchas

- `Svc.Init` logs "Services already initialized, skipping" on a second call but does not skip; it
  calls `pi.Create<Svc>()` again. Harmless, and only reachable if you call it yourself.
- `Init` with a non-plugin `instance` (for example a separate bootstrap class) throws. Pass the
  `IDalamudPlugin` object.
- `[Cmd]` methods are discovered on the instance passed to `Init` only, at `Init` time.
- Module failures are swallowed by `GenericHelpers.Safe` and only logged. If `DalamudReflector`
  calls misbehave, check the log for an exception during `Init` first.
- A NuGet consumer cannot see `internal` members that the source shows (for example
  `EzThrottler.Throttler`); do not write code against them.
