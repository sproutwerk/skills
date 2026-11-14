# ImGui helpers outside ImGuiEx

Covers: everything in `ECommons/ImGuiMethods/` except `ImGuiEx/` and `ImGuiDragDrop.cs`:
notifications, colour types, BBCode text, per-call-site state (`Ref<T>`), async image loading,
ready-made overlay,
changelog and selector windows, plus `ECommons.Funding.PatreonBanner`. Verified against ECommons
`bc38dc7` (2026-11-08, v3.2.1.26, Dalamud
API 15); nothing here was compiled. Paths are relative to the skill root.

`ImGuiEx` and `ImGuiDragDrop` are in `references/topic-imguiex-core.md` and `references/topic-imguiex-widgets.md`.
`EzConfigGui`, `ConfigWindow`, `PopupWindow` are in `references/topic-simplegui.md`. Stock `Window`,
`WindowSystem`, `ImRaii`, textures and fonts are in the `dalamudimgui` skill.

```csharp
using ECommons.ImGuiMethods;                     // everything below unless stated
using ECommons.ImGuiMethods.TerritorySelection;  // TerritorySelector
using ECommons.ImGuiMethods.ItemSelection;       // ItemSelectorWindow
using ECommons.Funding;                          // PatreonBanner
```

All types need `ECommonsMain.Init` (they use `Svc`). No `Module` flag is required by anything here.
Source files are `references/ecommons/ECommons/ImGuiMethods/<Name>.cs` unless a path is given.

## Notify

`public static class Notify`: Dalamud toast notifications (bottom-right), not chat and not a window.

```csharp
public static void Success(string s)
public static void Info(string s)
public static void Error(string s)
public static void Warning(string s)
public static void Plain(string s)      // NotificationType.None
```

Each call creates `new TickScheduler(...)` whose body runs
`Svc.NotificationManager.AddNotification(new Notification() { Content = s, Title =
DalamudReflector.GetPluginName(), Type = ... })`. The notification is therefore added on the next
framework update, not synchronously; the title is the plugin's `InternalName`. There is no overload
for duration, minimisation or a custom title: build a `Notification` yourself for those (`dalamud`
skill). Nothing is returned, so the notification cannot be dismissed or updated later.

## Colours

### EColor

`public static class EColor`: mutable `public static Vector4` fields, all alpha 1, built with
`ImGuiEx.Vector4FromRGB(0xRRGGBB)`.

| Name | RGB | Name | RGB |
|---|---|---|---|
| `RedBright` | FF0000 | `OrangeBright` | FF7F00 |
| `Red` | AA0000 | `Orange` | AA5400 |
| `RedDark` | 440000 | `CyanBright` | 00FFFF |
| `GreenBright` | 00FF00 | `Cya` (sic) | 00AAAA |
| `Green` | 00AA00 | `VioletBright` | FF00FF |
| `GreenDark` | 004400 | `Violet` | AA00AA |
| `BlueBright` | 0000FF | `VioletDark` | 440044 |
| `Blue` | 0000AA | `BlueSky` | 0085FF |
| `White` | FFFFFF | `BlueSea` | 0058AA |
| `Black` | 000000 | `PurpleBright` | FF0084 |
| `YellowBright` | FFFF00 | `Purple` | AA0058 |
| `Yellow` | AAAA00 | `PinkLight` | FFABD6 |
| `YellowDark` | 444400 | | |

The fields are not `readonly`; the doc comment invites plugins to reassign them after
`ECommonsMain.Init` to restyle. The medium cyan is spelled `Cya`; `EColor.Cyan` does not exist.
`EColor` is not obsolete (the attribute is commented out in source).

### EzColor

`public struct EzColor`: RGBA floats with conversions.

