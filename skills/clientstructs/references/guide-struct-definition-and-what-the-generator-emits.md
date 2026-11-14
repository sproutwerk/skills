## Struct definition and what the generator emits

```csharp
// FFXIV/Client/Game/UI/Telepo.cs — a complete real definition
[GenerateInterop]                                   // required for any generator feature
[Inherits<AtkEventInterface>]                       // base class(es); flattened into this struct
[StructLayout(LayoutKind.Explicit, Size = 0x58)]    // every struct is explicit-layout, sized
public unsafe partial struct Telepo {               // must be partial + unsafe
    [StaticAddress("48 8D 0D ?? ?? ?? ?? 48 8B 12", 3)]
    public static partial Telepo* Instance();       // static object in .data (isPointer:false)

    [FieldOffset(0x10)] public StdVector<TeleportInfo> TeleportList;
    [FieldOffset(0x51)] public bool ActiveTeleportRequest;

    [MemberFunction("E8 ?? ?? ?? ?? 48 8B 4B 10 84 C0 48 8B 01 74 2C")]
    public partial bool Teleport(uint aetheryteId, byte subIndex);
}
```

Attributes (all in `InteropGenerator.Runtime.Attributes`, global-using'd):

- `[GenerateInterop(isInherited: false)]` — mark the struct. `isInherited: true`
  when other structs `[Inherits<T>]` it (needed so its fields/vfuncs get
  flattened into children).
- `[Inherits<T>(parentOffset = 0)]` (AllowMultiple) — child gets a parent
  accessor field named `T` (or `TBase` on name clash) at the offset, and every
  public field/method of `T` re-emitted at `offset + field.Offset`. So
  `BattleChara` (`Inherits<Character>`, `Character` is
  `Inherits<GameObject>, Inherits<CharacterData>`) exposes `->Position`,
  `->Health`, `->GetStatusManager()` directly, and also `->Character`,
  `->GameObject`, `->CharacterData` sub-structs.
- `[VirtualFunction(n)]` on `partial` method — call through
  `VirtualTable->Name(this, …)`; generator emits `XVirtualTable` struct with
  `[FieldOffset(n*8)] delegate* unmanaged<X*, …>` and
  `[FieldOffset(0)] XVirtualTable* VirtualTable`.
- `[VirtualTable("sig", offset, functionCount)]` on struct — static vtable
  address; gives `StaticVirtualTablePointer` (hook vfuncs via
  `IGameInteropProvider.HookFromAddress<…>((nint)Addon.StaticVirtualTablePointer->OnSetup, detour)`;
  the static `Hook<T>.FromAddress` is `internal` in Dalamud. For addon and
  agent vfuncs prefer `IAddonLifecycle` / `IAgentLifecycle` over a hook).
- `[MemberFunction("sig")]` / `[MemberFunction("sig", relativeFollowOffset(s))]`
  — non-virtual member or `static` function. Sig starting `E8`/`E9` auto-follows
  the rel32 at offset 1. Generated body:
  `if (MemberFunctionPointers.Name is null) ThrowHelper.ThrowNullAddress(...)`
  then `MemberFunctionPointers.Name((X*)Unsafe.AsPointer(ref this), args)`.
- `[StaticAddress("sig", offset, isPointer)]` — `offset` = index of the rel32
  displacement in the sig (usually first `??`). `isPointer: true` when the
  static is `T*` (heap singleton: `Framework`, `AtkStage`), false when the
  object itself is in `.data` (`InventoryManager`, `PlayerState`, `Conditions`).
- `[GenerateStringOverloads]` + `[StringIgnore]` on params — for
  `CStringPointer`/`byte*` params, emits `string` and `ReadOnlySpan<byte>`
  overloads (stackalloc ≤511 bytes). Never for returns.
- `[FixedSizeArray]` / `[FixedSizeArray(isString: true)]` /
  `[FixedSizeArray(isBitArray: true, bitCount: n)]` on
  `internal FixedSizeArrayN<T> _camelName;` → generated
  `[UnscopedRef] public Span<T> CamelName`, plus `string CamelNameString`
  (get/set, throws if too long) for strings, plus `BitArray CamelNameBitArray`
  for bit arrays. `FixedSizeArrayN` types are generated on demand
  (`[InlineArray(N)]`).
- `[BitField<T>(nameof(Prop), index, length = 1)]` on the backing field +
  `public partial bool Prop { get; set; }` → `BitOps` accessors. Example:
  `AtkUnitBase.IsReady` = bit 0 of `Flags1A0`.
- `[Addon("Name", …)]` (`FFXIVClientStructs.Attributes`) — names the addon(s)
  a struct maps to; 204 files use it. `[Agent(AgentId.X)]` — 163 files; generates
  `AgentX.Instance()` and `AgentModule.GetAgentX()`. `[InfoProxy(InfoProxyId.X)]`
  likewise.
- `[CExporterIgnore]` (the real name; there is no `CExportIgnore`),
  `[CExporterUnion("Name")]` for overlapping fields, `[CExporterExcel("Sheet")]`
  for `void*` params that are Excel rows, `[CExporterForce]`,
  `[CExporterTypeForce]`. These only affect the C header export/validation.
- `[Obsolete("Renamed to X")]` — old name kept at the same offset alongside
  the new field for one cycle (`InventoryManager.ItemLevelInventoryType` →
  `EquippedItemsInventoryType`; `GameObject.YalmDistanceFromPlayerX` →
  `CurrentDistance`). Generated `Addresses.Register()` wraps in
  `#pragma warning disable 612, 618`.
- `ICreatable<T>` + a `Ctor()` → `IMemorySpace.GetDefaultSpace()->Create<T>()`
  allocates with game allocators (UI objects, `Utf8String`, `InventoryItem`).

Per-struct generated surface (see `references/generator/InteropGenerator.Rendering.cs`):
`public const int StructSize`, `static class Addresses { Address Name = new(...) }`,
`static class MemberFunctionPointers`, `static class StaticAddressPointers`,
`static partial class Delegates { delegate R Name(X* thisPtr, …) }` (use these
for Dalamud hooks: `Hook<AtkUnitBase.Delegates.FireCallback>`), the vtable
struct, and the method bodies. Assembly-wide:
`FFXIVClientStructs.Interop.Generated.Addresses.Register()/Unregister()`.

Types you will meet: `Pointer<T>` (8-byte wrapper because C# forbids `T*` as a
generic arg; implicit both ways, `.Value`, `.IsNull`), `CStringPointer`
(`byte*` with `.ToString()`, `.HasValue`, `.AsSpan()`), `Utf8String` (0x68,
game std::string-alike: `StringPtr`, `Length`, `AsSpan()`, `ToString()`,
`SetString(string)`, `Clear()`, `EqualToString(string)`; `new Utf8String("x")`
calls the game ctor and is `IDisposable` → `using var s = new Utf8String("x");
` `Utf8String.FromString(...)` heap-allocates in game memory — ownership rules
in the Strings pitfall), `StdVector<T>`
(`First/Last/End`, `Count`, indexer, `AsSpan()`, `foreach`), `StdMap<K,V>`
(`Tree`, indexer, `Keys/Values`, `foreach` KeyValuePair), `StdString`
(MSVC SSO, 0x20), `StdSet/StdList/StdDeque/StdLinkedList`, `Span<T>`
accessors for fixed arrays, `SpanExtensions.PointerEnumerator()` to iterate a
`Span<T>` as `T*`.

