## What you need on nearly every task

### Strings: three overloads everywhere

```csharp
using OtterGui.Text;

ImUtf8.Text("Static"u8);                          // ReadOnlySpan<byte>, zero cost
ImUtf8.Text(name);                                // ReadOnlySpan<char>, transcoded to a static buffer
ImUtf8.Text($"{count} items ({size:F1} MiB)");    // interpolated, formatted to UTF-8, no string allocated
if (ImUtf8.Button("Save"u8)) Save();
ImUtf8.HoverTooltip($"Saves {path}");             // holes are not evaluated unless the item is hovered
```

Labels and IDs passed as `ReadOnlySpan<byte>` must be null-terminated; `u8` literals are. The
`char` and interpolated overloads write into per-type static buffers (labels 128 KiB, hints
128 KiB, text 4 MiB) and throw `ImUtf8FormatException` on overflow. Draw thread only.

### Scopes

```csharp
using OtterGui.Raii;          // ImRaii: string-based, returns IEndObject
using OtterGui.Text;          // ImUtf8: UTF-8, returns ref structs

using var id    = ImRaii.PushId(index);
using var style = ImRaii.PushStyle(ImGuiStyleVar.ItemSpacing, Vector2.Zero)
    .Push(ImGuiStyleVar.FrameRounding, 0);
using var color = ImRaii.PushColor(ImGuiCol.Text, 0xFF8080FF, isError);   // third argument: condition

using var table = ImUtf8.Table("##t"u8, 3, ImGuiTableFlags.RowBg);
if (!table)
    return;

using (var child = ImUtf8.Child("##c"u8, size, true))
{
    if (child) { /* ... */ }
}                                                   // EndChild is always called
```

Scopes that always end (`Child`, `Group`, `Tooltip`) and scopes that end only on success
(`Table`, `Combo`, `Popup`, `TabBar`, `TabItem`, `TreeNode`, ...) are both handled by `Dispose`;
test the object before drawing the contents.

### Tooltips, disabled buttons, enum combos (`ImGuiUtil`, string API)

```csharp
using OtterGui;

if (ImGuiUtil.DrawDisabledButton("Delete", Vector2.Zero, "Hold Control to delete.", !ImGui.GetIO().KeyCtrl))
    Delete();
ImGuiUtil.HoverTooltip("Shown while the last item is hovered.");
if (ImGuiUtil.GenericEnumCombo("Mode", 150 * ImGuiHelpers.GlobalScale, config.Mode, out var mode))
    config.Mode = mode;
```

### Long lists

```csharp
ImGuiClip.ClippedDraw(items, DrawRow, ImGui.GetTextLineHeightWithSpacing());   // IReadOnlyList<T>
```

Every row must have exactly the height you pass.

