# Automation: Chat, Callback, UI input, macros, key presses, cutscene skipper

Covers: everything in `ECommons/Automation/` except the task managers: `Chat`, `Callback`,
`UIInput` (`ClickHelper`, `ClickHelperExtensions`, `EventData`, `InputData`, `EventType`),
`MacroManager`, the obsolete `Macro`/`UTF8String` structs, `WindowsKeypress` and
`AutoCutsceneSkipper`. Verified against ECommons `bc38dc7` (2026-11-08, v3.2.1.26, Dalamud
API 15); nothing here was compiled. Paths are relative to the skill root.

Task queues are in `references/topic-taskmanager.md`; typed addon wrappers in `references/topic-addonmaster.md`.
AtkValue/AtkUnitBase layout: see the `clientstructs` skill.

If a click throws `Unable to cast object of type 'ReceiveEvent' to type 'ReceiveEventDelegate'`,
the plugin is on ECommons 3.2.1.22 or older. Those versions of `ClickHelper.InvokeReceiveEvent`
turn the vtable entry into a managed delegate, which fails when Dalamud's addon hooks have
replaced the entry with a delegate of another type. 3.2.1.23 and later call the entry through an
unmanaged function pointer. Update ECommons; do not change addon argument order or substitute
guessed signatures to work around the cast error.

## Namespaces

```csharp
using ECommons.Automation;           // Chat, Callback, MacroManager, WindowsKeypress, AutoCutsceneSkipper
using ECommons.Automation.UIInput;   // ClickHelper, ClickHelperExtensions, EventData, InputData, EventType
using Callback = ECommons.Automation.Callback;   // alias, see below
```

`Callback` is a common type name. ECommons itself writes that alias line in `ECommonsMain.cs` and
in `UIHelpers/AddonMasterImplementations/SelectString.cs`, and AutoRetainer declares `global using
Callback = ECommons.Automation.Callback;` in its
`Global.cs`. Add the alias when the compiler reports an ambiguous `Callback`.

None of these classes needs a `Module` flag in `ECommonsMain.Init`. None of them checks the calling
thread. All of them call game code or touch UI memory, so call them from the framework thread
(a TaskManager task, `Svc.Framework.Update`, `Svc.Framework.RunOnFrameworkThread`); see the
`dalamud` skill for the thread rules.

## Chat

Source: `references/ecommons/ECommons/Automation/Chat.cs`. `public static unsafe class Chat`. At
this commit it is a plain static class; no signature scanning, it calls
`UIModule.Instance()->ProcessChatBoxEntry` from FFXIVClientStructs.

```csharp
public static void SendMessageUnsafe(byte[] message)
public static void SendMessage(string message)
public static string SanitiseText(string text)
public static void ExecuteCommand(string message)
public static void ExecuteGeneralAction(uint generalActionId)
public static void ExecuteAction(uint actionId)
```

| Method | Behaviour |
|---|---|
| `SendMessageUnsafe` | Null-terminates the bytes, builds a `Utf8String`, passes it to `ProcessChatBoxEntry`, destroys the string. No validation at all. Input is processed like text typed into the chat box: a leading `/` runs a command, anything else is sent to the current chat channel. |
| `SendMessage` | UTF-8 encodes, then throws `ArgumentException` when the message is empty, longer than 500 bytes, or when `SanitiseText(message).Length != message.Length` ("message contained invalid characters"). Then calls `SendMessageUnsafe`. |
| `SanitiseText` | `Utf8String.SanitizeString((AllowedEntities)0x27F)`; returns the text with disallowed characters removed. |
| `ExecuteCommand` | Throws `InvalidOperationException` unless the text starts with `/`, then `SendMessage`. Use this for commands so that a formatting mistake can never leak text into public chat. |
| `ExecuteGeneralAction` | `ExecuteCommand($"/generalaction \"{name}\"")` with the name from the `GeneralAction` sheet. |
| `ExecuteAction` | `ExecuteCommand($"/action \"{name}\"")` with the name from the `Action` sheet. |

`ExecuteGeneralAction`/`ExecuteAction` use `GetRowOrDefault(id)?.Name`; an unknown id yields an
empty quoted name and the game prints its own error. Names come from the sheet in the client
language, which is what the command parser expects.

