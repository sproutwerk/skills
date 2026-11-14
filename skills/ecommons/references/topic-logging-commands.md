# Logging, commands, chat output, DTR, context-menu prefix

Covers: `ECommons.Logging` (`PluginLog`, `DuoLog`, `InternalLog`), exception logging extensions,
`SimpleLogger`, the `[Cmd]` / `[SubCmd]` command attributes and `EzCmd`, `ECommons.ChatMethods`
(`ChatPrinter`, `Sender`, `UIColor`), `EzDtr` and `ContextMenuPrefixRemover`. Verified against
ECommons `bc38dc7` (2026-11-08, v3.2.1.26, Dalamud
API 15); nothing here was compiled. Paths are relative to the skill root.

Stock `IPluginLog`, `ICommandManager`, `IChatGui`, `IDtrBar`, `IContextMenu` semantics are in the
`dalamud` skill; ECommons reaches them as `Svc.Log`, `Svc.Commands`, `Svc.Chat`, `Svc.DtrBar`,
`Svc.ContextMenu`.

```csharp
using ECommons;                 // ECommonsMain, EzCmd, SimpleLogger, GenericHelpers (e.Log() etc.)
using ECommons.Logging;         // PluginLog, DuoLog, InternalLog, InternalLogMessage
using ECommons.Commands;        // CmdAttribute, SubCmdAttribute, CmdManager
using ECommons.ChatMethods;     // ChatPrinter, Sender, UIColor
using ECommons.EzDTR;           // EzDtr
using ECommons.EzContextMenu;   // ContextMenuPrefixRemover (RemovePrefix extension)
```

## What ECommonsMain.Init / Dispose do in this area

`references/ecommons/ECommons/ECommonsMain.cs`

`ECommonsMain.Init(pluginInterface, instance, params Module[] modules)`, regardless of modules:

- Logs a banner at Information (`This is ECommons v... and <InternalName> v.... Plugin's Instance
  ID is 0x........ Hello!`) unless `ECommonsMain.ReducedLogging` is `true`.
- Sets `Svc.Log.MinimumLogLevel = LogEventLevel.Verbose;` unconditionally. Dalamud's default for
  installed (non-dev) plugins is `Debug`, so with ECommons every plugin emits Verbose lines to the
  Dalamud log whenever Dalamud's global level allows them. Set `Svc.Log.MinimumLogLevel` again
  after `Init` to undo this.
- Calls `CmdManager.Init` (inside `GenericHelpers.Safe`), which registers every `[Cmd]` method of
  the plugin instance. No module flag is involved.

`public static bool ReducedLogging = false;` is a plain static field, changeable at any time. It
suppresses ECommons' own informational chatter: the banner and "module has been requested" lines
in `Init`, and registration messages in `EzIPC`, `EzHook` and `SingletonServiceManager`. It does
not change `MinimumLogLevel` and does not affect `PluginLog` calls made by the plugin. To silence
the banner it must be set before `ECommonsMain.Init`.

`ECommonsMain.Dispose()` for this area:

| Call | Effect |
|---|---|
| `CmdManager.Dispose()` | `Svc.Commands.RemoveHandler` for every `[Cmd]` command |
| `foreach(var x in EzCmd.RegisteredCommands) Svc.Commands.RemoveHandler(x)` | Removes every `EzCmd.Add` command |
| `ContextMenuPrefixRemover.Dispose()` | Unregisters the `ContextMenu` addon listener |
| `EzDtr.DisposeAll()` | Disposes every `EzDtr` (unhooks update, removes the DTR entry) |

`InternalLog.Messages` is not cleared, and `SimpleLogger` instances are not disposed; dispose
those yourself.

## PluginLog

`public static class PluginLog`, `references/ecommons/ECommons/Logging/PluginLog.cs`. Static
wrapper over `Svc.Log` (Dalamud `IPluginLog`).

```csharp
public static void Information(string s)
public static void Error(string s)
public static void Fatal(string s)
public static void Debug(string s)
public static void Verbose(string s)
public static void Warning(string s)
// aliases
public static void LogInformation(string s)   // -> Information
public static void LogError(string s)         // -> Error
public static void LogFatal(string s)         // -> Fatal
public static void LogDebug(string s)         // -> Debug
public static void LogVerbose(string s)       // -> Verbose
public static void LogWarning(string s)       // -> Warning
public static void Log(string s)              // -> Information
// 3.2.1.23+
public static bool TraceEnabled = false;
public static void Trace(string s)
```

