OtterGui topic: catalogue of every widget in `Widgets/` and `Text/Widget/` (filtered combos,
tutorial, changelog, tag buttons, tab bar, checkboxes, editors, key selectors, framed group, hex
viewer, colour pickers, toggle/split buttons) with constructor and draw signatures. Verified
against `Limiana/OtterGui` `a08fda6` (2026-05-30, branch `main`). Compared with upstream `79771ee`
the only difference in these folders is one line in `Widgets/FilteredCombo.cs`, flagged below.

Paths such as `references/...` and `references/topic-*.md` are relative to the skill root (the directory that
holds `SKILL.md`), not to this file. Signatures are copied from `references/api-index.txt`; nothing
was compiled.

## Namespaces

| Folder | Namespace |
|---|---|
| `Widgets/` | `OtterGui.Widgets` |
| `Text/Widget/` | `OtterGui.Text.Widget` |
| `Text/Widget/Editors/` | `OtterGui.Text.Widget.Editors` |

Two types exist in both worlds with the same simple name: `Changelog` and `ChangeLogDisplayType`
are declared in `OtterGui.Widgets` (string API) **and** in `OtterGui.Text.Widget` (UTF-8 API).
Importing both namespaces makes them ambiguous. `TristateCheckbox` (`OtterGui.Widgets`, values
`sbyte` -1/0/1) and `TriStateCheckbox` (`OtterGui.Text.Widget`, values `bool?`) differ by one
capital letter.

## Filtered combos (`Widgets/FilteredCombo.cs`, `FilterComboCache.cs`, `FilterComboColors.cs`)

A combo with a filter text field on top, clipped list, keyboard navigation (Up/Down wrap around,
Enter picks the highlighted or first entry, Escape closes) and optional mouse-wheel scrolling on
the closed combo.

```csharp
[Flags] public enum MouseWheelType : byte { None = 0, Unmodified = 1, Shift = 2, Control = 4, Alt = 8 }

public abstract class FilterComboBase<T>
{
    protected FilterComboBase(IReadOnlyList<T> items, bool keepStorage, Logger log);

    public readonly IReadOnlyList<T> Items;
    public LowerString Filter { get; }
    public void ResetFilter();                       // marks the filter dirty
    public virtual bool Draw(string label, string preview, string tooltip, ref int currentSelection,
        float previewWidth, float itemHeight, ImGuiComboFlags flags = ImGuiComboFlags.None);

    protected readonly Logger Log;
    protected bool SearchByParts;                    // split the filter at spaces, all parts must match
    protected int? NewSelection;
    protected MouseWheelType AllowMouseWheel { get; set; }

    protected virtual string ToString(T obj);                                // default obj?.ToString()
    protected virtual bool   IsVisible(int globalIndex, LowerString filter);
    protected virtual bool   DrawSelectable(int globalIdx, bool selected);   // default ImGui.Selectable(ToString(item))
    protected virtual float  GetFilterWidth();
    protected virtual void   DrawCombo(string label, string preview, string tooltip, int currentSelected,
        float previewWidth, float itemHeight, ImGuiComboFlags flags);
    protected virtual void   DrawFilter(int currentSelected, float width);
    protected virtual void   DrawList(float width, float itemHeight);
    protected virtual int    UpdateCurrentSelected(int currentSelected);
    protected virtual void   PostCombo(float previewWidth);
    protected virtual void   Cleanup();
    protected virtual void   OnClosePopup();
    protected virtual void   OnMouseWheel(string preview, ref int currentSelection, int steps);
    protected void DrawKeyboardNavigation();
    protected void ClosePopup(uint id, string label);
}
```

`Draw` returns `true` in the frame a new entry was chosen and writes its index in `Items` to
`currentSelection`. `itemHeight` must be the real height of one entry including spacing, because
the list is clipped with it. `previewWidth` is passed to `SetNextItemWidth`. The combo is always
opened with `ImGuiComboFlags.HeightLarge` added. `keepStorage: false` frees the filtered index
list when the popup closes.

