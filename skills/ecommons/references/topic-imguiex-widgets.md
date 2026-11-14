# ImGuiEx widgets: buttons, checkboxes, inputs, combos, drag-drop

Covers: the widget half of `ECommons.ImGuiMethods.ImGuiEx` (buttons, checkbox variants, collection
checkboxes, numeric/text inputs, sliders, enum and sheet-backed combos, row drag-drop) plus the
`ImGuiDragDrop` payload helpers. Verified against ECommons `bc38dc7` (2026-11-08, v3.2.1.26, Dalamud
API 15); nothing here was compiled. Paths are relative to the skill root.

Not covered here: text, alignment, tables, tooltips, `Scale()`, `Ctrl`/`Shift`/`Alt`,
`DragDropRepopulate*` and the rest of `ImGuiEx.cs` -> `references/topic-imguiex-core.md`. Other
`ImGuiMethods/` classes (`EColor`, `Box<T>`, `Ref<T>`, `ImGuiBB`, ...) -> `references/topic-imgui-helpers.md`.
Stock bindings, `ImRaii`, `ImU8String` -> `dalamudimgui` skill.

```csharp
using ECommons.ImGuiMethods;          // ImGuiEx, ImGuiDragDrop, EColor, Box<T>
using Dalamud.Bindings.ImGui;         // ImGui, ImGuiSliderFlags, ImGuiComboFlags, ImU8String ...
using Dalamud.Interface;              // FontAwesomeIcon
using System.Numerics;                // Vector2, Vector4
using Lumina.Excel;                   // ExcelSheet<T>, IExcelRow<T> (ExcelSheetCombo only)
```

Source files (all `public static [unsafe] partial class ImGuiEx` unless noted):

| File | Content |
|---|---|
| `references/ecommons/ECommons/ImGuiMethods/ImGuiEx/Button.cs` | buttons, `ButtonCheckbox`, `EzButtonGroup` |
| `references/ecommons/ECommons/ImGuiMethods/ImGuiEx/Checkbox.cs` | checkbox variants |
| `references/ecommons/ECommons/ImGuiMethods/ImGuiEx/CollectionCheckbox.cs` | collection-backed checkboxes |
| `references/ecommons/ECommons/ImGuiMethods/ImGuiEx/Inputs.cs` | ColorEdit/Drag/Input/Slider wrappers, `Combo<T>`, radio |
| `references/ecommons/ECommons/ImGuiMethods/ImGuiEx/EnumCombo.cs` | the `EnumCombo` overloads |
| `references/ecommons/ECommons/ImGuiMethods/ImGuiEx/ExcelCombos.cs` | `ExcelSheetCombo`, `SearchableCombo` |
| `references/ecommons/ECommons/ImGuiMethods/ImGuiEx/DragDrop.cs` | nested class `RealtimeDragDrop<T>` |
| `references/ecommons/ECommons/ImGuiMethods/ImGuiEx/ObsoleteOverloads.cs` | two obsolete forwarders |
| `references/ecommons/ECommons/ImGuiMethods/ImGuiEx/InternalUtilities.cs` | private `Execute(Action, bool)` |
| `references/ecommons/ECommons/ImGuiMethods/ImGuiDragDrop.cs` | `static class ImGuiDragDrop` |

## Conventions common to the whole area

- Everything must run inside an ImGui draw callback (framework/draw thread). No `Module` flag is
  needed. `ExcelSheetCombo` and `delayedOperation: true` use `Svc`, so `ECommonsMain.Init` must
  have run.
- Labels and IDs are plain `string`, not `ImU8String`. Interpolated strings allocate a managed
  string per frame. The only `ImU8String` parameters in this area are `label`/`format` of
  `SliderInt` and `format` of `DragInt`/`DragFloat`.
- Standard ImGui `##`/`###` ID rules apply to every `label`/`name`/`id` parameter unless a section
  below says otherwise. Several widgets derive internal IDs from the label, and several keep
  per-widget state in static dictionaries keyed by the label string.
- `enabled: false` is implemented with `PushStyleVar(Alpha, ...)` plus discarding the result, not
  with `ImGui.BeginDisabled()`. The item still hovers, still shows tooltips and still animates
  when pressed; it just returns `false` and leaves the value unchanged. Exceptions that do use
  `BeginDisabled`: `ButtonCtrl`, nullable `InputInt`/`InputFloat`, obsolete `ButtonCond`, and
  `CollectionCheckbox` with an empty value set.
- Nothing in this area is cleaned up by `ECommonsMain.Dispose()`; the static search/state
  dictionaries persist for the plugin's lifetime (no reference to them exists outside these
  files).

## Buttons (`Button.cs`)