Each level method does two things: `Svc.Log.<Level>($"{s}")`, then
`Svc.Framework?.RunOnFrameworkThread(() => InternalLog.Messages.PushBack(new(s, level)))`.

`Trace` is an opt-in channel, not a log level. It does nothing unless `PluginLog.TraceEnabled` is
`true`; then it writes `Svc.Log.Information($"[Trace] {s}")`. The entry it pushes to `InternalLog`
is the literal `"[Trace] "` without `s`, so the in-plugin log viewer shows only the marker; read
trace output in `/xllog`.

Differences from Dalamud `IPluginLog`:

| | `PluginLog` (ECommons) | `IPluginLog` (`Svc.Log`) |
|---|---|---|
| Access | static, no injection | service instance |
| Arguments | one `string` only | `(string messageTemplate, params object[] values)` and `(Exception?, string, params object[])` |
| Exception overloads | none; use `e.Log()` extensions or `e.ToStringFull()` | yes |
| Templates | none; the string is passed as the message template itself | Serilog message templates |
| In-plugin buffer | also appended to `InternalLog.Messages` | no |
| `MinimumLogLevel`, `Logger` | not exposed; use `Svc.Log` | yes |

- Use string interpolation: `PluginLog.Debug($"value={x}")`. Because the interpolated string is
  handed to Serilog as the template with no values, literal `{` and `}` in the final text (JSON
  dumps, for example) are parsed as template holes and may render oddly in the Dalamud log; the
  `InternalLog` copy is always the raw string.
- Callable from any thread. The `InternalLog` append is marshalled to the framework thread, so
  off-thread messages appear in the in-plugin viewer slightly later.
- Requires `Svc` to be initialised (`Svc.Log` is dereferenced without a null check). Do not call
  before `ECommonsMain.Init`.
- The in-plugin copy is recorded regardless of `MinimumLogLevel`.

## InternalLog and InternalLogMessage

`references/ecommons/ECommons/Logging/InternalLog.cs`,
`references/ecommons/ECommons/Logging/InternalLogMessage.cs`

In-memory log of the last 1000 messages plus an ImGui viewer. `public class InternalLog` (not
declared static, but every member is static).

| Member | Notes |
|---|---|
| `static readonly CircularBuffer<InternalLogMessage> Messages` | Capacity 1000, oldest entries overwritten. `ECommons.CircularBuffers.CircularBuffer<T>`, `IEnumerable<T>`. |
| `static void Information/Error/Fatal/Debug/Verbose/Warning(string s)` | Append to the buffer only. Nothing is written to the Dalamud log. |
| `static void LogInformation/LogError/LogFatal/LogDebug/LogVerbose/LogWarning(string s)`, `Log(string s)` | Aliases; `Log` is Information. |
| `static void PrintImgui()` | Draws the viewer into the current window. |
| `static (string, Action, Vector4, bool) ImGuiTab(bool draw = true)` | Tab tuple `("Log", PrintImgui, ImGuiColors.DalamudGrey3, false)`; name is `null` when `draw` is `false`. Shaped for `ImGuiEx.EzTabBar`. |
| `enum FilterType : long` (`[Flags]`) | `Verbose`, `Debug`, `Info`, `Warning`, `Error`, `Fatal`, `Default = ~0`. Only used by the viewer's private filter state. |

```csharp
public record struct InternalLogMessage
{
    public string Message;
    public LogEventLevel Level;      // Serilog.Events
    public DateTimeOffset Time;      // DateTimeOffset.Now at construction
    public InternalLogMessage(string Message, LogEventLevel Level = LogEventLevel.Information)
}
```

`PrintImgui()` draws: an autoscroll checkbox, `Copy all` (copies messages as
`[Level@Time] Message` lines), `Clear`, a text filter (matches message substring case-insensitively,
or the exact level name), a filter icon button opening a per-level checkbox popup, then a child
region listing `> [Time] Message` coloured by level; each line is click-to-copy
(`ImGuiEx.TextWrappedCopy`). There is no clipping: all retained messages are drawn every frame.
Filter, search and autoscroll state are static and not persisted.

- The direct `InternalLog.*` methods push on the calling thread and the buffer is not
  synchronised. Call them from the framework/draw thread, or use `PluginLog`, which marshals.
