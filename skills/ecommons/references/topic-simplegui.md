# SimpleGui: EzConfigGui, ConfigWindow, PopupWindow, EzCmd

Covers: the one-call plugin window setup in `ECommons.SimpleGui` (`EzConfigGui`, `ConfigWindow`,
`PopupWindow`), command registration through `EzCmd`, and what `ECommonsMain.Dispose()` tears down
for them. Verified against ECommons `bc38dc7` (2026-11-08, v3.2.1.26, Dalamud
API 15); nothing here was compiled. Paths are relative to the skill root.

Sources: `references/ecommons/ECommons/SimpleGui/EzConfigGui.cs`,
`references/ecommons/ECommons/SimpleGui/ConfigWindow.cs`,
`references/ecommons/ECommons/SimpleGui/PopupWindow.cs`, `references/ecommons/ECommons/EzCmd.cs`,
`references/ecommons/ECommons/ECommonsMain.cs`.

Related: `references/topic-configuration.md` (EzConfig in depth), `references/topic-imguiex-core.md` and
`references/topic-imguiex-widgets.md` (`ImGuiEx`), `references/topic-imgui-helpers.md` (other windows, `Notify`,
colours). Stock `Window`/`WindowSystem`/`IUiBuilder` semantics are in the `dalamudimgui` skill;
`ICommandManager` is in the `dalamud` skill.

```csharp
using ECommons;                 // ECommonsMain, EzCmd, GenericHelpers
using ECommons.SimpleGui;       // EzConfigGui, ConfigWindow, PopupWindow
using ECommons.Configuration;   // EzConfig
using ECommons.DalamudServices; // Svc
using ECommons.ImGuiMethods;    // ImGuiEx, EColor, Notify
using Dalamud.Interface.Windowing; // Window, WindowSystem
```

All of this requires `ECommonsMain.Init(pluginInterface, this)` to have run (it fills `Svc` and
`ECommonsMain.Instance`). No `Module` flag is needed for anything in this file.

## EzConfigGui

`public static class EzConfigGui` owns one `WindowSystem` and one "main" window per plugin.

| Member | Notes |
|---|---|
| `static WindowSystem WindowSystem { get; internal set; }` | `null` until `Init`; set back to `null` by `ECommonsMain.Dispose()`. |
| `static Window? Window { get; private set; }` | The main window (a `ConfigWindow` or the instance you passed). |
| `static WindowType Type { get; private set; }` | The `windowType` passed to `Init`. |
| `static void Init(Action draw, IPluginConfiguration config = null, string nameOverride = null, WindowType windowType = WindowType.Config)` | Creates a `ConfigWindow` that calls `draw`. |
| `static T Init<T>(T window, IPluginConfiguration config = null, string nameOverride = null, WindowType windowType = WindowType.Config) where T : Window` | Uses your window instance; returns it. |
| `static void Open()` | `Window.IsOpen = true` if `Window` is not null. |
| `static void Open(string cmd = null, string args = null)` | Calls `Open()`. Shape matches a command handler. |
| `static void Toggle()` | `Window?.Toggle()`. |
| `static T? GetWindow<T>() where T : Window` | First window in `WindowSystem.Windows` whose runtime type is exactly `T`. |
| `static void RemoveWindow<T>() where T : Window` | `WindowSystem.RemoveWindow` on the result of `GetWindow<T>()`, if any. |
| `enum WindowType { Config, Main, Both }` | Nested: `EzConfigGui.WindowType`. |

Internal, not reachable from a plugin: `Draw`, `OnClose`, `OnOpen` (all `Action`), `Config`
(`IPluginConfiguration`). There is no public setter for `OnOpen`/`OnClose`, so those `ConfigWindow`
callbacks are always null for a consuming plugin; subclass `ConfigWindow` or pass your own `Window`
if you need open/close hooks.

### What Init does

Both public overloads store their first argument (`Draw = draw` or `Window = window`) and then run
the same private routine:

1. Throws `new Exception("ConfigGui already initialized")` if `WindowSystem != null`. One call per
   plugin load.