| Member | Notes |
|---|---|
| `float R`, `G`, `B`, `A` `{ get; set; }` | 0..1 |
| `Vector4 Vector4` | same as the implicit conversion |
| `uint ARGB`, `uint RGBA` | packed from `(uint)(channel * 255)` |
| `uint U32` | `[Obsolete("Use ARGB")]`. Returns `ImGui.ColorConvertFloat4ToU32`, which is ImGui's ABGR packing, so the named replacement yields a different value. |
| `EzColor()`, `EzColor(Vector4 vec)`, `EzColor(float r, float g, float b, float a = 1)` | |
| `EzColor(uint col)` | `0xRRGGBB` or `0xRRGGBBAA`, see below |
| `static EzColor From(float r, float g, float b, float a = 1)` | |
| `static EzColor From(Vector3 col, float alpha = 1)`, `From(Vector4 col)`, `From(uint col)` | `From(uint)` = `new(col)` |
| `static EzColor From(ImGuiCol col)` | passes `ImGui.GetColorU32(col)` to `From(uint)`; see Gotchas |
| `static EzColor FromARGB(uint argb)`, `FromABGR(uint abgr)` | rearranged to RGBA then `From(uint)` |
| `static EzColor FromUiForeground(uint id)` | `UIColor` sheet row `id`, column `Dark` |
| `static EzColor FromUiGlow(uint id)` | `UIColor` sheet row `id`, column `Light` |
| `static EzColor FromStain(uint id)` | `Stain` sheet row `id`, `Color`, alpha forced to 1 |
| `implicit operator Vector4(EzColor)` | |
| `implicit operator uint(EzColor)` | `ImGui.ColorConvertFloat4ToU32` (ABGR, ready for draw lists and `PushStyleColor(uint)`) |

Static presets are `{ get; set; }` properties with the same names and values as `EColor`, except
the cyan is spelled `Cyan` and there is an extra `Transparent`.

The `uint` constructor decides between the two formats with
`(col & 0xFFFFFF) == col ? (col << 8) | 0xFF : col`: any value that fits in 24 bits is treated as
`0xRRGGBB` with alpha 255. Values whose top byte is zero are therefore never read as
`0xRRGGBBAA`. In particular `EzColor.Transparent` is declared as `new(0x00000000)` and evaluates to
opaque black (R=G=B=0, A=1). For a transparent colour write `new EzColor(0, 0, 0, 0)`. The same
ambiguity affects `FromARGB`/`FromABGR`/`FromUiForeground` when the rearranged value has a zero red
byte.

`ImGuiEx.Text(EzColor col, string s)` exists (`ImGuiEx/Text.cs`).

### GradientColor

```csharp
public static Vector4 Get(Vector4 start, Vector4 end, int Milliseconds = 1000)
public static bool IsColorInRange(Vector4 testedColor, Vector4 start, Vector4 end, float tolerance = 0.1f)
```

`Get` is a linear ping-pong driven by `Environment.TickCount64`: `start` to `end` over
`Milliseconds`, then back, full period `2 * Milliseconds`. Stateless; call it every frame.
`IsColorInRange` tests each component against `[min(a, b) - tolerance, max(a, b) + tolerance]`.

```csharp
ImGuiEx.Text(GradientColor.Get(EColor.RedBright, EColor.YellowBright, 500), "Attention");
```

## ImGuiBB

```csharp
public static void Text(string str, bool ignoreCache = false)
```

Renders a small BBCode dialect. The string is split with a regex on every call; `ignoreCache` and
the private `Cache` dictionary are unused. The whole body runs inside `GenericHelpers.Safe`.

| Tag | Effect |
|---|---|
| `[color=#RGB]`, `[color=#RRGGBB]` (`#` optional, case-insensitive) | `ImGui.PushStyleColor(ImGuiCol.Text, ...)`, alpha 1. Any other length parses to black. Colour names are not supported. |
| `[/color]` | `ImGui.PopStyleColor()` |
| `[size=N]` | `ImGui.SetWindowFontScale(N)` when `0.1 <= N <= 50` (`float.TryParse`, current culture) |
| `[/size]` | `ImGui.SetWindowFontScale(1f)` |
| `[img]URL-or-path[/img]` | `ThreadLoadImageHandler.TryGetTextureWrap`; drawn at native pixel size once loaded, nothing while loading |
| `\n` | new line |

Text runs are drawn with `ImGui.Text` and joined with `ImGui.SameLine(0, 0)`; there is no word
wrap. Tags are not balanced for you: a missing `[/color]` leaves the style colour pushed, a missing
`[/size]` leaves the window font scale changed.

## ImGuiTrans (obsolete, error)

`[Obsolete("Deprecated", true)] public class ImGuiTrans`: any use is a compile error. The source
names no replacement. Members, for recognising old code: `PushStyleColor(ImGuiCol, Vector4)`,
`PushTransparency(float)`, `PopTransparency()`, `WithTextColor(Vector4, Action)`,
`Text(Vector4, string)`, `TextWrapped(Vector4, string)`.

## ImGuiNoRef

```csharp
public static bool Checkbox(string label, bool value, out bool newResult)
```

Checkbox for a value that cannot be passed by `ref` (a property, a dictionary entry). Returns true
when clicked, with the new state in `newResult`. When it returns false `newResult` is `default`
(false), not the current value, so only assign inside the `if`:

