# Luna + ImSharp

Luna: "Utility library for the Dalamud plugin ecosystem for FFXIV" (the whole
upstream README). AGPL-3.0. Repo `github.com/Ottermandias/Luna`, branch `main`,
no tags/releases. ImSharp: "Hand-make, bespoke, C#-first wrapper and extension
library for Dear ImGui" (its whole README). **MIT**, separate repo
`github.com/Ottermandias/ImSharp`, consumed by Luna as a git submodule on
branch `gc1.88`. Luna also compiles a second submodule, `Luna.Targa` (MIT,
`github.com/Ottermandias/Luna.Targa`, branch `main`), a Targa reader and writer.

Pinned versions this skill was written against:

| Thing | Value |
|---|---|
| Luna | `52eb0fe` (2026-11-13), `main` |
| ImSharp | `4d09327` (2026-11-03), branch `gc1.88` — exactly the commit Luna `52eb0fe` pins (`git ls-tree HEAD ImSharp`) |
| Luna.Targa | `8f66a2d` (2026-11-07), `main` — exactly the commit Luna `52eb0fe` pins (`git ls-tree HEAD Luna.Targa`) |
| SDK | `Dalamud.NET.Sdk/15.0.0` (`Luna/Luna.csproj`); SDK sets `net10.0-windows`. CI uses .NET `10.x.x` |
| ImGui | `Im.Version.VersionString = "1.88"`, `VersionNumber = 18800`, DLL name `"cimgui"` |
| Dalamud compared against | `43b3411` (2026-11-10), `DalamudVersion 15.0.3.9` |
| Consumers read | Penumbra `51bf831` (2026-11-13, pins Luna `52eb0fe`), Glamourer `e165910` (2026-10-08, pins Luna `dcc4ee7`, which predates the service rework) |
| OtterGui (mapping only) | `79771ee` (2026-04-29), Apache-2.0 |

Everything named below was checked against those checkouts unless marked
*(unverified)*. Nothing here was compiled or run in-game; examples use only
members whose signatures were read from source. Curated source lives in
`references/Luna/…`, `references/ImSharp/…` and `references/Luna.Targa/…` (upstream paths preserved);
`references/api-index.txt` lists every public/protected declaration in the three
repos as `Repo/path:line: declaration` — grep it before guessing a name.

Luna applies to existing Luna/ImSharp consumers, including Penumbra and Glamourer, or an explicit request to adopt it. Introducing Luna changes the UI API and adds an AGPL-3.0 dependency; preserve a plugin's existing UI library unless adoption is part of the task.

Out of scope: stock Dalamud ImGui layer (`Dalamud.Bindings.ImGui`, `ImRaii`) → sibling
`dalamudimgui` skill.
Dalamud services (`IFramework`, `IChatGui`, …) → sibling `dalamud` skill.
FFXIVClientStructs → `clientstructs` skill.


## Guide sections

Paths written as `references/...` are relative to the skill directory. Read only the section needed for the task.

- [topics read on demand](guide-topics-read-on-demand.md)
- [repo layout](guide-repo-layout.md)
- [adding luna to a plugin](guide-adding-luna-to-a-plugin.md)
- [bootstrapping context logger window system](guide-bootstrapping-context-logger-window-system.md)
- [service registration](guide-service-registration.md)
- [imsharp](guide-imsharp.md)
- [luna folders](guide-luna-folders.md)
- [lunagenerators](guide-lunagenerators.md)
- [schemas](guide-schemas.md)
- [coming from ottergui](guide-coming-from-ottergui.md)
- [luna or stock dalamud](guide-luna-or-stock-dalamud.md)
- [pitfalls](guide-pitfalls.md)

## Recent API changes