| Signature | Notes |
|---|---|
| `bool Button(string label, bool enabled = true)` | `ImGui.Button`; when `!enabled` alpha is multiplied by 0.6 and the click is swallowed |
| `bool Button(string label, Vector2 size, bool enabled = true)` | same with explicit size |
| `bool ButtonScaled(string label, Vector2 size, bool enabled = true)` | `size.Scale()` (global scale and font size) then `Button` |
| `bool IconButton(FontAwesomeIcon icon, string id = "ECommonsButton", Vector2 size = default, bool enabled = true)` | pushes `UiBuilder.IconFont`; ImGui label is `$"{icon}##{icon}-{id}"` |
| `bool IconButton(string icon, string id = "ECommonsButton", Vector2 size = default, bool enabled = true)` | `icon` is an already converted icon string |
| `bool IconButtonScaled(FontAwesomeIcon icon, ...)`, `bool IconButtonScaled(string icon, ...)` | same parameters, `size.Scale()` |
| `bool IconButtonWithText(FontAwesomeIcon icon, string id, bool enabled = true, Vector4? defaultColor = null, Vector4? activeColor = null, Vector4? hoveredColor = null, Vector2? size = null, bool scaling = true)` | icon and text centred in the button; see below |
| `bool SmallButton(string label, bool enabled = true)` | `ImGui.SmallButton` with `enabled` |
| `bool SmallIconButton(FontAwesomeIcon icon, string id = "ECommonsButton")`, `bool SmallIconButton(string icon, string id = "ECommonsButton")` | no `enabled` parameter |
| `bool ButtonCopy(string buttonText, string copy)` | replaces `$COPY` in `buttonText` with `copy`; on click calls `GenericHelpers.Copy(copy)` (WinForms clipboard, shows a notification) and returns `true` |
| `void InvisibleButton(int width = 0)` | draws `ImGui.Button(" ")` at alpha 0 as a spacer; `width` is ignored |
| `bool ButtonWrapped(string label)` | calls `SameLine()` itself, then `NewLine()` if the button would not fit |
| `bool ButtonCtrl(string text, string affix = " (Hold CTRL)")` | see below |
| `bool ButtonCtrl(string text, Vector2? size, string affix = " (Hold CTRL)")` | `size == null` means auto size |
| `bool ButtonCtrlScaled(string text, Vector2? size, string affix = " (Hold CTRL)")` | `size.Scale()` |
| `[Obsolete("Use Button")] bool ButtonCond(string name, Func<bool> condition)` | `BeginDisabled` when the condition is false |

Behaviour:

- `IconButton`: the ID is built from the icon and `id`, so the same icon drawn twice in one ID scope
  with the default `id` collides. Pass a distinct `id` or wrap each row in `ImGui.PushID`.
  AutoRetainer relies on `PushID` per row and calls `ImGuiEx.IconButton(FontAwesomeIcon.Trash)`.
- `IconButtonWithText`: `id` is the visible text and the ImGui ID at once. Everything from the
  first `#` onward is removed from the displayed text, and the full string is pushed with
  `ImRaii.PushId(id)` around an empty-label button. `size` components equal to 0 fall back to the
  auto width / `ImGui.GetFrameHeight()`. `size` is multiplied by `ImGuiHelpers.GlobalScale` when
  `scaling` is true. Icon and text are drawn with the window draw list, so they are not clipped to
  the button rectangle.
- `ButtonCtrl`: wrapped in `BeginDisabled()` unless `ImGui.GetIO().KeyCtrl` is down. The affix is
  inserted before a `##` or `###` separator so the ID part is preserved. The text is split with
  `string.Split`, so only the first two segments are used; a label with two separators loses the
  rest.
- `ImGuiEx.Ctrl` (from `ImGuiEx.cs`) is commonly passed as `enabled` for the same effect without
  the affix.

```csharp
if(ImGuiEx.Button("Update", Player.Interactable)) { /* ... */ }
if(ImGuiEx.IconButtonWithText(FontAwesomeIcon.Paste, "Merge with Clipboard", ImGuiEx.Ctrl)) { }
if(ImGuiEx.ButtonCtrl("Delete this plan")) { /* only fires while CTRL is held */ }
if(ImGuiEx.IconButton(FontAwesomeIcon.ArrowUp, $"up{i}") && i > 0) { }
```

### ButtonCheckbox (toggle buttons)

A button that flips a `bool` and is tinted while "on". All return `true` on the frame the value
changed. Button, hovered and active colours are all set to the same colour.

| Signature | Active colour default |
|---|---|
| `bool ButtonCheckbox(string name, ref bool value, bool smallButton)` | `EColor.Red`; `smallButton` is required |
| `bool ButtonCheckbox(string name, ref bool value, uint color, bool smallButton = false)` | `color.ToVector4()` |
| `bool ButtonCheckbox(string name, ref bool value, Vector4 color, bool smallButton = false)` | given |
| `bool ButtonCheckbox(string name, ref bool value, Vector4? color = null, bool inverted = false, Vector2? size = null, bool noColor = false)` | `EColor.Green` |
| `bool ButtonCheckbox(FontAwesomeIcon icon, ref bool value, Vector4? color = null, bool inverted = false, Vector2? size = null)` | `EColor.Green`; pushes icon font |
| `bool ButtonCheckbox(string name, ref bool? value, Vector4? TrueColor = null, Vector4? FalseColor = null, bool smallButton = false)` | green for `true`, red for `false`, untinted for `null` |

- `inverted: true` tints the button when the value is `false` (use for "Exclude..." style fields).
  The click still simply negates the value.
- Which overload a two-argument call `ButtonCheckbox(name, ref b)` binds to is decided by the
  compiler among the optional-parameter overloads; by the parameter lists above only the
  `Vector4? color = null` overload is applicable, giving green. Use named arguments
  (`color:`, `inverted:`) as AutoRetainer does to make the choice explicit.
- The icon overload uses the bare icon string as label: two identical icons in one ID scope
  collide, use `ImGui.PushID`.
- `bool?` overload: left click sets `true` unless already `true`, in which case `false`; right
  click sets `false` unless already `false`, in which case `true`. Despite the XML comment, the
  code never writes `null` back.

