# EzHook, EzSignatureHelper, EzDelegate, EzPatch

Covers: `ECommons.EzHookManager`, the lazy, auto-disposed wrapper around Dalamud hooks, its
attribute-driven initializer, the delegate helper and the byte patcher. Verified against ECommons
`bc38dc7` (2026-11-08, v3.2.1.26, Dalamud API 15); nothing here was compiled. Paths are relative
to the skill root.

```csharp
using ECommons.EzHookManager;   // EzHook<T>, EzHookAttribute, EzHookFromCSAttribute,
                                // EzSignatureHelper, EzDelegate, EzPatch, EzPatchWithPointer<T>,
                                // EzHookCommon
```

Source: `references/ecommons/ECommons/EzHookManager/*.cs` (8 files). Stock `IGameInteropProvider`,
`Hook<T>`, `[Signature]` and `ISigScanner` behaviour: see the `dalamud` skill,
`references/topic-hooking.md`.

## What it wraps

`EzHook<T>` holds an address, a detour and (once created) a Dalamud `Hook<T>` obtained from
`Svc.Hook.HookFromAddress(Address, Detour)` (`Svc.Hook` is `IGameInteropProvider`). Signatures are
resolved with `Svc.SigScanner.ScanText(sig)`. Differences from a raw `Hook<T>`:

| Aspect | Dalamud `Hook<T>` | `EzHook<T>` |
|---|---|---|
| Creation | at `HookFromAddress` | lazily, on the first `Enable()` |
| `Disable()` | stops detour execution, hook stays installed | disposes the underlying hook |
| Pause without dispose | `Disable()` | `Pause()` |
| Call the game function while unhooked | not possible through the hook | `Original` falls back to `Delegate` |
| Disposal on unload | you must `Dispose()` | `ECommonsMain.Dispose()` disposes all |
| `IDisposable` | yes | no; there is no `Dispose()` method |

`ECommonsMain.Init(...)` must have run first (everything goes through `Svc`). No `Module` flag is
required for anything in this namespace.

## `EzHook<T>`

`public class EzHook<T> where T : System.Delegate` (`EzHook.cs`).

```csharp
public EzHook(string sig, T detour, bool autoEnable = true, int offset = 0)
public EzHook(nint address, T detour, bool autoEnable = true)
```

| Member | Meaning |
|---|---|
| `nint Address { get; }` | Resolved address (`ScanText(sig) + offset` for the string constructor). |
| `T Detour { get; }` | The detour passed to the constructor. |
| `T Delegate { get; }` | `Marshal.GetDelegateForFunctionPointer<T>(Address)`, created in the constructor. Callable without ever enabling the hook. While the hook is enabled, calling it enters the hooked function, so it runs your detour. |
| `T Original` | `HookDelegate?.Original ?? Delegate`: the trampoline when the hook exists (enabled or paused), otherwise the plain function pointer delegate. Use this inside the detour. |
| `void Enable()` | No hook yet: `HookFromAddress`, `Enable()`, register in `EzHookCommon.RegisteredHooks`. Hook exists but is paused: `Enable()` it. Already enabled: nothing. |
| `void Pause()` | If created and enabled: Dalamud `Hook<T>.Disable()`. The hook stays installed. |
| `void Disable()` | If created: `Hook<T>.Dispose()`, unregister, set the internal hook to null. A later `Enable()` creates a new hook. |
| `bool IsEnabled` | Hook exists and `Hook<T>.IsEnabled`. |
| `bool IsCreated` | Hook exists (enabled or paused). |

There is no member named `OriginalDelegate` on `EzHook<T>`; the two callable members are
`Original` and `Delegate`. The underlying `Hook<T>` is `internal` (`HookDelegate`) and not
reachable from plugin code.

Notes:

- `autoEnable` defaults to `true`, so `new EzHook<T>(sig, detour)` installs the hook at once. Pass
  `false` to get the lazy behaviour and call `Enable()` later.
- The string constructor calls `ScanText`, which throws when the signature is not found. The
  exception propagates out of the constructor. `ScanText` follows a leading `E8`/`E9` call (see
  `dalamud` `references/topic-hooking.md`); `offset` is added after that resolution.
