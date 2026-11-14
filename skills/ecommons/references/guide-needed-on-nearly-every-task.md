## Needed on nearly every task

### Logging and notifications

```csharp
PluginLog.Information("text");  // also Debug, Verbose, Warning, Error, Fatal. One string, no params
DuoLog.Error("text");           // Dalamud log AND game chat, prefixed with [InternalName]
Notify.Success("text");         // Dalamud notification: Success, Info, Warning, Error, Plain
InternalLog.Debug("text");      // in-plugin buffer only; InternalLog.PrintImgui() draws the viewer
exception.Log();                // extension: logs message and stack at Error. Also LogWarning(),
                                // LogDebug(), LogInfo(), LogVerbose(), LogFatal(), LogDuo()
```

`PluginLog` passes the string to Serilog as the message template, so literal braces can be
mangled. `DuoLog` prints to chat at every level, including Debug and Verbose.

### Safe, throttling, state checks

```csharp
GenericHelpers.Safe(() => Risky());                        // try/catch, logs the exception
GenericHelpers.Safe(() => Risky(), suppressErrors: true);  // swallow silently

if(EzThrottler.Throttle("MyPlugin.Click", 500)) Click();   // true at most once per 500 ms per name
if(EzThrottler.Check("MyPlugin.Click")) { }                // true if not throttled; does not arm
EzThrottler.Reset("MyPlugin.Click");
if(FrameThrottler.Throttle("MyPlugin.Frame", 10)) { }      // same, counted in frames (default 60)

GenericHelpers.IsOccupied()     // true for a fixed list of 35 busy Condition flags, and true when
                                // there is no local player. Does not cover InCombat or Mounted
GenericHelpers.IsScreenReady()  // false while NowLoading, FadeMiddle or FadeBack is visible
Player.Available                // local player object exists
Player.Interactable             // local player exists and is targetable
```

Signatures: `EzThrottler.Throttle(string name, int miliseconds = 500, bool rethrottle = false)` and
`Throttle(string name, TimeSpan ts, bool reThrottle = false)`. `true` means "not throttled, now
armed". With `rethrottle: true` a throttled call pushes the deadline out again. The parameter is
spelled `miliseconds`; pass it positionally. Throttle names are global per plugin: prefix them.
`EzThrottler<T>` and `FrameThrottler<T>` are instance versions with any key type.

### Player

`ECommons.GameHelpers.Player` is a static class over the local player. Most used: `Object`
(`IPlayerCharacter?`), `Available`, `Interactable`, `CID` (`ulong`), `Name`, `NameWithWorld`,
`Job` (`ECommons.ExcelServices.Job`), `Level`, `Position`, `Territory` (`RowRef<TerritoryType>`),
`HomeWorld` / `CurrentWorld` (`RowRef<World>`), `HomeWorldName` / `CurrentWorldName` (`string`),
`IsInDuty`, `IsInHomeWorld`, `IsMoving`, `Mounted`, `IsAnimationLocked`, `IsCasting`, `IsDead`,
`DistanceTo(Vector3 | Vector2 | IGameObject)`, plus raw pointers `Character`, `BattleChara`,
`GameObject`. `Player.Name` has no null check and throws without a player. The pre-rework class
is `ECommons.GameHelpers.LegacyPlayer.Player` (obsolete), where `Territory` is `uint` and the
worlds are `string`; AutoRetainer aliases it globally, so its code shows the legacy types →
`references/topic-game.md`.

### ImGuiEx essentials

`ECommons.ImGuiMethods.ImGuiEx` is a static partial class of several hundred members. Labels are
plain `string` almost everywhere. The most used, with exact signatures:

```csharp
ImGuiEx.Text(string s);  ImGuiEx.Text(Vector4 col, string s);  ImGuiEx.Text(uint col, string s);
ImGuiEx.TextWrapped(string s);  ImGuiEx.TextWrapped(Vector4? col, string s);
ImGuiEx.TextV(string s);             // vertically aligned to frame padding, for use next to widgets
ImGuiEx.TextCentered(string text);   ImGuiEx.TextCopy(string displayText, string? copyText = null);
ImGuiEx.Tooltip(string s);           // shows when the previous item is hovered; checks it itself
ImGuiEx.HelpMarker(string helpText, Vector4? color = null, string symbolOverride = null,
                   bool sameLine = true, bool preserveCursor = false);
bool ImGuiEx.IconButton(FontAwesomeIcon icon, string id = "ECommonsButton", Vector2 size = default,
                        bool enabled = true);
bool ImGuiEx.IconButtonWithText(FontAwesomeIcon icon, string id, bool enabled = true, ...);
bool ImGuiEx.Button(string label, bool enabled = true);
bool ImGuiEx.ButtonCtrl(string text, string affix = " (Hold CTRL)");
bool ImGuiEx.Checkbox(FontAwesomeIcon icon, string label, ref bool value, bool enabled = true);
bool ImGuiEx.ButtonCheckbox(string name, ref bool value, Vector4? color = null, ...);
bool ImGuiEx.CollectionCheckbox<T>(string label, T value, ICollection<T> collection,
                                   bool inverted = false, bool delayedOperation = false);
bool ImGuiEx.EnumCombo<T>(string name, ref T refConfigField, Func<T, bool> filter = null,
                          IDictionary<T, string> names = null, T[]? valuesList = null)
                          where T : Enum, IConvertible;   // valuesList: 3.2.1.23+
bool ImGuiEx.Combo<T>(string name, ref T refConfigField, IEnumerable<T> values,
                      Func<T, bool> filter = null, IDictionary<T, string> names = null);
ImGuiEx.SetNextItemFullWidth(int mod = 0);   ImGuiEx.SetNextItemWidthScaled(float width);
ImGuiEx.LineCentered(string id, Action func);
ImGuiEx.InputWithRightButtonsArea(string id, Action inputAction, Action rightAction);
bool ImGuiEx.HoveredAndClicked(string tooltip = null, ImGuiMouseButton btn = ImGuiMouseButton.Left,
                               bool requireCtrl = false);
ImGuiEx.EzTabBar(string id,
                 params (string name, Action function, Vector4? color, bool child)[] tabs);
bool ImGuiEx.BeginDefaultTable(string[] headers, bool drawHeader = true,
                 ImGuiTableFlags extraFlags = ImGuiTableFlags.None);   // then ImGui.EndTable()
ImGuiEx.EzTableColumns(string id, Action[] values, int? columns = null, ...);
ImGuiEx.TreeNodeCollapsingHeader(string name, Action action, ...);
bool ImGuiEx.Ctrl;  bool ImGuiEx.Shift;  bool ImGuiEx.Alt;     // properties
float .Scale() / Vector2 .Scale()            // extension: multiply by the UI scale
```

