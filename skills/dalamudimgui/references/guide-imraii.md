## ImRaii

`Dalamud.Interface.Utility.Raii.ImRaii`: factory methods returning disposable
scopes. Two families with different rules.

Must check (`if (!x) return;` or `if (x)`), implicit `bool`:
`Child`, `ChildFrame`, `Table`, `TabBar`, `TabItem`, `TreeNode`, `Header`, `Combo`, `ListBox`,
`Popup`, `PopupModal`, `ContextPopup`, `ContextPopupItem`, `Menu`, `MenuBar`, `MainMenuBar`,
`DragDropSource`, `DragDropTarget`. These are `ref struct`s (stack only). End is called
always for `Child`/`ChildFrame`, on success for the rest. No bool conversion, always ended:
`Group()`, `Tooltip()`, `Columns(count, id, border = false)`.

Push stacks are `sealed class` objects (chainable, storable in a field): `PushColor`,
`PushStyle`, `PushFont`, `PushId`, `PushIndent`, `PushClipRect`, `ItemWidth`, `TextWrapPos`,
`Disabled`, `Enabled`, `DefaultColors`/`DefaultStyle`/`DefaultFont`. Most take a trailing `condition`.

The pattern, including the early-out that keeps nesting flat:

```csharp
using var table = ImRaii.Table("items"u8, 3, ImGuiTableFlags.RowBg | ImGuiTableFlags.ScrollY);
if (!table)
    return;                      // Dispose still runs; EndTable is skipped because Begin failed

ImGui.TableSetupColumn("Name"u8, ImGuiTableColumnFlags.WidthStretch);
ImGui.TableSetupColumn("Qty"u8, ImGuiTableColumnFlags.WidthFixed, 60 * ImGuiHelpers.GlobalScale);
ImGui.TableSetupColumn("##actions"u8, ImGuiTableColumnFlags.WidthFixed, ImGui.GetFrameHeight());
ImGui.TableHeadersRow();
```

Nested form: `using (var tab = ImRaii.TabItem("General"u8)) if (tab) DrawGeneral();`

Every checked scope has `Success`, `Alive`, implicit `bool`, `!`, `&`, `|`
and a static `EndUnsafe()`/`PopUnsafe()` escape hatch. `Dispose()` is
idempotent, so calling it early to close a scope before the block ends is
fine (`ImGuiClip` does this with `Group`).

Skipping the check draws into the parent or asserts in native code. For
`Child` the body is merely wasted work when clipped, but check anyway.

Push stack example:

```csharp
using var color = ImRaii.PushColor(ImGuiCol.Button, ImGuiColors.DalamudRed, isDangerous)
                        .Push(ImGuiCol.ButtonHovered, ImGuiColors.DalamudRed.Lighten(0.1f), isDangerous);
using var style = ImRaii.PushStyle(ImGuiStyleVar.FramePadding, new Vector2(8, 4) * ImGuiHelpers.GlobalScale)
                        .Push(ImGuiStyleVar.FrameRounding, 4f);
using var disabled = ImRaii.Disabled(!canDelete);
if (ImGui.Button("Delete"u8)) Delete();
```

- One object tracks its own count; `Dispose()` pops exactly what it pushed,
  including zero when every `condition` was false. `PushColor` with a `null`
  `Vector4?` pushes nothing.
- `StyleDisposable.Push` validates the value type against the style var and
  throws `ArgumentException` for a mismatch (`float` into `FramePadding`).
- These are heap objects: one small allocation per `ImRaii.Push*` call per
  frame. Chain pushes on one object; in very hot loops use raw
  `ImGui.PushStyleColor`/`PopStyleColor`.
- `PushIndent(float)` scales by `GlobalScale` unless `scaled: false`.
- `EnabledDisposable` only knows about `ImRaii.Disabled`, not raw
  `ImGui.BeginDisabled`.

`Header(label, ref visible, flags)` leaked its `PushID(label)` on Dalamud builds
before `15.0.3.8` (the constructor never set `Alive`, so `Dispose()` returned
early). The bundled revision has the fix. On an older runtime use
`Header(label, flags)` or a manual `ImGui.CollapsingHeader(label, ref visible)`.

There is no `ImRaii.Window`; windows go through `WindowSystem` or raw
`Begin`/`End`.

Full factory catalog with arguments, chaining methods and ImPlot scopes → `references/topic-imraii.md`

