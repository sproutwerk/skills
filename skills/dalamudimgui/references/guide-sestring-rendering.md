## SeString rendering

`ImGui.Text` does not understand SeString payloads or soft hyphens; use
`ImGuiHelpers.SeStringWrapped` / `CompileSeStringWrapped`, and always close a `colortype`
push with `<colortype(0)>`. → `references/topic-sestring.md`

