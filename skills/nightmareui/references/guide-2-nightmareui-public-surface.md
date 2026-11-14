## 2. NightmareUI public surface

Namespaces: `NightmareUI` (`NuiTools`, `NightmareUIState`), `NightmareUI.PrimaryUI` (`NuiBuilder`,
`Nui`), `NightmareUI.ImGuiElements` (`WorldSelector`), `NightmareUI.Censoring` (`Censor`,
`CensorConfig`). `NightmareUI.PrimaryUI.Components` is entirely `internal` (`IWidget`,
`ImGuiWidget`, `Section`, `SeparatorWidget`, `CondIf`, `CondElse`, `CondEndIf`). AutoRetainer puts
`global using NightmareUI.PrimaryUI;` in its `Global.cs`.

### 2.1 `NuiBuilder` (`PrimaryUI/NuiBuilder.cs`)

A fluent list of sections, each a list of widgets. Every method returns `this`. Building records
delegates; `Draw` runs them. All widget methods throw `NullReferenceException("CurrentSection is
null")` if called before the first `Section(...)`.

```csharp
public string Filter = "";
public bool ShouldHighlight { get; }            // any ImGuiWidget label matches Filter
public NuiBuilder Section(string name, Vector4? color = null, bool collapsible = false,
                          Action? rightFloat = null, Func<bool>? visible = null)
public NuiBuilder Draw(bool noCollapse = false)

public NuiBuilder Widget(string name, Action<string> drawAction, string? help = null)
public NuiBuilder Widget(float width, string name, Action<string> drawAction, string? help = null)
public NuiBuilder Widget(float width, Action drawAction)
public NuiBuilder Widget(Action drawAction)
public NuiBuilder TextWrapped(string text)
public NuiBuilder TextWrapped(Vector4? col, string text)
public NuiBuilder Separator()   // likewise parameterless: Indent(), Unindent(), Else(), EndIf()
public NuiBuilder If(Func<bool> cond)

public NuiBuilder Checkbox(string name, RefBoolDelegate value, string? help = null)
public NuiBuilder CheckboxInverted(string name, RefBoolDelegate value, string? help = null)
public NuiBuilder InputInt(float width, string name, RefIntDelegate value, string? help = null)
public NuiBuilder InputInt(float width, string name, RefIntDelegate value, int step, int step_fast,
                           string? help = null)
public NuiBuilder InputFloat(float width, string name, RefFloatDelegate value, string? help = null,
                             string? format = null)
public NuiBuilder InputText(float width, string name, RefStringDelegate value, int maxLength = 512,
                            string? help = null)
public NuiBuilder DragInt(float width, string name, RefIntDelegate value, string? help = null)
// + 3 DragInt overloads inserting (float v_speed), (float v_speed, int v_min) or
//   (float v_speed, int v_min, int v_max) between value and help
public NuiBuilder SliderInt(float width, string name, RefIntDelegate value, int min, int max,
                            string? help = null, ImGuiSliderFlags flags = ImGuiSliderFlags.None)
public NuiBuilder SliderFloat(float width, string name, RefFloatDelegate value, float min,
                              float max, string? help = null,
                              ImGuiSliderFlags flags = ImGuiSliderFlags.None)
public NuiBuilder SliderIntAsFloat(float width, string name, RefIntDelegate value, int min, int max,
                                   float divider = 1000, string? help = null)
public NuiBuilder EnumCombo<T>(float width, string name, RefEnumDelegate<T> value,
        IDictionary<T, string>? names = null, string? help = null) where T : Enum, IConvertible
public NuiBuilder EnumComboFullWidth<T>(float? width, string name, RefEnumDelegate<T> value,
        Func<T, bool> filter = null, IDictionary<T, string>? names = null, string? help = null)
        where T : Enum, IConvertible

public delegate ref bool RefBoolDelegate();   // same shape: RefIntDelegate, RefFloatDelegate,
                                              // RefStringDelegate (all nested in NuiBuilder)
public delegate ref T RefEnumDelegate<T>() where T : Enum, IConvertible;
```