Obsolete: the nested `public static class Chat.Instance` with the same six methods, marked
`[Obsolete("Use Chat.<MethodName> directly instead of Chat.Instance.<MethodName>")]`. Older code
written as `Chat.Instance.SendMessage(...)` still compiles (warning CS0618); new code calls
`Chat.SendMessage(...)`. The XML comments still mention an `InvalidOperationException` for a
missing signature; the current code has no signature and cannot throw it.

```csharp
Chat.ExecuteCommand("/automove off");
Chat.ExecuteCommand($"/gearset change {index + 1}");
var safe = Chat.SanitiseText(userText);
if(safe.Length > 0) Chat.SendMessage("/echo " + safe);   // still subject to the 500-byte limit
```

Commands registered by Dalamud plugins are not game commands. The chat box entry path used here
is the game's; whether a plugin command runs through it depends on Dalamud's own chat hook
*(unverified from ECommons source)*. For Dalamud commands prefer
`Svc.Commands.ProcessCommand(string content)` (`ICommandManager`, returns `bool`).

## Callback

Source: `references/ecommons/ECommons/Automation/Callback.cs`. `public static unsafe class Callback`.

```csharp
public static readonly AtkValue ZeroAtkValue = new() { Type = 0, Int = 0 };
public static void Fire(AtkUnitBase* Base, bool updateState, params object[] values)
public static void FireRaw(AtkUnitBase* Base, int valueCount, AtkValue* values, byte updateState = 0)
public static void InstallHook()
public static void UninstallHook()
public static List<string> DecodeValues(int cnt, AtkValue* values)
public static string DecodeValue(AtkValue a)
internal static void Dispose()
```

`Fire` has exactly one overload. It converts `values` to an unmanaged `AtkValue` array, calls
`Base->FireCallback((uint)values.Length, atkValues, updateState)` and frees the memory.

| CLR type of the argument | `AtkValueType` | Field written |
|---|---|---|
| `uint` | `UInt` | `UInt` |
| `int` | `Int` | `Int` |
| `float` | `Float` | `Float` |
| `bool` | `Bool` | `Byte` (1/0) |
| `string` | `String` | `String` (UTF-8 copy, null-terminated, freed after the call) |
| `AtkValue` | copied unchanged | whole struct |
| anything else | - | `ArgumentException("Unable to convert type ... to AtkValue")` |

- Matching is on the boxed runtime type. `1` is `Int`, `1u` is `UInt`, `1f` is `Float`. `long`,
  `short`, `ushort`, `byte`, `double`, enums and `null` are rejected (`null` fails with a
  `NullReferenceException` while formatting the error).
- `Base == null` throws `Exception("Null UnitBase")`.
- `updateState` is forwarded as the last argument of `AtkUnitBase.FireCallback`. Consumers and
  AddonMaster pass `true` for ordinary interactions (`Callback.Fire(addon, true, -1)` to close,
  `Callback.Fire((AtkUnitBase*)Addon, true, Index)` for a `SelectString` entry).
- Use `Callback.ZeroAtkValue` or a hand-built `AtkValue` for positions that must carry a type
  the mapping cannot produce.
- Log output: one `PluginLog.Verbose` line `Firing callback: {name}, valueCount = ..., updateStatte
  = ..., values:`. The decoded value list is built but not appended to the message.
- `FireRaw` only forwards to `Base->FireCallback((uint)valueCount, values, updateState != 0)`. The
  source comment calls it kept for compatibility; call `FireCallback` on the addon directly.

Hook facility: `InstallHook()` lazily creates a Dalamud hook on
`AtkUnitBase.MemberFunctionPointers.FireCallback` and enables it. The detour calls the original
and then writes every callback in the game (addon name, count, `updateState`, decoded values) to
`PluginLog.Debug`. It is a logging aid for discovering callback values; it cannot block or modify
callbacks. `InstallHook()` when already enabled and `UninstallHook()` when already disabled log an
error. `UninstallHook()` before any `InstallHook()` dereferences a null hook. `Callback.Dispose()`
is internal and called by `ECommonsMain.Dispose()`; it disposes the hook.

`DecodeValue` renders `Int`, `UInt`, `Bool` and the string types (`String`, `String8`,
`WideString`, `ManagedString`); every other type, including `Float`, prints as
`Unknown Type: {Int}`.

