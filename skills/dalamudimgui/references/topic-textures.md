# Textures and icons

Covers `ITextureProvider`: shared `GetFrom*` vs owned `CreateFrom*` textures, game icons,
`ISharedImmediateTexture`, `IDalamudTextureWrap`, `UldWrapper`.
Verified against goatcorp/Dalamud `master` @ `43b3411` (2026-11-10, API level 15, Dear ImGui 1.88 docking).
Paths are relative to the skill root (`SKILL.md`, `references/topic-*.md`, `references/...`); a bare
`index/...` or `src/...` path means `references/index/...` or `references/src/...`.

`ITextureProvider` (`Dalamud.Plugin.Services`). Two families with opposite
lifetime rules:

| | `GetFrom*` | `CreateFrom*` |
|---|---|---|
| Returns | `ISharedImmediateTexture` | `IDalamudTextureWrap` or `Task<IDalamudTextureWrap>` |
| Ownership | Dalamud's shared cache | You; **must `Dispose()`** |
| Loading | Background; empty/null until ready | Synchronous or awaited |
| Per frame | Call again every frame | Create once, store, reuse |

```csharp
// Shared: look up every frame, never store the wrap
var tex = textures.GetFromGameIcon(new GameIconLookup(item.Icon, itemHq: true));
if (tex.TryGetWrap(out var wrap, out var ex))
    ImGui.Image(wrap.Handle, ImGuiHelpers.ScaledVector2(40));
else
    ImGuiHelpers.ScaledDummy(40);              // keep layout stable while loading

ImGui.Image(textures.GetFromGame("ui/icon/060000/060001_hr1.tex").GetWrapOrEmpty().Handle, size);
ImGui.Image(textures.GetFromFile(Path.Combine(dir, "logo.png")).GetWrapOrEmpty().Handle, size);
ImGui.Image(textures.GetFromManifestResource(typeof(Plugin).Assembly, "MyPlugin.logo.png").GetWrapOrEmpty().Handle, size);
```

| `ISharedImmediateTexture` | |
|---|---|
| `GetWrapOrEmpty()` | Never null, never throws; an empty texture while loading or on failure |
| `GetWrapOrDefault(defaultWrap = null)` | Your fallback instead |
| `TryGetWrap(out wrap, out exception)` | `false` while loading or failed; `exception` tells which |
| `RentAsync(ct)` | Owned reference for long-lived use; dispose it. Marked as a last resort |

- The wrap from `GetWrap*`/`TryGetWrap` is valid **for the current frame
  only**. Do not cache it, do not dispose it (`Dispose` is ignored). Caching
  the `ISharedImmediateTexture` itself is allowed but gains little.
- These calls assert the main thread.
- `GameIconLookup(uint iconId, bool itemHq = false, bool hiRes = true, ClientLanguage? language = null)`;
  implicit from `int`/`uint`. `GetFromGameIcon` throws (`IconNotFoundException`)
  for an icon id with no file; `TryGetFromGameIcon(lookup, out tex)` and
  `TryGetIconPath` do not. Use the `Try` form for ids that come from data.
- Game and icon paths respect texture substitution (mods).
- Owned textures: `CreateFromImageAsync(bytes | stream)`,
  `CreateFromRaw(RawImageSpecification.Rgba32(w, h), bytes)`,
  `CreateFromTexFile(texFile)`, `CreateFromExistingTextureAsync`,
  `CreateFromImGuiViewportAsync`, `CreateFromClipboardAsync`,
  `CreateEmpty`, `CreateDrawListTexture`, `CreateTextureFromSeString`.
  Store the wrap in a field, guard for "not loaded yet", dispose in
  `Dispose()`. Never call a `Create*` every frame.
- `IDalamudTextureWrap`: `Handle` (`ImTextureID`), `Width`, `Height`, `Size`.
  Texture sizes are pixels; scale the draw size yourself.
- `UldWrapper` from `pi.UiBuilder.LoadUld(path)`: `LoadTexturePart(texturePath, part)`
  returns an owned wrap (nullable); check `Valid`; dispose both.
- Draw-list use: `ImGui.GetWindowDrawList().AddImage(wrap.Handle, min, max)`.
