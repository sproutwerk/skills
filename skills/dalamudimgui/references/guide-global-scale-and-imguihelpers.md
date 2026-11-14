## Global scale and ImGuiHelpers

`ImGuiHelpers.GlobalScale` is the user's Dalamud UI scale
(`ImGui.GetIO().FontGlobalScale`, or the configured value before ImGui is
up). Fonts scale automatically; **every pixel constant you write does not**.

Most used: `GlobalScale`, `ScaledVector2(x)` / `(x, y)`, `ScaledDummy(size)`,
`ScaledIndent(size)`, `GetButtonSize(text)`, `CenteredText(text)`,
`ClickToCopyText(text, textCopy = default, color = null)`.

Rules of thumb: derive sizes from ImGui metrics where possible
(`ImGui.GetFrameHeight()`, `GetTextLineHeightWithSpacing()`,
`GetStyle().ItemSpacing`, `CalcTextSize`) since those already include the
scale; multiply everything else. `Window.Size`/`SizeConstraints` and the
`size` argument of `ImGuiComponents.IconButton*` are scaled for you, so do
not pre-multiply those. Values stored in config should be unscaled.

Full helper table (viewports, blur, key mapping, font helpers) → `references/topic-helpers-components.md`