2. `WindowSystem = new($"ECommons@{DalamudReflector.GetPluginName()}")`. `GetPluginName()` returns
   `Svc.PluginInterface?.InternalName ?? "Not initialized"`, so the ImGui namespace is
   `ECommons@<InternalName>`.
3. `Config = config`.
4. `Window ??= new ConfigWindow(nameOverride)`, then `WindowSystem.AddWindow(Window)`.
5. `Type = windowType`.
6. `Svc.PluginInterface.UiBuilder.Draw += WindowSystem.Draw` (always).
7. Hooks `EzConfigGui.Open` (the parameterless overload) by `windowType`:

| `WindowType` | `UiBuilder` events subscribed |
|---|---|
| `Config` (default) | `OpenConfigUi` |
| `Main` | `OpenMainUi` |
| `Both` | `OpenConfigUi` and `OpenMainUi` |

Consequences:

- `nameOverride` only affects the `ConfigWindow` created in step 4. With `Init<T>(window, ...)` the
  window already exists, so `nameOverride` is ignored and the title is whatever your constructor
  passed to `Window`.
- With the default `WindowType.Config`, Dalamud's plugin installer shows only the settings (cog)
  button. If you also want the "Open" button, pass `WindowType.Both`, or subscribe
  `OpenMainUi` yourself (YesAlready does the latter, see below).
- The window starts closed. Nothing opens it until `Open()`/`Toggle()` or a UiBuilder event.

### Window name / title

`ConfigWindow`'s constructor passes this to `Window`:

```csharp
$"{name ?? DefaultPluginName}###{DalamudReflector.GetPluginName()}"
// DefaultPluginName:
$"{DalamudReflector.GetPluginName()} v{ECommonsMain.Instance.GetType().Assembly.GetName().Version}"
```

So the default title is `<InternalName> v<plugin assembly version>` (four-part version of the
assembly that contains your `IDalamudPlugin` class, not the manifest's version string and not the
ECommons version), and the ImGui ID is always `###<InternalName>`. Because the ID is fixed, changing
the visible title does not reset saved window position or size.

## ConfigWindow

`public class ConfigWindow : Window`

| Member | Notes |
|---|---|
| `string? ConfigWindowName` | The `name` passed to the constructor (may be null). |
| `ConfigWindow(string? name = null)` | Sets `SizeConstraints` to min `(200, 200)`, max `(float.MaxValue, float.MaxValue)`. Flags are left at `ImGuiWindowFlags.None`. |
| `static string DefaultPluginName` | See above. |
| `void SetSuffix(string suffix)` | Rewrites `WindowName` to `"<name or default> <suffix>###<InternalName>"`; `null` removes the suffix. |
| `override void Draw()` | `GenericHelpers.Safe(EzConfigGui.Draw)`: exceptions from your draw delegate are caught and logged rather than propagated to Dalamud. |
| `override void OnOpen()` | Invokes internal `EzConfigGui.OnOpen` (null for plugins). |
| `override void OnClose()` | Saves configuration, see below. |

`ConfigWindow` is not sealed and not abstract. Subclassing it and passing the instance to
`EzConfigGui.Init<T>` keeps the autosave-on-close behaviour while letting you set flags, size and
title bar buttons in the constructor. Override `Draw()` in that case (the base implementation draws
`EzConfigGui.Draw`, which the window overload never sets), and call `base.OnClose()` if you override
`OnClose()`.

To customise the window created by the `Init(Action ...)` overload, mutate it afterwards:

```csharp
EzConfigGui.Init(DrawUi);
EzConfigGui.Window.Size = new(600, 400);
EzConfigGui.Window.SizeCondition = ImGuiCond.FirstUseEver;
if(EzConfigGui.Window is ConfigWindow w) w.SetSuffix("(beta)");
```

### Saving on close

`ConfigWindow.OnClose()` runs two independent checks, in this order:

1. If the `config` passed to `EzConfigGui.Init` is not null:
   `Svc.PluginInterface.SavePluginConfig(config)` and `PluginLog.Debug("Configuration saved")`.
