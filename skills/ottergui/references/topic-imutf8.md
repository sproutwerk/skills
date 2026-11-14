OtterGui topic: `OtterGui.Text.ImUtf8`, the UTF-8 / interpolated-string ImGui wrapper (every
partial file, the `EndObjects`, `HelperObjects` and `Extended` folders, buffer sizes, exceptions).
Verified against `Limiana/OtterGui` `a08fda6` (2026-05-30, branch `main`); `Text/` is identical to
upstream `79771ee` except for comments and two unused locals in `ImUtf8.RotatedText.cs`.

Paths such as `references/...` and `references/topic-*.md` are relative to the skill root (the directory that
holds `SKILL.md`), not to this file. Signatures are copied from
`references/api-index.txt`; nothing was compiled.

## Shape of the API

```csharp
using OtterGui.Text;                 // ImUtf8, TerminatedByteString
using OtterGui.Text.EndObjects;      // Child, Table, TabItem, Id, ListClipper, ... (return types)
using OtterGui.Text.HelperObjects;   // Utf8StringHandler<T>, buffers, exceptions
```

`ImUtf8` is one `public static partial class` spread over `Text/ImUtf8.<Widget>.cs` (core ImGui
calls) and `Text/Extended/ImUtf8.<Name>.cs` (composite helpers). Nearly every function exists in
three overloads that differ only in how the string is passed:

| Overload parameter | What you pass | Cost |
|---|---|---|
| `ReadOnlySpan<byte>` | `"Label"u8` or your own UTF-8 bytes | none, forwarded as is |
| `ReadOnlySpan<char>` | a `string` or char span | transcoded into a static buffer |
| `ref Utf8StringHandler<TBuffer>` | an interpolated string `$"Item {i}"` | formatted straight to UTF-8 in a static buffer, no `string` allocated |

```csharp
ImUtf8.Text("Static text"u8);
ImUtf8.Text(someString);
ImUtf8.Text($"{count} items, {size:F1} MiB");
if (ImUtf8.Button("Save"u8, new Vector2(120, 0))) Save();
```

The library itself calls the handler overloads with interpolated strings, for example
`ImUtf8.CopyOnClickSelectable($"0x{address:X}", UiBuilder.MonoFont)` in `Services/DynamisIpc.cs`.

