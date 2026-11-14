## Dependency injection

| Way | Form | Notes |
|---|---|---|
| Constructor | `public Plugin(IDalamudPluginInterface pi, ICommandManager cmd)` | Any service interface. Unknown parameter type = "An eligible ctor with satisfiable services could not be found". |
| Property | `[PluginService] public static IClientState ClientState { get; private set; } = null!;` | `PluginServiceAttribute` is `AttributeTargets.Property` only: **a field does not compile**. Static or instance, any visibility. |
| Service class | `pluginInterface.Create<Service>()` | Creates `T` via the same container (ctor + property injection). Returns `null` on failure and logs. `CreateAsync<T>` is the awaitable form. |
| Existing object | `pluginInterface.Inject(obj)` | Property injection only, returns `bool`. `InjectAsync`. |
| Resolve | `pluginInterface.GetService(typeof(IChatGui))` | `IDalamudPluginInterface : IServiceProvider` since API 14. |

`Create`/`Inject` take `params object[] scopedObjects`: extra objects that become resolvable by type
for that call. `[RequiredVersion]` was removed in API 10; delete it. All service interfaces derive
from the marker `IDalamudService` (API 14). Most services are plugin-scoped wrappers, so Dalamud
removes your commands, hooks, lifecycle listeners and DTR entries on unload as a safety net. Still
unregister everything yourself; reviewers require a complete dispose cycle.

