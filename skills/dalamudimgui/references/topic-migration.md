# Migrating from ImGui.NET

Covers the full ImGui.NET (API 12 and older) to Dalamud bindings (API 13+) mapping and API 16 notes.
Verified against goatcorp/Dalamud `master` @ `43b3411` (2026-11-10, API level 15, Dear ImGui 1.88 docking).
Paths are relative to the skill root (`SKILL.md`, `references/topic-*.md`, `references/...`); a bare
`index/...` or `src/...` path means `references/index/...` or `references/src/...`.

| ImGui.NET (API ≤ 12) | Dalamud bindings (API 13+) |
|---|---|
| `using ImGuiNET;` / `ImPlotNET` / `ImGuizmoNET` | `Dalamud.Bindings.ImGui` / `.ImPlot` / `.ImGuizmo` |
| `string` parameters | `ImU8String` (accepts `string`, `"..."u8`, `$"..."`, spans) |
| `ImGui.Text(fmt)` went through printf formatting | `ImGui.Text` is unformatted; `%` is literal |
| `ImGuiHelpers.SafeTextWrapped` / `SafeTextColoredWrapped` | `ImGui.TextWrapped` / `ImGui.TextColoredWrapped` (the old ones are `[Obsolete(..., true)]` = compile error) |
| `out` parameters | `ref` or pointer overloads (only `CalcTextSizeA` and `ImTextStrFromUtf8` still use `out`) |
| `IntPtr` texture ids, `wrap.ImGuiHandle` | `ImTextureID`, `wrap.Handle` |
| `ImGuiColorEditFlags.DisplayRGB`, `.HDR`, `.InputHSV` | PascalCase: `DisplayRgb`, `Hdr`, `InputHsv` (check `index/enums.txt`) |
| `new ImGuiListClipperPtr(ImGuiNative.ImGuiListClipper_ImGuiListClipper())` | `ImGui.ImGuiListClipper()` |
| `ImGui.InputInt(label, ref v)` showed +/- buttons (step 1) | `step` defaults to **0**: no buttons unless you pass `step: 1` |
| `ImGui.InputText(label, ref s, maxLen)` with `uint` length | `InputText(label, ref s, int maxLength = 512, flags, callback)` |
| `ImGui.PushID(int)` | No `int` overload; an `int` binds to `PushID(nint)`. Fine, but hashes differ from native `PushID(int)` |
| Internals via reflection / own P/Invoke | `ImGuiP` (unstable across ImGui upgrades) |
| `ImGuiNative.igFoo(...)` | `ImGuiNative.Foo(...)` (public, raw pointers, no `ig` prefix) |
| `IEndObject` from `ImRaii` (API ≤ 14) | Concrete types: `ref struct` for Begin/End scopes, `sealed class` for push stacks (API 15) |

API 16 (in development, not in this checkout; see `references/upstream-docs/v16/index.md`):

- The mutable `SeString` class goes away. `SeStringDrawResult.InteractedPayload` changes
  from `Payload?` to `ReadOnlySePayload?`, and any `SeString`-typed UI helper changes with
  it. `references/upstream-docs/v16/sestring.md` is the upstream conversion guide.
- `IDragDropManager` moves to `Dalamud.Plugin.Services`.
- `ImGuiHelpers.GlobalScaleSafe` is marked for removal (`[Api16ToDo]`); use `GlobalScale`.
- The notes list a move to .NET 11 and C# 15 as pending.