Null termination, from the XML comments: parameters documented as label, id, format or hint
"HAS to be null-terminated" in the `ReadOnlySpan<byte>` overload; text parameters ("Does not have
to be null-terminated") do not. A `"..."u8` literal is fine for both. The `ReadOnlySpan<char>` and
handler overloads terminate the buffer for you.

## Static buffers (`Text/HelperObjects`)

| Buffer type | `Size` | Used for |
|---|---|---|
| `LabelStringHandlerBuffer` | `128 * 1024` | labels, ids |
| `HintStringHandlerBuffer` | `128 * 1024` | hints, formats, previews, second string of a call |
| `TextStringHandlerBuffer` | `4 * 1024 * 1024` | text, tooltips |
| `InputStringHandlerBuffer` | `8 * 1024 * 1024` | the `...OnDeactivated` inputs; also has `LastId`, `IsActive`, `Update`, `Return` |

Each is a `struct` implementing `IStringHandlerBuffer` (`static abstract` members `Size`, `Buffer`,
`Span`, `Write`) whose static constructor allocates the block once with `Marshal.AllocHGlobal`.
Consequences that follow directly from the code:

- One buffer per type per assembly copy, shared by every call. A function that takes two
  strings uses two different buffer types (label + hint) so they do not overwrite each other.
  Two values of the *same* buffer type cannot be alive at once: finish one call before formatting
  the next.
- No locking anywhere. Use `ImUtf8` on the draw thread only.
- Overflow or a failed format does not truncate silently: the internal `Span<T>()` / `End<T>()`
  helpers throw `ImUtf8FormatException` ("Could not format UTF8 String."). `ImUtf8SizeException`
  ("Input data is longer than buffer size.") is the second exception type in
  `HelperObjects/ImUtf8Exception.cs`.

`Utf8StringHandler<T>` is an `[InterpolatedStringHandler] ref struct` wrapping
`System.Text.Unicode.Utf8.TryWriteInterpolatedStringHandler`; public members are the two
constructors `(int literalLength, int formattedCount, out bool shouldAppend)` and
`(..., IFormatProvider? provider, out bool shouldAppend)`, `AppendLiteral`, the `AppendFormatted`
overloads, `GetEnd(out byte* end)` and `GetSpan(out ReadOnlySpan<byte> span)`.

`HoverUtf8StringHandler` is the handler used by `ImUtf8.HoverTooltip`: its constructor calls
`ImGui.IsItemHovered(flags)` first and sets `shouldAppend = false` when the item is not hovered,
so the interpolation holes are **not evaluated** unless the tooltip will be shown.

Other helper types: `TerminatedByteString` (namespace `OtterGui.Text`, file
`Text/Extended/TerminatedByteString.cs`; owned null-terminated UTF-8, members `Empty`,
`TextWithNull`, `Length`, `Count`, `IsEmpty`, implicit conversion to `ReadOnlySpan<byte>`,
`ToString()`), `ImTextExtensions.CloneNullTerminated(this ReadOnlySpan<byte> / Span<byte>)`,
`SpanTextWriter` (a `ref struct` writer over a `Span<byte>` with `TryAppend` / `Append` overloads,
`GetWrittenSpan()`, `EnsureNullTerminated()`), `DataCache<T>` (internal).

`ImUtf8.GetBytes(ReadOnlySpan<char>)` and `ImUtf8.GetBytes(ref Utf8StringHandler<TextStringHandlerBuffer>)`
return a `TerminatedByteString` when you need to keep a converted string.

## EndObjects (`Text/EndObjects`, namespace `OtterGui.Text.EndObjects`)

`ref struct`s returned by the scope-opening functions. Use them with `using`; the ones with a
result expose `public readonly bool Success`, `public bool Disposed`, an implicit conversion to
`bool` and the operators `true`, `false`, `!`, `&`, `|`.

| Type | Opened by | `Dispose` calls the end function |
|---|---|---|
| `Child` | `ImUtf8.Child`, `ImUtf8.ResizableChild` | always (`EndChild`) |
| `Group` | `ImUtf8.Group()` | always |
| `Tooltip` | `ImUtf8.Tooltip()` | always |
| `CollapsingHeader` | `ImUtf8.CollapsingHeaderId(...)` | always `PopID` (the header pushes its label as ID whether open or not) |
| `Columns` | `ImUtf8.Columns(...)` | always resets to `ImGui.Columns(1, default, true)`; has `Next()` |
| `Combo`, `ListBox`, `Menu`, `Modal`, `Popup`, `PopupContextItem`, `TabBar`, `TabItem`, `Table` | same-named `ImUtf8` function | only when `Success` |
| `DragDropSource` | `ImUtf8.DragDropSource(flags)` | only when `Success`; static `SetPayload(label, ImGuiCond condition = None)` sets an empty payload |
| `DragDropTarget` | `ImUtf8.DragDropTarget()` | only when `Success`; `IsDropping(label, flags)`, static `CheckPayload(label, flags)` |
| `TreeNode` | `ImUtf8.TreeNode(...)` | `TreePop` when `Success` and `NoTreePushOnOpen` was not set |
| `Id` | `ImUtf8.PushId(...)` | pops everything it pushed; `Push(int / nint / bytes / chars / handler)`, `Pop(int count = 1)` |
| `ClipRect` | `ImUtf8.PushClipRect`, `PushClipRectSize` | pops all; `Push`, `PushSize`, `Pop(int count = 1)` |
| `ListClipper` | `ImUtf8.ListClipper(int itemsCount, float itemsHeight = -1f)` | `End` and `Destroy`; `Step()`, `DisplayStart`, `DisplayEnd`, `ItemsCount`, `ItemsHeight`, `StartPosY`, `ForceDisplayRangeByIndices(int, int)`, field `ClipperPtr` |

`ImUtf8.CollapsingHeader(...)` (returns `bool`) is the plain call; `CollapsingHeaderId(...)`
(returns the `CollapsingHeader` end object) additionally scopes an ID.

Because these are `ref struct`s they cannot be captured by lambdas, stored in fields or used
across `await`. `OtterGui.Raii.ImRaii` returns the boxed `IEndObject` instead
(`references/topic-raii.md`).

## Core functions (`Text/ImUtf8.*.cs`)

Only the `ReadOnlySpan<byte>` overload is listed; a `ReadOnlySpan<char>` and a handler overload
with the same remaining parameters exist unless noted. `T` in scalar functions is constrained
`where T : unmanaged, INumber<T>`.

| Function | Signature (byte-span form) |
|---|---|
| `Text` | `void Text(ReadOnlySpan<byte> text)`; `void Text(ReadOnlySpan<byte> text, uint color)` |
| `BulletText` | `void BulletText(ReadOnlySpan<byte> text)` |
| `LabelText` | `void LabelText(ReadOnlySpan<byte> label, ReadOnlySpan<byte> text)` |
| `CalcTextSize` | `Vector2 CalcTextSize(ReadOnlySpan<byte> text, bool hideTextAfterDashes = true, float wrapWidth = 0)`; char and handler forms also have an overload with `out TerminatedByteString formatted` |
| `Button` | `bool Button(ReadOnlySpan<byte> label, Vector2 size = default)` |
| `SmallButton` | `bool SmallButton(ReadOnlySpan<byte> label)` |
| `InvisibleButton` | `bool InvisibleButton(ReadOnlySpan<byte> id, Vector2 size, ImGuiButtonFlags flags = None)` |
| `ColorButton` | `bool ColorButton(ReadOnlySpan<byte> description, Vector4 color, ImGuiColorEditFlags flags = None, ...)`; also `uint color` |
| `Checkbox` | `bool Checkbox(ReadOnlySpan<byte> label, ref bool value)`; `(label, ref uint value, uint flags)`; `bool Checkbox<T>(label, ref T value, T flags) where T : unmanaged, Enum` |
| `RadioButton` | `bool RadioButton(ReadOnlySpan<byte> label, bool active)`; `bool RadioButton<T>(label, ref T value, T flag) where T : IEquatable<T>` |
| `Selectable` | `bool Selectable(ReadOnlySpan<byte> label, bool isSelected = false, ImGuiSelectableFlags flags = None, Vector2 size = default)` |
| `CollapsingHeader` / `CollapsingHeaderId` | `(ReadOnlySpan<byte> label, ImGuiTreeNodeFlags flags = None)`; `(label, ref bool visible, flags)` |
| `TreeNode` | `TreeNode TreeNode(ReadOnlySpan<byte> label, ImGuiTreeNodeFlags flags = None)`; `TreeNode TreeNode(nint id, ReadOnlySpan<byte> label, flags)` |
| `Child` | `Child Child(ReadOnlySpan<byte> id, Vector2 size, bool border = false, ImGuiWindowFlags flags = None)` |
| `Group`, `Tooltip` | `Group Group()`, `Tooltip Tooltip()` |
| `Combo` | `Combo Combo(ReadOnlySpan<byte> label, ReadOnlySpan<byte> preview, ImGuiComboFlags flags = None)`; preview also as chars or `Utf8StringHandler<HintStringHandlerBuffer>` |
| `ListBox` | `ListBox ListBox(ReadOnlySpan<byte> label, Vector2 size)` |
| `Menu` | `Menu Menu(ReadOnlySpan<byte> label, bool enabled = true)` |
| `MenuItem` | `bool MenuItem(ReadOnlySpan<byte> text, bool selected = false, bool enabled = true)`; `(text, shortcut, selected, enabled)` |
| `Popup` | `Popup Popup(ReadOnlySpan<byte> id, ImGuiWindowFlags flags = None)` |
| `Modal` | `Modal Modal(ReadOnlySpan<byte> id, ImGuiWindowFlags flags = None)`; `(id, ref bool open, flags)` |
| `PopupContextItem` | `PopupContextItem PopupContextItem(ReadOnlySpan<byte> id, ImGuiPopupFlags flags = ImGuiPopupFlags.MouseButtonDefault)` |
| `OpenPopup` | `void OpenPopup(ReadOnlySpan<byte> id, ImGuiPopupFlags flags = None)` |
| `IsPopupOpen` | `bool IsPopupOpen(ReadOnlySpan<byte> id, ImGuiPopupFlags flags = None)` |
| `TabBar` | `TabBar TabBar(ReadOnlySpan<byte> label, ImGuiTabBarFlags flags = None)` |
| `TabItem` | `TabItem TabItem(ReadOnlySpan<byte> label, ImGuiTabItemFlags flags = None)`; `(label, ref bool open, flags)` |
| `TabItemButton` | `bool TabItemButton(ReadOnlySpan<byte> label, ImGuiTabItemFlags flags = None)` |
| `Table` | `EndObjects.Table Table(ReadOnlySpan<byte> id, int columns, ImGuiTableFlags flags = None, Vector2 outerSize = default, float innerWidth = default)` |
| `TableSetupColumn` | `void TableSetupColumn(ReadOnlySpan<byte> label, ImGuiTableColumnFlags flags = None, float widthOrWeight = 0, uint userId = 0)` |
| `TableHeader` | `void TableHeader(ReadOnlySpan<byte> label)` |
| `Columns` | `Columns Columns(int count, ReadOnlySpan<byte> id, bool border = true)`; `Columns Columns(int count, bool border = true)` |
| `PushId` | `Id PushId(int id)`, `Id PushId(nint id)`, `Id PushId(ReadOnlySpan<byte> id)` |
| `GetId` | `uint GetId(ReadOnlySpan<byte> label)`; char/handler forms also with `out TerminatedByteString` |
| `InputText` | see below |
| `InputMultiLine` | `bool InputMultiLine(ReadOnlySpan<byte> label, ref string text, Vector2 size = default, ImGuiInputTextFlags flags = None)`; buffer forms `(label, Span<byte> buffer, out TerminatedByteString result, ...)` and `(label, Span<byte> buffer, out string result, ...)` |
| `InputScalar<T>` | `bool InputScalar<T>(ReadOnlySpan<byte> label, ref T value, T step = default, T stepFast = default, ImGuiInputTextFlags flags = None)`; with `format` as second string before `step` |
| `DragScalar<T>` | `bool DragScalar<T>(ReadOnlySpan<byte> label, ref T value, T? min = null, T? max = null, float speed = 1, ImGuiSliderFlags flags = None)`; with `format` after `value` |
| `Slider<T>` | `bool Slider<T>(ReadOnlySpan<byte> label, ref T value, T minValue, T maxValue, ImGuiSliderFlags flags = None)`; with `format` after `value` |
| `SliderVertical<T>` | `bool SliderVertical<T>(ReadOnlySpan<byte> label, ref T value, Vector2 size, T minValue, T maxValue, ImGuiSliderFlags flags = None)`; with `format` after `size` |
| `SliderAngle` | `bool SliderAngle(ReadOnlySpan<byte> label, ref float degreeRadians, float minDegrees = -360f, float maxDegrees = 360f, ImGuiSliderFlags flags = None)`; with `format` after the value |
| `ColorEdit` | `bool ColorEdit(ReadOnlySpan<byte> label, ref Vector3 color, ImGuiColorEditFlags flags = None)`; also `ref Vector4`, `ref uint` |
| `ColorPicker` | `bool ColorPicker(ReadOnlySpan<byte> label, ref Vector3 / Vector4 / uint color, ImGuiColorEditFlags flags = None)`; `Vector4` and `uint` forms also take a `referenceColor` |
| `ProgressBar` | `void ProgressBar(float fraction)`, `(fraction, Vector2 size)`, `(fraction, format)`, `(fraction, size, format)` |
| `RotatedText` | `void RotatedText(ReadOnlySpan<byte> text, bool alignToFrame = false)` |
| `DragDropSource` / `DragDropTarget` | `DragDropSource DragDropSource(ImGuiDragDropFlags flags = None)`, `DragDropTarget DragDropTarget()` |
| `PushClipRect` | `ClipRect PushClipRect(Vector2 minimum, Vector2 maximum, bool push = true, bool intersectWithCurrent = false)`; `PushClipRectSize(minimum, size, push, intersectWithCurrent)` |
| `ListClipper` | `ListClipper ListClipper(int itemsCount, float itemsHeight = -1f)` |
| `SetClipboardText` / `GetClipboardText` | `void SetClipboardText(ReadOnlySpan<byte> text)`; `string GetClipboardText()`; `TerminatedByteString GetClipboardTextUtf8()` |
| `AddText` (extension) | `void AddText(this ImDrawListPtr drawList, ReadOnlySpan<byte> text, Vector2 pos, uint color = 0)`; `(drawList, text, ImFontPtr font, float fontSize, Vector2 pos, ...)` |

Style shortcuts in `ImUtf8.Style.cs` (static properties): `Style` (`ImGuiStylePtr`), `IO`
(`ImGuiIOPtr`), `ItemSpacing`, `ItemInnerSpacing`, `FramePadding`, `GlobalScale`, `TextHeight`,
`FrameHeight`, `TextHeightSpacing`, `FrameHeightSpacing`; method `SameLineInner()`.

### InputText

Three families, each with `hint` as bytes (defaulted), chars or handler:

```csharp
bool InputText(ReadOnlySpan<byte> label, ref string text,
    ReadOnlySpan<byte> hint = default, ImGuiInputTextFlags flags = ImGuiInputTextFlags.None)
bool InputText(ReadOnlySpan<byte> label, Span<byte> buffer, out string result,
    ReadOnlySpan<byte> hint = default, ImGuiInputTextFlags flags = ImGuiInputTextFlags.None)
bool InputText(ReadOnlySpan<byte> label, Span<byte> buffer, out TerminatedByteString result,
    ReadOnlySpan<byte> hint = default, ImGuiInputTextFlags flags = ImGuiInputTextFlags.None)
bool InputText(ReadOnlySpan<byte> label, Span<byte> buffer, out int textLength,
    ReadOnlySpan<byte> hint = default, ImGuiInputTextFlags flags = ImGuiInputTextFlags.None)
```

Usage inside the library: `ImUtf8.InputText("##"u8, ref newFilter, "Filter..."u8,
ImGuiInputTextFlags.EnterReturnsTrue)` (`ItemSelector.cs`).

## Extended helpers (`Text/Extended`)

| Function | Signature (first overload) | Notes |
|---|---|---|
| `ButtonEx` | `bool ButtonEx(ReadOnlySpan<byte> label, ReadOnlySpan<byte> tooltip = default, Vector2 size = default, bool disabled = false, uint textColor = 0, uint buttonColor = 0)` | button with tooltip, disabled state and colours; shorter forms without tooltip |
| `IconButton` | `bool IconButton(FontAwesomeIcon icon, ReadOnlySpan<byte> tooltip, Vector2 size = default, bool disabled = false, uint textColor = 0, uint buttonColor = 0)` | also `(icon, Vector2 size = default)`, `(icon, bool disabled, Vector2 size = default)` |
| `Icon` | `void Icon(FontAwesomeIcon icon)`; `(icon, uint textColor)`; `(icon, tooltip)`; `(icon, tooltip, textColor)` | `Vector2 CalcIconSize(FontAwesomeIcon icon)`, `void IconDummy()` |
| `HoverTooltip` | `void HoverTooltip(ReadOnlySpan<byte> text)`; `(ImGuiHoveredFlags flags, ReadOnlySpan<byte> text)`; `(ref HoverUtf8StringHandler text)`; `(flags, ref HoverUtf8StringHandler text)` | call right after the item |
| `LabeledHelpMarker` | `void LabeledHelpMarker(ReadOnlySpan<byte> label, ReadOnlySpan<byte> tooltip, FontAwesomeIcon icon = FontAwesomeIcon.InfoCircle, bool withSameLine = true)` | |
| `SelectableHelpMarker` | `void SelectableHelpMarker(ReadOnlySpan<byte> tooltip, FontAwesomeIcon icon = FontAwesomeIcon.InfoCircle)` | |
| `CopyOnClickSelectable` | `void CopyOnClickSelectable(ReadOnlySpan<byte> text)`; `(text, copiedText)`; `(text, copiedText, tooltip)`; `(ref Utf8StringHandler<TextStringHandlerBuffer> text, ImFontPtr font)` | |
| `ColorPicker` | `bool ColorPicker(ReadOnlySpan<byte> label, ReadOnlySpan<byte> tooltip, Vector3 input, Action<Vector3> setter, ReadOnlySpan<byte> letter = default)` | setter-style, distinct from the core `ColorPicker(label, ref color)` |
| `InputTextOnDeactivated` | `bool InputTextOnDeactivated(ReadOnlySpan<byte> label, ref string text, ReadOnlySpan<byte> hint = default, ImGuiInputTextFlags flags = None)`; `(label, ReadOnlySpan<char> input, out string result, hint, flags)`; `(label, ReadOnlySpan<byte> input, out TerminatedByteString result, hint, flags)` | returns true once, when the field loses focus after an edit with a changed value |
| `InputScalarOnDeactivated<T>` | `bool InputScalarOnDeactivated<T>(ReadOnlySpan<byte> label, ref T value, T step = default, T stepFast = default, ImGuiInputTextFlags flags = None)` | same, for numbers |
| `InputMultiLineOnDeactivated` | `bool InputMultiLineOnDeactivated(ReadOnlySpan<byte> label, ref string text, Vector2 size = default, ImGuiInputTextFlags flags = None)` | |
| `ResizableChild` | `Child ResizableChild(ReadOnlySpan<byte> label, Vector2 size, out Vector2 currentSize, Action<Vector2> setSize, Vector2 minSize, Vector2 maxSize, ImGuiWindowFlags flags = default, bool resizeX = true, bool resizeY = false)` | bordered child with a drag handle; `setSize` fires when a resize is finalized |
| `TextFramed` | `void TextFramed(ReadOnlySpan<byte> text, uint frameColor, Vector2 size = default, uint textColor = 0, uint borderColor = 0)` | |
| `TextCentered`, `TextFrameAligned` | `(ReadOnlySpan<byte> text)` | |
| `TextRightAligned` | `void TextRightAligned(ReadOnlySpan<byte> text, float offset = 0)` | |
| `TextWrapped` | `void TextWrapped(ReadOnlySpan<byte> text, float wrapPos = 0)` | |
| `TextShadowed` | `void TextShadowed(ReadOnlySpan<byte> text, uint foregroundColor, uint shadowColor, byte shadowWidth = 1)`; also `(ImDrawListPtr drawList, Vector2 position, text, ...)` | |
| `BulletTextColored` | `void BulletTextColored(uint color, ReadOnlySpan<byte> text)` | |
| `DrawTableColumn`, `DrawFrameColumn` | `(ReadOnlySpan<byte> text)` | `TableNextColumn` plus text (frame-aligned in the second) |
| `Dummy`, `ScaledDummy` | `(float width)`, `(float width, float height)`, `(Vector2 size)` | |
| `Spinner` | `bool Spinner(ReadOnlySpan<byte> label, float radius, int thickness, uint color)` | |

The `...OnDeactivated` family keeps the text being edited in `InputStringHandlerBuffer` (or
`DataCache<T>` for scalars) keyed by the ImGui ID, so your variable is only written when editing
ends. Only one such input can be active at a time, which is how ImGui focus works anyway.

## `Text/ImUtf8Test.cs`

`ImUtf8TestWindow` (a `Dalamud.Interface.Windowing.Window`) and `ImUtf8Test` are a benchmark/demo
that ships inside the library. They are the only worked example of `ImUtf8` in the repository;
read `references/src/Text/ImUtf8Test.cs` for side-by-side calls of all three overload kinds.

## Finding an overload

```sh
grep -n 'ImUtf8.Slider' references/api-index.txt     # by file
grep -n ' InputTextOnDeactivated(' references/api-index.txt
```