```csharp
if(TryGetAddonByName<AtkUnitBase>("FreeCompanyCreditShop", out var addon) && IsAddonReady(addon))
{
    if(EzThrottler.Throttle("CloseShop")) Callback.Fire(addon, true, -1);
}
```

## UIInput

Source: `references/ecommons/ECommons/Automation/UIInput/*.cs`. These helpers send synthetic
`ReceiveEvent` calls to addons and components. Most plugin code should use an AddonMaster wrapper
(which calls these internally through `ClickButtonIfEnabled` and friends) and fall back to the
extensions only for addons without a wrapper.

### ClickHelperExtensions

`public static unsafe class ClickHelperExtensions`. The `this` parameter is the struct by value;
call through a pointer: `button->ClickAddonButton(addon)`.

```csharp
public static void ClickAddonButton(this AtkComponentButton target, AtkUnitBase* addon)
public static void ClickAddonButton(this AtkCollisionNode target, AtkUnitBase* addon)
public static void ClickAddonButton(this AtkComponentButton target, AtkUnitBase* addon, AtkEvent* eventData)
public static void ClickAddonButton(this AtkCollisionNode target, AtkUnitBase* addon, AtkEvent* eventData)
public static void ClickAddonButton(this AtkComponentButton target, AtkComponentBase* addon, uint which, EventType type = EventType.CHANGE, EventData? eventData = null)
public static void ClickRadioButton(this AtkComponentRadioButton target, AtkUnitBase* addon)
public static void ClickRadioButton(this AtkComponentRadioButton target, AtkComponentBase* addon, uint which, EventType type = EventType.CHANGE)
public static void ClickCheckBox(this AtkComponentCheckBox target, AtkUnitBase* addon)
```

| Overload | What it does |
|---|---|
| `ClickAddonButton(AtkUnitBase*)` on a button | Takes the first event registered on the button's owner node (`AtkResNode.AtkEventManager.Event`) and calls `addon->ReceiveEvent(evt->State.EventType, (int)evt->Param, evt)`. |
| `ClickAddonButton(AtkUnitBase*)` on a collision node | Walks the node's event list to the first `AtkEventType.MouseClick` event and sends that. No null check while walking; a node without a click event walks off the list. |
| `ClickAddonButton(AtkUnitBase*, AtkEvent*)` | Calls the sig-scanned `ClickHelper.Listener` with the supplied event. |
| `ClickAddonButton(AtkComponentBase*, uint which, ...)` / `ClickRadioButton(AtkComponentBase*, uint which, ...)` | For components nested in another component: `ClickHelper.ClickAddonComponent(addon, ownerNode, which, type, ...)`. |
| `ClickRadioButton(AtkUnitBase*)` | Same as the button form; also writes a `Svc.Log.Debug` line with the event type and param. |
| `ClickCheckBox(AtkUnitBase*)` | Sends the first registered event with a zero-initialised `AtkEventData`. It does not change the visual checked state; `AddonMasterBase.ClickCheckboxIfEnabled` follows it with `checkbox->SetChecked(true)`. |

There is no method named `ClickAddonCheckBox`; the checkbox method is `ClickCheckBox`. None of the
extensions checks whether the control is enabled or visible. The pattern used by AddonMaster:

```csharp
if(button->IsEnabled && button->AtkResNode->IsVisible()) button->ClickAddonButton(addon);
```

### ClickHelper

`public unsafe class ClickHelper` (static members only).

```csharp
public delegate nint InvokeListener(nint a1, AtkEventType a2, uint a3, AtkEvent* a4);
public static InvokeListener Listener { get; }
public unsafe delegate IntPtr ReceiveEventDelegate(AtkEventListener* eventListener, EventType evt, uint which, void* eventData, void* inputData);
public static ReceiveEventDelegate GetReceiveEvent(AtkEventListener* listener)
public static void InvokeReceiveEvent(AtkEventListener* eventListener, EventType type, uint which, EventData eventData, InputData inputData)
public static void ClickAddonComponent(AtkComponentBase* unitbase, AtkComponentNode* target, uint which, EventType type, EventData? eventData = null, InputData? inputData = null)
public static void SendClick(string clickName, nint addon = default)
public static string[] GetAvailableClicks()
```

