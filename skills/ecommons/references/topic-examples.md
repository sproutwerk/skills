# Worked examples

Covers: where each complete example lives, plus one skeleton plugin that combines bootstrap,
EzConfig, an EzConfigGui window, a command, a TaskManager and EzIPC. Verified against ECommons
`bc38dc7` (2026-11-08, v3.2.1.26, Dalamud API 15); nothing here was compiled. Paths are relative to
the skill root.

## Index of complete examples

| Example | Where |
|---|---|
| Minimal plugin bootstrap (`ECommonsMain.Init` / `Dispose`) | `references/topic-setup.md`, "Minimal plugin" |
| Config window: `EzConfig` + `EzConfigGui` + `EzCmd` + `ImGuiEx` tabs, combos, checkboxes | `references/topic-simplegui.md`, "Worked example" (with stock `IPluginConfiguration` and own-window variants) |
| TaskManager sequence that waits for `SelectString` / `SelectYesno` and drives them through AddonMaster with throttling | `references/topic-taskmanager.md`, "Worked example" |
| EzIPC provider class and consumer class in another plugin, with an event and the resulting channel names | `references/topic-ezipc.md`, "Worked example" |
| Reacting to an addon through `IAddonLifecycle` and wrapping `args.Addon` in an AddonMaster | `references/topic-addonmaster.md` |
| `[EzHook]` field + detour + `EzSignatureHelper.Initialize` | `references/topic-ezhook.md` |
| Subscribing to `ActionEffect`, `DirectorUpdate`, `MapEffect` | `references/topic-game-hooks.md` |
| Drawing through Splatoon from a plugin | `references/topic-splatoonapi.md` |
| Reading another plugin's state by reflection | `references/topic-reflection-interop.md` |

## Skeleton plugin

Every ECommons call below uses a signature that is quoted in the topic file named in the comment.
The game-specific parts (which menu entry to pick, which prompt to confirm) are placeholders.