The bundled sources moved from Luna `ff4467b` / ImSharp `fcf41fc` to Luna `52eb0fe` / ImSharp
`4d09327` on 2026-11-13
([Luna compare](https://github.com/Ottermandias/Luna/compare/ff4467b...52eb0fec6795b10c7e389c268486def9d18a9c36),
[ImSharp compare](https://github.com/Ottermandias/ImSharp/compare/fcf41fc...4d09327e3d7ea7d4cfda7fef2a0ca04dddda5cee)).
A consumer pinned to an older Luna commit still has the left-hand side of each item below, so
check its submodule before applying one. These are source comparisons, without a consumer build or
in-game test.

Services ([details](topic-services.md)):

- `IService` splits into `ISingletonService` and `IScopedService`. `IUiService` and `IApiService`
  are scoped, `IDataContainer` is a singleton, and `IAsyncService` no longer derives from
  `IService`. A Debug build of Luna throws when a registered `IService` type picks none or several
  of the lifetimes.
- `ServiceManager` adds `AddScoped<T>()`, `EnterScope()`, `ExitScope()`, `InScope`, `HasProvider`,
  `GetService(Type)` and `DisposeAsync()`. `EnsureRequiredServices()` now takes a
  `CancellationToken`. The public `Provider` property and `AddIServices<TInterface>(Assembly)` are
  gone.
- `ErrorWindow` takes the `ImSharpDalamudContext` as a fifth constructor argument and no longer
  creates its own; `DrawFailure()` is a new virtual.
- `IPluginLoader` and `PluginLoader<TPlugin>` add `GameVersion`.
- `HookManager` and `FastHook<T>` moved from `Luna/Services/` to `Luna/GameInterop/` (same
  namespace `Luna`). `HookManager.Diagnostics` is replaced by `Hooks` (`IEnumerable<IHookData>`)
  and the `HookStateChanged` event; `HookManagerDrawer` draws them.
- `DayChangeTracker` is an instance class implementing `IDayChangeTracker` and `IDisposable`
  (constructors take an optional `LunaLogger` and a `TimeProvider`). The static `DayChanged` event
  is gone.

Other Luna changes:

- `EventBase` drops the `ReaderWriterLockSlim`, the `Event` list and `Enumerate()`. Subscribers
  live in a copy-on-write array that `Invoke` reads without locking.
- `BasicWrapper` removes `VersionMethod`, `IsDisposedMethod` and `DisposedEventMethod`. Its
  wrappers use the adapter's `Version`, `IsDisposed` and `Disposed` members directly, and the
  wrapper's version and disposal properties are no longer virtual. `AdapterVersionMismatchException`
  and `ThrowAdapterVersionMismatch` report incompatible adapters, and
  `PluginSubscriber.InitializationFailure` records an attachment exception.
- The `IEnumerable<T>` helpers `IndexOf` and `FindFirst` moved into `IndexOfExtensions` and
  `FindFirstExtensions`, with array, list and context-passing overloads. New: `ArrayExtensions`
  (`Repeat`, `Without`), `VersionExtensions` (`Equals(major, minor[, build[, revision]])`) and
  `ReadOnlyWrapper<…>`.
- `BaseButton<T>` adds the virtuals `MenuItem(in T)` and `Button(Vector2, in T)`.
  `FileSystemDrawer` adds `DragDropAutoScrollMargin` and scrolls its list while a node is dragged
  near an edge.
- `ConfigurationFile<TProvider>` no longer queues a save while it is loading.
- Luna compiles a second submodule, [Luna.Targa](https://github.com/Ottermandias/Luna.Targa)
  (MIT, namespace `Luna.Targa`, a Targa reader and writer), into `Luna.dll`. Its source is
  bundled under `references/Luna.Targa/`.

ImSharp:

- `ImEx.Table.ActiveFilterColor` is removed. Set `ImSharpConfiguration.Colors.ActiveFilterBackground`
  and `ActiveFilterBorder`, or override `ActiveBackgroundColor` / `ActiveBorderColor` on a filter.
  Luna's `ColorCache` fills both from Dalamud's attention colors.
- New: `Im.ContentRegion.Width(y)` and `Height(x)`, `Im.DragDrop.AutoScroll(speed, marginHeight)`,
  `StringU8.EmptyLazy`, and equality and comparison on `StringPair`.
- The `HeaderDisposable` constructor that takes `ref bool visible` now sets `Alive`, so
  `Im.Tree.HeaderId` with a close button pops its ID. Pins before ImSharp `162c310` miss that pop.
- `Im.Style.TextHeight`, `TextHeightWithSpacing`, `FrameHeight` and `FrameHeightWithSpacing` are
  computed from the cached context and style instead of a native call.
