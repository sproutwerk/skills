## Textures and icons

`ITextureProvider.GetFrom*` is a shared cache: call it every frame and never store or
dispose the wrap (`textures.GetFromGameIcon(new GameIconLookup(id)).GetWrapOrEmpty().Handle`).
`CreateFrom*` is owned: create once, store, **must `Dispose()`**. → `references/topic-textures.md`

