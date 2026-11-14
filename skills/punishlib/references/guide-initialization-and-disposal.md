## Initialization and disposal

```csharp
namespace PunishLib;

public class PunishLibMain
{
    public static PunishConfig PunishConfig;
    public static void Init(IDalamudPluginInterface pluginInterface, string pluginName,
                            AboutPlugin about = null, params PunishOption[] opts);
    public static void Init(IDalamudPluginInterface pluginInterface, string pluginName,
                            params PunishOption[] opts);   // forwards with about = null
    public static void Dispose() { }                        // empty body
}
public enum PunishOption { DefaultKoFi }
public class AboutPlugin {
    public string Sponsor { get; set; }     // URL; null hides the "Sponsor" button
    public string Developer { get; set; }   // null selects the "by Puni.sh" header
}
```

What `Init` does, in order: stores `pluginName` and `pluginInterface` in internal statics; sets
the internal `About` to `about ?? new()`; loads `PunishConfig` from disk; constructs the internal
`SharedConfig` (creates `%AppData%/PunishConfig`, reads `PunishSharedConfig.json`, publishes each
stored value through `pluginInterface.GetOrCreateData`); and if `opts` contains
`PunishOption.DefaultKoFi` sets `About.Sponsor = "https://ko-fi.com/spetsnaz"`. The stored
`pluginName` is never read; the About tab shows the argument of `AboutTab.Draw`.

Ordering: call `ECommonsMain.Init(pluginInterface, this, ...)` **first**, then
`PunishLibMain.Init`. `Init` itself only touches the `IDalamudPluginInterface` you hand it, but
`AboutTab.Draw`, `EzConfig` and `GenericHelpers.Safe` go through ECommons' `Svc`, which
`ECommonsMain.Init` populates. Every consumer read calls them in that order.

Disposal: `PunishLibMain.Dispose()` has an empty body. PandorasBox and AutoRetainer call it;
AutoHook, Artisan and Saucy never do. Calling it is harmless; nothing is released either way.

