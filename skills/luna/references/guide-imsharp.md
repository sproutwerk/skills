## ImSharp

### What it is, relative to stock Dalamud

ImSharp is its own binding, not a wrapper over `Dalamud.Bindings.ImGui`. Every
native call is a `[LibraryImport("cimgui", EntryPoint = "ig…")]` partial method
under `Im.Native.Methods.*` (`ImSharp/ImSharp/ImSharp/Native/Methods/*.cs`), and
the ImGui structs are re-declared under `Im.Native.*` and `Im.Native.Internal.*`
(the latter mirror `imgui_internal.h`: `Context`, `Window`, `Table`, …).
`Dalamud.Bindings.ImGui` also resolves its library name to `"cimgui"`
(`imgui/Dalamud.Bindings.ImGui/ImGui.cs`), so both bindings talk to the same
loaded native library and the same ImGui context. Dalamud's cimgui is 1.88,
docking branch: Dalamud `43b3411` pins `lib/cimgui` to `goatcorp/gc-cimgui`
`ecd32e5`, whose `cimgui.h` header says so (details in the `dalamudimgui`
skill). That matches ImSharp's branch name `gc1.88` and its
`VersionString = "1.88"` / `VersionNumber = 18800` (`ImSharp/Version.cs`).
That the DLL loaded at runtime is that exact build is *(unverified)*.

| Concern | Stock Dalamud | ImSharp / Luna |
|---|---|---|
| Entry class | `ImGui.*` (generated) | `Im.*` hand-written, grouped into nested static classes (`Im.Table`, `Im.Popup`, …); extras in `ImEx.*` |
| Strings | `ImU8String` | `Utf8StringHandler<TBuffer>` writing into shared native buffers; `StringU8` owned UTF-8 |
| RAII | `ImRaii.*` (`Dalamud/Interface/Utility/Raii`) | `Im.*Disposable` returned by `Im.X.Begin(…)` / `.Push(…)` |
| Enums | `ImGuiCol`, `ImGuiStyleVar`, `ImGuiWindowFlags` | `ImGuiColor`, `ImStyleSingle`/`ImStyleDouble`/`ImStyleBorder`, `WindowFlags`, … (members renamed, e.g. `WindowBackground`, `FrameBackground`, `TableRowBackground`) |
| Colors | `uint` / `Vector4` | `Rgba32`, `ColorParameter`, `Vector4` |
| IDs / textures | `uint`, `ImTextureID` | `ImGuiId`, `ImTextureId` |
| Windows | `Dalamud.Interface.Windowing.Window` | `Luna.Window` derives from it and re-types `Flags` and the `*Condition` properties to ImSharp enums |

Enum values are cast between the two worlds in Luna (`(ImGuiWindowFlags)flags`,
`(MouseButton)button`), so they are numerically compatible. Helpers:
`WindowFlags.ToDalamudWindowFlags()`, `ImTextureId.ToDalamudId()`,
`VirtualKey.ToImGuiKey()` / `Key.ToVirtualKey()`.

Namespaces: `ImSharp` (`Im`, `ImEx`, `StringU8`, `Rgba32`, filters, combos,
`CacheManager`), `ImSharp.Table` (`TableData`, `DefaultTable`, columns,
`TableCache`), `ImSharp.Containers` (`ObservableList`, `CacheListAdapter`),
`ImSharp.ImNodes` (a C# node-editor implementation; `ImNodesColor` is used by
`Luna/Colors`). `ImSharp.csproj` defines `IMPLOT`, but `Luna.csproj` does not,
and `ImSharp/ImPlot/` holds only an `Enums` folder; treat ImPlot as unavailable
through Luna.

### Minimal window (worked example)

```csharp
using System.Numerics;
using Dalamud.Interface.Windowing;   // WindowSizeConstraints
using ImSharp;
using Luna;
using Window = Luna.Window;          // disambiguate from Dalamud's Window, as Glamourer does

public sealed class DemoWindow : Window
{
    private string _name    = string.Empty;
    private int    _count   = 3;
    private bool   _enabled = true;

    public DemoWindow()
        : base("Demo###DemoWindow", WindowFlags.NoCollapse)
        => SizeConstraints = new WindowSizeConstraints
        {
            MinimumSize = new Vector2(400, 300),
            MaximumSize = new Vector2(4096, 4096),
        };

    public override void Draw()                       // Dalamud's Window has already called Begin
    {
        Im.Text("Static label"u8);                    // no copy
        Im.Text($"Count is {_count}");                // formatted into the shared text buffer
        Im.Line.Same();
        LunaStyle.DrawAlignedHelpMarker("Explains the counter."u8);

        Im.Item.SetNextWidthScaled(200);
        Im.Input.Text("##name"u8, ref _name, "Name..."u8);
        Im.Checkbox("Enabled"u8, ref _enabled);

        using (Im.Disabled(!_enabled))
        {
            Im.Item.SetNextWidthScaled(200);
            Im.Slider("Count"u8, ref _count, 0, 10);
            if (ImEx.Icon.Button(LunaStyle.RefreshIcon, "Reset the counter."u8))
                _count = 0;
        }

        using (ImGuiColor.Text.Push(LunaStyle.ErrorForeground, _count > 8))
            Im.Text("High"u8);

        using var child = Im.Child.Begin("##list"u8, Im.ContentRegion.Available, true);
        if (!child)
            return;

        for (var i = 0; i < _count; ++i)
        {
            using var id = Im.Id.Push(i);
            if (Im.Selectable($"Row {i}"))
                _name = $"Row {i}";
            Im.Tooltip.OnHover($"Click to copy row {i} into the name.");
        }
    }
}
```

Register it: `windowSystem.AddWindow(new DemoWindow())` on a `Luna.WindowSystem`.
A free-standing window outside the window system uses
`using var w = Im.Window.Begin("Name"u8, ref open, flags); if (!w) return;`.

### Strings

Write literals as `"..."u8` (passed by pointer, no copy). `string` and `$"…"` are transcoded
into one shared buffer per handler kind (label, text, hint); two interpolated strings of the
*same* kind alive at once overwrite each other. → `references/topic-imsharp-strings.md`

### Scope objects

`Im.X.Begin(…)` returns a `ref struct` with implicit `bool` (check it, return early);
`.Push(…)` returns a chainable `sealed class` that pops exactly what it pushed. Use both with
`using`. → `references/topic-imsharp-scopes.md`

### `Im` by area and `ImEx` extras

`Im.Style.GlobalScale` is the scale factor to multiply pixel sizes by; `ImEx.ScaledVector(x, y)`
and the `…Scaled` helpers do it for you. Member lists → `references/topic-imsharp-widgets.md`

### Tables and table classes

`Im.Table.Begin(id, columns, flags)` returns `Im.TableDisposable`, whose instance methods are
the table API; `DrawColumn(text)` = `NextColumn()` then `Im.Text`, and `NextColumn` wraps rows.
→ `references/topic-imsharp-tables.md`

### Filter combos, filters and caches

`GetItems()` is only called when the cache is dirty with `IManagedCache.DirtyFlags.Custom`;
selection in `SimpleFilterCombo<T>` needs sensible equality on `T`.
→ `references/topic-imsharp-filters-combos-caches.md`

### Draw lists, colors, IDs, textures, fonts

`Rgba32` is ABGR-packed like ImGui (`Red = 0xFF0000FF`); an omitted draw-list color is white;
`IDalamudTextureWrap.Id` gives the `ImTextureId`. → `references/topic-imsharp-drawing.md`

