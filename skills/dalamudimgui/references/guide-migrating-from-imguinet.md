## Migrating from ImGui.NET

The rows generic ImGui.NET knowledge gets wrong most often:

| ImGui.NET (API ≤ 12) | Dalamud bindings (API 13+) |
|---|---|
| `using ImGuiNET;` / `ImPlotNET` / `ImGuizmoNET` | `Dalamud.Bindings.ImGui` / `.ImPlot` / `.ImGuizmo` |
| `string` parameters | `ImU8String` (accepts `string`, `"..."u8`, `$"..."`, spans) |
| `ImGui.Text(fmt)` went through printf formatting | `ImGui.Text` is unformatted; `%` is literal |
| `out` parameters | `ref` or pointer overloads (only `CalcTextSizeA` and `ImTextStrFromUtf8` still use `out`) |
| `IntPtr` texture ids, `wrap.ImGuiHandle` | `ImTextureID`, `wrap.Handle` |
| `ImGuiColorEditFlags.DisplayRGB`, `.HDR`, `.InputHSV` | PascalCase: `DisplayRgb`, `Hdr`, `InputHsv` (check `index/enums.txt`) |
| `ImGui.InputInt(label, ref v)` showed +/- buttons (step 1) | `step` defaults to **0**: no buttons unless you pass `step: 1` |
| `ImGui.InputText(label, ref s, maxLen)` with `uint` length | `InputText(label, ref s, int maxLength = 512, flags, callback)` |
| `ImGui.PushID(int)` | No `int` overload; an `int` binds to `PushID(nint)`. Fine, but hashes differ from native `PushID(int)` |

Full table (obsolete helpers, clipper construction, `ImGuiNative`, `IEndObject`) and API 16
changes → `references/topic-migration.md`