```csharp
ImGuiEx.ButtonCheckbox(FontAwesomeIcon.GasPump, ref chr.AutoFuelPurchase, color: ImGuiColors.TankBlue);
if(ImGuiEx.ButtonCheckbox(FontAwesomeIcon.Users, ref chr.ExcludeRetainer, inverted: true)) Save();
```

### EzButtonGroup

`public class ImGuiEx.EzButtonGroup`, a horizontal row of buttons, optionally centred.

| Member | Notes |
|---|---|
| `bool IsCentered { get; set; }` | centres the group in `GetContentRegionAvail().X` |
| `float Height { get; set; }` | initialised in the constructor from `ImGui.GetTextLineHeight()` and frame padding, so construct it inside a draw call |
| `float Width { get; private set; }` | computed in `Draw()` |
| `void Add(string label, Action action, ButtonStyle? style = null)` | plus overload with `string tooltip` before `style` |
| `void AddIconWithText(FontAwesomeIcon icon, string label, Action action, ButtonStyle? style = null)` | plus `tooltip` overload; drawn with `ImGuiComponents.IconButtonWithText` |
| `void AddIconOnly(FontAwesomeIcon icon, Action action, ButtonStyle? style = null)` | plus `tooltip` overload; drawn with `ImGuiEx.IconButton(icon)` |
| `void Draw()` | draws text buttons, then icon+text, then icon-only, regardless of insertion order |

`EzButtonGroup.ButtonStyle`: `bool? NoButtonBg`, `Vector4? ButtonColor`, `Vector4? TextColor`,
`string? Tooltip`, `Func<bool>? Condition` (button skipped when it returns false), and
`operator +` which merges two styles with the right operand winning per property.

`Draw()` adds to `Width` without resetting it, so an instance drawn on more than one frame
reports a growing `Width` and drifts when `IsCentered` is set. Create a new instance every frame.

## Checkboxes (`Checkbox.cs`)

| Signature | Behaviour |
|---|---|
| `bool Checkbox(string label, ref bool value, bool enabled = true)` | stock checkbox; when `!enabled` dimmed and any toggle is reverted |
| `bool Checkbox(string label, ref int value)` | `int` as bool: non-zero is checked, writes `1`/`0` |
| `bool Checkbox<T>(string id, ref T? value, T? defaultValue = null) where T : struct` | checked when `value != null`; checking assigns `defaultValue ?? default(T)`, unchecking assigns `null` |
| `bool Checkbox(string label, ref bool? value)` | tri-state, forwards to the next with `enabled: true, inverted: false` |
| `bool Checkbox(string label, ref bool? value, bool enabled, bool inverted)` | tri-state; `null` is drawn as a bullet |
| `bool Checkbox(FontAwesomeIcon icon, string label, ref bool value, bool enabled = true)` | icon drawn inside the box instead of a check mark |
| `bool Checkbox(FontAwesomeIcon icon, Vector4? selectedColor, string label, ref bool value, bool enabled = true)` | `selectedColor` used for both icon and text when checked |
| `bool Checkbox(FontAwesomeIcon icon, Vector4? selectedCheckboxIconColor, Vector4? unselectedCheckboxIconColor, Vector4? selectedText, Vector4? unselectedText, string label, ref bool value, bool enabled = true)` | full form |
| `bool CheckboxBullet(string label, ref bool value)` | shows a bullet (mixed-state mark) instead of a check mark when `true`; implemented with `ImGui.CheckboxFlags(label, ref flags, int.MaxValue)` |
| `bool CheckboxInverted(string label, ref bool value)` | checked when `value == false` |
| `bool CheckboxWrapped(string label, ref bool v)` | calls `SameLine()` itself and wraps to a new line if it would not fit |

All return `true` only on the frame the value changed.

- Tri-state cycle with `inverted: false`: `true` -> `false` -> `null` -> `true`. With
  `inverted: true` the displayed check is negated and the cycle is `false` -> `true` -> `null` ->
  `false`.
- Icon checkbox defaults: checked icon and text `ImGuiColors.ParsedGreen`, unchecked icon
  `ImGuiColors.DalamudGrey3`, unchecked text the current `ImGuiCol.Text`. It draws the stock
  checkbox with a dummy `false` value, overlays the icon via the window draw list, and ends with
  `ImGui.SameLine(0, 0)` followed by `ImGui.NewLine()`.
- `Checkbox(label, ref bool?)` binds to the non-generic tri-state overload, not to
  `Checkbox<T>`. For `int?`, `float?`, enum? etc. the generic overload is used.

```csharp
ImGuiEx.Checkbox("Enabled", ref data.TeleportOptionsOverride.Enabled);
ImGuiEx.Checkbox("Unique", ref intField);                 // int 0/1
ImGuiEx.Checkbox("Override delay", ref cfg.DelayOverride, 500);   // int? -> null or 500
ImGuiEx.Checkbox("Use buff", ref cfg.UseBuff);            // bool? tri-state
```

## Collection checkboxes (`CollectionCheckbox.cs`)

These take the collection itself (no `ref`) and add/remove elements. `T` needs sensible equality.

