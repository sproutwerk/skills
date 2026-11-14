## ImGuiComponents

Takes `string`, not `ImU8String`. `DisabledButton` only *looks* disabled and still returns
clicks (use `ImRaii.Disabled`); the toggle switch is `ToggleButton(string id, ref bool v)`;
`HelpMarker` calls `SameLine()` itself. → `references/topic-helpers-components.md`

