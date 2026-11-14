## Pitfalls

| Pitfall | Consequence / fix |
|---|---|
| Unbalanced push/pop or Begin/End | The ImGui context is shared by Dalamud and **all** plugins; a leaked style, ID or font corrupts everyone's UI, a missing End asserts or crashes. Use `ImRaii` with `using` everywhere |
| `using var x = ImRaii.Table(...)` without `if (!x)` | Drawing table cells with no table: native assert |
| `using (ImRaii.PushColor(...));` with a stray `;` or discarding the result | Scope ends immediately or never |
| `ImRaii.Header(label, ref visible, flags)` on a Dalamud build before `15.0.3.8` | Leaks one ID push per open header; fixed in the bundled revision (see `references/topic-imraii.md`) |
| Same label twice in one ID scope (`"Delete"` per row, empty labels, repeated icons) | Only the first responds. `##id` suffix, `ImRaii.PushId(i)`, or the `id` overloads |
| Window titles that change (`$"Items ({n})"`) | New window identity each change: position/size lost. `$"Items ({n})###items"` |
| Two windows with one `WindowName` in a `WindowSystem` | `AddWindow` throws. Across plugins, identical names merge into one ImGui window: give `WindowSystem` a namespace and use `###` IDs |
| Hard-coded pixel sizes | Broken layout at non-100 % scale. `* ImGuiHelpers.GlobalScale` |
| Caching `CalcTextSize`/layout results forever | Stale after scale or font change; invalidate on `DefaultGlobalScaleChanged`/`DefaultFontChanged` |
| Game-state work in `Draw` (object table scans, inventory walks, hooks, IPC) | Frame drops; also `Draw` does not run while the UI is hidden, so logic placed there silently stops. Compute in framework update (sibling `dalamud` skill), draw from cached state |
| Mutating a collection in `Draw` while another thread fills it | `InvalidOperationException` mid-frame and the plugin's error window. Swap immutable snapshots |
| `ImGui.*`, `IFontHandle.Push`, `TryGetWrap` from a `Task`, timer or hook on another thread | Asserts or crashes; ImGui is single-threaded. Only touch ImGui inside `Draw` and the window callbacks |
| Storing `ImFontPtr`, `ImDrawListPtr`, `ImGuiViewportPtr` or `*U8()` spans across frames | Dangling pointers after atlas rebuild or frame end |
| `ImGui.PopFont()` after `fontHandle.Push()` | Breaks the handle's bookkeeping (it logs a warning); use the returned disposable or `Pop()` |
| Creating font handles or textures inside `Draw` | Atlas rebuild or GPU upload every frame, leaked handles |
| Not disposing `IFontHandle`, owned `IDalamudTextureWrap`, `UldWrapper`, custom `IFontAtlas` | GPU/atlas resources leak; pass `debugName` to `Create*` so leaked textures are identifiable in Dalamud's texture debug widget |
| Disposing or caching wraps from `GetWrapOrEmpty`/`TryGetWrap` | Dispose is ignored; cached wraps can stop being valid on any later frame |
| Not unsubscribing `UiBuilder.Draw`/`OpenConfigUi`/`OpenMainUi` in `Dispose` | Callbacks into a disposed plugin |
| `InputText` `maxLength` too small | Input silently truncated; it is a UTF-8 **byte** count (buffer is `maxLength + 1`), so CJK text fits a third as many characters |
| Using Dear ImGui APIs newer than 1.88 from tutorials | Do not exist; grep the index |
| Writing `ImGui.GetIO().ConfigFlags`, fonts, `GetStyle()` values | Global side effects for every plugin |
| Throwing from `Draw` | `Window.Draw`: that window shows an error. Raw `UiBuilder.Draw` handler: your whole UI is replaced by the error window |
