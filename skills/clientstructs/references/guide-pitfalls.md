## Pitfalls

- **Null everywhere — three different states, three different checks.**
  1. *Unresolved sig*: every `[StaticAddress]`/`[MemberFunction]` throws
     `ThrowNullAddress` at call time; it never returns `null`
     (`InteropGenerator.Rendering.cs` `RenderStaticAddresses`). Catch or
     verify sig health after patches.
  2. *Null heap singleton*: `isPointer: true` statics (`Framework`, `AtkStage`)
     return the dereferenced pointer, which is `null` until the game creates
     the object; `UIModule.Instance()` returns `null` when `Framework.Instance()`
     is (then chains `GetUIModule()`); `GetAddonByName` returns `null` when
     closed; `IndexSorted` slots are null; `StatusManager.Owner` is documented
     as often null. Check before `->`.
  3. *Present but not loaded*: `isPointer: false` statics (`PlayerState`,
     `InventoryManager`, `Conditions`, `Telepo`) are objects in `.data`, so
     `Instance()` is non-null on the title screen too — a non-null address says
     nothing about player data. Gate on `PlayerState.Instance()->IsLoaded`
     (offset 0), `InventoryContainer.IsLoaded` /
     `InventoryManager.IsContainerLoaded`, or Dalamud `IClientState.IsLoggedIn`
     / `IPlayerState.IsLoaded` (both in the `dalamud` skill's service table).
- **`partial` is mandatory** for any struct with generator attributes
  (`CSIG0001`), including containing structs (`CSIG0002`); a struct with
  `[GenerateInterop]` must not be nested in a class (`CSIG0003`) and must have
  `[StructLayout]` (`CSIG0004`).
- **Fixed arrays are `Span<T>` now, not `fixed T[]`**. `_name` backing fields
  are `internal`; use the generated `Name` span / `NameString`. Old code using
  `fixed byte Name[64]` + `MemoryHelper.ReadStringNullTerminated` no longer
  compiles against current CS *(migration date unverified; the only `fixed`
  buffer left in the curated set is `StdString.Buffer`)*.
- **Pointers in generics** must be `Pointer<T>` (`StdVector<Pointer<AtkResNode>>`,
  `Span<Pointer<GameObject>>`); `.Value` or implicit conversion to get `T*`.
- **`bool` params**: CS methods take real `bool` (assembly has
  `DisableRuntimeMarshalling`); ECommons maps `bool` AtkValues to `Byte`.
- **Strings**: never store a `CStringPointer`/`byte*` from the game; copy
  with `.ToString()` immediately. `Utf8String` has three ownership cases
  (`Utf8String.cs`):
  1. *Local value* — `using var s = new Utf8String("x");`. The ctor calls the
     game `Ctor()`; `using` runs `IDisposable.Dispose()` = `Dispose(false)` =
     `Dtor()` only, which frees the string buffer, never the struct (it is on
     your stack). Never `Dtor(true)` a local.
  2. *Owned heap pointer* — `Utf8String.FromString(...)`, `FromSequence`,
     `FromUtf8String`, `CreateEmpty` allocate the struct itself via
     `IMemorySpace.Create<Utf8String>()`. Unless the callee took ownership,
     end with `str->Dtor(free: true)` (= `Dtor()` + `IMemorySpace.Free(str)`):
     `var msg = Utf8String.FromString("hi"); try { fn(msg); } finally { msg->Dtor(true); }`.
     Plain `Dtor()`/`Dispose` on it destroys the contents but leaks the
     0x68-byte struct.
  3. *Borrowed, game-owned* — `->NodeText`, `Director.Title`, anything
     returned by or embedded in a game struct: read (`.ToString()`), never
     destroy.
  `string` overloads exist only for
  `[GenerateStringOverloads]` methods (e.g. `GetAddonByName(string)`,
  `Utf8String.SetString(string)`, `AtkValue.SetManagedString(string)`).
- **Namespace moves** happen (the checkout groups by C++ namespace: e.g.
  `RaptureShellModule` is `Client.UI.Shell`, `RaptureGearsetModule` is
  `Client.UI.Misc`, `PlayerState/UIState/Telepo/TerritoryInfo` are
  `Client.Game.UI`, `Conditions` is `Client.Game`, `Control/TargetSystem` are
  `Client.Game.Control`). When a type vanishes, grep
  `references/all-structs.txt` — it was probably moved or renamed with an
  `[Obsolete]` shim *(specific historical moves unverified)*.
- **Main thread only.** Game structures are not thread-safe; call from
  `Framework.Update` / Dalamud `IFramework.RunOnFrameworkThread`. A plugin
  constructor and `Task.Run` bodies are not on that thread (threading rules:
  `dalamud` skill).
- **Account and content IDs.** `Character.AccountId` and `ContentId` are
  readable here, but Dalamud's official-repo rules forbid collecting other
  players' account IDs and persisting their content IDs unhashed (`dalamud`
  skill, Publishing). Reading a field being possible does not make the plugin
  acceptable.
- **Prefer Dalamud/ECommons wrappers** where they exist: `IGameGui.GetAddonByName`,
  `IObjectTable`, `IClientState`, `ICondition`, `IPlayerState`; ECommons
  `Svc.*`, `GenericHelpers.TryGetAddonByName<T>`, `IsAddonReady`,
  `AddonMaster.*`, `Callback.Fire`, `Chat.SendMessage`. Drop to raw CS for
  members those don't expose (AtkValues, node trees, agents, `InventoryManager`,
  `ActionManager`, `QuestManager`).
- **AtkValue index maps are per addon and per patch.** Log them
  (`Callback.InstallHook()` or a `FireCallback` hook) before hardcoding.
- **`GetAddonByName` index is 1-based** (`index = 1` default); the second
  `SelectYesno` is `index: 2`.
