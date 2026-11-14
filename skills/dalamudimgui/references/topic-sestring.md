# SeString rendering

Covers drawing game text (colors, icons, links) with `SeStringWrapped`,
`CompileSeStringWrapped`, `CenteredSeString` and `SeStringDrawParams`.
Verified against goatcorp/Dalamud `master` @ `43b3411` (2026-11-10, API level 15, Dear ImGui 1.88 docking).
Paths are relative to the skill root (`SKILL.md`, `references/topic-*.md`, `references/...`); a bare
`index/...` or `src/...` path means `references/index/...` or `references/src/...`.

Game text with colors, icons, links and soft hyphens:

```csharp
ImGuiHelpers.SeStringWrapped(readOnlySeString, new SeStringDrawParams { Color = 0xFFFFFFFF, WrapWidth = 300 });

var r = ImGuiHelpers.CompileSeStringWrapped(
    "<icon(1)> <colortype(710)>Warning<colortype(0)>: text", default, "myLink"u8);
if (r.Clicked) { /* r.InteractedPayloadOffset, r.InteractedPayloadEnvelope */ }
```

- `SeStringWrapped(ReadOnlySpan<byte> sss, in SeStringDrawParams style = default, ImGuiId imGuiId = default, ImGuiButtonFlags buttonFlags = MouseButtonDefault)`
  draws encoded bytes (Dalamud itself passes a Lumina `ReadOnlySeString`
  directly); `CompileSeStringWrapped(string text, ...)` compiles macro text
  and caches the result. Always close a `colortype`/`edgecolortype` push with
  `<colortype(0)>`.
- Pass an `ImGuiId` (implicit from string, UTF-8 span, integers, pointers)
  to make links interactive. Result: `Size`, `Clicked`,
  `InteractedPayloadOffset`, `InteractedPayloadEnvelope`, `InteractedPayload`.
- `SeStringDrawParams` (all optional): `Font`, `FontSize`, `LineHeight`,
  `WrapWidth`, `Opacity`, `Color`, `EdgeColor`, `ShadowColor`, `Bold`,
  `Italic`, `Edge`, `Shadow`, `ThemeIndex`, link colors, `GetEntity`,
  `TargetDrawList`, `ScreenOffset`. Colors are `uint` as ImGui packs them.
- Setting `TargetDrawList` means no `Dummy` is emitted and the ID is ignored,
  and since API 14 you must also set `Font`, `FontSize` and `ScreenOffset`.
- `CenteredSeString(ReadOnlySpan<byte> sss, scoped in SeStringDrawParams style = default)`
  (first tagged in `15.0.3.8`) centers the text horizontally in the current
  window. With `TargetDrawList` set it centers inside the box that starts at
  `ScreenOffset` and is `WrapWidth` wide; both must be set or it throws
  `ArgumentException`.
- `ITextureProvider.CreateTextureFromSeString(text, drawParams, debugName)`
  renders to an owned texture (main thread).
- `ImGui.Text` does not understand SeString payloads or soft hyphens.
