## Pitfalls

- **Clone recursively.** `Luna/ImSharp` and `Luna/Luna.Targa` are submodules; without ImSharp
  `Luna.csproj` compiles no `Im` at all. Luna's `.gitmodules` uses ssh URLs for both.
- **Do not add a reference to `ImSharp.csproj`.** Its sources are already inside `Luna.dll`; a
  second copy produces duplicate `ImSharp.*` types.
- **Nothing formats before the context exists.** Buffers come from `ImSharpContext`; with the empty
  context the buffer pointers are null. Construct `ImSharpDalamudContext` first and draw through
  `Luna.WindowSystem`, which waits for `ImSharpConfiguration.IsInitialized`. The context is set on a
  later framework tick.
- **Drawing outside Dalamud's `Draw`** (your own `IUiBuilder.Draw` handler that runs before
  ImSharp's): `Im.Style`, `Im.Io`, `Im.Context` are cached pointers refreshed by
  `ImSharpPerFrame.OnUpdate()`.
- **Same-kind interpolated strings clobber each other.** Each handler kind has one shared buffer.
  Never hold a `Utf8TextHandler` across another call that formats text.
- **`u8` for literals.** A plain `"literal"` works but is transcoded every frame and triggers
  `Luna01`.
- **Labels must be null-terminated.** A sliced `ReadOnlySpan<byte>` is not. Wrap it in
  `new StringU8(span)` or interpolate it.
- **Check the scope result before drawing into it**, and return early: `if (!table) return;`. With
  `STATECHECKS` undefined (the default; it is commented out in `ImSharp.csproj`)
  `ImGuiStateException.CheckState` compiles away and misuse goes straight to native ImGui.
- **`using var` order matters.** Disposal is reverse declaration order; a style pushed before
  `Im.Table.Begin` stays active for the whole table. Use a `using (…) { }` block to limit it.
- **Stack disposables are classes.** Chain pushes on one object instead of allocating one per push
  in a hot loop.
- **Leaked pushes are reported, not fixed.** `Luna.WindowSystem` logs a stack imbalance at most
  every 5 seconds per kind.
- **`Filesystem/Selector/` is dead code** (excluded in `Luna.csproj`), though it is in the repo and
  in `references/`.
- **`GetItems()` is cached.** Tables and combos do not see data changes until their cache is marked
  `DirtyFlags.Custom`.
- **Enum member names differ from ImGui's** (`Background` not `Bg`, `Thickness` not `BorderSize`,
  `Minimum`/`Maximum`). Grep `references/ImSharp/ImSharp/ImSharp/Enums/`.
- **`Rgba32` is ABGR in the `uint`** (`0xAABBGGRR`), like ImGui's `ImU32`.
- **`Window` is ambiguous** when both `Luna` and `Dalamud.Interface.Windowing` are imported; alias
  it. `WindowSizeConstraints` still comes from Dalamud.
- **Generator attributes are internal per assembly.** A library project using `[NamedEnum]` needs
  its own analyzer reference.
- **Glamourer and Penumbra pin different Luna commits.** At the revisions read here Penumbra is on
  the service rework and Glamourer is not. When copying a pattern from a consumer, check the member
  against the Luna commit you build.
- **Luna01–Luna04 are warnings**; with warnings-as-errors they fail the build.
- **Every `IService` type needs exactly one lifetime marker** (`ISingletonService`, `IScopedService`
  or `IConstructedService`). A Debug build of Luna throws on registration otherwise; a Release build
  registers a bare `IService` as a singleton without complaint. `IUiService` and `IApiService` are
  already scoped.
- **Scoped services end with their scope.** Without `EnterScope()` they resolve from the
  root provider and are disposed with the manager. After `ExitScope()` they are disposed, so do not
  keep a reference to one in a singleton.
- **`EnsureRequiredServices` takes a `CancellationToken`** from Luna `52eb0fe` on; code copied from a
  consumer on an older pin calls it without one.
