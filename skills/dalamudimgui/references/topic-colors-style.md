# Colors and style

Covers `ImGuiColors`, `ColorHelpers`, packed color layout, `StyleModel` and the `Easing`
animation helpers.
Verified against goatcorp/Dalamud `master` @ `43b3411` (2026-11-10, API level 15, Dear ImGui 1.88 docking).
Paths are relative to the skill root (`SKILL.md`, `references/topic-*.md`, `references/...`); a bare
`index/...` or `src/...` path means `references/index/...` or `references/src/...`.

- `ImGuiColors` (`Vector4`): `DalamudRed`, `DalamudGrey`, `DalamudGrey2`,
  `DalamudGrey3`, `DalamudWhite`, `DalamudWhite2`, `DalamudOrange`,
  `DalamudYellow`, `DalamudViolet`, `TankBlue`, `HealerGreen`, `DPSRed`,
  `ParsedGrey` .. `ParsedGold`, and semantic pairs `Info`/`Success`/`Warning`/
  `Error`/`Attention` + `Foreground`/`Background`. They follow the user's
  Dalamud style, so prefer them to literals.
- `ColorHelpers`: `RgbaVector4ToUint`, `RgbaUintToVector4`, `RgbaToHsv`,
  `HsvToRgb`, `Lighten`, `Darken`, `Saturate`, `Desaturate`, `Fade`,
  `WithAlpha`, `NormalizeToUnitRange`, `ApplyOpacity`, `SwapRedBlue`,
  `SwapEndianness`, `KnownColor.Vector()`. The `Vector4` ones are extension
  methods.
- ImGui `uint` colors are `0xAABBGGRR`. `ImGui.GetColorU32(ImGuiCol | Vector4 | uint)`
  applies the current style alpha; `ColorConvertFloat4ToU32` does not.
- `StyleModel` (`Dalamud.Interface.Style`): `GetFromCurrent()`,
  `GetConfiguredStyle()`, `GetConfiguredStyles()`, `Deserialize(string)`,
  `Serialize()`, `Apply()`, `Push()`/`Pop()`. `StyleModelV1.DalamudStandard`
  is the stock look. `Pop()` uses static counters recorded on the first
  `Push()`, so pair every `Push()` with exactly one `Pop()` in the same frame.
  Do not call `Apply()` from a plugin: it changes the global style for every
  plugin.
- Never write to `ImGui.GetStyle()` or `ImGui.GetIO()` fields directly; the
  context is shared. Use push/pop.
- `Dalamud.Interface.Animation`: `Easing` subclasses in `EasingFunctions`
  (`InOutCubic`, `OutCubic`, `InOutSine`, `OutElastic`, ...), constructed with
  a `TimeSpan`; `Start()`, call `Update()` each frame, read `ValueClamped`/
  `ValueUnclamped`/`EasedPoint`, check `IsDone` or `IsRunning`; `Elapsed`
  (`15.0.4.0` and later) gives the running time. `AnimUtil.Lerp`. Skip
  animation when `UiBuilder.ShouldUseReducedMotion`.