```csharp
if(ImGuiNoRef.Checkbox("Enabled", C.Enabled, out var v)) C.Enabled = v;
```

## CImGui

`public partial class CImGui`: five raw `[LibraryImport("cimgui")]` cdecl imports. Nothing else in
ECommons calls them.

```csharp
public static partial void igBringWindowToDisplayFront(nint ptr);
public static partial void igBringWindowToDisplayBack(nint ptr);
public static partial nint igGetCurrentWindow();
public static partial int igFindWindowDisplayIndex(nint ptr);
public static partial void igClosePopupToLevel(int remaining, bool restore_focus_to_window_under_popup);
```

`nint ptr` is an `ImGuiWindow*`. Dalamud's bindings also name their native library `cimgui`;
whether these imports resolve at run time from a plugin's load context is *(unverified)*. Prefer
`ImGuiP` from `Dalamud.Bindings.ImGui` (`dalamudimgui` skill) for internals.

## Ref&lt;T&gt; and Box&lt;T&gt;

`public static class Ref<T>`: keyed static storage that hands out a `ref T`, so immediate-mode
widgets can have state without declaring a field.

```csharp
public static ref T? Get()                                   // key = GenericHelpers.GetCallStackID()
public static ref T? Get(T? defaultValue)                    // key = call stack
public static ref T? Get(Func<T?>? defaultValueGenerator)    // key = call stack
public static ref T? Get(string key)
public static ref T? Get(string key, T? defaultValue)
public static ref T? Get(string s, Func<T?>? defaultValueGenerator)
```

- One dictionary per closed `T`; the same key under `Ref<int>` and `Ref<bool>` are separate slots.
- The default is used only when the key is first seen. The generator overload invokes the
  delegate only then.
- For `T == string` with a null default (or null generator) the slot is initialised to `""`, so
  it can go straight into `ImGui.InputText`.
- The keyless overloads build the key from a `StackTrace` of the caller (3 frames by default);
  `GetCallStackID` is documented in source as expensive. Two calls on different lines get different
  slots; calls from the same line inside a loop share one slot. Prefer explicit keys in per-frame
  code, for example `$"{ImGui.GetID(label)}_filter"` as `ImGuiEx` does.
- With `T == string`, `Ref<string>.Get("x")` binds to `Get(string key)`, not `Get(T? defaultValue)`.
  Use `Get(key, defaultValue)` to supply a string default.
- The static constructor registers the type with `Purgatory`; `ECommonsMain.Dispose()`
  (`Purgatory.Purge`) nulls its static reference fields. Do not touch `Ref<T>` after dispose.
- Entries are never evicted while the plugin is loaded; do not generate unbounded key sets.
- Not thread-safe (plain `Dictionary`). Use from the draw thread.

```csharp
ref var filter = ref Ref<string>.Get("myplugin.filter");
ImGui.InputTextWithHint("##f", "Filter...", ref filter, 100);

ref var page = ref Ref<int>.Get("myplugin.page", 1);
ImGui.InputInt("Page", ref page);
```

`public class Box<T>`: reference wrapper with `public T? Value`, `Box(T? value)` and
`Box<T> Clone()` (shallow: new box, same `Value`). `Ref<T>` stores its slots as `Box<T>`.

## FontAwesome

`public static class FontAwesome`

| Member | Value |
|---|---|
| `static readonly string Save` | `` |
| `Cross` | `` |
| `Check` | `` |
| `Info` | `` |
| `Trash` | `` |
| `Plus` | `` |
| `Layers` | `` |

```csharp
public static void Print(Vector4? col, string icon)
public static void PrintV(Vector4? col, string icon)
public static void Print(Vector4? col, FontAwesomeIcon icon)
public static void PrintV(Vector4? col, FontAwesomeIcon icon)
```

Each pushes `UiBuilder.IconFont`, draws with `ImGuiEx.Text` (`Print`) or `ImGuiEx.TextV` (`PrintV`,
the frame-padding-aligned variant for use beside buttons and inputs), and pops. `col` null means
the current text colour. The string constants are only meaningful while the icon font is pushed.
The typed overloads take `Dalamud.Interface.FontAwesomeIcon`.

## EnumMemberNameAttribute

```csharp
[AttributeUsage(AttributeTargets.All)]
public class EnumMemberNameAttribute(string name) : Attribute { public string Name = name; }
```

