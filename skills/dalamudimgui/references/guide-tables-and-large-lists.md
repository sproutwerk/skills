## Tables and large lists

Tables are stock ImGui through `ImRaii.Table`. Over a few hundred rows, clip with
`ImGuiClip.ClippedDraw(...)`; row height must be constant and include item spacing. A
hand-made `ImGui.ImGuiListClipper()` must be `Destroy()`ed. → `references/topic-tables.md`

