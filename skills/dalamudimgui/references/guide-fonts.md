## Fonts

Icons: inside `using (ImRaii.PushFont(UiBuilder.IconFont))` draw
`ImGui.Text(FontAwesomeIcon.Cog.ToIconString())`, or use
`ImGuiComponents.IconButton("settings", FontAwesomeIcon.Cog)`. Create `IFontHandle`s
**once** from `pi.UiBuilder.FontAtlas`, draw with `using (handle.Push())`, dispose with the
plugin; never store a raw `ImFontPtr` across frames. → `references/topic-fonts.md`

