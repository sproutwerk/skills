## Colors and style

Prefer `ImGuiColors` (follows the user's style) to literals; `uint` colors are `0xAABBGGRR`.
Never write to `ImGui.GetStyle()` or `ImGui.GetIO()` fields or call `StyleModel.Apply()`:
the context is shared. → `references/topic-colors-style.md`

