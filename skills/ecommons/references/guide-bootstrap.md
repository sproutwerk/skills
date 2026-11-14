## Bootstrap

```csharp
using Dalamud.Plugin;
using ECommons;
using ECommons.DalamudServices;

public sealed class MyPlugin : IDalamudPlugin
{
    public MyPlugin(IDalamudPluginInterface pluginInterface)
    {
        ECommonsMain.Init(pluginInterface, this);   // first. Optional: Module.DalamudReflector, ...
        // Svc.*, EzConfig, EzConfigGui, TaskManager, EzIPC are usable from here on
    }

    public void Dispose()
    {
        // own teardown first
        ECommonsMain.Dispose();                     // last
    }
}
```

`public static void Init(IDalamudPluginInterface pluginInterface, object instance, params Module[]
modules)`. `instance` must implement `IDalamudPlugin` or `IAsyncDalamudPlugin`, otherwise
`InvalidOperationException`. `Init` injects `Svc`, registers `[Cmd]` methods found on `instance`,
sets `Svc.Log.MinimumLogLevel` to Verbose and initialises the requested modules. Set
`ECommonsMain.ReducedLogging = true` before `Init` to silence ECommons' own info lines.

| `Module` | Needed for |
|---|---|
| none | almost everything, including all `ECommons.Hooks` classes (they install lazily) |
| `DalamudReflector` | `TryGetDalamudPlugin` with its default cache, `RegisterOnInstalledPluginsChangedEvents`, `SetKeyState` |
| `SplatoonAPI` | `ECommons.SplatoonAPI`; implies `DalamudReflector` |
| `ObjectLife` | `ObjectLife` lifetime queries (hooks the game object constructor) |
| `VfxTracking` | `VfxManager` tracked VFX lists |
| `ObjectFunctions` | nothing strictly; only front-loads a signature scan that `IsHostile()` would do lazily |
| `All` | all of the above, including hooks you may not need |

`ECommonsMain.Dispose()` saves `EzConfig`, tears down the EzConfigGui windows, removes `EzCmd` and
`[Cmd]` commands, disposes every TaskManager, `EzHook`, built-in hook, `Ez*` event subscription,
IPC registration, DTR entry and singleton service, and nulls the static throttlers. Skipping it
leaves hooks enabled with detours in an unloading assembly (crash on unload or reload), TaskManagers
still ticking, commands and IPC providers still registered and the last config changes unsaved.
After it, `ECommonsMain.Disposed` is true and `EzThrottler` throws `NullReferenceException`.

Other `ECommonsMain` members: `Instance`, `InstanceUniqueId`, `MainWindowHandle`, `Disposed`.