- `Listener` resolves lazily on first access through
  `EzDelegate.Get<InvokeListener>("48 89 5C 24 ?? 48 89 74 24 ?? 57 48 83 EC 30 0F B7 FA")`; a
  failed scan surfaces at that point, not at plugin load.
- `GetReceiveEvent` reads `listener->VirtualTable->ReceiveEvent` and returns a delegate for it.
  `InvokeReceiveEvent` no longer uses it: it casts the same vtable entry to
  `delegate* unmanaged<AtkEventListener*, EventType, uint, void*, void*, nint>` and calls that.
  `GetReceiveEvent` itself can still throw the cast exception described at the top of this file
  on a hooked addon.
- `ClickAddonComponent` creates `EventData.ForNormalTarget(target, unitbase)` and
  `InputData.Empty()` when the arguments are null, invokes the component's `ReceiveEvent`, and
  disposes only the objects it created.
- `GetAvailableClicks()` reflects over the ECommons assembly and returns
  `"{AddonMasterClass}_{Method}"` for every public method of every class nested directly in
  `AddonMaster` (methods declared on `object` and property accessors excluded).
- `SendClick(clickName, addon)` is meant for user-entered strings. It splits at the first `_`,
  finds a type of that name in the ECommons assembly, constructs it with the `(nint)` constructor
  and invokes the method without arguments. Failures are only logged with `Svc.Log.Debug`.
  Consequences: methods that take parameters cannot be invoked (the reflection call throws), and
  AddonMaster classes whose name begins with `_` (for example `_TitleMenu`) cannot be addressed
  because the class name is cut at the first underscore.

### EventData, InputData, EventType

```csharp
public sealed unsafe class EventData : IDisposable
    public AtkEvent* Data { get; }
    public static EventData ForNormalTarget(void* target, void* listener)

public sealed unsafe class InputData : IDisposable
    public void** Data { get; }
    public static InputData Empty()
    public static InputData ForPopupMenu(PopupMenu* popupMenu, ushort index)

public enum EventType : ushort
```

- `EventData` allocates `sizeof(AtkEvent)` with `Marshal.AllocHGlobal` (not zeroed) and sets only
  `Target` and `Listener`. `InputData` allocates `0x40` bytes and nulls eight pointer slots;
  `ForPopupMenu` fills slot 0 with the list item renderer at `index` and slot 2 with
  `index | ((ulong)index << 48)`. Constructors are private. Both free their memory in `Dispose()`
  and in a finalizer.
- `EventType` is ECommons' own enum, not `FFXIVClientStructs...AtkEventType`. Selected values:
  `MOUSE_CLICK = 9`, `BUTTON_PRESS = 22`, `BUTTON_CLICK = 23`, `CHANGE = 25`,
  `LIST_ITEM_CLICK = 33`, `LIST_INDEX_CHANGE = 35`, `ICON_TEXT_CLICK = 56`; the full list (1..76)
  is in `references/ecommons/ECommons/Automation/UIInput/EventType.cs`. Spelling in source:
  `UNIT_SCALE_CHANEGED`, `UNIT_RESOLUTION_SCALE_CHANEGED`.
- `UIInput/AtkEvent.cs` is entirely commented out. There is no
  `ECommons.Automation.UIInput.AtkEvent` type; `AtkEvent` in these signatures is the
  FFXIVClientStructs struct.

## MacroManager

Source: `references/ecommons/ECommons/Automation/MacroManager.cs`.
`public static unsafe class MacroManager`. Builds a temporary `RaptureMacroModule.Macro` and runs it
with `RaptureShellModule.Instance()->ExecuteMacro`.

```csharp
public static void Execute(string multilineString)
public static void Execute(params string[] commands)
public static void Execute(IEnumerable<string> commands)
public static RaptureMacroModule.Macro? FromLines(IEnumerable<string> commands)
public static RaptureMacroModule.Macro? FromLines(IEnumerable<SeString> commands)
```

- `Execute(string)` strips `\r` and splits on `\n`.
- `FromLines` throws `InvalidOperationException` for more than 15 lines, for a line whose encoded
  SeString is longer than 180 bytes, or for a line containing `\n`, `\r`, `\0` or characters that
  `Chat.SanitiseText` would remove. Empty lines become empty macro lines. Other failures while
  filling the macro are logged with `Svc.Log.Error` and return `null`.
