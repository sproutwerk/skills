# ImPlot and ImGuizmo

Covers the `Dalamud.Bindings.ImPlot` and `Dalamud.Bindings.ImGuizmo` bindings and the ImAnim
status.
Verified against goatcorp/Dalamud `master` @ `43b3411` (2026-11-10, API level 15, Dear ImGui 1.88 docking).
Paths are relative to the skill root (`SKILL.md`, `references/topic-*.md`, `references/...`); a bare
`index/...` or `src/...` path means `references/index/...` or `references/src/...`.

Both ship with Dalamud and are initialised against the shared ImGui context
by the backend (`ImPlot.SetImGuiContext` + `CreateContext`,
`ImGuizmo.SetImGuiContext`, and `ImGuizmo.BeginFrame()` every frame), so
plugins call them directly and must not create contexts.

- `using Dalamud.Bindings.ImPlot;` ImPlot 0.14, fully generated: labels are
  `string`, `ReadOnlySpan<byte>` or `ref byte`, **not `ImU8String`**, so
  interpolated strings allocate. Data is passed as `ref T` to the first
  element plus a count. Use the `ImRaii.Plot(...)` family for scopes.
- `using Dalamud.Bindings.ImGuizmo;` `Manipulate`, `ViewManipulate`,
  `DecomposeMatrixToComponents`, `RecomposeMatrixFromComponents`, `SetRect`,
  `SetDrawlist`, `SetOrthographic`, `SetID`, `IsOver`, `IsUsing`, `Enable`,
  `AllowAxisFlip`, `DrawGrid`, `DrawCubes`; enums `ImGuizmoOperation`,
  `ImGuizmoMode`.
- ImAnim bindings (`Dalamud.Bindings.ImAnim`) are announced in the v14 notes
  but are not present in this checkout *(unverified)*.