- Use `InternalLog.*` for high-volume diagnostics you want visible in the plugin's own log tab
  without writing to `dalamud.log`; use `PluginLog.*` for both.

```csharp
// a Dalamud Window (AutoRetainer UI/Windows/LogWindow.cs)
internal class LogWindow : Window
{
    public LogWindow() : base("AutoRetainer log") { }
    public override void Draw() => InternalLog.PrintImgui();
}

// as a tab
ImGuiEx.EzTabBar("tabs", ("Settings", DrawSettings, null, true), InternalLog.ImGuiTab());
```

## DuoLog

`public static class DuoLog`, `references/ecommons/ECommons/Logging/DuoLog.cs`. Logs to the Dalamud
log (through `PluginLog`, hence also `InternalLog`) and prints the same text to game chat.

```csharp
public static void Information(string s)   // chat colour key 3
public static void Debug(string s)         // 4
public static void Verbose(string s)       // 5
public static void Warning(string s)       // 540
public static void Error(string s)         // 17
public static void Fatal(string s)         // 19
```

- The message is prefixed `[<InternalName>] ` (`DalamudReflector.GetPluginName()`, which returns
  `Svc.PluginInterface?.InternalName ?? "Not initialized"`; no `Module.DalamudReflector` needed).
  The prefix is included in the log line as well as in chat.
- Chat output is `Svc.Chat.Print(new XivChatEntry { Message = <UiForeground-coloured text>,
  Type = Svc.PluginInterface.GeneralChatType })`, issued from a `TickScheduler`, so it runs on the
  next framework update. Safe to call from any thread.
- The chat line is printed at every level, including `Verbose` and `Debug`, regardless of log
  level settings. It is visible to the user; do not use it in loops.
- No `Log*` aliases here; only the six names above.

## Exception logging extensions

`references/ecommons/ECommons/GenericHelpers/LogHelpers.cs`, in `public static unsafe partial class
GenericHelpers`, namespace `ECommons` (so `using ECommons;` brings them in). All the extensions
are on `Exception`; there are no string extensions such as `"text".Log()` in this file.

```csharp
public static string ToStringFull(this Exception e)
public static void Log(this Exception e, Action<string> exceptionFunc)
public static void Log(this Exception e)                              // PluginLog.Error
public static void Log(this Exception e, string ErrorMessage)
public static void LogFatal(this Exception e)                         // PluginLog.Fatal
public static void LogFatal(this Exception e, string ErrorMessage)
public static void LogWarning(this Exception e)                       // PluginLog.Warning
public static void LogWarning(this Exception e, string errorMessage)
public static void LogVerbose(this Exception e)                       // PluginLog.Verbose
public static void LogVerbose(this Exception e, string ErrorMessage)
public static void LogInternal(this Exception e)                      // InternalLog.Error only
public static void LogInternal(this Exception e, string ErrorMessage)
public static void LogDebug(this Exception e)                         // PluginLog.Debug
public static void LogDebug(this Exception e, string ErrorMessage)
public static void LogInfo(this Exception e)                          // PluginLog.Information
public static void LogInfo(this Exception e, string ErrorMessage)
public static void LogDuo(this Exception e)                           // DuoLog.Error (log + chat)
```

- `ToStringFull()` returns `"{Message}\n{StackTrace}"` followed, for each inner exception, by
  `"\nAn inner exception (n) was thrown: {Message}\n{StackTrace}"`. The exception type name is not
  included.
- The `(string ErrorMessage)` overloads emit two separate log entries at the same level: first the
  message, then the exception text.
- `Log(Action<string>)` lets any sink be used: `e.Log(InternalLog.Warning)`,
  `e.Log(myLogger.LogError)`.
- There is no `LogError` extension; `e.Log()` is the Error-level one.

```csharp
try { DoWork(); }
catch(Exception e) { e.Log(); }                       // Error
catch(Exception e) { e.LogWarning("DoWork failed"); } // two Warning lines
```

## SimpleLogger

`public class SimpleLogger : IDisposable`, namespace `ECommons`,
`references/ecommons/ECommons/SimpleLogger.cs`. Append-only text file logger with its own writer
thread.

```csharp
public SimpleLogger(string dir, string filename)
public void Log(string s)
public void LogError(string s)     // Log($"Error: {s}")
public void Dispose()
```

