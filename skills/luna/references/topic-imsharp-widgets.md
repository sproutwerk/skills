# ImSharp widgets: `Im` by area and `ImEx` extras

Covers the member catalog of the `Im` nested classes (text, buttons, sliders, input, combo, window,
child, popup, tab bar, tree, menu, tooltip, layout, item, ID, input state, drag-drop, color, style,
font, image, list clipper) and the `ImEx` extras. Verified against Luna `52eb0fe` and ImSharp
`4d09327` (branch `gc1.88`, ImGui 1.88). Paths are relative to the skill root. Grep
`references/api-index.txt` before using a member that is not shown here.

## `Im` by area

Read properties as live ImGui state. Full overload lists: grep
`ImSharp/ImSharp/ImSharp/<File>.cs` in `references/api-index.txt`.

| Area | Members |
|---|---|
| Text | `Im.Text(text)`, `Im.Text(text, Rgba32)`, `Im.Text(text, in Vector4)`, `Im.TextDisabled`, `Im.TextWrapped(text, wrapPosition = 0)`, `Im.BulletText`, `Im.PushTextWrapPosition` |
| Buttons etc. | `Im.Button(label, size = default, flags)`, `SmallButton`, `InvisibleButton`, `ArrowButton(id, Direction)`, `RadioButton`, `Selectable(label, isSelected, flags, size)`, `Checkbox(label, ref bool)`, `Checkbox<T>(label, ref T value, T flags) where T : unmanaged, Enum`, `ProgressBar`, `Bullet`, `Separator` |
| Sliders | `Im.Slider<T>(label, ref value, min, max, …)`, `SliderN<T>(label, Span<T>, …)`, `SliderVertical<T>`, `SliderAngle`, `Im.Drag<T>(label, ref value, [format,] T? min, T? max, float speed, …)`, `DragN<T>` |
| `Im.Input` | `Text(label, ref string, hint = default, flags, maxLength)`, `Text(label, ref StringU8, …)`, `Text(label, Span<byte> buffer, out StringU8 result, …)`, `MultiLine(…)`, `Scalar<T>(label, ref value, [format,] step, stepFast, …)`, `ScalarN<T>` |
| `Im.Combo` | `Begin(label, preview, flags)` → `ComboDisposable`; `Draw(label, ref int, params IReadOnlyList<StringU8|string>)`; `DrawEnum<T>(label, ref T, [toName], …)`; `DrawItems<T>(label, ref T, …)`; `DrawPreview` + `DrawPopup` for custom combos |
| `Im.Window` | `Begin(name[, ref open], flags)`; `SetNextPosition/Size/SizeConstraints/ContentSize/Collapsed/Scroll/BackgroundAlpha/Viewport`, `FocusNext`; `Appearing`, `Collapsed`, `Focused(flags)`, `Hovered(flags)`, `DrawList`, `Position`, `Size`, `Width`, `Height`, `Viewport`, `DpiScale` |
| `Im.Child` | `Begin(id, size, border = false, flags)` and overloads without size / with `ImGuiId` |
| `Im.Popup` | `Open`, `OpenOnItemClick`, `IsOpen`, `CloseCurrent`, `Begin`, `BeginModal`, `BeginResizable`, `BeginContextItem`, `BeginContextWindow`, `BeginContextVoid` |
| `Im.TabBar` | `Begin(label, flags)`, `BeginItem(label[, ref open], flags)`, `Button`, `SetItemClosed`; or `tabBar.Item(…)` / `tabBar.Button(…)` on the disposable |
| `Im.Tree` | `Node(label, flags)` → `TreeNodeDisposable`, `Leaf`, `Push`, `Header(label, flags)` (bool, collapsing header), `HeaderId` (collapsing header + ID scope), `SetNextOpen`, `ToggledOpen` |
| `Im.Menu` | `Bar()`, `Main()`, `Begin(label, enabled)`, `Item(label[, shortcut], selected/ref selected, enabled)` |
| `Im.Tooltip` | `Begin()`, `Set(text)`, `OnHover(text…)`, `OnHover(HoveredFlags, …)`, `ImageOnHover(image, size, …)` |
| Layout | `Im.Line.Same()`, `Same(offsetFromStartX, spacing)`, `SameInner()`, `NoSpacing()`, `New()`, `Spacing()`; `Im.Dummy(…)`, `ScaledDummy`, `FrameDummy`; `Im.Group()`; `Im.Indent(…)`; `Im.Cursor.Position/X/Y/ScreenPosition/StartPosition`, `Cursor.FrameAlign()`; `Im.ContentRegion.Available/Maximum`, `ContentRegion.Width(y = 0)` and `Height(x = 0)` (the available size with the other axis replaced); `Im.Scroll.X/Y/MaximumX/MaximumY/SetHereY(…)` |
| `Im.Item` | `SetNextWidth`, `SetNextWidthFull`, `SetNextWidthScaled`, `PushWidth`, `CalculateWidth`; state `Hovered(flags)`, `Clicked(button)`, `RightClicked()`, `MiddleClicked()`, `Active`, `Focused`, `Visible`, `Edited`, `Activated`, `Deactivated`, `DeactivatedAfterEdit`; geometry `UpperLeftCorner`, `LowerRightCorner`, `Bounds`, `Size`; low level `Add`, `SetSize`, `AllowOverlap` |
| `Im.Id` | `Push`, `Get(label)`, `Get(nint)`, `Calculate(id, seed)`, `Current`, `Active`, `IsActive`, `SetActive`, `ClearActive` |
| Input state | `Im.Mouse.IsDown/IsClicked/IsReleased/IsDoubleClicked/IsDragging/GetDragDelta/Position/Cursor`; `Im.Keyboard.IsDown/IsPressed/IsReleased/SetFocusHere`; `Im.Io.KeyControl/KeyShift/KeyAlt/KeyModifiers/MouseWheel/DeltaTime/DisplaySize` |
| `Im.DragDrop` | `Source(flags)`, `Target()`, `SetPayload<T>(id, data)`, `AcceptPayload`, `TryAcceptPayload(type, out Payload)`, `PeekPayload`; `Payload.TryGetData<T>(out T)`; `AutoScroll(speed, marginHeight)` scrolls the hovered window while a drag is near its top or bottom edge |
| `Im.Color` | `Editor(label, ref Vector3|Vector4|Rgba32, flags)`, `Picker`, `Button`, `Tooltip`, `Push`, `Get(ImGuiColor, alphaMultiplier)`, `ToHsv`, `ToRgb` |
| `Im.Style` | Read-only view of the style (`ItemSpacing`, `FramePadding`, `FrameHeight`, `FrameHeightWithSpacing`, `TextHeight`, `TextHeightWithSpacing`, `GlobalScale`, indexers by `ImGuiColor`/`ImStyleSingle`/`ImStyleDouble`) plus `Push…` helpers; `AsWritable()` for direct writes |
| `Im.Font` | `Current`, `Mono`, `Default`, `Push(font)`, `PushMono()`, `PushDefault()`, `CurrentSize`, `CalculateSize(text, hideTextAfterDashes, wrapWidth)`, `CalculateButtonSize(text)` |
| `Im.Image` | `Draw(ImTextureId image, Vector2 size, …)`, `DrawScaled(image, iconSize, imageSize)`, `Button(…)` |
| `Im.ListClipper` | struct, `IDisposable`, `IEnumerable<int>`; `new Im.ListClipper(count, itemHeight)`, `Iterate(list)`, `Step()`, static `Draw(…)` / `DrawGrouped(…)` |
| Misc | `Im.Clipboard.Set/Get/GetCopy/GetUtf16`, `Im.State.FrameCount/Time/Storage`, `Im.Viewport.Main`, `Im.Debug.ShowDemoWindow()`, `Im.Behavior.Button/Splitter`, `Im.Render.Frame/FrameBorder/Checkmark/Cross` |