Behaviour that differs from what the names suggest:

- `enabled: false` on most widgets is alpha dimming plus a discarded result, not
  `ImGui.BeginDisabled`.
- `IconButton` has a constant default id. Two icon buttons with the same icon in one ID scope need
  explicit ids or `ImGui.PushID`.
- `ImGuiEx.SetNextItemWidth(float percent, int mod = 0)` takes a fraction of the available width,
  unlike `ImGui.SetNextItemWidth`. `ImGuiEx.SetNextItemWidth(200f)` means 200 times the width.
- `ImGuiEx.Spacing()` moves the cursor; it is not `ImGui.Spacing()`.
- `BeginDefaultTable` header prefixes: `~` makes the column stretch, `^` sets `NoSort`.
- `EnumCombo` and `Combo` have a search box and keep their search text in public static
  dictionaries keyed by label; `SearchableCombo` / `ExcelSheetCombo` share one search state
  across all instances.
- There is no non-nullable `ImGuiEx.InputInt`, and the non-nullable `ImGuiEx.InputFloat(width, ...)`
  exists only from 3.2.1.23; stock `ImGui` covers plain inputs on every version.
- Colors: `EColor.RedBright`, `EColor.GreenBright`, ... are `Vector4` fields (`EColor.Cya` is the
  real spelling for cyan); Dalamud's `ImGuiColors` works everywhere a `Vector4` is taken.

### Config, window, command

```csharp
C = EzConfig.Init<Config>();        // where T : new(). Loads DefaultConfig.json from the plugin
                                    // config directory or creates a new instance
EzConfig.Save();                    // also called by ECommonsMain.Dispose()
EzConfigGui.Init(Draw);             // creates a ConfigWindow, hooks UiBuilder.OpenConfigUi
EzConfigGui.Init(Draw, windowType: EzConfigGui.WindowType.Both);   // also OpenMainUi
EzConfigGui.WindowSystem.AddWindow(new OtherWindow());
EzConfigGui.Open();  EzConfigGui.Toggle();
EzCmd.Add("/myplugin", OnCommand, "help text");   // handler is (string command, string arguments)
```

`EzConfig` needs no interface (`IEzConfig` is obsolete). Use public fields in config classes so
they can be passed by `ref` to ImGui. A corrupt file throws out of `Init<T>()`; write errors are
only logged. `EzConfigGui.Init(Action draw, IPluginConfiguration config = null, string
nameOverride = null, WindowType windowType = WindowType.Config)`: pass `config` only for a stock
Dalamud configuration; with EzConfig leave it null, the `ConfigWindow` saves EzConfig on close by
itself. `EzCmd.Add` without a help message hides the command from help.

### TaskManager

```csharp
TaskManager = new(new(timeLimitMS: 20000, abortOnTimeout: true, showDebug: true));
TaskManager.Enqueue(() => Step(), "name");              // Func<bool>, Func<bool?> or Action
TaskManager.EnqueueDelay(500);
TaskManager.Insert(() => SubStep());                    // front of the queue
if(TaskManager.IsBusy) TaskManager.Abort();
```

One task function runs per framework tick on the framework thread. Return `true` = done, `false` =
call again next tick, `null` = abort request. Defaults: `TimeLimitMS` 30000, `AbortOnTimeout`
true, `AbortOnError` true, `ShowDebug` false. A per-task `TaskManagerConfiguration` overrides only
the properties it sets. `abortOnTimeout: false` drops the timed-out task and continues with the
rest of the queue. Full surface, stacks, events, `NeoTasks`, legacy class → `references/topic-taskmanager.md`.

### Addons

```csharp
if(TryGetAddonMaster<AddonMaster.SelectYesno>(out var m) && m.IsAddonReady) m.Yes();
if(TryGetAddonByName<AtkUnitBase>("Shop", out var addon) && IsAddonReady(addon))
    Callback.Fire(addon, true, 0, 3, 1);        // values: int, uint, float, bool, string, AtkValue
Callback.Fire(addon, true, -1);                 // the usual "close" callback
```

`TryGetAddonByName<T>(string Addon, out T* AddonPtr) where T : unmanaged` looks up instance 1 only
and checks neither visibility nor load state. `TryGetAddonMaster<T>(out T)` derives the addon name
from the class name and returns true for hidden or loading addons too, so always test
`IsAddonReady`. `GenericHelpers.IsAddonReady(AtkUnitBase*)` and the extension `addon->IsReady()`
are `IsVisible && UldManager.LoadedState == AtkLoadState.Loaded && IsFullyLoaded()`; the
ClientStructs property `addon->IsReady` (no parentheses) is a different check. 117 AddonMaster
classes exist at this pin → `references/topic-addonmaster.md`.