| Signature | Behaviour |
|---|---|
| `bool CollectionCheckbox<T>(string label, T value, ICollection<T> collection, bool inverted = false, bool delayedOperation = false)` | checked when `collection.Contains(value)` |
| `bool CollectionCheckbox<T>(string label, IEnumerable<T> values, ICollection<T> collection, bool inverted = false, bool delayedOperation = false)` | group checkbox over a set of values |
| `delegate bool CollectionCoreDelegate(ref bool contains)` | drawing callback for `CollectionCore` |
| `bool CollectionCore<T>(CollectionCoreDelegate draw, T value, ICollection<T> collection, bool inverted = false, bool delayedOperation = false)` | same logic with a custom widget |
| `bool CollectionButtonCheckbox<T>(string name, T value, ICollection<T> collection, bool smallButton = false, bool inverted = false)` | toggle button, colour `EColor.Red` |
| `bool CollectionButtonCheckbox<T>(string name, T value, ICollection<T> collection, Vector4 color, bool smallButton = false, bool inverted = false)` | explicit colour |
| `bool CollectionButtonCheckbox<T>(FontAwesomeIcon icon, T value, ICollection<T> collection, bool smallButton = false, bool inverted = false)` | pushes icon font; plus overload with `Vector4 color` after `collection` |
| `[Obsolete("Please switch to CollectionCheckbox")] bool HashSetCheckbox<T>(string label, T value, HashSet<T> collection)` | forwards to `CollectionCheckbox` |

- Single value: removal loops `while(collection.Contains(value))`, so every duplicate in a `List<T>`
  is removed. `inverted` shows a check when the value is absent.
- Set of values: check mark when all are present, bullet when some are present, empty when none.
  Clicking anything other than the fully checked state adds all (each value is removed first,
  then added, so lists get no duplicates); clicking the fully checked state removes all. With
  `inverted: true` the check mark means none present. The XML comment says `inverted` is not
  implemented for this overload, but the code implements it. An empty `values` draws a disabled
  unchecked box and returns `false`.
- `delayedOperation: true` wraps the mutation in `new TickScheduler(action)`, which runs it on the
  next `Svc.Framework.Update`. Use it when the checkbox is drawn while enumerating the same
  collection. The return value is still `true` on the click frame, before the collection changes.
  The `CollectionButtonCheckbox` overloads have no delayed mode and mutate immediately.

```csharp
// group header + children, AutoRetainer pattern; key type is a tuple
ImGuiEx.CollectionCheckbox(chara.Name, chara.RetainerData.Select(r => (chara.CID, r.Name)), Selected);
foreach(var r in chara.RetainerData)
    ImGuiEx.CollectionCheckbox(r.Name, (chara.CID, r.Name), Selected);

Modified |= ImGuiEx.CollectionCheckbox($"{x.Name}##{x.RowId}", x.RowId, SelectedCategories);
ImGuiEx.CollectionCheckbox(label, key, plan.ExcludedRoutes, true);   // inverted
```

## Value editors (`Inputs.cs`)

### ColorEdit4, DragInt, DragFloat (width-first, `enabled`, nullable forms)

Width is the first parameter. For the non-nullable `Drag*` and the nullable forms, `width == 0`
skips `SetNextItemWidth`; non-nullable `ColorEdit4` always calls `SetNextItemWidth(width)`.

```csharp
bool ColorEdit4(float width, string id, ref Vector4 valueRef,
    ImGuiColorEditFlags flags = ImGuiColorEditFlags.NoInputs, bool enabled = true)
bool ColorEdit4(float width, string id, ref uint valueRef,
    ImGuiColorEditFlags flags = ImGuiColorEditFlags.NoInputs, bool enabled = true)
bool ColorEdit4(float width, string id, ref Vector4? nullableValueRef,
    ImGuiColorEditFlags flags = ImGuiColorEditFlags.NoInputs, bool enabled = true,
    Vector4 defaultValue = default, bool isLabelPrefix = false, bool showCheckbox = true)
bool ColorEdit4(float width, string id, ref uint? nullableValueRef,
    ImGuiColorEditFlags flags = ImGuiColorEditFlags.NoInputs, bool enabled = true,
    uint defaultValue = default, bool isLabelPrefix = false, bool showCheckbox = true)
// 3.2.1.23+: no width, flags default None; stock ImGui.ColorEdit4 for a packed uint colour
bool ColorEdit4(ImU8String label, scoped ref uint col,
    ImGuiColorEditFlags flags = ImGuiColorEditFlags.None)

bool DragInt(float width, string id, ref int valueRef, float vSpeed = 1.0f, int vMin = 0,
    int vMax = 0, ImU8String format = default, ImGuiSliderFlags flags = ImGuiSliderFlags.None,
    bool enabled = true)
bool DragInt(float width, string id, ref int? nullableValueRef, float vSpeed = 1.0f, int vMin = 0,
    int vMax = 0, ImU8String format = default, ImGuiSliderFlags flags = ImGuiSliderFlags.None,
    bool enabled = true, int defaultValue = default, bool isLabelPrefix = false,
    bool showCheckbox = true)

bool DragFloat(float width, string id, ref float valueRef, float vSpeed = 1.0f, float vMin = 0.0f,
    float vMax = 0.0f, ImU8String format = default,
    ImGuiSliderFlags flags = ImGuiSliderFlags.None, bool enabled = true)
bool DragFloat(float width, string id, ref float? nullableValueRef, float vSpeed = 1.0f,
    float vMin = 0.0f, float vMax = 0.0f, ImU8String format = default,
    ImGuiSliderFlags flags = ImGuiSliderFlags.None, bool enabled = true,
    float defaultValue = default, bool isLabelPrefix = false, bool showCheckbox = true)
```