```csharp
public abstract class FilterComboCache<T> : FilterComboBase<T>
{
    protected FilterComboCache(IEnumerable<T> items, MouseWheelType allowMouseWheel, Logger log);
    protected FilterComboCache(Func<IReadOnlyList<T>> generator, MouseWheelType allowMouseWheel, Logger log);

    public T? CurrentSelection { get; protected set; }
    public event Action<T?, T?>? SelectionChanged;          // (old, new)
    public bool Draw(string label, string preview, string tooltip, float previewWidth, float itemHeight,
        ImGuiComboFlags flags = ImGuiComboFlags.None);

    protected int  CurrentSelectionIdx = -1;
    protected bool IsInitialized { get; }
    protected virtual void UpdateSelection(T? newSelection);
}
```

`FilterComboCache` wraps the items in a `TemporaryList<T>` or `LazyList<T>` (`Classes/`) and
clears that list whenever the popup closes, so the item list is only materialised while the combo
is open. It tracks the selection itself; `SelectionChanged` fires when the reference changes.

Minimal subclass (not compiled):

```csharp
public sealed class WorldCombo(IEnumerable<string> worlds, Logger log)
    : FilterComboCache<string>(worlds, MouseWheelType.Control, log)
{
    public bool Draw(float width)
        => Draw("##world", CurrentSelection ?? "Select World...", string.Empty, width,
            ImGui.GetTextLineHeightWithSpacing());
}
```

`FilterComboColors : FilterComboCache<KeyValuePair<byte, (string Name, uint Color, bool Gloss)>>`
is a colour-swatch combo whose entries are `(Name, Color, Gloss)` keyed by a `byte`: constructor `(float comboWidth, MouseWheelType
allowMouseWheel, Func<IReadOnlyList<KeyValuePair<byte, (string Name, uint Color, bool Gloss)>>>
colors, Logger log)`, draw `Draw(string label, uint color, string name, bool found, bool gloss,
float previewWidth, MouseWheelType mouseWheel = MouseWheelType.Control)` and an overload without
`previewWidth`.