Defined but not read anywhere in the library at this commit. `ImGuiEx.EnumCombo` takes display
names through its `IDictionary<T, string> names` parameter, not through this attribute. Applying it
has no effect unless your own code reads it.

## ThreadLoadImageHandler

`public class ThreadLoadImageHandler` (all members static): lazy image cache for URLs, files, game
texture paths and game icons, loaded on a worker thread. `ImGuiBB`, `ItemSelectorWindow` and parts
of `ImGuiEx` use it.

```csharp
public static bool TryGetTextureWrap(string url, out IDalamudTextureWrap textureWrap)
public static bool TryGetIconTextureWrap(uint icon, bool hq, out IDalamudTextureWrap textureWrap)
public static bool TryGetIconTextureWrap(int icon, bool hq, out IDalamudTextureWrap textureWrap)
public static void ClearAll()
public static Action<Exception?, string?>? ErrorAction = null;
public static void AddConversionToBitmap(Func<byte[], byte[]> conversion)
public static void RemoveConversionToBitmap(Func<byte[], byte[]> conversion)
```

Usage, every frame:

```csharp
if(ThreadLoadImageHandler.TryGetIconTextureWrap(iconId, false, out var tex))
    ImGui.Image(tex.Handle, new(24, 24));

if(ThreadLoadImageHandler.TryGetTextureWrap("https://example.org/banner.png", out var banner))
    ImGui.Image(banner.Handle, new(banner.Width, banner.Height));
```

Behaviour:

- First request for a key inserts an empty entry, starts the worker thread if it is not running,
  and returns false with `textureWrap == null`. Later calls return true once the texture exists.
- Key interpretation for `TryGetTextureWrap`: starts with `http:` or `https:` (case-insensitive):
  downloaded with a shared `HttpClient` (10 s timeout), then `Svc.Texture.CreateFromImageAsync`;
  else if `File.Exists(key)`: `Svc.Texture.GetFromFile`; otherwise `Svc.Texture.GetFromGame` (game
  path such as `ui/icon/...tex`).
- Icons use `Svc.Texture.GetFromGameIcon(new(id, hiRes: hq))`. Despite the parameter name, `hq`
  selects the high-resolution icon, not the HQ-item overlay.
- File, game and icon entries hold an `ISharedImmediateTexture` and return
  `GetWrapOrDefault()` on every call. The source comment is explicit: do not cache the returned
  wrap, call the function every frame. Do not dispose the wrap either.
- An entry is marked completed before loading starts. A failed download or decode leaves the
  entry with a null texture permanently: the function keeps returning false and does not retry
  until `ClearAll()`.
- Errors go to `ErrorAction(exception, message)` if set (either argument may be null; invoked on
  the worker thread), otherwise to `PluginLog.Error` / `Exception.Log()`.
- `AddConversionToBitmap` registers a byte-array transform tried on downloaded content, in order,
  after the built-in identity conversion, until `CreateFromImageAsync` succeeds. Only URL loads use
  conversions.
- The worker exits after 100 iterations with nothing to load and is restarted by the next uncached
  request. It sleeps 100 ms per iteration only while no cache entry at all is completed; once
  anything has loaded, the idle iterations run without sleeping and the thread ends almost at once.
- `ClearAll()` disposes URL-created wraps and clears both caches; the next request reloads. Use it
  to force a reload or free memory.
- `ECommonsMain.Dispose()` calls `ClearAll()` and disposes the `HttpClient`.
- `ImageLoading/ImageLoadingResult` is `internal`.

## Window base classes

None of these register themselves unless stated. "Self-registering" means the class creates a
private `WindowSystem` and subscribes `Svc.PluginInterface.UiBuilder.Draw`.

| Class | Kind | Self-registering | Cleaned by `ECommonsMain.Dispose()` |
|---|---|---|---|
| `NotifyWindow` | abstract | no | no |
| `EzOverlayWindow` | abstract | no | no |
| `EzFullscreenOverlayWindow` | abstract | no | no |
| `MiddleOverlayWindow` | concrete, `IDisposable` | yes | no, call `Dispose()` |
| `TimedMiddleOverlayWindow` | concrete | yes | no, disposes itself on timeout |
| `ChangelogWindow` | concrete | yes, conditionally | no, detaches in `OnClose` |
| `TerritorySelector` | concrete | yes, one tick after construction | no, detaches in `OnClose` |
| `ItemSelectorWindow` | concrete | into the `WindowSystem` you pass | only if that is `EzConfigGui.WindowSystem` |