- The `uint` colour forms convert with the ECommons `ToVector4()`/`ToUint()` extensions.
- Non-nullable forms edit a local copy and only write back when `enabled`, so a disabled widget
  never changes the value.
- Nullable forms render `[checkbox] [editor] label`. The checkbox (ID `##{id}_checkbox`) switches
  between `null` and `defaultValue`; the editor is inert while the value is `null` and shows
  `defaultValue`. `showCheckbox: false` hides the checkbox, leaving the editor inert whenever
  the value is `null`.
- In nullable forms `id` is rendered as text with `ImGuiEx.TextV` (after the editor, or before it
  with `isLabelPrefix: true`) unless it starts with `##`. The inner editor's ID is
  `"##coloredit4_" + id` or `"##dragfloat_" + id` (the latter for both `DragInt` and `DragFloat`),
  so an `int?` and a `float?` editor with the same `id` in one scope collide.
- Returns `true` when either the checkbox or the editor changed the value.

### InputInt / InputFloat

```csharp
bool InputInt(float width, string label, ref int? valueNullable, int step = 1,
    int step_fast = 100, ImGuiInputTextFlags flags = ImGuiInputTextFlags.None)
bool InputFloat(float width, string label, ref float? valueNullable, int step = 1,
    int step_fast = 100, ImGuiInputTextFlags flags = ImGuiInputTextFlags.None)
bool InputFloat(float width, string label, ref float value, int step = 1,      // 3.2.1.23+
    int step_fast = 100, ImGuiInputTextFlags flags = ImGuiInputTextFlags.None)
```

The nullable forms draw a checkbox followed by the stock input inside `BeginDisabled()` when the
value is `null`. Enabling always assigns `0` (there is no default-value parameter).
`step`/`step_fast` are `int` even for the float forms. `SetNextItemWidth(width)` is always called.
The non-nullable `InputFloat` is `SetNextItemWidth(width)` plus stock `ImGui.InputFloat`. There is
no non-nullable `ImGuiEx.InputInt`; use stock `ImGui` for that.

### InputFancyNumeric

```csharp
bool InputFancyNumeric(string label, ref int number, int step, Action? afterInput = null)
bool InputFancyNumeric(float width, string label, ref int number, int step, Action? afterInput = null)
```

Text field showing the number formatted `N0` (thousands separators), with optional `-`/`+` buttons
when `step > 0`. Typed input accepts a leading `-`, decimals, thousands separators and any number
of trailing `K` (x1000) / `M` (x1000000) suffixes, case-insensitive, e.g. `1.5m`. Empty input
becomes 0; unparsable input leaves the number unchanged. Holding a button for more than 0.5 s
repeats every 50 ms through `EzThrottler.Throttle("FancyInputHold", 50)` (one throttle name shared
by every instance). `afterInput` is invoked every frame right after the text field, before the
buttons, which is the place for a tooltip on the field. The visible label is drawn separately with
`ImGuiEx.Text` unless `label` starts with `##`. In the width overload `width` is the total
including the two buttons. Returns `true` on text edit or button press. No clamping is applied.

### Sliders

```csharp
bool SliderInt(ImU8String label, scoped ref int v, int vMin = 0, int vMax = 0,
    ImU8String format = default, ImGuiSliderFlags flags = ImGuiSliderFlags.None)
bool SliderFloat(string label, ref float v, float v_min, float v_max)
bool SliderFloat(string label, ref float v, float v_min, float v_max, string format)
bool SliderFloat(string label, ref float v, float v_min, float v_max, string format, ImGuiSliderFlags flags)
bool SliderIntAsFloat(string id, ref int value, int min, int max, float divider = 1000)
```

`SliderInt` and `SliderFloat` are the stock sliders followed by `ImGuiEx.ActivateIfDoubleClicked()`
(double-click switches the slider to keyboard entry; stock ImGui needs Ctrl+click). That helper is
public and can follow any stock slider. `SliderIntAsFloat` stores an `int` and shows
`value / divider` (milliseconds shown as seconds by default); it truncates on write-back. From
3.2.1.23 it goes through `ImGuiEx.SliderFloat`, so it has the double-click behaviour too; older
packages call the stock slider. None of these take a width; call `ImGui.SetNextItemWidth` first.

```csharp
ImGui.SetNextItemWidth(100f);
ImGuiEx.SliderInt("Index override", ref d.IndexOverride, 0, 4,
    d.IndexOverride == 0 ? "Disabled" : $"{d.IndexOverride}");
ImGuiEx.SliderIntAsFloat("Activation Time", ref C.RetainerSenseThreshold, 1000, 100000);
```

### Text-backed numeric inputs

| Signature | Notes |
|---|---|
| `bool InputLong(string id, ref long num)` | `InputText` + `long.TryParse`; parses every frame, so unparsable text sets `num` to 0 |
| `void InputUint(string name, ref uint uInt)` | max 16 chars; value kept when parsing fails |
| `void InputHex(string name, ref uint hexInt)` | shows `X` format, strips `0x` |
| `void InputHex(string name, ref long hexInt)` | same for `long` |
| `void InputHex(string name, ref byte hexByte)` | buffer length 2, no `0x` handling |

`InputUint` and `InputHex` return `void`; compare the value before and after to detect changes.

### Multiline text