```csharp
using Dalamud.Bindings.ImGui;
using Dalamud.Plugin;
using ECommons;
using ECommons.Automation.NeoTaskManager;
using ECommons.Automation.NeoTaskManager.Tasks;
using ECommons.Configuration;
using ECommons.DalamudServices;
using ECommons.EzIpcManager;
using ECommons.GameHelpers;
using ECommons.ImGuiMethods;
using ECommons.Logging;
using ECommons.SimpleGui;
using ECommons.Throttlers;
using ECommons.UIHelpers.AddonMasterImplementations;
using System;
using static ECommons.GenericHelpers;

namespace Courier;

public class Config
{
    public bool Enabled = true;            // fields: ImGui widgets take them by ref
    public string MenuEntry = "Purchase";
    public string ConfirmText = "Exchange";
}

public sealed unsafe class Courier : IDalamudPlugin
{
    internal static Courier P = null!;
    internal static Config C = null!;
    internal TaskManager TaskManager = null!;
    internal Ipc Ipc = null!;

    public Courier(IDalamudPluginInterface pluginInterface)
    {
        P = this;
        ECommonsMain.Init(pluginInterface, this);                       // topics/setup.md
        C = EzConfig.Init<Config>();                                    // topics/configuration.md
        EzConfigGui.Init(Draw, windowType: EzConfigGui.WindowType.Both); // topics/simplegui.md
        EzCmd.Add("/courier", OnCommand, "Open the Courier window\n/courier run - start");
        TaskManager = new(new(timeLimitMS: 20000, showDebug: true));    // topics/taskmanager.md
        Ipc = new();                                                    // topics/ezipc.md
    }

    public void Dispose()
    {
        TaskManager.Abort();
        ECommonsMain.Dispose();   // last: saves EzConfig, detaches the TaskManager, unregisters IPC,
                                  // removes the command and the window
    }

    private void OnCommand(string command, string arguments)
    {
        if(arguments == "run") Run();
        else EzConfigGui.Toggle();
    }

    private void Draw()
    {
        ImGui.Checkbox("Enabled", ref C.Enabled);
        ImGuiEx.SetNextItemWidthScaled(200f);
        ImGui.InputText("Menu entry", ref C.MenuEntry, 100);
        ImGuiEx.HelpMarker("Text of the SelectString entry to pick");
        if(ImGuiEx.Button("Run", !TaskManager.IsBusy)) Run();           // second argument: enabled
        ImGui.SameLine();
        if(ImGuiEx.Button("Abort", TaskManager.IsBusy)) TaskManager.Abort();
        ImGuiEx.Text(TaskManager.IsBusy ? EColor.YellowBright : EColor.GreenBright,
            TaskManager.IsBusy ? $"Working, {TaskManager.NumQueuedTasks} tasks left" : "Idle");
    }

    internal void Run()
    {
        if(!C.Enabled || TaskManager.IsBusy || !Player.Interactable) return;
        TaskManager.EnqueueTask(NeoTasks.WaitForNotOccupied());
        TaskManager.Enqueue(SelectEntry, "SelectEntry", new TaskManagerConfiguration());
        TaskManager.Enqueue(Confirm, "Confirm", new(timeLimitMS: 60000));
        TaskManager.Enqueue(() => !TryGetAddonMaster<AddonMaster.SelectYesno>(out _), "WaitClosed");
        TaskManager.Enqueue(() => Ipc.OnFinished(), "Notify subscribers");
    }

    // Func<bool?>: true = done, false = run again next frame, null = abort the queue.
    // A task that may return null gets a non-null configuration (see topics/taskmanager.md).
    private bool? SelectEntry()
    {
        if(!TryGetAddonMaster<AddonMaster.SelectString>(out var m) || !m.IsAddonReady) return false;
        foreach(var entry in m.Entries)
        {
            if(entry.Text.Contains(C.MenuEntry))
            {
                if(!EzThrottler.Throttle("Courier.SelectEntry", 500)) return false;
                entry.Select();
                return true;
            }
        }
        DuoLog.Error($"Menu entry {C.MenuEntry} not found");
        return null;
    }

    private bool Confirm()
    {
        if(TryGetAddonMaster<AddonMaster.SelectYesno>(out var m) && m.IsAddonReady
            && m.Text.Contains(C.ConfirmText))
        {
            if(EzThrottler.Throttle("Courier.Yes", 500))
            {
                m.Yes();
                return true;
            }
        }
        return false;
    }
}

public class Ipc
{
    public Ipc() => EzIPC.Init(this);           // prefix defaults to the plugin's InternalName

    [EzIPC] public bool IsBusy() => Courier.P.TaskManager.IsBusy;       // Courier.IsBusy
    [EzIPC] public void Run() => Courier.P.Run();                       // Courier.Run
    [EzIPCEvent] public Action OnFinished;                              // Courier.OnFinished
}
```

Consumer side, in another plugin that also uses ECommons:

```csharp
public class CourierIpc
{
    public CourierIpc() => EzIPC.Init(this, "Courier", SafeWrapper.AnyException);

    [EzIPC] public Func<bool> IsBusy;
    [EzIPC] public Action Run;
    [EzIPCEvent] private void OnFinished() => PluginLog.Information("Courier finished");
}
```

Notes on the choices made above:

- `ECommonsMain.Dispose()` is last; `TaskManager.Abort()` before it stops queued work from running
  during teardown. The TaskManager itself needs no explicit `Dispose`, `ECommonsMain.Dispose()`
  detaches every instance.
- `EzConfigGui.Init(Draw, ...)` is called without a `config` argument on purpose. The window it
  creates is a `ConfigWindow`, whose `OnClose` calls `EzConfig.Save()` whenever `EzConfig.Config`
  is not null.
- `WindowType.Both` hooks both `OpenConfigUi` and `OpenMainUi`. The default `WindowType.Config`
  hooks only `OpenConfigUi`, and Dalamud's plugin validation warns about a plugin with a main UI
  that has no `OpenMainUi` handler.
- `ImGuiEx.Button(label, enabled)` dims and ignores clicks when `enabled` is false; it is not
  `ImGui.BeginDisabled`.
- The consumer's `[EzIPC]` delegates are `null` until `EzIPC.Init` ran, and with
  `SafeWrapper.AnyException` they return `default` silently when Courier is not loaded.
- Addon pointers and AddonMaster wrappers are fetched inside each task, every frame. Never keep
  one in a field across frames.
