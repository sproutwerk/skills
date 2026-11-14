# Luna service registration, DI and plugin loader

Covers the `ServiceManager` worked example and member table, the `IService` marker interfaces,
singleton and scoped lifetimes, `Luna/PluginLoader/` and `ErrorWindow`. Verified against Luna
`52eb0fe` and ImSharp `4d09327` (branch `gc1.88`, ImGui 1.88). Paths are relative to the skill root.
Grep `references/api-index.txt` before using a member that is not shown here.

Luna `52eb0fe` merges the service rework: services are now singleton or scoped, and several
`ServiceManager` members changed. A consumer pinned before that merge (Glamourer `e165910` pins
Luna `dcc4ee7`) still has the older API; see [recent API changes](guide.md#recent-api-changes).

## Service registration (worked example)

Synchronous entry point, shaped after `Glamourer/Services/ServiceManager.cs` and
`Glamourer/Glamourer.cs` and adjusted to the `52eb0fe` signatures:

```csharp
using Dalamud.Plugin;
using ImSharp;
using Luna;

public sealed class MyPlugin : IDalamudPlugin
{
    public static readonly MainLogger Log = new("MyPlugin");
    private readonly ServiceManager _services;

    public MyPlugin(IDalamudPluginInterface pi)
    {
        try
        {
            _services = new ServiceManager(Log)
                .AddDalamudServices(pi)               // pi, pi.UiBuilder, every IDalamudService type
                .AddExistingService(Log)
                .AddSingleton<MessageService>()       // not an IService, so add it by hand
                .AddExistingService(this);
            _services.AddIServices(typeof(MyPlugin).Assembly);        // all concrete IService types
            _services.AddIServices(typeof(ServiceManager).Assembly);  // Luna's own IService types
            _services.BuildProvider();

            _ = _services.GetService<ImSharpDalamudContext>();
            _services.EnsureRequiredServices(CancellationToken.None); // every IRequiredService
            _services.GetService<MyWindowSystem>();   // your IUiService that adds windows
        }
        catch { Dispose(); throw; }
    }

    public void Dispose() => _services?.Dispose();
}
```

This example never calls `EnterScope()`. The provider is built with `ValidateScopes = false`, so
scoped registrations then resolve from the root provider and live as long as the manager. Penumbra
`51bf831` uses the scope instead: its plugin constructor calls `Services.EnterScope()` and its
`DisposeAsync` returns `Services.ExitScope()`, which disposes every scoped service while the
singletons survive for the loader and the error window.

`ServiceManager` (`Luna/Services/ServiceManager.cs`, `IDisposable` and `IAsyncDisposable`) wraps
`Microsoft.Extensions.DependencyInjection`:

| Member | Notes |
|---|---|
| `ServiceManager(MainLogger logger, string? name = null)` | Registers `ILogger<>` → `MainLogger<>`, `ILogger`, `LunaLogger`, `MainLogger`, itself; with `name`, a scoped `WindowSystem`. |
| `AddSingleton<T>()` / `AddScoped<T>()` | Constructed through the constructor with the **most parameters**; construction is timed in `Timers` (`StartTimeTracker`). |
| `AddSingleton<T>(Func<IServiceProvider, T> factory)` / `AddScoped<T>(Func<IServiceProvider, T> factory)` | |
| `AddGenericSingleton(Type queryType, Type? implementationType = null)` | Open generics. |
| `AddExistingService<T>(T service, bool takeOwnership = false)` | Singleton. With ownership, an `IDisposable` or `IAsyncDisposable` is disposed with the manager. |
| `AddDalamudServices(IDalamudPluginInterface pi)` | Resolves Dalamud services lazily through `pi.GetRequiredService(type)`. |
| `AddIServices(Assembly)` | Exported, non-abstract types implementing `IService`, skipping `IConstructedService`. A type that implements `IScopedService` is added scoped, anything else as a singleton. |
| `BuildProvider()` | `ValidateOnBuild = true`, `ValidateScopes = false`. `HasProvider` reports whether it ran. |
| `EnsureRequiredServices(CancellationToken cancel)` | Builds the provider if needed, then resolves every registered `IRequiredService`. |
| `EnterScope()` / `ExitScope()` (`ValueTask`) | One scope at a time; `EnterScope` throws without a provider or inside a scope. `InScope` reports the state. While a scope is open, `GetService` resolves from it. |
| `GetService<T>()`, `GetService(Type)`, `GetServicesImplementing<T>()` | Throw when the service is missing. |
| `Dispose()` / `DisposeAsync()` | Dispose the scope, the provider and owned objects. |

The provider itself is private at this pin; there is no public `Provider` property.

### Marker interfaces

`Luna/Services/IService.cs`:

| Interface | Meaning |
|---|---|
| `IService` | Base marker collected by `AddIServices`. Do not implement it alone. |
| `ISingletonService : IService` | One instance for the lifetime of the manager. |
| `IScopedService : IService` | One instance per scope, disposed by `ExitScope()`. |
| `IConstructedService : IService` | Implements a service interface but must not be registered. |
| `IRequiredService : IService` | Instantiated by `EnsureRequiredServices`. Carries no lifetime, so pair it with `ISingletonService` or `IScopedService`. |
| `IAsyncService` | `Awaiter`, `Finished`. No longer derives from `IService`. |
| `IAwaitedService : IRequiredService, IAsyncService` | |
| `IHookService : IAwaitedService` | `Address`, `Enable()`, `Disable()`. |
| `IDataContainer : ISingletonService`, `IAsyncDataContainer` | `Name`, `Time`, `Memory`, `TotalCount`. |
| `IApiService : IScopedService`, `IUiService : IScopedService` | API and UI markers. |

When Luna is compiled in Debug, registration runs a `[Conditional("DEBUG")]` check that throws if
an `IService` type implements none or more than one of `ISingletonService`, `IScopedService` and
`IConstructedService`, if an `IConstructedService` is registered, or if `AddSingleton<T>()` is used
for an `IScopedService` type (or `AddScoped<T>()` for one that is not). A Release build of Luna
skips the check, so a type that only implements `IService` is silently registered as a singleton.

Lifetimes of Luna's own services at this pin: `ImSharpDalamudContext`, `DynamisIpc`,
`KeyboardManager` and `FileCompactor` are singletons. `FrameworkManager`, `HookManager`,
`FastHook<T>`, `StartTimeTracker`, both `EventBase<…>` classes, `BaseFilePathProvider` and
`ConfigurationFile<TProvider>` are scoped, as is everything marked `IUiService` (`WindowSystem`,
`Window`, `TabBar<T>`, `ITab`, `IPanel`, `IHeader`, `IFooter`, `TwoPanelLayout`, `DragDropManager`,
`HookManagerDrawer`) or `IApiService` (`IpcObjectManager`).

## Plugin loader

`Luna/PluginLoader/` is the async entry point. Penumbra's whole loader class is
`public sealed class PenumbraLoader(IDalamudPluginInterface pluginInterface) : PluginLoader<Penumbra>(pluginInterface, "Penumbra")`.

`PluginLoader<TPlugin> : IPluginLoader, IAsyncDalamudPlugin` (abstract; constructor
`PluginLoader(IDalamudPluginInterface pluginInterface, string? pluginName)`) creates the
`MainLogger`, calls `TPlugin.CreateServiceManager`, adds itself as `IPluginLoader`, and builds the
provider. `CreateServiceManager` must therefore return a manager that has **not** built one; the
loader throws otherwise. It exposes `PluginInterface`, `Log`, `Services`, `Source`
(`PluginLoadSource`: `Normal`, `Testing`, `Dev`), `Version`, `CommitHash`, `GameVersion`
(`"Unknown"` until the game framework exists), `PluginName`, `GetWindowLabel(fixedLabel)`,
`IntermediaryPlugin`, `Plugin`, `HasIntermediaryPlugin` and `HasPlugin`.

`TPlugin` implements `IPluginDefinition<TSelf>` and `IAsyncDisposable`:

- Static abstract: `CreateServiceManager(pluginInterface, log)`, `PreInitializeAsync`,
  `InitializeLoggingAsync`, `LaunchIntermediaryAsync` (return `new NopIntermediary<TPlugin>(loader)`
  when no validation step is wanted) and `LaunchPluginAsync`.
- Instance stages, run by `InitializePluginAsync` in this order: `CreateBackupsAsync` and
  `LoadGameDataAsync` together, then `LoadAndMigrateConfigurationAsync`, `LoadPluginObjectsAsync`,
  `CreateGameInteropAsync`, `CreateUiAsync`, `CreateApiAsync`. Each takes
  `(PluginLoader<TSelf> loader, CancellationToken cancel)`. The stages are optional structure; the
  service provider may order construction instead.

`IIntermediaryPlugin<TPlugin>` runs before the main plugin: `ValidateAsync` returns `true` to
launch immediately, otherwise `OnValidationFailedAsync` runs and the intermediary stays loaded. If
a stage throws, the loader disposes the half-built plugin and calls
`OnPluginInitializationFailedAsync(PluginInitializationFailure, cancel)`; the
`PluginInitializationFailure(Stage, Exception)` record names the `InitializationStage`, and the
returned `FailureHandling` (`Handled` or `Rethrow`) decides whether the load fails.
`StubPluginLoader<TPlugin>` is an `IPluginLoader` for a plugin constructed without the loader.

## Error window

`Luna.ErrorWindow(IDalamudPluginInterface pluginInterface, LunaLogger log, string label, string name, ImSharpDalamudContext context)`
is the fallback window that shows a load failure and an update check. It creates its own
`WindowSystem` and no longer creates or disposes a context: pass the registered one, as Penumbra's
intermediary does with `Parent.Services.GetService<ImSharpDalamudContext>()`. Override
`DrawFailure()` to replace the message and `DrawDebugUtilities()` to add Debug-only tools.