- Logging: every configure/create/enable/pause/dispose writes `PluginLog.Debug("[EzHook] ...")`
  unless `ECommonsMain.ReducedLogging` is `true`.
- `EzHookCommon.TrackMemory` (`public static int`, default `0`): when greater than 0, that many
  bytes at `Address` are read before and after each create/pause/dispose and logged as `Before:` /
  `After:` hex. Debug aid only.
- Detours are not wrapped in try/catch by ECommons. An exception escaping a detour crashes the game.

```csharp
private delegate void RetainerItemCommandDelegate(nint module, uint slot, InventoryType type,
    uint a4, RetainerItemCommand command);
private EzHook<RetainerItemCommandDelegate> RetainerItemCommandHook;

// in the constructor; installs immediately because autoEnable defaults to true
RetainerItemCommandHook = new("48 89 5C 24 ?? 48 89 6C 24 ?? 48 89 74 24 ?? 57 48 83 EC 30 48 8B 5C 24 ?? 41 8B F0",
    RetainerItemCommandDetour);
```

(Pattern from AutoRetainer `Internal/Memory.cs`; signature strings are game-version specific.)

## `[EzHook]` and `EzSignatureHelper`

`EzHookAttribute` (`EzHookAttribute.cs`), `[AttributeUsage(Field | Property)]`:

```csharp
public EzHookAttribute(string signature, bool autoEnable = true)
public EzHookAttribute(string signature, int offset, bool autoEnable = true)
public EzHookAttribute(string signature, string? detourName, bool autoEnable = true)
public EzHookAttribute(string signature, int offset, string? detourName, bool autoEnable = true)
```

Read-only properties: `string Signature`, `int Offset` (0), `string? Detour` (null),
`bool AutoEnable` (true). `ToString()` returns
`EzHook(Signature=..., Offset=..., Detour=..., AutoEnable=...)`.

`EzSignatureHelper` (`EzSignatureHelper.cs`):

```csharp
public static void Initialize(Object obj)   // instance and static fields of obj.GetType()
public static void Initialize(Type type)    // static fields of a (static) class
```

### The exact convention

For each field returned by `GetFields(ReflectionHelper.AllFlags)` (public, non-public, instance,
static) that carries `[EzHook]`:

1. Detour name = `attr.Detour` if given; otherwise the field name with a trailing `Hook` removed,
   plus `Detour`. `SellItemHook` looks for `SellItemDetour`. A field named `Foo` (no `Hook`
   suffix) looks for `FooDetour`.
2. The method is looked up by name only: `GetType().GetMethod(detourName, AllFlags)`. Any
   visibility, instance or static. Overloads of the detour name make `GetMethod` throw
   `AmbiguousMatchException` (caught and logged, field stays null).
3. The delegate type is taken from the field type: `x.FieldType.GenericTypeArguments[0]`, the `T`
   of `EzHook<T>`. The delegate type's own name plays no part in the matching; any name works,
   and FFXIVClientStructs delegate types such as `ActionManager.Delegates.UseAction` are fine.
4. The delegate is built with `Delegate.CreateDelegate(T, method.IsStatic ? null : obj, method)`,
   so the detour's parameters and return type must match `T`.
5. The hook is built with `Activator.CreateInstance(fieldType, attr.Signature, delegate,
   attr.AutoEnable, attr.Offset)`, that is, the `EzHook(string sig, T detour, bool autoEnable,
   int offset)` constructor, and stored with `x.SetValue(obj, hook)`.

`Initialize(Type)` does the same with `null` as target, so fields and detours must be static.

Failure handling: a `TargetInvocationException` (for example the signature was not found inside
the `EzHook` constructor) is logged together with its inner exception; any other exception logs
`An error during initialization of attribute ... on <field>`. `Initialize` itself does not throw,
the affected field is left `null`, and the remaining fields are still processed. Null-check hooks
that may be missing after a game patch.

Limits that follow from the code:

- Only fields are scanned. The attribute's `AttributeUsage` allows properties, but
  `EzSignatureHelper` never calls `GetProperties`, so `[EzHook]` on a property does nothing.
  `[field: EzHook(...)]` on an auto-property is *(unverified)*; use a field.
- Reflection is run on the runtime type (`obj.GetType()`). Standard .NET reflection rules apply:
  private fields and private methods declared in a base class are not returned for a derived
  type. Hooks declared in a base class need non-private fields and detours (PandorasBox declares
  them `public` / `public virtual` in `BaseFeature`).