Windows added to `EzConfigGui.WindowSystem` are removed by `ECommonsMain.Dispose()`
(`references/topic-simplegui.md`).

### NotifyWindow

```csharp
public abstract class NotifyWindow : Window
public NotifyWindow(string name)     // flags NoCollapse | AlwaysAutoResize, forceMainWindow true
public abstract void DrawContent();
```

Blocking-notice style window: `RespectCloseHotkey = false`, `ShowCloseButton = false`. `Draw()`
calls `DrawContent()` inside try/catch (logged), then sets
`Position = ImGuiHelpers.MainViewport.Size / 2f - ImGui.GetWindowSize() / 2f`, which re-centres the
window every frame (it cannot be moved). `Draw` is not sealed, but override `DrawContent`. You must
set `IsOpen` and provide the `WindowSystem`; close it from your own button with `IsOpen = false`.
AutoRetainer's `UI/Windows/SingletonNotifyWindow.cs`:

```csharp
public class SingletonNotifyWindow : NotifyWindow
{
    private WindowSystem ws;
    public SingletonNotifyWindow() : base("AutoRetainer - warning!")
    {
        IsOpen = true;
        ws = new();
        Svc.PluginInterface.UiBuilder.Draw += ws.Draw;
        ws.AddWindow(this);
    }
    public override void OnClose() => Svc.PluginInterface.UiBuilder.Draw -= ws.Draw;
    public override void DrawContent()
    {
        ImGuiEx.Text("...");
        if(ImGui.Button("Close")) IsOpen = false;
    }
}
```

### EzOverlayWindow

```csharp
public abstract class EzOverlayWindow : Window
public EzOverlayWindow(string name, HorizontalPosition hPos, VerticalPosition vPos, Vector2? offset = null)
public HorizontalPosition HPos; public VerticalPosition VPos; public Vector2 Offset; public Vector2 WindowSize;
public virtual void PreDrawAction() { }
public abstract void DrawAction();
public enum VerticalPosition { Top, Middle, Bottom }
public enum HorizontalPosition { Left, Middle, Right }
```

Flags are `ImGuiEx.OverlayFlags | ImGuiWindowFlags.AlwaysAutoResize`, `forceMainWindow` true,
`RespectCloseHotkey = false`, `IsOpen = true`. `ImGuiEx.OverlayFlags` is `NoNav | NoMouseInputs |
NoDecoration | NoBackground | NoSavedSettings | NoFocusOnAppearing`, so the overlay is click-through
and has no background. `PreDraw()` and `Draw()` are `sealed`: override `PreDrawAction` (runs first
in `PreDraw`; push styles here and pop them in `PostDraw`) and `DrawAction`. Position is anchored
to the main viewport edge or centre using the window size measured on the previous frame, plus
`Offset`, through `ImGuiHelpers.SetNextWindowPosRelativeMainViewport`; the first frame is placed as
if the size were zero. To make it interactive, change `Flags` in your constructor.

```csharp
public class StatusOverlay : EzOverlayWindow
{
    public StatusOverlay() : base("##myplugin.status", HorizontalPosition.Right, VerticalPosition.Top, new(-10, 10)) { }
    public override bool DrawConditions() => C.Enabled;
    public override void DrawAction() => ImGuiEx.Text(EColor.GreenBright, "Running");
}
// EzConfigGui.WindowSystem.AddWindow(new StatusOverlay());
```

### EzFullscreenOverlayWindow

```csharp
public abstract class EzFullscreenOverlayWindow : Window
public EzFullscreenOverlayWindow(string name)     // ImGuiEx.OverlayFlags, forceMainWindow true
public virtual void PreDrawAction() { }
```

`PreDraw()` is sealed: calls `PreDrawAction()`, sets the next window size to
`ImGuiHelpers.MainViewport.Size` and position to the viewport origin. `Draw()` is not overridden
here; implement `Draw()` yourself (there is no `DrawAction`). Intended for draw-list rendering over
the whole game window. `IsOpen = true` and `RespectCloseHotkey = false` are set.

### MiddleOverlayWindow, TimedMiddleOverlayWindow

```csharp
public class MiddleOverlayWindow : Window, IDisposable
public MiddleOverlayWindow(string name, Action draw, int? topOffset = null, Vector4? bgCol = null)

public class TimedMiddleOverlayWindow : MiddleOverlayWindow
public TimedMiddleOverlayWindow(string name, long destroyAfterMS, Action draw, int? topOffset = null, Vector4? bgCol = null)
```