```csharp
bool InputTextMultilineExpanding(string id, ref string text, int maxLength = 500,
    int minLines = 2, int maxLines = 10, int? width = null)
bool InputTextWrapMultilineExpanding(string id, ref string text, int maxLength = 500,
    int minLines = 2, int maxLines = 10, int? width = null)
```

Height follows the number of `\n` separated lines plus one, clamped to `minLines..maxLines`.
`width == null` uses the available content width. The `Wrap` variant additionally installs a
`CallbackEdit` callback (private `TextEditCallback` in `ImGuiEx.cs`) that wraps at the width.

### List editors

```csharp
void InputListString(string name, List<string> list,
    Dictionary<string, string> overrideValues = null, string defaultValue = null)
void InputListUint(string name, List<uint> list, Dictionary<uint, string> overrideValues = null)
void InputList<T>(string name, List<T> list, Dictionary<T, string> overrideValues, Action addFunction)
```

A combo whose preview is `- No values -`, the single element, or `- N elements -`. Inside: an
input plus `Add` button, then one selectable per element; right-click an element for `Delete` and
`Clear (hold shift+ctrl)`. `overrideValues` maps elements to display names. The pending input is
kept in a private static dictionary keyed by `name`. `defaultValue` of `InputListString` is unused.
`InputList<T>` is the generic core; `addFunction` draws the add row.

## Combos and radio buttons (`EnumCombo.cs`, `Inputs.cs`)

The `EnumCombo` overloads and `EnumComboSearch` live in `EnumCombo.cs` from 3.2.1.23, which also
added the trailing `T[]? valuesList = null` parameter to each. Older packages declare them in
`Inputs.cs` without it.

```csharp
bool EnumCombo<T>(string name, ref T refConfigField, Func<T, bool> filter = null,
    IDictionary<T, string> names = null, T[]? valuesList = null) where T : Enum, IConvertible
bool EnumCombo<T>(string name, ref T refConfigField, IDictionary<T, string> names,
    T[]? valuesList = null) where T : Enum, IConvertible
bool EnumCombo<T>(float width, string name, ref T refConfigField, Func<T, bool> filter = null,
    IDictionary<T, string> names = null, T[]? valuesList = null) where T : Enum, IConvertible
bool EnumCombo<T>(string name, ref Nullable<T> refConfigField, Func<T, bool> filter = null,
    IDictionary<T, string> names = null, string nullName = "Not selected",
    T[]? valuesList = null) where T : struct, Enum, IConvertible
bool EnumCombo<T>(string name, ICollection<T> collection, Func<T, bool> filter = null,
    IDictionary<T, string> names = null, string nameOfNoneSelected = "Not selected",
    T[]? valuesList = null) where T : Enum, IConvertible

bool Combo<T>(string name, ref T refConfigField, IEnumerable<T> values,
    Func<T, bool> filter = null, IDictionary<T, string> names = null)

bool EnumRadio<T>(ref T refConfigField, bool sameLine = false, Func<T, bool> filter = null,
    IDictionary<T, string> names = null) where T : Enum, IConvertible

void RadioButtonBool(string labelTrue, string labelFalse, ref bool value, bool sameLine = false,
    Action prefix = null, Action suffix = null, bool inverted = false)
void RadioButtonBool(string? text, string labelTrue, string? helpTrue, string labelFalse,
    string? helpFalse, ref bool value, bool sameLine = false, Action prefix = null,
    Action suffix = null, bool inverted = false)

public static Dictionary<string, Box<string>> EnumComboSearch;   // filter text per combo name
public static Dictionary<string, Box<string>> ComboSearch;       // same for Combo<T>
```

- `valuesList` replaces `Enum.GetValues(typeof(T))` as the set and order of entries; `filter` still
  applies on top. Pass it to reorder or subset an enum without a predicate.
- Display name: `names[value]` when present, otherwise `value.ToString()` with `_` replaced by a
  space (`EnumCombo`, `EnumRadio`) or plain `ToString()` (`Combo<T>`).
- All combos use `ImGuiComboFlags.HeightLarge`. When there are more than 10 values a
  `Filter...` text box appears at the top; its text persists in the static dictionary under the
  combo's `name`, so two combos with the same `name` share filter text.
- `filter` hides values for which it returns false. The currently selected value is scrolled into
  view when the popup appears.
- Return value is `true` on the frame a selection was made (collection overload: an element was
  added or removed).
- The width overload exists only for the plain `ref T` form; otherwise call
  `ImGui.SetNextItemWidth` first.
- Nullable overload adds a `nullName` entry at the top that assigns `null`.
- Collection overload draws checkboxes instead of selectables, previews the selected names joined
  by the ECommons `Print()` extension, and lists values present in the collection but not defined
  in the enum in red so they can be removed.
- `Combo<T>` calls `refConfigField.ToString()` for the preview, so a `null` current value without
  a `names` entry throws. `values.Count()` is evaluated each frame the popup is open; pass a
  materialised collection.
- `EnumRadio` uses the display name as the radio label with no ID suffix: wrap in `ImGui.PushID`
  when the same enum is drawn twice in one scope.
- `RadioButtonBool`: two radio buttons for one `bool`; `inverted` swaps their order. `prefix` and
  `suffix` run before/after each button. `text` is a heading; `helpTrue`/`helpFalse` add
  `ImGuiEx.HelpMarker`.