- `filename` has invalid file-name characters replaced by `_`; `dir` is used as given and is not
  created. If the file cannot be opened, the failure is logged at Debug level only and every
  later `Log` call queues into a buffer nobody drains.
- The constructor starts a foreground `Thread` that opens `Path.Combine(dir, filename)` in append
  mode (UTF-8), writes `Log begins at <time>`, then blocks on a `BlockingCollection<string>`.
- `Log` prefixes `DateTimeOffset.Now` in `s` format and a space. After `Dispose` it logs
  `Can not log, collection is marked as completed` at Debug and drops the line.
- The stream is not flushed per line; content reaches disk when the writer's buffer fills or on
  `Dispose`.
- `Dispose()` calls `CompleteAdding()`; the thread drains the queue, writes `Log ends at <time>`
  and closes the file. Because the thread is not a background thread, a logger that is never
  disposed keeps a thread blocked for the life of the process. `ECommonsMain.Dispose()` does not
  track these.

```csharp
var log = new SimpleLogger(Svc.PluginInterface.ConfigDirectory.FullName, "events.log");
log.Log("started");
// plugin Dispose:
log.Dispose();
```

## Commands

### [Cmd] and [SubCmd]

`references/ecommons/ECommons/Commands/CmdAttribute.cs`, `SubCmdAttribute.cs`, `CmdManager.cs`

```csharp
[AttributeUsage(AttributeTargets.Method)]
public class CmdAttribute : Attribute
{
    public string Command { get; }
    public string HelpMessage { get; }
    public bool ShowInHelp { get; }
    public bool ShowInHelpPanel { get; }
    public CmdAttribute(string command, string helpMessage = "", bool showInHelp = true,
        bool showInHelpPanel = true)
}

[AttributeUsage(AttributeTargets.Method, AllowMultiple = true)]
public class SubCmdAttribute : Attribute
{
    public string SubCommand { get; }
    public string HelpMessage { get; }
    public SubCmdAttribute(string subCommand, string helpMessage = "")
}
```

Discovery (`CmdManager.Init`, internal, run by every `ECommonsMain.Init`):

- Scans `ECommonsMain.Instance.GetType().GetRuntimeMethods()`: methods of the plugin class (the
  object passed as `instance` to `ECommonsMain.Init`) only, including private, static and
  inherited ones. Methods on other classes are never found.
- A method qualifies when it has exactly two parameters, both of type `string`, and carries
  `[Cmd]`. The return type is not checked. A `[Cmd]` method with any other parameter list is
  skipped silently.
- Registered with `Svc.Commands.AddHandler(cmdAttr.Command, new CommandInfo(...) { HelpMessage =
  cmdAttr.HelpMessage, ShowInHelp = cmdAttr.ShowInHelp })`; the handler calls
  `m.Invoke(plugin, [command, arguments])`.
- `Command` must include the leading slash (`"/myplugin"`), as for stock `AddHandler`.
- Only one `[Cmd]` per method (`AllowMultiple` is not set). For aliases, add a second method or
  use `EzCmd.Add`.
- Registration happens inside `ECommonsMain.Init`, that is, usually early in the plugin
  constructor. A command can be invoked before the rest of the constructor has run only if
  something processes commands in between; still, handlers should tolerate uninitialised state.
- Exceptions thrown by the handler surface as `TargetInvocationException` from `Invoke` and are
  handled by Dalamud's command dispatcher, not by ECommons.

`[SubCmd]` registers nothing and does no argument parsing. It is documentation metadata for
`CmdManager.DrawHelp`; the handler must parse `arguments` itself.

`HelpMessage` on both attributes is passed through `.Loc()`
(`ECommons.LanguageHelpers`) in the attribute constructor, so it is translated if a localization
is loaded at the time the attribute is instantiated, otherwise returned unchanged.

`ShowInHelp` maps to Dalamud's `CommandInfo.ShowInHelp` (listing in `/xlhelp` and the plugin
installer). `ShowInHelpPanel` only affects `CmdManager.DrawHelp`.

```csharp
public sealed class Plugin : IDalamudPlugin
{
    public Plugin(IDalamudPluginInterface pi)
    {
        ECommonsMain.Init(pi, this);      // registers /myplugin here
    }

    [Cmd("/myplugin", "Open the main window")]
    [SubCmd("config", "Open settings")]
    [SubCmd("toggle", "Enable or disable")]
    private void OnCommand(string command, string arguments)
    {
        switch(arguments.Trim())
        {
            case "config": /* ... */ break;
            case "toggle": /* ... */ break;
            default: /* ... */ break;
        }
    }

    public void Dispose() => ECommonsMain.Dispose();   // removes /myplugin
}
```