- Calling `Initialize` twice on the same object creates a second set of hooks and overwrites the
  fields; the first set stays registered until `ECommonsMain.Dispose()`. PandorasBox guards with
  `if (SendActionHook is null) EzSignatureHelper.Initialize(this);`.
- `EzDelegate` is not processed by `EzSignatureHelper`; there is no attribute for plain delegates.
  Use `EzDelegate.Get<T>` in a field initializer, or Dalamud's `[Signature]` with
  `Svc.Hook.InitializeFromAttributes(this)` (AutoRetainer calls both initializers on the same
  object).

### Example

Assembled from AutoRetainer `Services/InnCutsceneSkip.cs` and `Internal/Memory.cs`:

```csharp
using ECommons.EzHookManager;

public unsafe class SellWatcher
{
    private delegate void SellItemDelegate(uint a1, InventoryType a2, uint a3);

    [EzHook("48 89 6C 24 ?? 48 89 74 24 ?? 57 48 83 EC 20 8B F2 8B E9", false)]  // autoEnable: false
    private EzHook<SellItemDelegate> SellItemHook;

    private void SellItemDetour(uint inventorySlot, InventoryType a2, uint a3)
    {
        try
        {
            PluginLog.Debug($"SellItem: {inventorySlot}, {a2}, {a3}");
        }
        catch(Exception e)
        {
            e.Log();
        }
        SellItemHook.Original(inventorySlot, a2, a3);
    }

    public SellWatcher()
    {
        EzSignatureHelper.Initialize(this);   // fills SellItemHook, does not install it yet
    }

    public void SetActive(bool active)
    {
        if(SellItemHook == null) return;      // signature not found
        if(active) SellItemHook.Enable();     // installs on first call
        else SellItemHook.Pause();            // or Disable() to remove the hook entirely
    }
}
```

Explicit detour name, as PandorasBox does:

```csharp
[EzHook("E8 ?? ?? ?? ?? B0 01 EB B6", detourName: "UseActionDetour")]
public EzHook<ActionManager.Delegates.UseAction>? UseActionHook;
```

## `[EzHookFromCS]`

`EzHookFromCSAttribute` (`EzHookFromCS.cs`), `[AttributeUsage(Field | Property)]`:

```csharp
public EzHookFromCSAttribute(string? detourName, bool autoEnable = true)
public EzHookFromCSAttribute(bool autoEnable = true)
```

Properties: `string? Detour`, `bool AutoEnable`. Intended use: the field is
`EzHook<SomeStruct.Delegates.SomeFunction>` with a FFXIVClientStructs generated delegate type, and
the address comes from FFXIVClientStructs instead of a signature string. Resolution in
`EzSignatureHelper`:

1. Detour name: same rule as `[EzHook]`. A missing method throws
   `InvalidOperationException("Appropriate detour method could not be found for <field>")`.
2. `T` = first generic argument of the field type; `T.DeclaringType` must be the `Delegates` class;
   its `DeclaringType` is the parent struct.
3. The parent struct's nested type `Addresses` is looked up, then the static field in it with the
   same name as the delegate type (`genericType.Name`).
4. That field's value must be an `InteropGenerator.Runtime.Address`; `addr.Value` is passed to the
   `EzHook(nint address, T detour, bool autoEnable)` constructor.

**Defect at this commit:** both `Initialize` overloads build the hook in the `[EzHookFromCS]`
branch but never call `x.SetValue(...)`. The `EzHook<T>` instance is created (and enabled when
`AutoEnable` is true) but the field remains `null`. Consequences: with `autoEnable: true` the
detour runs and cannot reach `Original` through the field; with `autoEnable: false` nothing
usable is produced. AutoRetainer uses `[EzHookFromCS(false)]` together with `Hook?.Enable()`,
which at this commit is a no-op. Until upstream fixes it, construct the hook yourself:

```csharp
private EzHook<PacketDispatcher.Delegates.HandleMarketBoardItemRequestStartPacket> MarketHook;

MarketHook = new(PacketDispatcher.Addresses.HandleMarketBoardItemRequestStartPacket.Value,
    MarketDetour, autoEnable: false);
```

