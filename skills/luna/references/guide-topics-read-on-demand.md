## Topics (read on demand)

This file is enough to set Luna up and write a basic window. Everything deeper lives in
`references/topic-*.md`; open the file whose trigger matches **before** writing that kind of code. Paths are
relative to the skill root. Before using a member that is not shown in this file or a topic,
grep `references/api-index.txt` for it.

| Topic file | Read this when ... |
|---|---|
| `references/topic-imsharp-strings.md` | a label/text/hint parameter does not accept what you pass, you hold or build strings across calls, or you need `StringU8` / `StringPair` members |
| `references/topic-imsharp-scopes.md` | pushing colors, styles, borders, IDs, fonts, widths or disabled state; you need the list of `*Disposable` types or the `ImStyleSingle` / `ImStyleDouble` / `ImStyleBorder` members |
| `references/topic-imsharp-widgets.md` | you need any `Im.*` or `ImEx.*` widget beyond the minimal window: inputs, sliders, combos, popups, tab bars, trees, menus, tooltips, layout/cursor, item state, mouse/keyboard, drag-drop, color editors, clipboard, list clipper |
| `references/topic-imsharp-tables.md` | drawing any table: `Im.Table.Begin`, `ImEx.Table.Draw`, or a filterable/sortable `DefaultTable<T>` with column classes |
| `references/topic-imsharp-filters-combos-caches.md` | writing a searchable combo (`SimpleFilterCombo<T>`, `FilterComboBase`), a filter, or caching per-frame derived data with `CacheManager` |
| `references/topic-imsharp-drawing.md` | drawing with a draw list, handling `Rgba32` / `ColorParameter` values, showing a texture or icon, or choosing fonts |
| `references/topic-services.md` | writing the plugin entry point, registering or resolving services, implementing an `IService` marker, or using `PluginLoader` / `ErrorWindow` |
| `references/topic-luna-folders.md` | you need a Luna class outside ImSharp: `LunaStyle`, `Layout` (tab bars, two-panel layout, buttons, editors), `Widgets`, `Colors`, `Keyboard`, `Filesystem`, `Events`, `Messaging`, `Ipc`, `Files` (save service), `Json`, `Logging`, `Containers`, `DirectX`, … |
| `references/topic-generators.md` | using `[NamedEnum]`, `[StrongType<T>]`, `[ConfigProperty]`, the IPC generator attributes, or fixing a `Luna01`-`Luna04` warning |
| `references/topic-schemas.md` | serializing or validating a `Luna/Conditions` condition tree as JSON |
| `references/topic-ottergui-mapping.md` | porting from OtterGui or reading code that uses `ImUtf8`, `ImRaii` or `OtterGui.*` names |