Self-registering, non-interactive banner, horizontally centred. Flags: `NoInputs | NoNav |
NoTitleBar | NoFocusOnAppearing | NoSavedSettings | NoScrollbar | AlwaysAutoResize`,
`forceMainWindow` true. Y position is `topOffset` pixels, or one third of the viewport height when
null. `bgCol` is pushed as `ImGuiCol.WindowBg` in `PreDraw` and popped in `PostDraw`.
`DrawConditions()` is `Svc.Objects.LocalPlayer != null`. `draw` is invoked without try/catch.
Position uses `ImGui.SetNextWindowPos` with viewport size only (viewport origin is not added).

`Dispose()` unsubscribes `UiBuilder.Draw` and removes the window; idempotent. The timed variant
overrides `Update()` and calls `Dispose()` once `Environment.TickCount64` passes the deadline.
Keep a reference to a plain `MiddleOverlayWindow` and dispose it in plugin `Dispose()`.

```csharp
new TimedMiddleOverlayWindow("##myplugin.done", 5000,
    () => ImGuiEx.Text(EColor.GreenBright, "All retainers processed"),
    bgCol: EColor.Black with { W = 0.7f });
```

### ChangelogWindow

```csharp
public class ChangelogWindow : Window
public ChangelogWindow(IPluginConfiguration Configuration, int version, Action func, Action onClose = null)
```

One-shot "was updated" window. Title `"{InternalName} was updated"`, flags `NoSavedSettings |
AlwaysAutoResize | NoCollapse`.

- The configuration type must have a public instance field `int ChangelogWindowVer`. It is read and
  written by reflection (`GetType().GetField("ChangelogWindowVer")`); a property or a missing field
  causes a `NullReferenceException` that the constructor catches and logs, and the window never
  shows.
- If `Svc.PluginInterface.Reason == PluginLoadReason.Installer` (fresh install) the constructor
  calls `OnClose()` immediately: the version is stored and nothing is displayed.
- Otherwise, if the stored value differs from `version`, the window opens and self-registers.
- `DrawConditions()` is `Svc.ClientState.IsLoggedIn`; the window waits for login.
- `OnClose()` writes `version` into the field, then invokes `onClose` if given, else
  `Svc.PluginInterface.SavePluginConfig(config)`. It then unsubscribes `UiBuilder.Draw`. With an
  `EzConfig`-managed configuration pass `onClose: EzConfig.Save`; note the constructor still
  requires the object to implement `IPluginConfiguration`.

```csharp
new ChangelogWindow(C, 3, () =>
{
    ImGuiEx.Text("Version 3:\n- Added zone filter\n- Fixed overlay position");
});
```

## TerritorySelector

`references/ecommons/ECommons/ImGuiMethods/TerritorySelection/TerritorySelector.cs`.
`public unsafe class TerritorySelector : Window`: a self-registering zone picker with category tabs
and a filterable table.

```csharp
// single selection (radio buttons)
public TerritorySelector(uint SelectedTerritory, Action<TerritorySelector, uint> Callback, string TitleName = null)
public TerritorySelector(Action<TerritorySelector, uint> Callback, string TitleName = null)
// multiple selection (checkboxes)
public TerritorySelector(IEnumerable<uint> SelectedTerritories, Action<TerritorySelector, HashSet<uint>> Callback, string TitleName = null)
public TerritorySelector(Action<TerritorySelector, HashSet<uint>> Callback, string TitleName = null)
public void Close()
```

| Member | Default | Notes |
|---|---|---|
| `static bool Singleton` | `true` | When true, opening a selector closes all others. When false, a second selector with the same title is refused with `Notify.Error("Territory selector is already open")`. |
| `static Dictionary<Category, TerritoryIntendedUseEnum[]> Categories` | World, Housing, Inn, Dungeon, Raid, Trial, Deep_Dungeon | Category to intended-use mapping; mutable. `Other` and `All` are computed. |
| `static readonly List<TerritorySelector> Selectors` | | Open instances. |
| `bool OnlySelected`, `string Filter`, `bool ShowUnnamed` | `false`, `""`, `false` | Bound to the header controls. |
| `uint[] HiddenTerritories`, `Category[] HiddenCategories` | empty | Excluded from the list. |
| `Category? SelectedCategory` | `null` | Tab to select on first draw; reset to null after each draw. |
| `Column[] ExtraColumns` | `[ID, Zone, Region, IntendedUse]` | |
| `DisplayMode Mode` | `PlaceNameAndDuty` | |
| `Action<TerritoryType, Vector4?, string> ActionDrawPlaceName` | `ImGuiEx.Text(col, name)` | Custom cell renderer, used in `PlaceNameAndDuty` mode. |