(`<Struct>.Addresses.<Function>.Value` is the pattern ECommons itself uses in
`Hooks/MapEffect.cs` and `Hooks/StaticVfx.cs`. Member names on the FFXIVClientStructs side: see the
`clientstructs` skill.)

## `EzDelegate`

`public static class EzDelegate` (`EzDelegate.cs`): a function-pointer delegate without a hook.

```csharp
public static T Get<T>(string sig, int offset = 0)  // ScanText(sig) + offset
public static T Get<T>(nint address)
```

Both return `Marshal.GetDelegateForFunctionPointer<T>(...)`. The string overload throws when the
signature is not found. There is no generic constraint; `T` must be a delegate type at run time.
Typical use is a field initializer (AutoRetainer `Internal/Memory.cs`):

```csharp
public delegate nint SetExchangeModeDelegate(nint addon, int mode);
public SetExchangeModeDelegate SetExchangeMode = EzDelegate.Get<SetExchangeModeDelegate>("E8 ?? ...");
```

A field initializer runs at construction; make sure `ECommonsMain.Init` ran before the object is
created.

## `EzPatch`

`public class EzPatch : IDisposable` (`EzPatch.cs`): writes bytes at an address with
`SafeMemory.WriteBytes` and restores the original bytes on disable.

```csharp
public EzPatch(nint addr, Data patchData, bool autoEnable = true, bool silent = false)
public EzPatch(string signature, nint offset, Data patchData, bool autoEnable = true, bool silent = false)
public EzPatch(ICollection<byte?> signature, nint offset, Data patchData, bool autoEnable = true, bool silent = false)
```

| Member | Meaning |
|---|---|
| `bool Disposed { get; }` | Set by `Dispose()`. |
| `bool Enabled { get; }` | Patch bytes are currently written. |
| `Data PatchData { get; }` | Original and patch bytes. `null` if initialization failed. |
| `bool Silent { get; set; }` | Suppresses debug logging and hides the patch bytes in error logs. |
| `nint Address { get; }` | Patched address; `0` if initialization failed. |
| `void Enable()` | Writes `PatchData.PatchData`. Logs an error and returns if disposed. Exceptions are caught and logged. |
| `void Disable()` | Writes `PatchData.OriginalData` if enabled. Exceptions are caught and logged. |
| `virtual void Dispose()` | `Disable()`, `Disposed = true`, removes itself from the internal list. |

`EzPatch.Data`:

```csharp
public readonly IReadOnlyList<byte> OriginalData;
public readonly IReadOnlyList<byte> PatchData;
public Data(IReadOnlyList<byte> originalData, IReadOnlyList<byte> patchData)
public Data(IReadOnlyList<byte> patchData)             // original bytes read from memory at init
public Data(string originalData, string patchData)     // parsed by GenericHelpers.ParseByteArray
public Data(string patchData)
```

The two-list constructor throws `InvalidOperationException` when `originalData.Count <
patchData.Count`.

Behaviour:

- The signature constructors use `Svc.SigScanner.ScanAllText` and take the first match plus
  `offset`. The string overload logs a warning when more than one match exists. With zero matches
  the constructor returns without initializing: no exception, `Address == 0`, `PatchData == null`,
  `Enabled == false`. A later `Enable()` then fails inside its try/catch (logged). Check `Address`
  or `PatchData` after construction.
- The `ICollection<byte?>` overload turns `null` entries into `??` wildcards.
- When `Data` has no original bytes, `PatchData.Count` bytes are read from the address and used as
  the original.
- When original bytes are given, memory is compared with them. A mismatch throws
  `Could not create patch ...: unexpected data` internally; this is caught, logged, and the patch
  disposes itself (`Disposed == true`). This is the guard against patching a changed game build.
- Auto-dispose: every successfully initialized patch is added to an internal list whose cleanup is
  registered with `Purgatory`; `ECommonsMain.Dispose()` runs `Purgatory.Purge`, which disposes all
  remaining patches and so restores the original bytes.

```csharp
var patch = new EzPatch("48 89 ?? ...", 0x12, new EzPatch.Data("74 05", "EB 05"), autoEnable: false);
if(patch.PatchData != null && !patch.Disposed) patch.Enable();
```