The `Ref*Delegate` types return a `ref`, so the argument is written `() => ref C.SomeField`. It
must be a field (or ref-returning member), not an auto-property.

What it renders (`Components/Section.cs`, internal):

- Each section is a one-column `ImGui.BeginTable(Name, 1, Borders | SizingFixedFit)` with
  `CellPadding` pushed to `7f`. Row 1 is the header: the section name on a background of `color`,
  or `ImGuiCol.FrameBg` when `color` is null. Row 2 holds all widgets. After the table comes
  `ImGui.Dummy(new(5f))`.
- `Separator()` starts a new table row (a border line); consecutive separators collapse.
- `collapsible: true` draws a FontAwesome minus/plus before the name and toggles on a left click
  anywhere on the header row while the window is focused. The collapsed set is
  `NuiTools.State.CollapsedHeaders`, keyed by `ImGui.GetID(Name + "NuiSection")`.
  `Draw(noCollapse: true)` ignores clicks but still honours an already collapsed state.
- `rightFloat` is drawn after the table, moved back to the header line through
  `ImGuiEx.RightFloat`. `visible` returning false skips the whole section.
- Each widget runs as `DrawAction(Label)` inside its own `try`/`catch` that calls `e.Log()`; the
  original cell padding is restored around the call. When `help` is non-null,
  `ImGuiEx.HelpMarker(help)` follows the widget.
- Width: the typed helpers and `Widget(float, Action)` call `ImGui.SetNextItemWidth(width)`
  themselves; `Widget(float, string, Action<string>, ...)` stores the width and `Section` applies
  it before the draw action. Widths are passed through as given, with no global-scale multiply.
- `EnumComboFullWidth` draws the name as wrapped text (plus help marker), then an indented combo
  labelled `##name`, full width when `width` is null.
- Filter: when `Filter != ""`, widgets whose label contains the filter (ordinal, ignore case) get
  text colour `ImGuiColors.ParsedGreen`, all others `ImGuiColors.DalamudGrey3`. Nothing is hidden.
  Label-less widgets (`Widget(Action)`, `TextWrapped`, `Indent`) never match.
- `If(cond)` ... `Else()` ... `EndIf()` gate the widgets (and separators) between them; the
  predicate is re-evaluated for every gated widget on every frame.

### 2.2 `Nui` (`PrimaryUI/Nui.cs`)

```csharp
public static void Checkbox(string label, ref bool value, string? help = null,
                            Action? callback = null)
public static void Widget(string label, Action<string> drawAction, string? help = null)
```

Immediate-mode helpers. `Checkbox` invokes `callback` on change and appends a help marker;
`Widget` wraps the draw action in `try`/`catch` + `e.Log()`. Neither consumer calls them.

### 2.3 `NuiTools` and `NightmareUIState` (`NuiTools.cs`, `NightmareUIState.cs`)

```csharp
public static NightmareUIState State { get; private set; } = new();
public static void SetState(NightmareUIState state)
public static void ButtonTabs(ButtonInfo[][] buttons2d, int maxButtons = int.MaxValue)
public static void ButtonTabs(string id, ButtonInfo[][] buttons2d, int maxButtons = int.MaxValue,
        bool child = true, float heightMultiplier = 1f, bool scaling = true)
public static bool RenderResidentialIcon(this uint residentialAetheryte, float? size = null)

public record class ButtonInfo   // nested: NuiTools.ButtonInfo
{
    public readonly string Name, InternalName;  public readonly Action Action;
    public readonly FontAwesomeIcon Icon;
    public ButtonInfo(string name, Action action)
    public ButtonInfo(string name, string internalName, Action action)
    public ButtonInfo(string name, Action action, FontAwesomeIcon icon)
    public ButtonInfo(string name, string internalName, Action action, FontAwesomeIcon icon)
}

[Serializable] public sealed unsafe class NightmareUIState
{
    public Dictionary<string, string> ActiveTab = [];     // tab bar id -> InternalName
    public HashSet<uint> CollapsedHeaders = [];           // ImGui IDs of collapsed sections
}
```

