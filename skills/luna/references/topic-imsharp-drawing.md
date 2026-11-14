# ImSharp draw lists, colors, IDs, textures and fonts

Covers `Im.DrawList` (shapes, paths, text, images, clip rects), `Rectangle`, `Rgba32`,
`ColorParameter`, `ImTextureId`, fonts and FontAwesome icons. Verified against Luna `52eb0fe` and
ImSharp `4d09327` (branch `gc1.88`, ImGui 1.88). Paths are relative to the skill root. Grep
`references/api-index.txt` before using a member that is not shown here.

## Draw lists

`Im.DrawList` (readonly struct over `Im.Native.ImDrawList*`) from
`Im.Window.DrawList`, `Im.Drawing.ForegroundDrawList`,
`Im.Drawing.BackgroundDrawList`, `Im.Viewport.Main.Foreground/Background`.

| Group | Members |
|---|---|
| `drawList.Shape` | `Line`, `Rectangle`, `RectangleFilled`, `RectangleMulticolor`, `Quad`, `QuadFilled`, `Triangle`, `TriangleFilled`, `Circle`, `CircleFilled`, `NGon`, `NGonFilled`, `Polyline`, `ConvexPolygonFilled`, `BezierCubic`, `BezierQuadratic`. Colors are `ColorParameter`; an omitted color is white (`Line` does `color.CheckDefault(Rgba32.White)`) |
| `drawList.Path` | fluent `LineTo`, `ArcTo`, `ArcToFast`, `BezierCubicCurveTo`, `BezierQuadraticCurveTo`, `Rectangle`, then `FinishStroke(color, flags, thickness)` or `FinishFillConvex(color)`; `Clear()` |
| `drawList.Render` | `Arrow`, `Bullet`, `Checkmark` |
| Direct | `Text(position, Rgba32 color, text)`, `Text(font, fontSize, position, color, text, wrapWidth)`, `TextClipped`, `Image`, `ImageQuad`, `ImageRounded`, `PushClipRect(…)`, `PushClipRectFullScreen()`, `Splitter` (`Split`, `SetChannel`, `Merge`) |

Geometry helper: `Rectangle(Vector2 Minimum, Vector2 Maximum)` with
`FromSize`, `Contains`, `Expand`, `Size`, `Center`, `Overlaps`.

## Colors, IDs, textures, fonts

- `Rgba32(uint Color)`: ABGR-packed like ImGui (`Red = 0xFF0000FF`). Implicit
  from `uint`, `Vector4`, `Vector3`. `ToVector()`, `WithAlpha`, `FullAlpha`,
  `HalfTransparent`, `Mix`, `Tinted`, `ContrastColor`, `TryRead`. Named
  constants `Transparent`, `Black`, `White`, `Red`, `Green`, `Blue`, `Cyan`,
  `Magenta`, `Yellow`, `Gray`.
- `ColorParameter(Rgba32? Color = null)`: "use the default" when empty.
  Implicit from `uint?`, `Rgba32`, `Rgba32?`, `Vector4`, `ImGuiColor`. Pushing
  a default `ColorParameter` pushes nothing.
- `ImGuiColor.X.Push(…)`, `.PushDefault()`, `.Get(alphaMultiplier)` are C# 14
  extension members on the enum.
- `ImTextureId(nint Value)`. From Dalamud: `IDalamudTextureWrap.Id` (extension
  property in `Luna/Extensions/TextureWrapExtensions.cs`). Back:
  `id.ToDalamudId()`. Draw with `Im.Image.Draw(wrap.Id, size)`.
  `Luna/GameInterop/SafeTextureHandle` wraps a game `Texture` as an
  `IDalamudTextureWrap`.
- Fonts come from Dalamud through `ImSharpDalamudContext`. FontAwesome icons:
  `Luna.AwesomeIcon` (implicit from `FontAwesomeIcon`) implements
  `IIconStandIn`; `ImEx.Icon.Draw(icon)`, `ImEx.Icon.Button(icon, tooltip, …)`,
  `ImEx.Icon.LabeledButton(icon, label, …)`, `ImEx.Icon.CalculateSize(icon)`.
  `FontAwesomeIcon` extension methods `.Icon()`, `.Draw()`, `.Button()` exist
  in `Luna/Extensions/AwesomeIconExtensions.cs`.