2. If `EzConfig.Config != null` (that is, `EzConfig.Init<T>()` or `EzConfig.Set` was used):
   `EzConfig.Save()` and the same debug line.
3. Invokes internal `EzConfigGui.OnClose`.

Therefore:

- A plugin that uses `EzConfig` gets autosave on close without passing anything as `config`. Leave
  `config` null; passing the same object there as well would additionally write it through
  Dalamud's own serializer to `pluginConfigs/<InternalName>.json`, and the parameter type requires
  `IPluginConfiguration` anyway.
- A plugin that uses stock Dalamud configuration passes its `IPluginConfiguration` as `config`.
- Autosave on close happens only when the main window is a `ConfigWindow` (or subclass that calls
  `base.OnClose()`). With an arbitrary `Window` passed to `Init<T>`, closing saves nothing.
- `OnClose` fires when the window transitions from open to closed. If the plugin is unloaded while
  the window is open, the save comes from `ECommonsMain.Dispose()` instead (next section).

## Adding more windows

`EzConfigGui.WindowSystem` is a normal Dalamud `WindowSystem`; add further windows straight after
`Init`. They are drawn by the same `UiBuilder.Draw` subscription and removed in
`ECommonsMain.Dispose()`.

```csharp
EzConfigGui.Init(new MainWindow().Draw);
EzConfigGui.WindowSystem.AddWindow(new ZoneListWindow());
EzConfigGui.WindowSystem.AddWindow(new ConditionsListWindow());
// later, anywhere:
EzConfigGui.GetWindow<ZoneListWindow>()?.Toggle();
```

That is YesAlready's pattern (`YesAlready/YesAlready.cs`, `UI/Tabs/YesNo.cs`). Notes:

- `WindowSystem.AddWindow` throws `ArgumentException` when another window has the same
  `WindowName`.
- `GetWindow<T>` compares `w.GetType() == typeof(T)`; asking for a base class does not find a
  derived window. It dereferences `WindowSystem`, so it throws `NullReferenceException` before
  `Init` and after `ECommonsMain.Dispose()`.
- Extra windows do not autosave on close unless they derive from `ConfigWindow`.
- Windows that implement `IDisposable` are not disposed by ECommons; `RemoveAllWindows()` only
  clears the list. Dispose them yourself before `ECommonsMain.Dispose()`.

## PopupWindow

`public class PopupWindow : Window`: a single-use window that registers itself and unregisters
when closed. It does not use `EzConfigGui.WindowSystem`.

| Constructor | Title |
|---|---|
| `PopupWindow(Action content)` | `"{Svc.PluginInterface.Manifest.Name} - Notification##{Guid.NewGuid()}"` |
| `PopupWindow(string title, Action content)` | `"{title}##{Guid.NewGuid()}"` |

Both use flags `ImGuiWindowFlags.NoSavedSettings | ImGuiWindowFlags.NoCollapse` and
`forceMainWindow: true`. Construction sets size constraints min `new(100)`, max
`new(float.MaxValue)`, `RespectCloseHotkey = false`, `AllowPinning = false`, creates a private
`WindowSystem`, adds itself, sets `IsOpen = true` and subscribes `UiBuilder.Draw`. `OnClose()`
unsubscribes `UiBuilder.Draw`. `Draw()` invokes `content` without a try/catch.

```csharp
new PopupWindow(() =>
{
    ImGuiEx.Text("Feature X was disabled because it conflicts with plugin Y.");
});
```

(AutoRetainer, `Helpers/Utils.cs`.) Just constructing it is enough; do not keep or re-open the
instance. `ECommonsMain.Dispose()` does nothing for popups that are still open; Dalamud drops the
`UiBuilder.Draw` subscribers when the plugin's UiBuilder is disposed.

## EzCmd

`public static class EzCmd` (namespace `ECommons`).

```csharp
public static void Add(string command, IReadOnlyCommandInfo.HandlerDelegate action,
    string helpMessage = null, int displayOrder = -1)
```

- `HandlerDelegate` is Dalamud's `void (string command, string arguments)`.
- Builds `new CommandInfo(action) { HelpMessage = helpMessage ?? "", ShowInHelp = helpMessage !=
  null, DisplayOrder = displayOrder }` and calls `Svc.Commands.AddHandler(command, cInfo)`.