Enums: `Category { World, Housing, Inn, Dungeon, Raid, Trial, Deep_Dungeon, Other, All }`,
`Column { ID, Zone, Region, IntendedUse }`,
`DisplayMode { PlaceNameDutyUnion, PlaceNameAndDuty, PlaceNameOnly }`.

Behaviour:

- The constructor only schedules setup with `new TickScheduler(...)`. Setup (building the zone
  cache, registering the window, `IsOpen = true`) runs on the next framework update, which is why
  options can be set with an object initialiser after the constructor call. Setting
  `HiddenTerritories`, `HiddenCategories` or `Categories` after that tick has no effect on the
  instance.
- Title defaults to `"Select zones"` when `TitleName` is null.
- In multi-select mode the callback receives the selector's own `HashSet<uint>` (a copy of the
  enumerable you passed) after every change; copy it back into your configuration. In single-select
  mode it receives the chosen id.
- Callbacks from checkbox and radio clicks are wrapped in try/catch; callbacks from the "Add all
  visible" / "Remove all visible" buttons are not.
- The header row with the current zone and the bulk buttons is drawn only when `Player.Available`.
- Instances are single-use. `IsOpen` is hidden by an `[Obsolete(..., true)]` private member; close
  with `Close()` and create a new instance to reopen.
- `OnClose` unsubscribes `UiBuilder.Draw` and removes the instance from `Selectors`.

PandorasBox (`FeaturesSetup/FeatureConfigEditor.cs`):

```csharp
if(ImGuiEx.IconButton(FontAwesomeIcon.List))
{
    new TerritorySelector(territories, (sel, selected) =>
    {
        territories.Clear();
        territories.AddRange(selected);
    })
    {
        SelectedCategory = TerritorySelector.Category.All,
        ExtraColumns = [TerritorySelector.Column.ID, TerritorySelector.Column.IntendedUse],
    };
}
```

## ItemSelectorWindow (incomplete at this commit)

`references/ecommons/ECommons/ImGuiMethods/ItemSelection/ItemSelectorWindow.cs`.

```csharp
public class ItemSelectorWindow : Window
public ItemSelectorWindow(ICollection<uint> itemCollection, WindowSystem? windowSystem = null, string name = "Select Items")
public ItemSelectorWindow(Func<uint> itemSelected, Predicate<uint>? isItemSelected, WindowSystem? windowSystem = null, string name = "Select Items")
public virtual Predicate<uint> ItemFilter { get; set; } = null;
public OrderedDictionary<string, uint[]> CategoryGroups { get; set; }
public enum DisplayCategory { Weapons, Armor, Consumable, Materials, Housing, Crystals, Special, Other }
public string SearchStr = ""; public int MinILvl = 0; public int MaxILvl /* highest LevelItem row */;
public bool? Tradeable = null; public bool? Marketable = null; public bool? Desynth = null;
public List<ItemRarity> Rarities = []; public List<uint> Categories = [];
public void Redefine(ICollection<uint> itemCollection)
public void Redefine(Func<uint> itemSelected, Predicate<uint>? isItemSelected = null)
public sealed override void Draw()
public void RebuildItemList()
```

Do not use it as shipped. Reading the source:

- The `CategoryGroups` getter writes `field[...] = ...` while `field` is null and never allocates
  the dictionary, so the first read throws `NullReferenceException`. `Draw()` reads it every frame.
  Assigning your own `OrderedDictionary<string, uint[]>` before the first draw avoids that.
- The item table body is an empty loop: no item rows are drawn and nothing can be selected.
  `ItemCollection`, `ItemSelected`, `IsItemSelected` and `ItemFilter` are stored but never used.
- The doc comment says the `itemSelected` callback returns a bool to close the window, but its type
  is `Func<uint>`.
- The search box, category checkboxes with icons and `RebuildItemList()` filtering are implemented.
  `RebuildItemList` filters on name (`SearchStr`, case-sensitive `Contains`), `Categories`, item
  level range, `Tradeable`, `Marketable`, `Desynth`, `Rarities`.

What does work as plumbing: flags `NoSavedSettings`; a `##<guid>` suffix is appended to the name
when it has no `##`; when `windowSystem` is given the window adds itself and opens, and `OnClose`
removes it from that `WindowSystem`.

## PatreonBanner