`Im.Style.GlobalScale` is the scale factor to multiply pixel sizes by;
`ImEx.ScaledVector(x, y)` and the `…Scaled` helpers do it for you.

## `ImEx` extras

| Member | Purpose |
|---|---|
| `ImEx.Button(label, tooltip, in ButtonConfiguration config)` and size/disabled overloads | Button with tooltip, disabled state and per-state colors (`ButtonConfiguration`: `Size`, `ButtonColor`, `HoveredColor`, `ActiveColor`, `TextColor`, `BorderColor`, `Flags`, `Disabled`). |
| `ImEx.SplitButton(id, SplitButtonData left, SplitButtonData right, size, …)` | Two-halved button, returns `SplitButtonHalf`. |
| `ImEx.TextFramed`, `TextBordered`, `TextFrameAligned`, `TextRightAligned`, `TextCentered`, `TextShadowed`, `MonoText`, `TextLabel`, `RotatedText`, `TextMultiColored(…).Then(…).End()` | Text variants. |
| `ImEx.InputOnDeactivation.Text/MultiLine/Scalar/Drag` | Inputs that only report on deactivation-after-edit or Enter. |
| `ImEx.TwoStateCheckbox`, `TriStateCheckbox` (`bool?` or on/off flag pair), `XCheckbox`, `IconCheckbox` | Built on `MultiStateCheckbox<T>`. |
| `ImEx.CopyOnClickSelectable`, `GuidInput`, `LogarithmicInput`, `HexViewer`, `Spinner`, `ColorPickerButton`, `HelpPopup` | |
| `ImEx.FramedGroup(label, tooltip, …)`, `ResizableChild(…)`, `IconTreeNode(…)`, `TreeLine.Draw(IReadOnlyList<IFlattenedTreeNode>, …)` | Containers and tree guides. |
| `ImEx.ScaledVector`, `ScaledVectorX/Y`, `SplitLabel`, `VisibleLabel`, `GetButtonColor`, `GetFrameBackgroundColor`, `HoverHelpLines` | Utilities. |