### CmdManager public members

```csharp
public static void DrawHelp(float indent = 0)
public static void DisplayCommandHelp(string command, string extraCommand = "",
    string helpMessage = "", float indent = 0)
```

- `DrawHelp` must be called inside an ImGui window. For every registered `[Cmd]` it draws the main
  command (if `ShowInHelpPanel`) and then one row per `[SubCmd]` (always, even when
  `ShowInHelpPanel` is `false`).
- Each row is a button labelled with the command text. Left click executes it through
  `Svc.Commands.ProcessCommand(command)`; right click while hovered copies it to the clipboard.
  The help text follows after ` → `.
- `indent`: `0` puts help on the same line, `> 0` same line with that indent, `< 0` puts it on the
  next line.
- The button executes the command string exactly as written in the attributes, so a sub-command
  such as `"set <value>"` would be run with the literal placeholder.
- `EzCmd.Add` commands are not included in `DrawHelp`.

### EzCmd

`public static class EzCmd`, namespace `ECommons`, `references/ecommons/ECommons/EzCmd.cs`.

```csharp
public static void Add(string command, IReadOnlyCommandInfo.HandlerDelegate action,
    string helpMessage = null, int displayOrder = -1)
```

- `HandlerDelegate` is Dalamud's `void (string command, string arguments)`; pass a method group.
- Creates `new CommandInfo(action) { HelpMessage = helpMessage ?? "", ShowInHelp = helpMessage !=
  null, DisplayOrder = displayOrder }` and calls `Svc.Commands.AddHandler`. Omitting
  `helpMessage` therefore hides the command from help, which is the usual way to add aliases.
- The command name is recorded in the internal `RegisteredCommands` list before `AddHandler` is
  called and regardless of its result; `ECommonsMain.Dispose()` removes every recorded name.
  There is no `EzCmd.Remove`; to remove one earlier call `Svc.Commands.RemoveHandler(command)`.
- Not obsolete. The `[Obsolete]` line pointing to `[Cmd]` is commented out in source. In the
  pinned consumers, AutoRetainer and YesAlready use `EzCmd.Add`; none of the three uses `[Cmd]`.
- Works with handlers on any class, unlike `[Cmd]`.

```csharp
// YesAlready
EzCmd.Add(Command, OnCommand, "Opens the plugin window.", int.MinValue);
Aliases.Each(a => EzCmd.Add(a, OnCommand, $"{Command} alias"));

// AutoRetainer: multi-line help on the main command, hidden alias
EzCmd.Add("/autoretainer", CommandHandler, """
    Open plugin interface
    /ays - alias for /autoretainer
    /autoretainer e|enable → Enable plugin
    """);
EzCmd.Add("/ays", CommandHandler);

private void CommandHandler(string command, string arguments) { /* ... */ }
```

## Chat helpers

### ChatPrinter

`public static class ChatPrinter`, `references/ecommons/ECommons/ChatMethods/ChatPrinter.cs`

```csharp
public static void Red(string text)
public static void Orange(string text)
public static void Yellow(string text)
public static void Green(string text)
public static void PrintColored(UIColor col, string text)
```

`PrintColored` calls `Svc.Chat.Print(new XivChatEntry { Message = new
SeStringBuilder().AddUiForeground(text, (ushort)col).Build() })` immediately on the calling thread.
Unlike `DuoLog` it adds no plugin-name prefix, writes nothing to the log, does not set `Type`
(the `XivChatEntry` default is used, not `GeneralChatType`) and is not deferred to the framework
thread. AutoRetainer adds its own tag: `ChatPrinter.Red("[AutoRetainer] " + t);`.

### UIColor

`public enum UIColor : short`, `references/ecommons/ECommons/ChatMethods/UIColor.cs`. Row IDs of
the game's `UIColor` sheet for use as UI foreground colour keys. Named members:

| Name | Value | Name | Value |
|---|---|---|---|
| `WhiteNormal` | 0 | `Green` | 45 |
| `White` | 1 | `DarkGreen` | 47 |
| `Grey1` | 2 | `WarmSeaBlue` | 52 |
| `Grey2` | 3 | `Orange` | 500 |
| `Grey3` | 4 | `LightBlue` | 502 |
| `Grey4` | 5 | `Yellow` | 514 |
| `Grey5` | 6 | `Gold` | 540 |
| `Black` | 7 | `DarkBlue` | 543 |
| `LightYellow` | 8 | `LightGreen` | 551 |
| `Red` | 17 | `Pink` | 561 |
| `DarkRed` | 19 | | |

All other members are placeholders named `_color<N>` (for example `_color504`); the enum is not
contiguous. Any `ushort` key can be used by casting: `(UIColor)37`. Cast to `ushort` when passing
to `SeStringBuilder.AddUiForeground`.

### Sender

`public struct Sender : IEquatable<Sender>`, `references/ecommons/ECommons/ChatMethods/Sender.cs`.
A player identity as name plus home-world row ID, usable as a dictionary key and serializable
(two public fields).

```csharp
public string Name;
public uint HomeWorld;

public Sender(string Name, uint HomeWorld)
public Sender(SeString Name, uint HomeWorld)
public Sender(SeString Name, RowRef<World> HomeWorld)
public Sender(string Name, RowRef<World> HomeWorld)
public Sender(IPlayerCharacter pc)

public static bool TryParse(string nameWithWorld, out Sender s)
public IPlayerCharacter? Find()
public bool TryFind([NotNullWhen(true)] out IPlayerCharacter pc)
public override string ToString()
// Equals(object), Equals(Sender), GetHashCode(), operator ==, operator !=
```

- `TryParse` expects exactly `Name@World`; the world name is resolved with
  `ExcelWorldHelper.Get(string)`. Returns `false` (and `default`) for any other shape or an
  unknown world.
- `ToString()` returns `Name@WorldName`, looking the world up in the `World` sheet; an unknown row
  yields `Name@` with an empty world.
- `Find()` iterates `Svc.Objects` and returns the first `IPlayerCharacter` with the same name and
  home world, or `null`. It reads the object table, so call it on the framework thread.
- Equality is ordinal on `Name` plus `HomeWorld`.
- `World` and `RowRef<>` are Lumina types (`Lumina.Excel.Sheets.World`, `Lumina.Excel.RowRef<T>`).

## EzDtr

`public class EzDtr : IDisposable`, namespace `ECommons.EzDTR` (capital `DTR` in the namespace and
folder, `EzDtr` for the class), `references/ecommons/ECommons/EzDTR/EzDtr.cs`. Wraps one server
info bar (DTR) entry and refreshes it every framework tick.

```csharp
public IDtrBarEntry? Entry;

public EzDtr(Func<SeString> text, Action? onClick = null, string? title = null,
    Func<bool> showCondition = null)
public EzDtr(Func<SeString> text, Action<DtrInteractionEvent>? onClick = null,
    string? title = null, Func<bool> showCondition = null)

public void Dispose()
public static void DisposeAll()
```

- `title` defaults to the plugin's `InternalName`; it is the key passed to `Svc.DtrBar.Get(title)`.
  Two `EzDtr` instances need different titles.
- The constructor subscribes to `Svc.Framework.Update`. Every tick: if `showCondition` is set and
  returns `false`, `Entry.Shown = false` and nothing else happens; otherwise `Entry.Shown = true`,
  `Entry.Text = text()`, and `Entry.OnClick` is assigned when a click handler was given.
- `text` and `showCondition` therefore run every frame on the framework thread. Keep them cheap
  and allocation-light; cache the `SeString` if building it is expensive.
- Because `Shown` is written every tick, setting `Entry.Shown` yourself has no lasting effect; use
  `showCondition`. Tooltip and other `IDtrBarEntry` properties can be set through `Entry`.
- `Dispose()` unsubscribes, removes the instance from the registry and calls `Entry?.Remove()`.
  `ECommonsMain.Dispose()` does this for all instances.
- Both constructors have identical optional parameters apart from the delegate type of `onClick`.
  Passing a lambda selects one by its shape (`() => ...` versus `e => ...`). Omitting `onClick`,
  or passing a bare `null`, matches both overloads and is expected to be rejected by the compiler
  as ambiguous *(unverified, not compiled)*; pass a typed value such as `(Action?)null`.