```csharp
ImGui.SetNextItemWidth(150f);
ImGuiEx.EnumCombo("##behavior", ref MassBehavior);
ImGuiEx.EnumCombo("##unlockmode", ref MassUnlockMode, Lang.UnlockModeNames);   // IDictionary names
ImGuiEx.Combo("##sacc", ref selected, accounts,
    names: accounts.ToDictionary(x => x, x => x == -1 ? "All service accounts" : $"Service account {x + 1}"));
ImGuiEx.RadioButtonBool("Desynthese", "Skip", ref plan.EnableDesynth, true);
```

## Sheet-backed and searchable combos (`ExcelCombos.cs`)

Unlike the combos above these return the pick through `out`, do not take the current value by
`ref`, and need the preview text supplied by the caller.

```csharp
// core forms
bool ExcelSheetCombo<T>(string id, [NotNullWhen(true)] out T selected,
    Func<ExcelSheet<T>, string> getPreview, Func<T, string> toName,
    Func<T, string, bool> searchPredicate, Func<T, bool> preFilter,
    ImGuiComboFlags flags = ImGuiComboFlags.None) where T : struct, IExcelRow<T>

bool SearchableCombo<T>(string id, [NotNullWhen(true)] out T? selected, string preview,
    IEnumerable<T> possibilities, Func<T, string> toName,
    Func<T, string, bool> searchPredicate, Func<T, bool> preFilter,
    ImGuiComboFlags flags = ImGuiComboFlags.None) where T : notnull
```

`ExcelSheetCombo<T>` has eight overloads: `(id, out selected, getPreview, ...)` followed by any of
these tails, each ending with `ImGuiComboFlags flags = ImGuiComboFlags.None`:

| Tail after `getPreview` | Defaults applied |
|---|---|
| (nothing) | `toName = t.ToString()`, name-contains search, no pre-filter |
| `Func<T, string, bool> searchPredicate` | `toName = t.ToString()` |
| `Func<T, bool> preFilter` | `toName = t.ToString()` |
| `Func<T, string, bool> searchPredicate, Func<T, bool> preFilter` | `toName = t.ToString()` |
| `Func<T, string> toName` | search = `toName(t).Contains(s, CurrentCultureIgnoreCase)` |
| `Func<T, string> toName, Func<T, string, bool> searchPredicate` | pre-filter `_ => true` |
| `Func<T, string> toName, Func<T, bool> preFilter` | default search |
| `Func<T, string> toName, Func<T, string, bool> searchPredicate, Func<T, bool> preFilter` | none |

`SearchableCombo<T>` has three: `toName` only (search uses `InvariantCultureIgnoreCase`), plus
`searchPredicate`, plus `searchPredicate` and `preFilter`.

Behaviour:

- `ExcelSheetCombo` fetches `Svc.Data.GetExcelSheet<T>()` (client language), passes the sheet to
  `getPreview` to produce the preview string, and forwards to `SearchableCombo`. A null sheet
  returns `false`.
- The overloads without `toName` list rows by `T.ToString()`, which for Lumina row structs is not
  a display name. Pass `toName`.
- The popup has a search box that takes keyboard focus on open; Up/Down move the highlight, Enter
  picks the highlighted row, Escape closes. Closing after Enter/Escape is done by toggling a
  `##x` suffix on the combo ID on the next frame.
- Returns `true` with `selected` set only on the pick frame; otherwise `selected` is `default`.
- Search text, filtered set and highlight index are single private statics shared by every
  searchable combo. The filtered set is rebuilt only when the search text changes or the popup
  reopens, so changes to `possibilities` or `preFilter` while the popup is open are not seen.
- `preFilter` and `searchPredicate` run over the whole sheet on each rebuild; with an empty search
  every row passing `preFilter` is listed without clipping, so pre-filter large sheets.

```csharp
if(ImGuiEx.ExcelSheetCombo<Lumina.Excel.Sheets.Item>("##item", out var row,
    sheet => sheet.GetRowOrDefault(cfg.ItemId)?.Name.ToString() ?? "None",
    x => x.Name.ToString(),
    x => x.RowId != 0))
{
    cfg.ItemId = row.RowId;
}

if(ImGuiEx.SearchableCombo("##profile", out var picked, current?.Name ?? "None",
    profiles, p => p.Name))
{
    current = picked;
}
```

No consumer in the three reference plugins calls these; the examples are assembled from the
signatures only.

## Drag and drop

### `ImGuiDragDrop` (payload helpers)

`public static class ECommons.ImGuiMethods.ImGuiDragDrop`:

```csharp
static unsafe void SetDragDropPayload<T>(string type, T data, ImGuiCond cond = 0) where T : struct
static unsafe bool AcceptDragDropPayload<T>(string type, out T payload,
    ImGuiDragDropFlags flags = ImGuiDragDropFlags.None) where T : struct
static unsafe void SetDragDropPayload(string type, string data, ImGuiCond cond = 0)
static unsafe bool AcceptDragDropPayload(string type, out string payload,
    ImGuiDragDropFlags flags = ImGuiDragDropFlags.None)
```

- Call `Set...` between `ImGui.BeginDragDropSource` and `EndDragDropSource`, and `Accept...`
  between `BeginDragDropTarget` and `EndDragDropTarget`; the helpers do not open those scopes.
- Struct form copies the raw bytes of `data` via `MemoryMarshal.AsBytes`, so `T` must not contain
  managed references. Reading uses `Unsafe.Read<T>` on the payload pointer.