## `EzPatchWithPointer<T>`

`public unsafe class EzPatchWithPointer<T> : EzPatch, IDisposable where T : unmanaged`
(`EzPatchWithPointer.cs`): a patch plus a typed pointer into the patched region, for patches that
embed a value you want to change at run time.

```csharp
public EzPatchWithPointer(nint addr, Data patchData, nint pointerOffset,
    bool autoEnable = true, bool silent = false, bool restoreOriginal = false)
public EzPatchWithPointer(string signature, nint offset, Data patchData, nint pointerOffset,
    bool autoEnable = true, bool silent = false, bool restoreOriginal = false)
```

| Member | Meaning |
|---|---|
| `readonly T* PointerRaw` | `(T*)(Address + pointerOffset)`. |
| `T PointerValue { get; set; }` | Getter dereferences `PointerRaw` always. Setter writes only while `Enabled`; otherwise the assignment is silently dropped. |
| `override void Dispose()` | Optionally restores page protection, then `EzPatch.Dispose()`. |

The constructor calls `VirtualProtectEx` to make `sizeof(T)` bytes at `PointerRaw`
`PAGE_EXECUTE_READWRITE` and remembers the old protection. `restoreOriginal: true` makes
`Dispose()` put the old protection back; if the protection found at that point differs from the
remembered one, an error is logged and the protection found is re-applied instead. If the base
patch failed to initialize, `Address` is 0 and `PointerRaw` equals `pointerOffset`: do not touch
`PointerValue` in that case.

## Disposal: `EzHookCommon` and `ECommonsMain.Dispose()`

`public static class EzHookCommon` (`EzHookCommon.cs`) has one public member,
`public static int TrackMemory = 0`. `RegisteredHooks` (`List<IDalamudHook>`) and `DisposeAll()`
are `internal`; plugin code cannot call `EzHookCommon.DisposeAll()`.

`ECommonsMain.Dispose()` (`references/ecommons/ECommons/ECommonsMain.cs`) runs, in this order and
each inside `GenericHelpers.Safe`:

1. the built-in hooks of `ECommons.Hooks` (`DirectorUpdate`, `ActionEffect`, `MapEffect`,
   `SendAction`, `ActorVfx`, `StaticVfx`, `GameObjectCtor`), see `references/topic-game-hooks.md`;
2. later, `EzHookCommon.DisposeAll`: disposes every Dalamud hook currently registered by any
   `EzHook<T>` and sets the list to `null`;
3. last, `Purgatory.Purge`: disposes all `EzPatch` instances (and `Svc.GameNetwork`).

So a plugin that calls `ECommonsMain.Dispose()` in its own `Dispose()` does not need to disable
`EzHook` or `EzPatch` instances itself.

## Gotchas

- `EzHook<T>.Disable()` disposes the hook; `Pause()` is the equivalent of Dalamud's `Disable()`.
  Porting code from `Hook<T>` by renaming the type changes behaviour.
- After `ECommonsMain.Dispose()` the `EzHook<T>` objects still hold their (now disposed) inner
  hook, and the registry list is `null`. Calling `Enable()`/`Disable()`/`Original` on them
  afterwards fails. Do not touch hooks after ECommons is disposed.
- `[EzHookFromCS]` leaves the field `null` at this commit (see above).
- `[EzHook]` on a property is ignored; private hook fields or detours in a base class are not
  found.
- `EzSignatureHelper.Initialize` swallows all errors into the log. A missing signature shows up as
  a `NullReferenceException` at first use unless you null-check.
- `autoEnable` defaults to `true` everywhere (constructors and attributes). The positional `bool`
  in `[EzHook(sig, false)]` is `autoEnable`; the positional `int` in `[EzHook(sig, 3)]` is
  `offset`.
- `Delegate` on an enabled hook calls the hooked function (detour included); inside a detour use
  `Original`, otherwise the detour recurses.
- `EzPatch` with an unmatched signature fails silently (no exception, no log unless more than one
  match).
- `ECommons.DalamudServices.Legacy.SignatureHelper.Initialise(object which, bool log = false)` is
  unrelated to `EzSignatureHelper`: it forwards to `Svc.Hook.InitializeFromAttributes(which)`
  (Dalamud `[Signature]`), and the `log` argument is ignored.
