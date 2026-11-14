# ImSharp scope objects and style/color/ID stacks

Covers the Begin/End `ref struct` disposables, the chainable stack disposables, the `ImStyleSingle`
/ `ImStyleDouble` / `ImStyleBorder` enums and `IdDisposable`. Verified against Luna `52eb0fe` and
ImSharp `4d09327` (branch `gc1.88`, ImGui 1.88). Paths are relative to the skill root. Grep
`references/api-index.txt` before using a member that is not shown here.

## Scope objects

Two families, both used with `using`:

| Family | Kind | Examples | Dispose |
|---|---|---|---|
| Begin/End | `ref struct`, `Success`, `Alive`, implicit `bool` | `WindowDisposable`, `ChildDisposable`, `TableDisposable`, `TabBarDisposable`, `TabItemDisposable`, `PopupDisposable`, `ComboDisposable`, `TreeNodeDisposable`, `ListBoxDisposable`, `MenuDisposable`, `MenuBarDisposable`, `MainMenuBarDisposable`, `TooltipDisposable`, `GroupDisposable`, `DragDropSourceDisposable`, `DragDropTargetDisposable`, `HeaderDisposable`, `ColumnsDisposable` | Window and Child always call `End`/`EndChild`; Table only calls `EndTable` when `Success`. Each has a static `EndUnsafe()` / `PopUnsafe()`. |
| Stacks | `sealed class`, chainable, tracks its own count | `ColorDisposable`, `StyleDisposable`, `ColorStyleDisposable`, `IdDisposable`, `FontDisposable`, `IndentDisposable`, `ItemWidthDisposable`, `TextWrapDisposable`, `ClipRectDisposable`, `DrawListClipRectDisposable`, `ButtonRepeatDisposable`, `KeyboardFocusDisposable`, `DisabledDisposable`, `EnabledDisposable` | Pops exactly what that object pushed. `Pop(n)` is clamped to its count. |

All are nested in `Im` (`Im.TableDisposable`, `Im.ColorDisposable`). Stack
objects are heap allocations; reuse one by chaining rather than creating many.

```csharp
using var color = ImGuiColor.Text.Push(LunaStyle.ErrorForeground)      // Im.ColorDisposable
    .Push(ImGuiColor.Button, Rgba32.Transparent)
    .Push(ImGuiColor.Border, someColor, condition);                    // conditional push
using var style = ImStyleDouble.ItemSpacing.Push(Vector2.Zero)         // Im.StyleDisposable
    .Push(ImStyleSingle.FrameRounding, 4f)
    .PushX(ImStyleDouble.FramePadding, 0);
using var both  = Im.ColorStyle()                                      // Im.ColorStyleDisposable
    .Push(ImGuiColor.Text, Rgba32.White).Push(ImStyleSingle.Alpha, 0.5f);
using var border = ImStyleBorder.Frame.Push(Rgba32.Red, 2f);           // border color + thickness
using var id    = Im.Id.Push("row"u8);                                 // also Push(ImGuiId) / Push(nint)
using var font  = Im.Font.PushMono();                                  // or AwesomeIcon.Font.Push()
using var dis   = Im.Disabled(!enabled);
using var ind   = Im.Indent();
using var w     = Im.Item.PushWidth(200 * Im.Style.GlobalScale);
```

Style enums: `ImStyleSingle` (float: `Alpha`, `DisabledAlpha`, `WindowRounding`,
`WindowBorderThickness`, `ChildRounding`, `ChildBorderThickness`,
`PopupRounding`, `PopupBorderThickness`, `FrameRounding`,
`FrameBorderThickness`, `IndentSpacing`, `ScrollbarSize`, `ScrollbarRounding`,
`MinimumGrabSize`, `GrabRounding`, `TabRounding`), `ImStyleDouble` (Vector2:
`WindowPadding`, `MinimumWindowSize`, `WindowTitleAlignment`, `FramePadding`,
`ItemSpacing`, `ItemInnerSpacing`, `CellPadding`, `ButtonTextAlign`,
`SelectableTextAlign`), `ImStyleBorder` (`Window`, `Child`, `Popup`, `Frame`).
The split makes a float/Vector2 mismatch a compile error. `PushDefault(…)`
re-pushes the value that was in effect before anything was pushed.

`IdDisposable.Push(ImGuiId)` calls `igPushID_Int` with the raw value; an `int`
converts implicitly to `ImGuiId`, so `Im.Id.Push(idx)` works for loop indices.
`IdDisposable.PushNext()` pushes an auto-incrementing counter.