- String form encodes UTF-8 into a `stackalloc` buffer sized to the string (keep payloads short)
  and decodes with `Encoding.Default`. On failure `payload` is `null` (`default` for structs).
- `type` is the ImGui payload type name; ImGui limits it to 32 characters.

### `ImGuiEx.RealtimeDragDrop<T>` (reorder rows by dragging)

Nested class in `ImGuiEx`. Draws a move handle per row and reorders an `IList<T>` while the drag
is in progress. Rows are identified by a caller-supplied unique string.

| Member | Notes |
|---|---|
| `RealtimeDragDrop(string dragDropId, Func<T, string> getUniqueId, bool smallButton = false)` | `dragDropId` is the payload type and part of button IDs |
| `void Begin()` | step 1, before the table begins; clears queued commands |
| `void NextRow()` | step 2, at the start of the row's first column; stores the cursor |
| `void DrawButtonDummy(T item, IList<T> list, int targetPosition)` | step 3, where the handle should be; reserves space with a `Dummy` |
| `void DrawButtonDummy(string uniqueId, IList<T> list, int targetPosition)` | same by id |
| `void DrawButtonDummy(string uniqueId, Action<string> onAcceptDragDropPayload)` | custom move; callback receives the dragged row's id |
| `bool SetRowColor(string uniqueId, bool setColor = true)`, `bool SetRowColor(T element, bool setColor = true)` | optional, after `TableNextRow`; highlights the dragged row, returns whether this row is being dragged |
| `uint HighlightColor { get; }` | pulsing green from `GradientColor.Get` |
| `bool AcceptPayload([NotNullWhen(true)] out string? uniqueId, ImGuiDragDropFlags flags = ImGuiDragDropFlags.None)` | makes the last item an extra drop target, e.g. to drop a row onto a folder |
| `void End(int numRows = 1)` | step 4, after the table ends; draws the real handles and row-wide drop targets |

- The handle buttons and drop targets are drawn in `End()`, by moving the cursor back to the
  stored positions. The list is therefore mutated in `End()`, outside the caller's row loop.
- `targetPosition` is the index of the row being drawn. The list overloads call
  `GenericHelpers.MoveItemToPosition(list, x => getUniqueId(x) == payload, targetPosition)`.
- Payloads are accepted with `AcceptBeforeDelivery`, so the row moves as soon as the cursor is
  over another row, not on release.
- `numRows` is the number of text lines in the tallest row; it sizes the row-wide drop target.
- Keep one instance per list for the lifetime of the UI (field or `Ref<T>`), not per frame: the
  instance tracks the current drag.
- Logs `DragDropSource = ...` through `InternalLog.Verbose` every frame of a drag.

```csharp
private ImGuiEx.RealtimeDragDrop<GCExchangeItem> DragDrop = new("GCELDD", x => x.ID);

DragDrop.Begin();
if(ImGui.BeginTable("##list", 3))
{
    for(var i = 0; i < plan.Items.Count; i++)
    {
        var item = plan.Items[i];
        ImGui.PushID(item.ID);
        ImGui.TableNextRow();
        DragDrop.SetRowColor(item);
        ImGui.TableNextColumn();
        DragDrop.NextRow();
        DragDrop.DrawButtonDummy(item, plan.Items, i);
        ImGui.TableNextColumn();
        ImGuiEx.TextV(item.ID);
        ImGui.PopID();
    }
    ImGui.EndTable();
}
DragDrop.End();
```

## Gotchas

- `Checkbox(string, ref bool?, bool enabled, bool inverted)` with `enabled: false` pushes
  `ImGuiStyleVar.Alpha` and never pops it, and does not block the toggle. Do not pass
  `enabled: false` to the tri-state checkbox at this commit; wrap it in `ImRaii.Disabled`
  instead.
- `enabled: false` elsewhere is visual dimming plus a swallowed result, not a disabled item;
  tooltips attached afterwards still show, which is usually what the consumers want.
- `IconButton`/`SmallIconButton` default `id` is the constant `"ECommonsButton"`; identical icons
  need distinct ids or a `PushID` scope. Same for the icon overloads of `ButtonCheckbox` and
  `CollectionButtonCheckbox`.
- `IconButtonWithText` truncates the visible text at the first `#`, including a literal `#` that
  was meant to be displayed.
- `InvisibleButton(int width)` ignores `width`.
- `EzButtonGroup.Width` accumulates across `Draw()` calls; do not cache the instance.
- `InputLong` resets the value to 0 while the text is unparsable (including while empty).
- `InputFloat(float width, string label, ref float? ...)` takes `int` steps.
- `DragInt` and `DragFloat` nullable forms share the inner ID prefix `##dragfloat_`.
- Combo filter text lives in public static dictionaries keyed by the combo label; same-label
  combos in different windows share it, and entries are never removed.
- `SearchableCombo`/`ExcelSheetCombo` share one static search state across all instances and
  return the selection through `out`; assigning the config field is the caller's job.
- `RealtimeDragDrop` list overloads throw from `MoveItemToPosition` (index -1) when the dragged
  id is not in the target list, e.g. when the same `dragDropId` is used for two different lists
  or the list is filtered. Use a distinct `dragDropId` per list.
- `CollectionCheckbox` returns `true` before the collection changes when `delayedOperation` is
  set; do not save configuration by reading the collection on that same frame.
- PandorasBox and YesAlready barely call this area through `ImGuiEx.`; the realistic examples above
  come from AutoRetainer and were checked against the pinned ECommons signatures.