`ECommons/ECommons/Funding/PatreonBanner.cs`, namespace `ECommons.Funding`. This file is not in the
skill's `references/` copy (`Funding/*` is excluded); it exists only in an ECommons checkout.

```csharp
public static class PatreonBanner
public static Func<bool> IsOfficialPlugin = () => false;
public static string Text = "♥ Patreon/KoFi";
public static string DonateLink => "https://www.patreon.com/NightmareXIV";
public static void DrawRaw()                                  // = DrawButton()
public static void DrawButton()
public static void RightTransparentTab(string? text = null)
public static void DrawRight()
```

- `DrawButton` draws a button labelled `Text` with an animated background (a
  `GradientColor.Get` between `0x022594` and `0x940238`, shared between plugins through
  `EzSharedData` key `"ECommonsPatreonBannerRandomColor"`). Left click opens `DonateLink` with
  `GenericHelpers.ShellStart`; hovering shows a tooltip naming the plugin
  (`Svc.PluginInterface.Manifest.Name`); right click opens a popup with Patreon, Ko-Fi and
  cryptocurrency links, plus Discord and "other plugins" links when `IsOfficialPlugin()` is false.
- `DrawRight` moves the cursor so the button is right-aligned on the current line, draws it, then
  restores the cursor position. Call it immediately before a tab bar to overlay the button on the
  tab bar's right edge.
- `RightTransparentTab` must be called inside a tab bar: it adds a disabled, fully transparent
  trailing tab as wide as `text` (default `Text`) to reserve that space.
  `ImGuiEx.EzTabBar(id, KoFiTransparent, ...)` calls `PatreonBanner.RightTransparentTab()` with no
  argument whenever `KoFiTransparent` is non-null; the string's content is ignored.
- `DonateLink` is get-only and all URLs are hard-coded to NightmareXIV. `Text` and
  `IsOfficialPlugin` are assignable. This is author branding; a plugin by someone else should not
  show it. For a Ko-fi button aimed at Puni.sh plugins see the `punishlib` skill.

AutoRetainer (`UI/MainWindow/AutoRetainerWindow.cs`):

```csharp
PatreonBanner.IsOfficialPlugin = () => true;     // in the window constructor
// in Draw:
PatreonBanner.DrawRight();
ImGuiEx.EzTabBar("tabbar", PatreonBanner.Text,
    ("Retainers", MultiModeUI.Draw, null, true),
    ("About", CustomAboutTab.Draw, null, true));
```

## Gotchas

- `EzColor.Transparent` is opaque black; `new EzColor(0x000000FF)` is opaque blue, not opaque
  black. Any `uint` below `0x01000000` is read as `0xRRGGBB`.
- `EzColor.From(ImGuiCol)` feeds ImGui's ABGR-packed `GetColorU32` result into the RGBA parser, so
  the channels come out reversed (R gets alpha, G gets blue, B gets green, A gets red). Use
  `EzColor.FromABGR(ImGui.GetColorU32(col))` or `new EzColor(ImGui.GetStyle().Colors[(int)col])`.
- `implicit operator uint(EzColor)` and the obsolete `U32` give ImGui ABGR; `ARGB` and `RGBA` are
  different packings. Pass the struct itself (or `.Vector4`) to ImGui calls.
- `EColor.Cya`, not `Cyan`. `EzColor.Cyan` exists.
- `Notify.*` is deferred by one framework tick and uses the plugin `InternalName` as title.
- `ThreadLoadImageHandler` never retries a failed load, and returned wraps must not be cached or
  disposed.
- `ImGuiNoRef.Checkbox` sets `newResult` to false when nothing was clicked.
- `Ref<T>.Get()` without a key walks the stack on every call.
- Self-registering windows (`MiddleOverlayWindow`, `ChangelogWindow`, `TerritorySelector`,
  `PopupWindow`) each create their own `WindowSystem` without a namespace; give them distinct names.
- `ECommonsMain.Dispose()` does not dispose `MiddleOverlayWindow` instances or close open
  selectors. Their `UiBuilder.Draw` subscriptions end when Dalamud disposes the plugin's UiBuilder.
- `EzOverlayWindow` and `EzFullscreenOverlayWindow` ignore mouse input by default
  (`NoMouseInputs`); buttons inside them do nothing unless `Flags` is changed.
- `ImGuiTrans` is `[Obsolete(..., true)]`: compile error, no named replacement.
- `ItemSelectorWindow` throws on first draw and draws no items.
- `EnumMemberNameAttribute` is inert.
