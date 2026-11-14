# Fonts

Covers FontAwesome icons, `IFontAtlas` / `IFontHandle`, the build toolkit, game fonts, glyph
ranges and the font chooser dialog.
Verified against goatcorp/Dalamud `master` @ `43b3411` (2026-11-10, API level 15, Dear ImGui 1.88 docking).
Paths are relative to the skill root (`SKILL.md`, `references/topic-*.md`, `references/...`); a bare
`index/...` or `src/...` path means `references/index/...` or `references/src/...`.

## Icons

```csharp
using (ImRaii.PushFont(UiBuilder.IconFont))                 // or pi.UiBuilder.IconFontHandle.Push()
    ImGui.Text(FontAwesomeIcon.Cog.ToIconString());

if (ImGuiComponents.IconButton("settings", FontAwesomeIcon.Cog)) this.config.Toggle();
```

`FontAwesomeIcon` is an enum of code points (Font Awesome 7.1.0 free solid;
API 14 removed `InstagramSquare` and `VectorSquare`). `ToIconChar()` returns
the `char`, `ToIconString()` allocates a new one-character string on every
call, so cache it or interpolate the char (`$"{icon.ToIconChar()}"`).
`FontAwesomeHelpers.GetIcons()`, `SearchIcons(search, category)`,
`GetCategories()` enumerate them. Text pushed with the icon font can only
show icons; pop before drawing normal text. `IconFontFixedWidth` has uniform
advance and vertically centred glyphs.

## Font handles

All custom fonts go through the plugin's `IFontAtlas`
(`pi.UiBuilder.FontAtlas`, auto-rebuild mode `Async`). Create handles **once**
(constructor), keep them, dispose them with the plugin.

```csharp
private readonly IFontHandle bigFont;
private readonly IFontHandle axis18;
private readonly IFontHandle custom;

// constructor
var atlas = pi.UiBuilder.FontAtlas;
using (atlas.SuppressAutoRebuild())                                  // one rebuild for all three
{
    this.bigFont = atlas.NewDelegateFontHandle(
        e => e.OnPreBuild(tk => tk.AddDalamudDefaultFont(UiBuilder.DefaultFontSizePx * 2)));

    this.axis18 = atlas.NewGameFontHandle(new GameFontStyle(GameFontFamilyAndSize.Axis18));

    this.custom = atlas.NewDelegateFontHandle(e => e.OnPreBuild(tk =>
    {
        var config = new SafeFontConfig { SizePx = 20 };
        config.MergeFont = tk.AddFontFromFile(fontPath, config);      // base font
        tk.AddGameSymbol(config);                                     // merge game symbols
        tk.AttachExtraGlyphsForDalamudLanguage(config);               // CJK fallback for the UI language
        tk.Font = config.MergeFont;                                   // the font this handle resolves to
    }));
}

// Draw
using (this.bigFont.Push())
    ImGui.Text("Heading"u8);

// Dispose
this.bigFont.Dispose(); this.axis18.Dispose(); this.custom.Dispose();
```

| `IFontHandle` | |
|---|---|
| `Push()` | Returns `IDisposable`; main thread only (asserts). Works even when the font is not built yet or failed: it then leaves the current font in place |
| `Pop()` | Matches a bare `Push()`. **Never** `ImGui.PopFont()` for a handle push |
| `Available`, `LoadException`, `WaitAsync([ct])` | Build state |
| `Lock()` / `TryLock(out error)` | `ILockedImFont` with `.ImFont`; must be disposed; for metrics or draw-list text |
| `ImFontChanged` | May fire off the main thread |

| `IFontAtlasBuildToolkitPreBuild` | |
|---|---|
| `AddDalamudDefaultFont(sizePx, glyphRanges = null)` | Default font + icon font + language glyphs. Negative `sizePx` = multiplier of the default size (`-1` = default) |
| `AddFontFromFile(path, in config)`, `AddFontFromMemory(span, in config, debugTag)`, `AddFontFromStream(stream, in config, leaveOpen, debugTag)` | TTF/OTF |
| `AddDalamudAssetFont(DalamudAsset, in config)`, `AddFontAwesomeIconFont(in config)` | Bundled fonts |
| `AddGameSymbol(in config)`, `AddGameGlyphs(GameFontStyle, glyphRanges, mergeFont)` | Game font glyphs |
| `AttachExtraGlyphsForDalamudLanguage(in config)` | The `IFontAtlas` XML example calls this `AddExtraGlyphsForDalamudLanguage`; that name does not exist |
| `SetFontScaleMode(font, FontScaleMode)` | `Default`, `SkipHandling`, `UndoGlobalScale` |
| `Font`, `Scale`, `IsAsyncBuildOperation`, `DisposeWithAtlas(...)`, `DisposeAfterBuild(...)`, `RegisterPostBuild(action)` | |

Post-build (`e.OnPostBuild(tk => ...)`) offers `FitRatio`, `BuildLookupTable`,
`StoreTexture`, and is where cross-font glyph copying goes.

- Build delegates run on a worker thread (`Async`); do not touch ImGui state
  or game state in them, only the toolkit.
- Calling `NewGameFontHandle`/`NewDelegateFontHandle` from inside a build
  callback throws `InvalidOperationException`.
- Sizes passed to the toolkit are unscaled; the atlas applies global scale
  when `IsGlobalScaled` (default).
- `GameFontStyle(GameFontFamily family, float sizePx)` picks the nearest baked
  size; families: `Axis`, `Jupiter`, `JupiterNumeric`, `Meidinger`,
  `MiedingerMid`, `TrumpGothic`. `Bold`/`Italic` are synthesized. Game fonts
  look blurry away from their baked sizes (see `GameFontFamilyAndSize`).
- `SafeFontConfig.GlyphRanges`: build with `"abc".ToGlyphRange()`,
  `UnicodeRanges.X.BeginGlyphRange().With(...).Build()`
  (`FluentGlyphRangeBuilder`).
- `pi.UiBuilder.CreateFontAtlas(FontAtlasAutoRebuildMode.Disable | OnNewFrame | Async, ...)`
  gives an isolated atlas when you rebuild fonts often; with `Disable` you
  call `BuildFontsAsync()`/`BuildFontsOnNextFrame()` yourself. Dispose it.
- User font choice: `SingleFontChooserDialog.CreateAuto((UiBuilder)pi.UiBuilder)`
  (takes the concrete class); await `ResultTask` for a `SingleFontSpec`,
  then `spec.CreateFontHandle(atlas)`.
- Raw `ImFontPtr` values (`UiBuilder.DefaultFont`, `ImGui.GetFont()`,
  `locked.ImFont`) are invalidated by atlas rebuilds: never store one in a
  field across frames.