- Omitting `helpMessage` therefore hides the command from `/xlhelp` and the installer's command
  list. Use that for aliases you do not want listed.
- `command` must include the leading slash (`"/myplugin"`).
- The command string is appended to the internal `RegisteredCommands` list before `AddHandler` is
  called; `ECommonsMain.Dispose()` calls `Svc.Commands.RemoveHandler` for each entry. Do not remove
  these handlers yourself.
- The return value of `AddHandler` is discarded, so a failed registration (duplicate command) is
  silent.
- Not obsolete. The source carries a commented-out `[Obsolete]` that pointed to the `[Cmd]`
  attribute (`ECommons.Commands.CmdAttribute`, handled by `CmdManager`, see
  `references/ecommons/ECommons/Commands/CmdAttribute.cs`); both mechanisms are live.

`EzConfigGui.Open` can be passed directly as the handler because of its
`(string cmd, string args)` overload:

```csharp
EzCmd.Add("/myplugin", EzConfigGui.Open, "Open the configuration window");
```

## Teardown in ECommonsMain.Dispose()

Relevant steps, in source order:

1. `EzConfig.Save()` if `EzConfig.Config != null`, then `EzConfig.Dispose()` (clears `OnSave`).
2. If `EzConfigGui.WindowSystem != null`:
   - `UiBuilder.OpenMainUi -= EzConfigGui.Open` when `Type` is `Main` or `Both`;
   - `UiBuilder.OpenConfigUi -= EzConfigGui.Open` when `Type` is `Config` or `Both`;
   - `UiBuilder.Draw -= EzConfigGui.Draw`;
   - if the `config` given to `Init` is not null: `Svc.PluginInterface.SavePluginConfig(config)` and
     `Notify.Info("Configuration saved")`;
   - `EzConfigGui.WindowSystem.RemoveAllWindows()`; `EzConfigGui.WindowSystem = null`.
3. `Svc.Commands.RemoveHandler(x)` for every command added through `EzCmd.Add`.

Notes on that code:

- The `Draw` unsubscription removes the internal `EzConfigGui.Draw` delegate (your draw callback),
  which was never subscribed; the delegate that `Init` subscribed is `WindowSystem.Draw`. The
  subscription is therefore left in place until Dalamud disposes the plugin's UiBuilder. It is
  harmless because the window list has been emptied, but do not rely on ECommons to detach it.
- `Window`, `Type`, `Draw` and `Config` are not reset. `Init` cannot be called again after
  `Dispose()` in the same load context in any meaningful way; the static state only goes away with
  the plugin's assembly.
- Any handler you attached yourself (for example `OpenMainUi += EzConfigGui.Toggle`) is yours to
  remove, before calling `ECommonsMain.Dispose()`.
- Call `ECommonsMain.Dispose()` last in your plugin's `Dispose()`; `Svc` is used by your own
  cleanup and by the steps above.

## How real plugins wire it

| Plugin | Window setup | Config | Commands |
|---|---|---|---|
| YesAlready (c531fb8) | `EzConfigGui.Init(new MainWindow().Draw)`, extra windows via `EzConfigGui.WindowSystem.AddWindow`, manual `OpenMainUi += EzConfigGui.Toggle` | Dalamud `GetPluginConfig()`; saves explicitly with its own `C.Save()`; passes no `config` to `Init` | `EzCmd.Add(Command, OnCommand, "Opens the plugin window.", int.MinValue)` plus aliases |
| AutoRetainer (cc020df) | Own `WindowSystem`, own `UiBuilder` subscriptions; does not use `EzConfigGui` | `EzConfig.Migrate<Config>(); config = EzConfig.Init<Config>();` | `EzCmd.Add("/autoretainer", CommandHandler, "<multi-line help>")`, `EzCmd.Add("/ays", CommandHandler)` |
| PandorasBox (3a35634) | Own windows; does not use `EzConfigGui` or `EzCmd` | own | own |