- `Execute(IEnumerable<string>)` wraps everything in `GenericHelpers.Safe`, so the validation
  exceptions above are caught and logged; the call returns normally and nothing runs. Call
  `FromLines` yourself to detect invalid input.
- The macro runs in the game's own macro executor, so macro-only syntax such as `<wait.N>` is
  available *(unverified: game behaviour, not visible in ECommons source)*.
- A `string` overload ambiguity: `Execute("a")` binds to `Execute(string multilineString)`;
  `Execute("a", "b")` binds to the `params` form.

```csharp
MacroManager.Execute("/ac \"Basic Synthesis\" <wait.3>", "/ac \"Basic Touch\" <wait.3>");
```

## Macro and UTF8String (obsolete)

Sources: `references/ecommons/ECommons/Automation/Macro.cs`, `.../UTF8String.cs`.

- `public readonly struct Macro : IDisposable`,
  `[Obsolete("Use RaptureMacroModule.Macro instead.")]`. Constants `numLines = 15`, `size = 0x8 +
  (UTF8String.size * (numLines + 1))`; constructor
  `Macro(IntPtr loc, string t, IReadOnlyList<string> commands)`.
- `public readonly struct UTF8String : IDisposable`,
  `[Obsolete("Use FFXIVClientStructs.FFXIV.Client.System.String.UTF8String instead.")]`.
  Constant `size = 0x68`; constructors `UTF8String(IntPtr loc, string text)` and
  `UTF8String(IntPtr loc, byte[] text)`.

Hand-rolled layouts from before FFXIVClientStructs covered these types. Do not use them in new
code; use `MacroManager` or the ClientStructs types. Note the replacement named in the attribute is
spelled `UTF8String`, while the ClientStructs type used by `Chat.cs` is `Utf8String`.

## WindowsKeypress

Sources: `references/ecommons/ECommons/Automation/WindowsKeypress.cs`,
`.../WindowsKeypress.Forms.cs`. `public static unsafe partial class WindowsKeypress`. Sends window
messages to the game window found by `ECommons.Interop.WindowFunctions.TryFindGameWindow` using
`SendMessage` (synchronous) from TerraFX.

```csharp
public static bool SendKeypress(int key)
public static bool SendKeypress(LimitedKeys key)
public static bool SendKeypress(Keys key)                                   // ECommons.WindowsFormsReflector.Keys
public static bool SendKeypress(int key, IEnumerable<int>? modifiers)
public static bool SendKeypress(VirtualKey key, IEnumerable<VirtualKey>? modifiers)
public static bool SendKeypress(LimitedKeys key, IEnumerable<LimitedKeys>? modifiers)
public static bool SendKeyHold(int key, IEnumerable<int>? modifiers)
public static bool SendKeyHold(VirtualKey key, IEnumerable<VirtualKey>? modifiers)
public static bool SendKeyHold(LimitedKeys key, IEnumerable<LimitedKeys>? modifiers)
public static bool SendKeyRelease(int key, IEnumerable<int>? modifiers)
public static bool SendKeyRelease(VirtualKey key, IEnumerable<VirtualKey>? modifiers)
public static bool SendKeyRelease(LimitedKeys key, IEnumerable<LimitedKeys>? modifiers)
public static void SendMousepress(int key)
public static void SendMousepress(Keys key)
public static bool SendMousepress(LimitedKeys key)
```

- `VirtualKey` is `Dalamud.Game.ClientState.Keys.VirtualKey`; `LimitedKeys` is
  `ECommons.Interop.LimitedKeys`; `Keys` is ECommons' own copy of the WinForms enum in
  `ECommons.WindowsFormsReflector` (no Windows Forms reference is needed). The two enum source
  files are not in `references/`; they exist only in an ECommons checkout.
- `SendKeypress(key)` sends `WM_KEYDOWN` then `WM_KEYUP` with `lParam` 0. The modifier overloads
  send key-down for each modifier, the key down/up, then key-up for each modifier.
  `SendKeyHold` sends only the downs, `SendKeyRelease` only the ups (modifiers first).
- There is no single-argument `SendKeypress(VirtualKey)`; pass `null` for modifiers or cast to
  `int`.
- Return value: `true` when the window was found and the messages were sent, `false` plus
  `PluginLog.Error("Couldn't find game window!")` otherwise. It says nothing about whether the
  game acted on the key.