```csharp
// YesAlready
_ = new EzDtr(
    () => new SeString(new TextPayload($"{Name}: {(C.Enabled ? "On" : "Off")}")),
    () => { C.Enabled ^= true; C.Save(); });
```

## ContextMenuPrefixRemover

`public static unsafe class ContextMenuPrefixRemover`, namespace `ECommons.EzContextMenu`,
`references/ecommons/ECommons/EzContextMenu/ContextMenuPrefixRemover.cs`. Dalamud requires a prefix
icon on plugin context-menu items; this removes it from the rendered menu.

```csharp
public static void Initialize()                       // optional; RemovePrefix calls it
public static MenuItem RemovePrefix(this MenuItem m)  // Dalamud.Game.Gui.ContextMenu.MenuItem
```

How it works:

- First use (lazy) picks a random `SeIconChar` value that is not a defined enum member and a random
  colour key, makes the pair unique across plugins through
  `EzSharedData.GetOrCreate<HashSet<(int,
  ushort)>>("ECommons.ContextMenuPrefixRemover.TakenContextMenuPrefixes", [])`,
  logs the choice at Debug, and registers
  `Svc.AddonLifecycle.RegisterListener(AddonEvent.PreRequestedUpdate, "ContextMenu", ...)`.
- `RemovePrefix` sets `m.Prefix` and `m.PrefixColor` to that marker pair and returns the same
  `MenuItem`.
- On each `ContextMenu` update the listener reads the entry strings from the addon's `AtkValues`
  (count at index 0, entries from index 7), finds a text payload beginning with the marker
  character directly after a `UIForegroundPayload` with the marker colour, removes that payload,
  and writes the SeString back in place.
- Initialisation errors are caught and logged with `e.Log()`; `Initialized` stays `true`, so it is
  not retried.
- `ECommonsMain.Dispose()` unregisters the listener. The taken prefix pair is not removed from the
  shared set.
- It depends on the `ContextMenu` addon's AtkValue layout and on how Dalamud renders prefixes, so
  it can break on game or Dalamud updates. It applies to `MenuItem` entries in the default
  `ContextMenu` addon only.
- The source comment asks that the item still indicate to the user that it comes from a plugin.
  Removing the prefix works around a Dalamud requirement; check the repository rules that apply
  to the plugin before using it (see the `dalamud` skill, publishing rules).

```csharp
// AutoRetainer Internal/ContextMenuManager.cs, shape only
args.AddMenuItem(new MenuItem()
{
    Name = "...",
    OnClicked = a => { /* ... */ },
}.RemovePrefix());
```

## Gotchas

- `ECommonsMain.Init` raises `Svc.Log.MinimumLogLevel` to `Verbose` for every plugin that uses
  ECommons. `ReducedLogging` does not revert that.
- `PluginLog` takes a single string. `PluginLog.Information("x = {0}", x)` does not compile; there
  is no params overload and no exception overload.
- `PluginLog` passes the text as a Serilog template. Text containing braces can be mangled in
  `dalamud.log`. Use `Svc.Log.Information("{Text}", text)` when the exact text matters.
- `InternalLog.Information(...)` and friends do not write to the Dalamud log. `e.LogInternal()`
  likewise only reaches the in-plugin buffer.
- `DuoLog.Verbose` and `DuoLog.Debug` print to the user's chat unconditionally.
- There are no string logging extensions; `.Log()` and its variants exist on `Exception` only.
- `e.Log()` loses the exception type name; include it in the message when it matters
  (`e.Log($"{e.GetType().Name} in X")`).
- `[Cmd]` methods are found only on the plugin instance class and only with the exact
  `(string, string)` parameter list. A mistake produces no error, just no command.
- `[SubCmd]` is help text only. It neither registers nor dispatches anything.
- A command registered twice (both `[Cmd]` and `EzCmd.Add`, or by two plugins) is rejected by
  Dalamud's `AddHandler`; ECommons ignores the return value and will still call `RemoveHandler`
  for that name on dispose.
- `ECommons.ChatMethods.UIColor` collides by simple name with Lumina's `UIColor` sheet type when
  both namespaces are imported; qualify or alias one of them.
- `ChatPrinter` prints on the calling thread with no deferral; call it from the framework thread.
- `EzDtr` rewrites `Shown`, `Text` and `OnClick` every tick; external changes to those are
  overwritten.
- `SimpleLogger` must be disposed by the plugin. Its writer thread is a foreground thread.