`EzConfigGui` fits plugins with one main window plus a few auxiliaries. Plugins with overlays that
need separate lifetime or draw conditions tend to keep their own `WindowSystem`.

## Worked example

Assembled from the verified signatures above, mirroring YesAlready's window wiring and
AutoRetainer's `EzConfig` usage. `EzConfig.Init<T>()` only requires `where T : new()`;
`IEzConfig` is `[Obsolete("You are no longer required to implement IEzConfig")]`.

```csharp
using Dalamud.Bindings.ImGui;
using Dalamud.Plugin;
using ECommons;
using ECommons.Configuration;
using ECommons.DalamudServices;
using ECommons.ImGuiMethods;
using ECommons.SimpleGui;
using System.Collections.Generic;

namespace MyPlugin;

public enum Mode { Off, Passive, Active }

public class Config
{
    public bool Enabled = true;                 // fields, not properties: they are passed by ref
    public Mode Mode = Mode.Passive;
    public int Delay = 500;
    public string Note = "";
    public HashSet<uint> Territories = [];
}

public sealed class MyPlugin : IDalamudPlugin
{
    internal static MyPlugin P = null!;
    internal static Config C = null!;

    public MyPlugin(IDalamudPluginInterface pluginInterface)
    {
        P = this;
        ECommonsMain.Init(pluginInterface, this);          // first: fills Svc, ECommonsMain.Instance

        // EzConfig.Migrate<Config>();                     // optional, only before Init, only when
                                                           // moving off Dalamud's own config file
        C = EzConfig.Init<Config>();                       // loads or creates the default config

        EzConfigGui.Init(DrawUi, windowType: EzConfigGui.WindowType.Both);
        // config: left null on purpose, ConfigWindow.OnClose calls EzConfig.Save() by itself

        EzCmd.Add("/myplugin", OnCommand, "Open the configuration window\n/myplugin toggle - enable or disable");
        EzCmd.Add("/myp", OnCommand);                      // no help text: hidden alias
    }

    private void OnCommand(string command, string arguments)
    {
        if(arguments == "toggle")
        {
            C.Enabled = !C.Enabled;
            EzConfig.Save();
            Notify.Info(C.Enabled ? "Enabled" : "Disabled");
        }
        else
        {
            EzConfigGui.Toggle();
        }
    }

    private void DrawUi()
    {
        ImGuiEx.EzTabBar("##main",
            ("General", DrawGeneral, null, true),
            ("Zones", DrawZones, null, true),
            ("About", DrawAbout, EColor.OrangeBright, false));
    }

    private void DrawGeneral()
    {
        ImGui.Checkbox("Enabled", ref C.Enabled);
        ImGuiEx.HelpMarker("Master switch. Also available as /myplugin toggle.");

        ImGuiEx.SetNextItemWidthScaled(150f);
        ImGuiEx.EnumCombo("Mode", ref C.Mode);

        ImGuiEx.SetNextItemWidthScaled(150f);
        ImGui.InputInt("Delay, ms", ref C.Delay);

        ImGuiEx.SetNextItemFullWidth();
        ImGui.InputTextWithHint("##note", "Note...", ref C.Note, 100);

        if(ImGuiEx.ButtonCtrl("Reset"))
        {
            C = EzConfig.Set(new Config());
            EzConfig.Save();
        }
        ImGuiEx.Tooltip("Restores default settings");
    }

    private void DrawZones()
    {
        var current = Svc.ClientState.TerritoryType;
        ImGuiEx.CollectionCheckbox($"Active in current zone ({current})", (uint)current, C.Territories);
        ImGuiEx.Text(EColor.GreenBright, $"{C.Territories.Count} zones selected");
    }

    private void DrawAbout()
    {
        ImGuiEx.TextWrapped($"{ConfigWindow.DefaultPluginName}. Settings are saved when this window closes.");
    }

    public void Dispose()
    {
        // own cleanup first
        ECommonsMain.Dispose();   // saves EzConfig, removes windows, removes EzCmd commands
    }
}
```

`ImGuiEx` members used, all in `references/ecommons/ECommons/ImGuiMethods/ImGuiEx/`:

| Member | File |
|---|---|
| `EzTabBar(string id, params (string name, Action function, Vector4? color, bool child)[] tabs)` | `ImGuiEx.cs` |
| `HelpMarker(string helpText, Vector4? color = null, string symbolOverride = null, bool sameLine = true, bool preserveCursor = false)` | `ImGuiEx.cs` |
| `SetNextItemWidthScaled(float width)`, `SetNextItemFullWidth(int mod = 0)` | `ImGuiEx.cs` |
| `Tooltip(string s)` (checks `ImGui.IsItemHovered()` itself) | `ImGuiEx.cs` |
| `EnumCombo<T>(string name, ref T refConfigField, Func<T, bool> filter = null, IDictionary<T, string> names = null, T[]? valuesList = null) where T : Enum, IConvertible` | `EnumCombo.cs` |
| `ButtonCtrl(string text, string affix = " (Hold CTRL)")` | `Button.cs` |
| `CollectionCheckbox<T>(string label, T value, ICollection<T> collection, bool inverted = false, bool delayedOperation = false)` | `CollectionCheckbox.cs` |
| `Text(Vector4 col, string s)`, `TextWrapped(string s)` | `Text.cs` |

`EzTabBar` with `child: true` wraps the tab body in `ImGui.BeginChild(name + "child")`, pushes ID
`name + "tab"`, and catches and logs exceptions thrown by the tab function. Tuples whose `name` is
null are skipped, which is the idiom for conditional tabs.

Variant with stock Dalamud configuration (`Configuration : IPluginConfiguration`):

```csharp
C = Svc.PluginInterface.GetPluginConfig() as Configuration ?? new Configuration();
EzConfigGui.Init(DrawUi, C);   // ConfigWindow.OnClose -> Svc.PluginInterface.SavePluginConfig(C)
```

Variant with your own window class:

```csharp
public class MainWindow : ConfigWindow
{
    public MainWindow() : base("My Plugin")   // title "My Plugin###<InternalName>"
    {
        Flags = ImGuiWindowFlags.NoScrollbar;
    }
    public override void Draw() { /* ... */ }
}

var window = EzConfigGui.Init(new MainWindow());   // returns the same instance, typed
```

## Gotchas

- `EzConfigGui.Init` twice throws `Exception("ConfigGui already initialized")`.
- `nameOverride` is ignored by the `Init<T>(T window, ...)` overload.
- The `config` parameter is `IPluginConfiguration`; an `EzConfig` class that does not implement it
  cannot be passed and does not need to be.
- Autosave on close is a `ConfigWindow` feature, not an `EzConfigGui` feature. A plain `Window`
  subclass passed to `Init<T>` never triggers it.
- When a Dalamud `config` was passed, unloading the plugin shows a "Configuration saved"
  notification (`Notify.Info`) from `ECommonsMain.Dispose()`. The `EzConfig` path saves silently.
- Default `WindowType.Config` hooks only `OpenConfigUi`. Dalamud's `PluginValidator` reports
  `NoMainUiProblem` when nothing is subscribed to `OpenMainUi`; use `WindowType.Both` or `Main`, or
  subscribe it
  yourself.
- `EzConfigGui.OnOpen`/`OnClose`/`Draw`/`Config` are `internal`; they cannot be assigned from a
  plugin.
- The default title uses the plugin assembly's `AssemblyName.Version`, which is `0.0.0.0` unless
  the csproj sets `Version`/`AssemblyVersion`.
- `ConfigWindow.Draw` swallows exceptions (logged through `GenericHelpers.Safe`). An ImGui
  push/begin left unbalanced by the exception is not repaired; keep pushes and pops exception-safe
  (`ImRaii`, see `dalamudimgui`).
- `EzCmd.Add` without `helpMessage` sets `ShowInHelp = false`.
- Commands added with `Svc.Commands.AddHandler` directly are not tracked by `EzCmd` and must be
  removed by the plugin.
- `PopupWindow` instances each hold their own `WindowSystem` and `UiBuilder.Draw` subscription
  until closed; do not create one per frame.