`ButtonTabs` draws each inner array as one row of square-cornered buttons with zero horizontal
spacing, each `ContentRegionAvail.X / min(maxButtons, count)` wide; a row wraps after `maxButtons`
buttons. The active button is painted with `ImGuiCol.ButtonActive`. Entries with a non-default
`Icon` use `ImGuiEx.IconButtonWithText`. Then, inside `ImGui.BeginChild($"NuiTabs{id}")` when
`child` is true, it invokes the `Action` of every button whose `InternalName` equals the stored
active name, every frame, inside a `try`/`catch` that logs. `null` entries are skipped, which is
how a tab is made conditional. The two-argument overload uses `GenericHelpers.GetCallStackID()`
as the id. `InternalName` defaults to `Name`; pass it explicitly when `Name` is localised.

`SetState` swaps the static state object for one owned by the plugin's configuration, which makes
the active tabs and collapsed sections survive a reload once the configuration is saved. Without
it the state lives only for the session.

`RenderResidentialIcon` is an extension on `uint`. It matches the value against ECommons'
`ResidentalAreas` territory constants (any other value gets a fallback icon), draws that region
of `ui/uld/Teleport_hr1.tex` at `size` (default: button height) and returns false while the
texture is not loaded yet.

### 2.4 `WorldSelector` (`ImGuiElements/WorldSelector.cs`)

```csharp
public static WorldSelector Instance { get; }            // lazily created shared instance
public WorldSelector(string id = "##world")
public string EmptyName { get; set; } = null;            // label for 0; also adds a row for it
public bool DisplayCurrent { get; set; } = false;        // adds "Current: <world>" when logged in
public bool DefaultAllOpen { get; set; } = false;
public Predicate<uint>? ShouldHideWorld { get; set; } = null;
public void Draw(ref int worldConfig, ImGuiComboFlags flags = ImGuiComboFlags.HeightLarge)
public void DrawInternal(ref int worldConfig)
```

`Draw` is a combo whose preview is the world name (`"Not selected"` or `EmptyName` for 0) and
whose popup is a search box plus a Region > Data Center > World tree of public worlds.
`DrawInternal` is the popup body, for embedding in your own popup. The value is the World row id
stored as `int`. The combo has an empty label, so set the width with `ImGui.SetNextItemWidth`
before `Draw`. Real use: `Splatoon/SplatoonScripts/Generic/WorldWaiter.cs` keeps
`private WorldSelector WorldSelector = new();` and calls `WorldSelector.Draw(ref World);`.

### 2.5 `Censor` and `CensorConfig` (`Censoring/`)

```csharp
public static CensorConfig Config { get; set; } = new();
public static string Character(string name, string world)  // "Adjective Animal@World"
public static string Character(string name)                // splits "Name@World" when it has '@'
public static string World(string world)
public static string Hide(string s, string placeholder = "Hidden")
public static string Hide(object s, string placeholder = "Hidden")
public static readonly IReadOnlyDictionary<CensorType, string[]> CensorData
public enum CensorType { Animal, Adjective, World }        // nested: Censor.CensorType

public class CensorConfig { public string Seed = Guid.NewGuid().ToString();
                            public bool Enabled = false; public bool LesserCensor = false; }
```

Deterministic streamer-mode pseudonyms: index = `Lumina.Misc.Crc32.Get(Config.Seed + input)`
modulo the word list. With `Enabled == false` every method returns its input (`Character(name,
world)` returns `name@world`). `LesserCensor` restricts candidates to words starting with the same
letter as the original. Store a `CensorConfig` in your configuration and assign it to
`Censor.Config` after loading, otherwise the seed changes on every load. AutoRetainer does **not**
use this class; its `Censor.Character`/`Censor.Retainer` calls go to its own internal
`Helpers/Censor.cs`, so beware the name clash when importing `NightmareUI.Censoring` there.