Fork versus upstream: in `UpdateFilter` the fork reads `_available[_lastSelection]` whenever
`_lastSelection != -1`; upstream first checks `_available.Count <= _lastSelection` (commit
`5a9ca81`, "Fix bug when opening filtered combo after mousewheel-scrolling it in a successive
frame"). On the fork that sequence can throw `ArgumentOutOfRangeException`.

`ClippedSelectableCombo<T>` (`Widgets/ClippedSelectableCombo.cs`) is an older, simpler variant:

```csharp
public ClippedSelectableCombo(string id, string label, float previewSize, IList<T> items, Func<T, string> itemToName)
public bool Draw(string currentName, out int newIdx, ImGuiComboFlags flags = ImGuiComboFlags.None)
public bool Draw(int currentIdx, out int newIdx, ImGuiComboFlags flags = ImGuiComboFlags.None)
public const int ItemsAtOnce = 12;
```

`previewSize` is unscaled (multiplied by `ImGuiHelpers.GlobalScale` inside). Enter selects the
first remaining entry.

## Tutorial (`Widgets/Tutorial.cs`)

```csharp
public class Tutorial
{
    public record struct Step(string Name, string Text, bool Enabled);
    public uint   HighlightColor { get; init; } = 0xFF20FFFF;
    public uint   BorderColor    { get; init; } = 0xD00000FF;
    public string PopupLabel     { get; init; } = "Tutorial";
    public int EndStep { get; }                          // number of registered steps
    public IReadOnlyList<Step> Steps { get; }

    public Tutorial Register(string name, string text);  // append an enabled step
    public Tutorial Deprecated();                        // append a disabled placeholder
    public Tutorial EnsureSize(int size);                // throws if the step count differs
    public void Open(int id, int current, Action<int> setter);
    public void Skip(int id, int current, Action<int> setter);
    public int  CurrentEnabledId(int current);
}
```

Usage: build the step list once with chained `Register` calls, keep one `int` in your
configuration, and call `Open(stepId, config.TutorialStep, v => { config.TutorialStep = v; Save(); })`
**immediately after drawing the widget that step explains**; `Open` highlights the last item
(`GetItemRectMin/Max`) and shows a popup next to it. Values written through `setter`: the next
enabled step id, `EndStep` for "Skip Tutorial" or after the last step, `-1` for "Disable
Tutorial". Step ids are positions in the list, so never remove a step: replace it with
`Deprecated()` to keep later ids stable.

## Changelog (`Widgets/Changelog.cs`, `Text/Widget/Changelog.cs`)

A `Dalamud.Interface.Windowing.Window` that opens itself when there are versions the user has not
seen.

```csharp
public enum ChangeLogDisplayType { New, HighlightOnly, Never }

public sealed class Changelog : Window
{
    public Changelog(string label, Func<(int, ChangeLogDisplayType)> getConfig,
        Action<int, ChangeLogDisplayType> setConfig);

    public const int  FreshInstallVersion   = int.MaxValue;
    public const uint DefaultHeaderColor    = 0xFF60D0D0;
    public const uint DefaultImportantColor = 0xFF6060FF;
    public const uint DefaultHighlightColor = 0xFFFF9090;
    public uint HeaderColor { get; set; }
    public bool ForceOpen   { get; set; }

    // OtterGui.Widgets version (string):
    public Changelog NextVersion(string title);
    public Changelog RegisterEntry(string text, ushort level = 0);
    public Changelog RegisterHighlight(string text, ushort level = 0, uint color = DefaultHighlightColor);
    public Changelog RegisterImportant(string text, ushort level = 0, uint color = DefaultImportantColor);
    public static string ToName(ChangeLogDisplayType type);
}
```

The `OtterGui.Text.Widget` version has the same members with `ReadOnlySpan<byte>` instead of
`string` (`NextVersion("Version 1.2"u8)`), and `ToName` returns `ReadOnlySpan<byte>`.

Behaviour of `PreOpenCheck` (read from the source): the stored `int` is the number of versions
already seen. `FreshInstallVersion` stores the current count and stays closed. `New` opens when
the stored count is lower than the number of registered versions. `HighlightOnly` opens only if
an unseen version contains a `RegisterHighlight` / `RegisterImportant` entry. `ForceOpen` opens
unconditionally. The window must be added to your `WindowSystem`; `level` indents an entry.

## Tag buttons (`Widgets/TagButtons.cs`)

```csharp
public class TagButtons
{
    public int Draw(string label, string description, IReadOnlyCollection<string> tags, out string editedTag,
        bool editable = true, float xOffset = 0, float rightEndOffset = 0);
}
```

Returns `-1` when nothing changed, otherwise the index of the edited tag, or `tags.Count` for a
newly added one, with the text in `editedTag`. The instance holds the edit state, so keep one
instance per tag list.

## Tab bar (`Widgets/TabBar.cs`)

```csharp
public interface ITab
{
    ReadOnlySpan<byte> Label { get; }
    bool IsVisible => true;
    ImGuiTabItemFlags Flags => ImGuiTabItemFlags.None;
    void DrawContent();
    void DrawHeader() { }
}

public static class TabBar
{
    public static bool Draw(string label, ImGuiTabBarFlags flags = ImGuiTabBarFlags.None, params ITab[] tabs);
    public static bool Draw(string label, ReadOnlySpan<byte> selectTab, params ITab[] tabs);
    public static bool Draw(string label, ImGuiTabBarFlags flags, ReadOnlySpan<byte> selectTab,
        out ReadOnlySpan<byte> currentTab, Action buttons, params ITab[] tabs);
}
```

`Label` is passed to ImGui through `fixed`, so it must be null-terminated: return a `u8` literal.
`selectTab` forces the tab with that exact label to be selected this frame; the return value is
`true` when such a tab was found. `buttons` runs after the tabs, inside the bar (for
`TabItemButton`s).

## Checkboxes (`Text/Widget`)

```csharp
public abstract class MultiStateCheckbox<T>
{
    protected abstract void RenderSymbol(T value, Vector2 position, float size);
    protected abstract T NextValue(T value);        // left click
    protected abstract T PreviousValue(T value);    // right click
    public bool Draw(ReadOnlySpan<byte> label, ref T value);                  // + char and handler overloads
    public bool Draw(ReadOnlySpan<byte> label, T currentValue, out T newValue);
}
```

| Type | `T` | Notes |
|---|---|---|
| `TwoStateCheckbox` | `bool` | `TwoStateCheckbox.Instance` |
| `XCheckbox` | `bool` | draws a cross instead of a check; `XCheckbox.Instance` |
| `TriStateCheckbox` | `bool?` | `TriStateCheckbox.Instance`; extra `Draw<T>(ReadOnlySpan<byte> label, ref T value, T onFlag, T offFlag) where T : unmanaged, Enum` maps two flags to true / false / null (neither flag set is treated like both set) |
| `FontAwesomeCheckbox(FontAwesomeIcon icon, uint? color = null)` | `bool` | shows the icon when true |
| `FontAwesomeCheckbox<T>` (abstract) | any | implement `GetIcon(T value)` returning `(FontAwesomeIcon? Icon, uint? Color)` |
| `OtterGui.Widgets.TristateCheckbox(uint crossColor = 0xFF0000FF, uint checkColor = 0xFF00FF00, uint dotColor = 0xFFD0D0D0)` | `sbyte` | -1 cross, 1 check, 0 dot; order 0 -> 1 -> -1 -> 0 |

These are drawn with ImGui internals (`ImGuiInternal.ItemAdd`, `ButtonBehavior`), not with
`ImGui.Checkbox`.

## Editors (`Text/Widget/Editors`)

Pre-configured components that edit a `Span<T>` of contiguous values in place.

```csharp
public interface IEditor<T> where T : unmanaged
{
    bool Draw(Span<T> values, bool disabled);
    IEditor<TStored> Reinterpreting<TStored>() where TStored : unmanaged;
    sealed IEditor<TStored> Converting<TStored>(Func<TStored, T> convert, Func<T, TStored> convertBack)
        where TStored : unmanaged;
}
```

ID and label handling are left to the caller (push an ID, draw the label yourself); `Draw`
honours `ImGui.SetNextItemWidth`.

| Type | Factory methods |
|---|---|
| `DragEditor<T>` | `CreateInteger(T? minimum, T? maximum, float speed, float relativeSpeed, unit, ImGuiSliderFlags flags)`, `CreateFloat(T? minimum, T? maximum, float speed, float relativeSpeed, byte precision, unit, ImGuiSliderFlags flags)` |
| `SliderEditor<T>` | `CreateInteger(T minimum, T maximum, unit, ImGuiSliderFlags flags)`, `CreateFloat(T minimum, T maximum, byte precision, unit, ImGuiSliderFlags flags)` |
| `InputEditor<T>` | `CreateInteger(T? minimum, T? maximum, T step, T stepFast, bool hex, unit, ImGuiInputTextFlags flags)`, `CreateFloat(T? minimum, T? maximum, T step, T stepFast, byte precision, unit, ImGuiInputTextFlags flags)` |
| `ColorEditor` | `ColorEditor.StandardDynamicRange`, `ColorEditor.HighDynamicRange`, `ColorEditor.Get(bool hdr)`; implements `IEditor<float>`, `IEditor<Vector3>`, `IEditor<Vector4>` |
| `EnumEditor<T>` | constructor `(IReadOnlyList<(ReadOnlyMemory<byte> Label, T Value, ReadOnlyMemory<byte> Description)> domain)` |
| `Editors` (static) | `DefaultFloat`, `DefaultInt`, `AsEditor<T>(this MultiStateCheckbox<T> inner)`, `PrepareMultiComponent(int numComponents)`, `TryAppendPrintfLiteral(...)`, nested `FormatBuffer` (`Size = 32`), `IdBuffer` (`Size = 16`), `ComponentHelper` |

`unit` is `scoped ReadOnlySpan<byte>` or `scoped ReadOnlySpan<char>` (two overloads each).

## Static `Widget` helpers (`OtterGui.Widgets.Widget`, partial class over several files)

| Function | Signature | Returns |
|---|---|---|
| `BeginFramedGroup` | `float BeginFramedGroup(string label, string description = "", uint headerColor = 0, FontAwesomeIcon headerPreSymbol = FontAwesomeIcon.None)`; overload with `Vector2 minSize` after `label` | a width; pair with `EndFramedGroup`, or use `ImRaii.FramedGroup(...)` |
| `EndFramedGroup` | `void EndFramedGroup(uint borderColor = 0)` | |
| `DrawHexViewer` | `void DrawHexViewer(ReadOnlySpan<byte> data, uint offsetColor = 0, uint color1 = 0, uint color2 = 0)` | |
| `ColorPicker` | `bool ColorPicker(string label, string tooltip, uint currentColor, Action<uint> setter, uint defaultColor)` | true when `setter` was called; draws a "Default" button |
| `PaletteColorPicker` | `bool PaletteColorPicker(string label, Vector2 iconSize, int currentColorIdx, int defaultColorIdx, IDictionary<int, uint> colors, out int newColorIdx, int boxesPerLine = 10)` | true on change; `newColorIdx` is `-1` when cleared |
| `DrawChatTypeSelector` | `bool DrawChatTypeSelector(string label, string description, XivChatType currentValue, Action<XivChatType> setter)` | true on change |
| `KeySelector` | `bool KeySelector(string label, string description, VirtualKey currentValue, Action<VirtualKey> setter, IReadOnlyList<VirtualKey> keys)` | true on change |
| `ModifierSelector` | `bool ModifierSelector(string label, string description, ModifierHotkey currentValue, Action<ModifierHotkey> setter)` | |
| `DoubleModifierSelector` | `bool DoubleModifierSelector(string label, string description, float width, DoubleModifier currentValue, Action<DoubleModifier> setter)` | |
| `ModifiableKeySelector` | `bool ModifiableKeySelector(string label, string description, float width, ModifiableHotkey currentValue, Action<ModifiableHotkey> setter, IReadOnlyList<VirtualKey> keys)` | |
| `RenameableCombo` | `bool RenameableCombo(string label, ref int currentItem, out string newName, string[] items, int numItems)` | |
| `ResizingTextInput` | `bool ResizingTextInput(string label, ref string input, uint maxLength)`; `(ref string input, uint maxLength)` | true on Enter |
| `InputOrText` | `bool InputOrText(bool editable, string label, ref string text, uint maxLength)` | |

The hotkey types (`ModifierHotkey`, `DoubleModifier`, `ModifiableHotkey`) are in
`OtterGui.Classes`, see `references/topic-services-classes.md`.

## Toggle and split buttons (`Widgets/ToggleButton.cs`)

```csharp
public static class ToggleButton
{
    public static bool ButtonEx(ReadOnlySpan<char> label, Vector2 sizeArg, ImGuiButtonFlags flags, ImDrawFlags corners);
    public static int  SplitButton(ImGuiId id, in SplitButtonData left, in SplitButtonData right, Vector2 size, uint sharedBorder);

    public ref struct SplitButtonData
    {
        public required ReadOnlySpan<byte> Label;
        public ReadOnlySpan<byte> Tooltip;
        public uint Background, Active, Hovered, Border, Text;
    }
}
```

`SplitButton` draws one rectangle split along its diagonal and returns `1` when the left half was
clicked, `2` for the right half, `0` otherwise. `ImGuiId` is `OtterGuiInternal.Enums.ImGuiId`;
`Custom/CustomGui.cs` passes `(ImGuiId)5`. `ToggleButton.ButtonEx` renders its label through
`ImGuiInternal.RenderTextClipped`, which at the fork's `OtterGuiInternal` pin under-allocates for
three-byte UTF-8 characters (`references/topic-fork-delta.md`).

`WidgetUtil` (`Widgets/WidgetUtil.cs`) has no public members.
