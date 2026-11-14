## Entry point and lifecycle

```csharp
using Dalamud.Game.Command;
using Dalamud.IoC;
using Dalamud.Plugin;
using Dalamud.Plugin.Services;

namespace MyPlugin;

public sealed class Plugin : IDalamudPlugin
{
    [PluginService] internal static IDalamudPluginInterface PluginInterface { get; private set; } = null!;
    [PluginService] internal static ICommandManager CommandManager { get; private set; } = null!;
    [PluginService] internal static IFramework Framework { get; private set; } = null!;
    [PluginService] internal static IPluginLog Log { get; private set; } = null!;

    private const string CommandName = "/myplugin";

    public Plugin()
    {
        CommandManager.AddHandler(CommandName, new CommandInfo(OnCommand) { HelpMessage = "Toggle the thing" });
        Framework.Update += OnUpdate;
        Log.Information("Loaded, reason {Reason}", PluginInterface.Reason);
    }

    public void Dispose()
    {
        Framework.Update -= OnUpdate;
        CommandManager.RemoveHandler(CommandName);
    }

    private void OnCommand(string command, string args) { }
    private void OnUpdate(IFramework framework) { }
}
```

Facts from `LocalPlugin.cs` and `ServiceContainer.cs`:

- Entry class = the first type in the assembly assignable to `IDalamudPlugin` or
  `IAsyncDalamudPlugin` (`FindPluginImpl`). Have exactly one.
- `IDalamudPlugin : IDisposable` has no members. `IDalamudPlugin.Name` was removed in API 9.
- Creation order: pick a **public** constructor whose every parameter is a known service →
  `GetUninitializedObject` → inject `[PluginService]` properties → invoke the constructor. So
  injected properties are already set inside the constructor.
- The constructor runs on a dedicated `LongRunning` thread, **not** the framework thread, unless
  `LoadSync` + `LoadRequiredState` 0/1. Touching `IObjectTable` there throws. Load is cancelled
  after 60 s.
- `Dispose()` runs on the framework thread unless the manifest has `CanUnloadAsync`.
- `PluginInterface.Reason` is a `[Flags] PluginLoadReason` (`Unknown`, `Installer`, `Update`,
  `Reload`, `Boot`); test with `.HasFlag()`.
- A non-dev plugin with `DalamudApiLevel` below the running level is refused ("incompatible API
  level"); dev plugins are allowed to load.

### IAsyncDalamudPlugin (API 15, interface stable, behaviour experimental)

```csharp
public sealed class Plugin(IFramework framework) : IAsyncDalamudPlugin   // LoadThingsAsync etc. are your own methods
{
    public async Task LoadAsync(CancellationToken cancellationToken)
    {
        var data = await LoadThingsAsync(cancellationToken);              // off the main thread
        await framework.Run(() => ApplyToGame(data), cancellationToken);  // main-thread work
    }

    public async ValueTask DisposeAsync() => await framework.Run(RemoveFromGame);
}
```

`Task LoadAsync(CancellationToken)`; inherits `IAsyncDisposable`. Load and unload are fully off the
main thread; the plugin is not "loaded" until `LoadAsync` completes; the token is cancelled after 60
s; `DisposeAsync` is not cancellable and must finish promptly.