- `SendMousepress(int)` handles only XBUTTON1 (`1 | 4` = 5) and XBUTTON2 (`2 | 4` = 6) through
  `WM_XBUTTONDOWN`/`WM_XBUTTONUP`; any other value logs `Invalid key`. `SendMousepress(LimitedKeys)`
  does not send a mouse message at all: its body is `SendKeypress((int)key)`.
- The key is delivered as a window message, so it acts on whatever has keyboard focus inside the
  game, including an open chat input.

```csharp
WindowsKeypress.SendKeypress(ECommons.Interop.LimitedKeys.Escape);
WindowsKeypress.SendKeypress(VirtualKey.F, [VirtualKey.CONTROL]);
```

## AutoCutsceneSkipper

Source: `references/ecommons/ECommons/Automation/AutoCutsceneSkipper.cs`.
`public unsafe class AutoCutsceneSkipper` with static members only.

```csharp
public static Func<nint, bool> Condition;
public static void Init(Func<nint, bool> cutsceneSkipCondition)
public static void Disable()
public static void Enable()
```

- `Init` creates an `EzHook` on the cutscene input handler (signature in source) and an `EzPatch`
  that turns a conditional jump into an unconditional one, then enables the hook. All exceptions
  inside `Init`, including the "already initialized" one raised on a second call and any failed
  signature scan, are caught and logged; `Init` never throws to the caller.
- The detour acts only while `Svc.Condition[ConditionFlag.OccupiedInCutSceneEvent]` is set. It
  evaluates `Condition` with the handler's first argument (`nint`); a `null` delegate or a `true`
  result means skip. When the cutscene is skippable (`*(nint*)(a1 + 56) != 0`) the patch is
  enabled around the original call and disabled again.
- It triggers the skip request only. The source summary states it does not include the skip
  confirmation; the resulting `SelectString`/confirmation dialog must be answered separately
  (AddonMaster plus a task or an addon listener).
- `Condition` can be reassigned at any time. `Disable()`/`Enable()` toggle the hook and throw
  `NullReferenceException` if `Init` was never called or failed before creating the hook.
- `AutoCutsceneSkipper.Dispose()` is internal, called by `ECommonsMain.Dispose()`, and its body is
  empty; the hook is released through the EzHook registry (`EzHookCommon.DisposeAll`, also called
  by `ECommonsMain.Dispose()`). No plugin-side cleanup is required.

```csharp
AutoCutsceneSkipper.Init(_ => C.SkipCutscenes && Svc.ClientState.TerritoryType == wantedTerritory);
```

## What ECommonsMain.Dispose() does for this area

| Item | Cleanup |
|---|---|
| `Callback` logging hook | `Callback.Dispose()` disposes it. |
| `AutoCutsceneSkipper` hook | Disposed with all `EzHook`s. |
| Task managers | Both `TaskManager.DisposeAll` variants (see `references/topic-taskmanager.md`). |
| `Chat`, `MacroManager`, `WindowsKeypress`, `ClickHelper` | No state to clean up. |

## Gotchas

- `Chat.SendMessage` with text that does not start with `/` posts to the active chat channel on the
  live server. Use `ExecuteCommand` for commands, and sanitise any user-provided fragment.
- `Callback.Fire` throws on unsupported argument types after allocating. In that failure path the
  `finally` block inspects `AtkValue` slots that were never initialised (the buffer is not
  zeroed) and may call `FreeHGlobal` on a garbage pointer. Pass only `int`, `uint`, `float`,
  `bool`, `string` or `AtkValue`; cast enums and other integer types explicitly.
- Integer literal type matters in `Callback.Fire`: many game callbacks distinguish `Int` from
  `UInt`. Use `Callback.InstallHook()` in a debug build to record what the game itself sends.
- `Callback.UninstallHook()` without a prior `InstallHook()` throws `NullReferenceException`.
- The `ClickHelperExtensions` methods do not test `IsEnabled` or visibility and do not verify that
  the addon is ready; check `GenericHelpers.IsAddonReady(addon)` first.
- `MacroManager.Execute` swallows validation errors (logged only).
- `WindowsKeypress.SendMousepress(LimitedKeys)` sends a key press, not a mouse press.
- Everything here fires every time it is called. Inside per-frame code wrap calls in
  `EzThrottler.Throttle(name, ms)` (`ECommons.Throttlers`).
