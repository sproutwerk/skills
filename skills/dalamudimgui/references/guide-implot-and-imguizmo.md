## ImPlot and ImGuizmo

Both are initialised by Dalamud; plugins must not create contexts. ImPlot labels are **not
`ImU8String`**, so interpolated strings allocate. → `references/topic-implot-imguizmo.md`

