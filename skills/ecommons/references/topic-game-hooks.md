# Built-in game hooks, network messages, VFX and object-life tracking

Covers: the ready-made hooks in `ECommons.Hooks` (`ActionEffect`, `DirectorUpdate`, `MapEffect`,
`SendAction`, `ActorVfx`, `StaticVfx`, `GameObjectCtor`), `Svc.GameNetwork` and the other
`Legacy` shims, `OpcodeUpdater`, `VfxManager` and `ObjectLife`. Verified against ECommons
`bc38dc7` (2026-11-08, v3.2.1.26, Dalamud API 15); nothing here was compiled. Paths are relative
to the skill root.

```csharp
using ECommons.Hooks;                    // ActionEffect, DirectorUpdate(+Category), MapEffect,
                                         // SendAction, ActorVfx, StaticVfx, GameObjectCtor
using ECommons.Hooks.ActionEffectTypes;  // ActionEffectSet, EffectHeader, EffectEntry,
                                         // TargetEffect, ActionEffectType, AttackType
using ECommons.DalamudServices;          // Svc
using ECommons.DalamudServices.Legacy;   // IGameNetwork, NetworkMessageDirection, shims
using ECommons.GameHelpers;              // VfxManager, VfxInfo, VfxStruct, Quat
using ECommons.ObjectLifeTracker;        // ObjectLife extension methods
```

Writing your own hook: `references/topic-ezhook.md`. Stock Dalamud hooking: `dalamud` skill,
`references/topic-hooking.md`. FFXIVClientStructs types named below: `clientstructs` skill.

## Overview

| Class | How it is installed | Public entry point | Multiple subscribers | `Module` flag |
|---|---|---|---|---|
| `ActionEffect` | lazily, on first event subscription | events | yes | none |
| `DirectorUpdate` | explicit `Init(callback)` | one callback | no, second `Init` throws | none |
| `MapEffect` | explicit `Init(callback)` | one callback | no, second `Init` throws | none |
| `SendAction` | explicit `Init(callback)` | one callback | no, second `Init` throws | none |
| `ActorVfx` | lazily, per event | events | yes | none |
| `StaticVfx` | lazily, per event | events | yes | none |
| `GameObjectCtor` | lazily, on first event subscription | event | yes | none |
| `Svc.GameNetwork` | lazily, on first property access | events | yes | none |
| `VfxManager` | `ECommonsMain.Init` | query methods | n/a | `Module.VfxTracking` or `Module.All` |
| `ObjectLife` | `ECommonsMain.Init` | extension methods | single `Action` field | `Module.ObjectLife` or `Module.All` |

`Module` is `enum Module { All, DalamudReflector, ObjectFunctions, ObjectLife, SplatoonAPI,
VfxTracking }` (`references/ecommons/ECommons/Module.cs`), passed as
`ECommonsMain.Init(pluginInterface, this, Module.VfxTracking, Module.ObjectLife)`.

All hooks use `Svc.Hook.HookFromAddress` and therefore need `ECommonsMain.Init` first. All detours
run on the thread the game calls the function from (the game's main thread for these functions
in practice; the source does not marshal anything except `GameObjectCtor`). All of them are
disposed by `ECommonsMain.Dispose()`; you never need to call their `Dispose()` yourself.

Unsubscribing from an event never removes or disables the hook. Once installed, a hook stays until
`Disable()`/`Dispose()` or plugin unload.

**No `ProcessPacket`-style hook class exists at this commit.** A search of the checkout for
`ProcessPacket`, `ProcessZonePacket` and `ProcessActorControlPacket` matches only the private
internals of `DalamudServices/Legacy/GameNetwork.cs` and `GameNetworkAddressResolver.cs`. There is
no `ProcessActorControlPacket`/ActorControl hook in ECommons. Raw packets are available only
through `Svc.GameNetwork` (below); for ActorControl write your own hook with `EzHook`.

## `ActionEffect`

`public static unsafe class ActionEffect` (`references/ecommons/ECommons/Hooks/ActionEffect.cs`).

```csharp
public const string Sig = "40 55 53 56 57 41 54 41 55 41 56 41 57 48 8D AC 24 ?? ?? ?? ?? 48 81 EC ?? ?? ?? ?? 48 8B 05 ?? ?? ?? ?? 48 33 C4 48 89 45 70 4C 8B BD";

public delegate void ProcessActionEffect(uint sourceId, Character* sourceCharacter, Vector3* pos,
    EffectHeader* effectHeader, EffectEntry* effectArray, ulong* effectTail);
public delegate void ActionEffectCallback(ActionEffectSet set);

public static event ActionEffectCallback ActionEffectEvent;
public static event Action<uint, ushort, ActionEffectType, uint, ulong, uint> ActionEffectEntryEvent;

[Obsolete] public static void Init(Action<uint, ushort, ActionEffectType, uint, ulong, uint> fullParamsCallback, bool logging = false)
public static void Enable()
public static void Disable()
```

- Both events install and enable the hook in their `add` accessor (`TryScanText(Sig)`). If the
  signature is not found, `Could not find ActionEffect signature` is logged as an error, the
  handler is still added, and nothing is ever raised.
- `ActionEffectEvent` is raised once per action with the whole `ActionEffectSet`.
- `ActionEffectEntryEvent` is raised once per non-`Nothing` effect entry per target. Arguments in
  order: `uint ActionID`, `ushort AnimationId`, `ActionEffectType type`, `uint sourceID`,
  `ulong targetID`, `uint damage` (`EffectEntry.Damage`).
- `Init` is `[Obsolete("Please use ActionEffectEntryEvent instead.")]`; it only does
  `ActionEffectEntryEvent += fullParamsCallback` and ignores `logging`.
- Handlers run before the game's original function. The whole dispatch is inside one try/catch
  that logs `An error has occurred in Action Effect hook.`; an exception in one subscriber stops
  the remaining subscribers and entries for that call, the original is still called.
- `Dispose()` is `internal`.

### `ActionEffectSet`

`public readonly unsafe record struct ActionEffectSet`
(`references/ecommons/ECommons/Hooks/ActionEffectTypes/ActionEffectSet.cs`).

| Member | Type | Meaning |
|---|---|---|
| `Action` | `Lumina.Excel.Sheets.Action?` | Row for `Header.ActionID`; set only when `Header.ActionType` is not `EventItem`, `Item` or `Mount`. |
| `Item` | `Item?` | Set for `ActionType.Item`. IDs above 1000000 have 1000000 subtracted (HQ). |
| `EventItem` | `EventItem?` | Set for `ActionType.EventItem`. |
| `Mount` | `Mount?` | Set for `ActionType.Mount`. |
| `IconId` | `ushort` | Icon of the row above; for actions: 101 if `ActionCategory` row 1 (auto attack), 104 for action 3 (Sprint), 118 for action 4, else `Action.Icon`. |
| `Name` | `string` | Name of the row above, `string.Empty` if the row is missing. |
| `Target` | `IGameObject?` | `Svc.Objects.SearchById(Header.AnimationTargetId)`. |
| `Source` | `IGameObject?` | `Svc.Objects.SearchById(sourceID)`. |
| `SourceCharacter` | `FFXIVClientStructs...Character?` | A by-value copy of `*sourceCharacter`, not a pointer. |
| `TargetEffects` | `TargetEffect[]` | One per target, length `Header.TargetCount`. |
| `Position` | `Vector3` | `*pos`. |
| `Header` | `EffectHeader` | Copy of the header. |
| `GetSpecificTypeEffect(ActionEffectType type)` | `Dictionary<ulong, uint>` | Target ID to `EffectEntry.value` (the raw 16-bit value, not `Damage`) of the first entry of that type per target. |
| `ToString()` | `string` | Multi-line dump. |

Constructor (public, normally not called by plugins): `ActionEffectSet(uint sourceID, Character*
sourceCharacter, Vector3* pos, EffectHeader* effectHeader, EffectEntry* effectArray, ulong*
effectTail)`.

### `EffectHeader`

`[StructLayout(LayoutKind.Explicit)] public struct EffectHeader` (`.../EffectHeader.cs`).

| Offset | Field |
|---|---|
| 0 | `ulong AnimationTargetId` |
| 8 | `uint ActionID` |
| 12 | `uint GlobalEffectCounter` |
| 16 | `float AnimationLockTime` |
| 20 | `uint SomeTargetID` |
| 24 | `ushort SourceSequence` |
| 26 | `ushort Rotation` |
| 28 | `ushort AnimationId` |
| 30 | `byte Variation` |
| 31 | `byte ActionTypeByte`. `ActionType ActionType` (`FFXIVClientStructs.FFXIV.Client.Game.ActionType`) is a computed property that casts it. Up to 3.2.1.25 `ActionType` was the field at this offset and `ActionTypeByte` did not exist |
| 33 | `byte TargetCount` |

Computed: `float RealRotation` and `float RealRotationD` (radians, `Rotation * 0.0095875263 *
0.0099999998 - PI`, single and double precision maths), `float RotationFromNorth` (returns
`-Rotation + MathF.PI` using the raw `ushort Rotation`, not `RealRotation`; treat with care).

### `TargetEffect`, `EffectEntry`, enums

`public unsafe struct TargetEffect` (`.../TargetEffect.cs`):

```csharp
public ulong TargetID { get; }
public TargetEffect(ulong targetId, EffectEntry* effects)
public EffectEntry this[int index]                 // index outside 0..7 returns default
public bool GetSpecificTypeEffect(ActionEffectType type, out EffectEntry effect)  // first match
public void ForEach(Action<EffectEntry> act)       // always 8 calls, including Nothing entries
public override string ToString()
```

`TargetEffect` holds a raw pointer into the packet buffer. It is valid only during the event
callback. Copy the `EffectEntry` values you need; do not store `ActionEffectSet`/`TargetEffect`
and read entries later.

`public struct EffectEntry` (`.../EffectEntry.cs`): fields `ActionEffectType type`, `byte param0`,
`byte param1`, `byte param2`, `byte mult`, `byte flags`, `ushort value` (lower-case names);
`byte AttackType => param1 & 0xF`; `uint Damage => mult == 0 ? value : value + 65536 * mult`.

`public enum ActionEffectType : byte` (`.../ActionEffectType.cs`): `Nothing = 0`, `Miss = 1`,
`FullResist = 2`, `Damage = 3`, `Heal = 4`, `BlockedDamage = 5`, `ParriedDamage = 6`,
`Invulnerable = 7`, `NoEffectText = 8`, `Unknown_0 = 9`, `MpLoss = 10`, `MpGain = 11`,
`TpLoss = 12`, `TpGain = 13`, `ApplyStatusEffectTarget = 14`, `ApplyStatusEffectSource = 15`,
`RecoveredFromStatusEffect = 16`, `LoseStatusEffectTarget = 17`, `LoseStatusEffectSource = 18`,
`StatusNoEffect = 20`, `ThreatPosition = 24`, `EnmityAmountUp = 25`, `EnmityAmountDown = 26`,
`StartActionCombo = 27`, `ComboSucceed = 28`, `Retaliation = 29`, `Knockback = 32`,
`Attract1 = 33`, `Attract2 = 34`, `Mount = 40`, `FullResistStatus = 52`,
`FullResistStatus2 = 55`, `VFX = 59`, `Gauge = 60`, `JobGauge = 61`, `SetModelState = 72`,
`SetHP = 73`, `PartialInvulnerable = 74`, `Interrupt = 75`.
Obsolete aliases: `GpGain` (use `ApplyStatusEffectTarget`), `Unknown0` (use `StartActionCombo`),
`Unknown1` (use `ComboSucceed`).

`public enum AttackType` (`.../AttackType.cs`): `Unknown = 0`, `Slashing`, `Piercing`, `Blunt`,
`Shot`, `Magical`, `Unique`, `Physical`, `LimitBreak = 8`. Note `EffectEntry.AttackType` is a
`byte`; cast it: `(AttackType)entry.AttackType`.

```csharp
ActionEffect.ActionEffectEvent += OnActionEffect;

private void OnActionEffect(ActionEffectSet set)
{
    if(set.Source?.Address != Svc.Objects.LocalPlayer?.Address) return;
    foreach(var target in set.TargetEffects)
    {
        if(target.GetSpecificTypeEffect(ActionEffectType.Damage, out var entry))
            PluginLog.Debug($"{set.Name} hit {target.TargetID:X} for {entry.Damage}");
    }
}
// on unload: ActionEffect.ActionEffectEvent -= OnActionEffect; (hook itself: ECommonsMain.Dispose)
```

## `DirectorUpdate`

`public static class DirectorUpdate` (`references/ecommons/ECommons/Hooks/DirectorUpdate.cs`).
Signature (private field `Sig`):
`40 53 57 48 83 EC 58 48 8B 05 ?? ?? ?? ?? 48 33 C4 48 89 44 24 ?? 41 8B F9`.

```csharp
public delegate void ProcessDirectorUpdate(nint a1, uint a2, DirectorUpdateCategory a3,
    uint a4, uint a5, int a6, int a7, int a8, int a9);
public static ProcessDirectorUpdate Delegate { get; }

public static void Init(Action<nint, uint, DirectorUpdateCategory, uint, uint, int, int, int, int> fullParamsCallback)
public static void Init(Action<DirectorUpdateCategory> categoryOnlyCallback)
public static void Enable()
public static void Disable()
public static void Dispose()
```

`public enum DirectorUpdateCategory : uint` (`.../DirectorUpdateCategory.cs`):
`Commence = 0x40000001`, `Recommence = 0x40000006`, `Complete = 0x40000003`,
`Wipe = 0x40000005`. Other values occur at run time and arrive as unnamed numbers.

- There are no events. Exactly one `Init` overload can be called once per plugin; a second call
  (either overload) throws `Exception("Director Update Hook is already initialized!")` until
  `Dispose()` is called. Fan out to several consumers in your own callback.
- If the signature is not found, `Init` logs an error and returns; the hook stays uninstalled and a
  later `Init` is allowed.
- The callback runs before the original, inside try/catch (`e.Log()`).
- `Delegate` returns the hook's `Original` while the hook exists, otherwise a cached
  `EzDelegate.Get<ProcessDirectorUpdate>(Sig)`. It lets you invoke the game function yourself and
  works without `Init`.

```csharp
DirectorUpdate.Init(category =>
{
    if(category == DirectorUpdateCategory.Commence) OnPullStart();
    if(category == DirectorUpdateCategory.Wipe) OnWipe();
});
```

## `MapEffect`

`public static unsafe class MapEffect` (`references/ecommons/ECommons/Hooks/MapEffect.cs`).

```csharp
public const string ExsMultisig = "40 55 41 57 48 83 EC ?? 48 83 B9";
public static ContentDirector.Delegates.ApplyMapEffect Delegate { get; }

public static void Init(Action<long, uint, ushort, ushort> fullParamsCallback)
public static void Enable()
public static void Disable()
public static void Dispose()
```

- There is no event and no ECommons-declared delegate type. The hook type is FFXIVClientStructs'
  `ContentDirector.Delegates.ApplyMapEffect`
  (`FFXIVClientStructs.FFXIV.Client.Game.InstanceContent`), and the address is
  `ContentDirector.Addresses.ApplyMapEffect.Value`, not a signature scan.
- Callback arguments: `long a1` (the `ContentDirector*` cast to `long`), `uint a2` (position /
  index), `ushort a3`, `ushort a4` (the two state words).
- One `Init` per plugin; a second call throws `Exception("MapEffect is already initialized!")`.
- `Init` also installs three further hooks found with `ScanAllText(ExsMultisig)` (the batched
  "map effect ex" packet handlers, as private `EzHook`s). They decode each entry and call the
  same callback with `a1 == 0`, and log `Map Effect Ex Hook: ...` at debug level per entry. If
  the scan does not return exactly 3 matches a warning is logged and only the main hook works.
- `Enable()`/`Disable()`/`Dispose()` act on the main hook only. The three extra hooks are
  `EzHook`s and are removed by `EzHookCommon.DisposeAll` inside `ECommonsMain.Dispose()`. After a
  manual `MapEffect.Dispose()` they stay installed, and a following `Init` installs a second set,
  which makes batched effects arrive twice. Avoid re-initializing.
- `Delegate` returns the hook's `Original` while hooked, else a cached delegate for
  `ContentDirector.Addresses.ApplyMapEffect.Value`.

```csharp
MapEffect.Init((director, position, data1, data2) =>
    PluginLog.Debug($"MapEffect {position}: {data1}, {data2}"));
```

## `SendAction`

`public static class SendAction` (`references/ecommons/ECommons/Hooks/SendAction.cs`). Signature
(private field `Sig`):
`48 89 5C 24 ?? 48 89 6C 24 ?? 48 89 74 24 ?? 57 48 81 EC ?? ?? ?? ?? 48 8B 05 ?? ?? ?? ?? 48 33 C4
48 89 84 24 ?? ?? ?? ?? 48 8B E9 41 0F B7 D9`.

```csharp
public delegate long SendActionDelegate(long targetObjectId, byte actionType, uint actionId,
    ushort sequence, long a5, long a6, long a7, long a8, long a9);
public delegate void SendActionCallbackDelegate(long targetObjectId, byte actionType, uint actionId,
    ushort sequence, long a5, long a6, long a7, long a8, long a9);

public static void Init(SendActionCallbackDelegate fullParamsCallback)
public static void Enable()
public static void Disable()
public static void Dispose()
```

Observation only: the callback returns `void`, runs before the original inside try/catch, and the
original is always called with unchanged arguments. One `Init` per plugin (second call throws
`Exception("SendAction Hook is already initialized!")`). No `Delegate` property. To block or
alter the action, hook the function yourself (PandorasBox does so with `[EzHook]` on the same
signature).

## `ActorVfx`

`public static unsafe class ActorVfx` (`references/ecommons/ECommons/Hooks/ActorVfx.cs`).

```csharp
public static readonly string CreateSig = "40 53 55 56 57 48 81 EC ?? ?? ?? ?? 0F 29 B4 24 ?? ?? ?? ?? 48 8B 05 ?? ?? ?? ?? 48 33 C4 48 89 84 24 ?? ?? ?? ?? 0F B6 AC 24 ?? ?? ?? ?? 0F 28 F3 49 8B F8";
public static readonly string DtorSig = "48 89 5C 24 ?? 57 48 83 EC ?? 48 8D 05 ?? ?? ?? ?? 48 8B D9 48 89 01 8B FA 48 8D 05 ?? ?? ?? ?? 48 89 81 ?? ?? ?? ?? 48 8B 89 ?? ?? ?? ?? 48 85 C9 74 ?? 48 8B 01 48 8B D3";

public delegate void ActorVfxCreateCallbackDelegate(nint vfxPtr, nint vfxPathPtr, nint casterAddress,
    nint targetAddress, float a4, byte a5, ushort a6, byte a7);
public delegate void ActorVfxDtorCallbackDelegate(nint actorVfxAddress);

public static event ActorVfxCreateCallbackDelegate ActorVfxCreateEvent;
public static event ActorVfxDtorCallbackDelegate ActorVfxDtorEvent;

public static void EnableCreate()   public static void DisableCreate()
public static void EnableDtor()     public static void DisableDtor()
public static void Dispose()
```

- Each event installs its own hook on first subscription. Signature not found: error logged, no
  events.
- `ActorVfxCreateEvent` is raised after the original ran; `vfxPtr` is its return value (castable
  to `VfxStruct*`). `vfxPathPtr` is a C string: `MemoryHelper.ReadString(vfxPathPtr,
  Encoding.ASCII, 256)`. `casterAddress`/`targetAddress` are game object addresses (compare with
  `IGameObject.Address`), not entity IDs.
- `ActorVfxDtorEvent` is raised before the original destructor; the struct is still readable.
- Subscribers are invoked one by one, each in its own try/catch; one failing subscriber does not
  stop the others.
- Do not call any `Original` from a subscriber; these are notifications, not detours.

## `StaticVfx`

`public static unsafe class StaticVfx` (`references/ecommons/ECommons/Hooks/StaticVfx.cs`).

```csharp
public delegate void StaticVfxCreateCallbackDelegate(nint vfxPtr, string path, string systemSource);
public delegate void StaticVfxDtorCallbackDelegate(nint staticVfxAddress);
public delegate void StaticVfxRunCallbackDelegate(nint staticVfxAddress, float a1, uint a2);

public static event StaticVfxCreateCallbackDelegate StaticVfxCreateEvent;
public static event StaticVfxRunCallbackDelegate StaticVfxRunEvent;
public static event StaticVfxDtorCallbackDelegate StaticVfxDtorEvent;

public static void EnableCreate()   public static void DisableCreate()
public static void EnableRun()      public static void DisableRun()
public static void EnableDtor()     public static void DisableDtor()
public static void Dispose()
```

| Event | Hooked function | Address source | Raised |
|---|---|---|---|
| `StaticVfxCreateEvent` | `VfxObject.Delegates.Create` | `VfxObject.Addresses.Create.Value` | after the original; `vfxPtr` is the returned `VfxObject*` |
| `StaticVfxRunEvent` | `VfxObject.Delegates.Update` | `VfxObject.Addresses.Update.Value` | before the original |
| `StaticVfxDtorEvent` | `VfxObject.Delegates.CleanupRender` | signature `40 53 48 83 EC 20 48 8B D9 48 8B 89 ?? ?? ?? ?? 48 85 C9 74 28 33 D2` | before the original |

`VfxObject` is `FFXIVClientStructs.FFXIV.Client.Graphics.Scene.VfxObject`. There are no public
signature constants on this class. `path` and `systemSource` are already converted to `string`.
Per-subscriber try/catch as in `ActorVfx`. Caster and target of a static VFX are usually unset at
creation and filled by the time it runs (source comment in `VfxManager`).

## `GameObjectCtor`

`public class GameObjectCtor` (non-static class with static members only;
`references/ecommons/ECommons/Hooks/GameObjectCtor.cs`).

```csharp
public delegate void GameObjectConstructorCallbackDelegate(nint objectAddress);
public const string Sig = "48 8D 05 ?? ?? ?? ?? 48 89 01 33 C0 48 89 41 10 48 89 41 18 89 81 ?? ?? ?? ?? 48 89 81";
public static event GameObjectConstructorCallbackDelegate GameObjectConstructorEvent;
public static void Enable()
public static void Disable()
public static void Dispose()
```

- Hook installed on first subscription.
- The event is not raised inside the detour. The detour schedules
  `Svc.Framework.RunOnTick(..., delayTicks: 1)`; on that later framework tick the address is
  looked up in `Svc.Objects`, and subscribers are called only if an object with that address is
  present in the object table. Handlers therefore run on the framework thread, one tick late, and
  can use `Svc.Objects` safely. Objects that are gone again or never enter the table produce no
  event.
- Per-subscriber try/catch.
- Same signature as `ObjectLife` (below), which calls back synchronously instead.

## `Svc.GameNetwork` and the `Legacy` namespace

Dalamud removed its `IGameNetwork` service; ECommons carries its own copy of the implementation.
Source: `references/ecommons/ECommons/DalamudServices/Legacy/`.

```csharp
// ECommons.DalamudServices.Svc
public static Legacy.IGameNetwork GameNetwork { get; }   // field ??= new GameNetwork()

// ECommons.DalamudServices.Legacy
public interface IGameNetwork
{
    public delegate void OnNetworkMessageDelegate(nint dataPtr, ushort opCode, uint sourceActorId,
        uint targetActorId, NetworkMessageDirection direction);
    public event OnNetworkMessageDelegate NetworkMessage;
    public event OnNetworkMessageDelegate NetworkMessageReplayed;
}
public enum NetworkMessageDirection { ZoneDown, ZoneUp }
```

`GameNetwork` and `GameNetworkAddressResolver` are `internal`; use only the interface through
`Svc.GameNetwork`. It is not a `[PluginService]`: it is created on first access of the property.

Construction installs and enables three hooks at once, whether or not anybody subscribes:

| Hook | Address | Raises |
|---|---|---|
| zone down | `PacketDispatcher.StaticVirtualTablePointer->OnReceivePacket` (FFXIVClientStructs) | `NetworkMessage`, `ZoneDown` |
| zone up | sig `48 89 5C 24 ?? 48 89 74 24 ?? 4C 89 64 24 ?? 55 41 56 41 57 48 8B EC 48 83 EC 70` | `NetworkMessage`, `ZoneUp` |
| replay | sig `48 89 5C 24 ?? 48 89 74 24 ?? 57 48 83 EC 60 48 8B FA 48 8B F1` | `NetworkMessageReplayed`, `ZoneDown` |

Argument values actually passed:

| Event / direction | `dataPtr` | `opCode` | `sourceActorId` | `targetActorId` |
|---|---|---|---|---|
| `NetworkMessage`, `ZoneDown` | packet body (header start + 0x20) | `ushort` at header + 0x12 | always 0 | `targetId` from the game |
| `NetworkMessage`, `ZoneUp` | `dataPtr + 0x20` | `ushort` at `dataPtr` | always 0 | always 0 |
| `NetworkMessageReplayed` | `*packet` | `*opcode` | always 0 | always 0 |

- The two signature scans use `ScanText`, which throws when a signature is missing; the exception
  leaves the `Svc.GameNetwork` getter and is thrown again on every later access.
- Subscriber exceptions are caught and logged (down and replay: per subscriber; up: one try/catch
  around the multicast invoke).
- Disposal is registered with `Purgatory` in the constructor and runs at the end of
  `ECommonsMain.Dispose()`. There is no public `Dispose`.
- Handlers run inside the packet handlers on every packet. Keep them short, filter by `opCode`
  first. `dataPtr` is valid only during the call.
- Opcodes change with every game patch. ECommons ships no opcode table.

```csharp
Svc.GameNetwork.NetworkMessage += OnNetworkMessage;

private void OnNetworkMessage(nint dataPtr, ushort opCode, uint sourceActorId, uint targetActorId,
    NetworkMessageDirection direction)
{
    if(direction != NetworkMessageDirection.ZoneDown || opCode != MyOpcode) return;
    var value = Marshal.ReadInt32(dataPtr, 4);
}
// on unload: Svc.GameNetwork.NetworkMessage -= OnNetworkMessage;
```

### Other shims in `ECommons.DalamudServices.Legacy`

Compatibility helpers for API removed or renamed in Dalamud. They need
`using ECommons.DalamudServices.Legacy;`.

| Type | Member | Does |
|---|---|---|
| `SignatureHelper` | `static void Initialise(object which, bool log = false)` | `Svc.Hook.InitializeFromAttributes(which)`. `log` is ignored. Not related to `EzSignatureHelper`. |
| `LegacyHelpers` | `static void PrintChat(this IChatGui chatGui, XivChatEntry entry)` | `Svc.Chat.Print(entry)` (ignores the receiver). |
| `LegacyHelpers` | `static void SetTarget(this ITargetManager targetManager, IGameObject obj)` | `targetManager.Target = obj`. |
| `LegacyHelpers` | `static void AddNotification(this IUiBuilder builder, string message, string? pluginName = null, NotificationType type = NotificationType.Info, int timeout = 3000)` | `Svc.NotificationManager.AddNotification(...)`; title defaults to `Svc.PluginInterface.InternalName`, `timeout` in milliseconds. |
| `IClientStateExtensions` | extension property `IPlayerCharacter LocalPlayer` on `IClientState` | `Svc.Objects.LocalPlayer`. |
| `IClientStateExtensions` | extension property `ulong LocalContentId` on `IClientState` | `Svc.PlayerState.ContentId`. |
| `IChatMessageExtensions` | extension property `XivChatType Type` on `IChatMessage` | Rebuilds the legacy chat type from `LogKind`, `SourceKind`, `TargetKind`. |

The extension properties use C# 14 `extension(...)` blocks; with the `using` in place
`Svc.ClientState.LocalPlayer` compiles again. In new code use `Svc.Objects.LocalPlayer` and
`Svc.PlayerState.ContentId` directly (or ECommons' `Player` helper).

## `OpcodeUpdater`

`public static class OpcodeUpdater`, namespace `ECommons.Opcodes`
(`references/ecommons/ECommons/Opcodes/OpcodeUpdater.cs`). The whole class is
`[Obsolete("Scheduled to be removed. Please contact the developer if you have an use for it.")]`;
no replacement is named. Do not use it in new code.

```csharp
public static void DownloadOpcodes(string url, Action<Dictionary<string, uint>> successCallback,
    Action<string>? failureCallback = null)
```

Downloads `url` on a new thread and parses lines of the form `Name|HEX` into a dictionary.
`successCallback` and the HTTP-status failure path are marshalled to the framework thread through
`TickScheduler`; the exception failure path calls `failureCallback` directly on the worker thread.
With no `failureCallback`, failures are logged as warnings.

## `VfxManager` (requires `Module.VfxTracking`)

`public static class VfxManager`, namespace `ECommons.GameHelpers`
(`references/ecommons/ECommons/GameHelpers/VfxManager.cs`). `ECommonsMain.Init` calls the internal
`VfxManager.Init` only when `modules` contains `Module.VfxTracking` or `Module.All` (verified in
`references/ecommons/ECommons/ECommonsMain.cs`). `Init` subscribes to all five `ActorVfx`/`StaticVfx`
events and to `GameObjectCtor.GameObjectConstructorEvent`, which installs six hooks, plus
`Svc.ClientState.TerritoryChanged` and `Svc.Framework.Update`. Without the module nothing throws:
the list simply stays empty and every `TryGet...` returns `false`.

| Member | Meaning |
|---|---|
| `static readonly List<VfxInfo> TrackedEffects` | The live list. Take `lock(VfxManager.TrackedEffects)` when enumerating it directly. |
| `static readonly TimeSpan VfxExpiryDuration` | 30 seconds; not configurable. |
| `static readonly bool EnableDtorCulling` | `true`; not configurable. |
| `static bool EnableStaticVfxCreationTracking { get; set; }` | Default `false`: static VFX are tracked from their first run event, not from creation. |
| `static bool Logging { get; set; }` | Default `false`. |
| `static string? LoggingFilter { get; set; }` | Only log lines containing this text (case-insensitive). |
| `static List<string> WhitelistedVfxPathSearches { get; set; }` | When non-empty, only VFX whose path contains one of the strings are tracked. |
| `static bool TryGetVfxFor(ulong objectID, out List<VfxInfo> vfxList, bool searchCasterAsWell = false)` | Matches `TargetID` (and `CasterID` when asked). `objectID == 0` returns `false`. |
| `static bool TryGetVfxLike(string pathSearch, out List<VfxInfo> vfxList)` | Path contains `pathSearch`, case-insensitive. |

Both return `true` when at least one entry matched; `vfxList` is a new list (empty on `false`).
Object IDs are `IGameObject.GameObjectId` values.

Filter extensions on `List<VfxInfo>` (each returns a new list): `FilterToReady()` (non-static, or
static that has run), `FilterToTargeted()`, `FilterToTarget(ulong targetID)`,
`FilterToCaster(ulong casterID)`, `FilterToNoTarget()`, `FilterToNoCaster()`,
`FilterToPath(string pathSearch)`, `FilterToExactPath(string path)`,
`FilterYoungerThan(TimeSpan duration)`, `FilterToTargetRole(CombatRole role)`. They are declared
in an `extension(List<VfxInfo>)` block inside `VfxManager`; `using ECommons.GameHelpers;` brings
them in.

Removal rules: on the VFX destructor events; all entries targeting a newly constructed game object
(only while logged in, screen ready, player available and not busy); everything on territory
change; everything when `ConditionFlag.InCombat` goes from set to unset; entries older than 30 s,
checked about every 140 ms and only while in combat. Tracking is thus meant for combat.

`public record struct VfxInfo`: fields `ulong CasterID`, `bool HasRun`, `bool IsStatic`,
`string Path`, `PlacementData? Placement`, `long SpawnTick` (`Environment.TickCount64`; refreshed
when a static VFX runs), `ulong TargetID`, `long VfxID` (the VFX address); computed `long Age`
(ms), `float AgeSeconds`, `TimeSpan AgeDuration`. An unknown caster or target is `ulong.MaxValue`.
Nested `public class PlacementData { Vector3 Position; Quat Rotation; Vector3 Scale; }`.

`public struct VfxStruct` (explicit layout): `Vector3 Position` (0x50), `Quat Rotation` (0x60),
`Vector3 Scale` (0x70), `ulong ActorCasterID` (0x128), `ulong ActorTargetID` (0x130),
`ulong StaticCasterID` (0x1B8), `ulong StaticTargetID` (0x1C0). Use as
`var vfx = (VfxStruct*)vfxPtr;` on pointers from the `ActorVfx`/`StaticVfx` events.

`public struct Quat`: fields declared in the order `X`, `Z`, `Y`, `W` (sequential layout), with an
implicit conversion to `Vector4(X, Y, Z, W)`.

```csharp
// ECommonsMain.Init(pi, this, Module.VfxTracking);
if(VfxManager.TryGetVfxFor(Svc.Objects.LocalPlayer.GameObjectId, out var list))
{
    var markers = list.FilterToPath("vfx/lockon/").FilterYoungerThan(TimeSpan.FromSeconds(5));
}
```

`VfxManager.Dispose()` (internal, from `ECommonsMain.Dispose()`) clears the lists and
unsubscribes; note it runs after the hook classes were already disposed.

## `ObjectLife` (requires `Module.ObjectLife`)

`public static class ObjectLife`, namespace `ECommons.ObjectLifeTracker`
(`references/ecommons/ECommons/ObjectLifeTracker/ObjectLife.cs`). Enabled by `Module.ObjectLife`
or `Module.All`.

```csharp
public static Action<nint> OnObjectCreation = null;          // field, not an event
public static long GetLifeTime(this IGameObject o)           // ms since construction
public static float GetLifeTimeSeconds(this IGameObject o)
public static long GetSpawnTime(this IGameObject o)          // Environment.TickCount64 at construction
```

- `Init` is internal and deferred with a `TickScheduler`: the hook (same signature as
  `GameObjectCtor.Sig`, resolved with `ScanText`) is installed on the next framework tick after
  `ECommonsMain.Init`, not during it. Objects already in `Svc.Objects` at that moment are recorded
  with that moment as their spawn time.
- `GetSpawnTime` (and both `GetLifeTime...`) throw `Exception("Object life tracker was not
  initialized")` when the hook does not exist: module not requested, called in the first tick
  after init, or after dispose.
- An address without a record logs a warning and returns spawn time `0`, which makes
  `GetLifeTime` return the full `Environment.TickCount64` value.
- Records are keyed by object address and never removed; a reused address is overwritten by the
  next construction at that address.
- `OnObjectCreation` is invoked synchronously inside the detour, after the original constructor,
  with the object address, inside try/catch. The object is not in the object table yet at that
  point. It is a plain delegate field: use `+=`/`-=`, do not assign with `=` if other code in the
  plugin also uses it. For a table-safe, deferred notification use
  `GameObjectCtor.GameObjectConstructorEvent`.
- `ObjectLife.Dispose()` is internal and called by `ECommonsMain.Dispose()`.

```csharp
// ECommonsMain.Init(pi, this, Module.ObjectLife);
foreach(var o in Svc.Objects)
    if(o.GetLifeTimeSeconds() < 2f) { /* freshly spawned */ }
```

## Dispose order in `ECommonsMain.Dispose()`

Relevant calls, in source order, each wrapped in `GenericHelpers.Safe`: `ObjectLife.Dispose`
(early), then `DirectorUpdate.Dispose`, `ActionEffect.Dispose`, `MapEffect.Dispose`,
`SendAction.Dispose`, `ActorVfx.Dispose`, `StaticVfx.Dispose`, `GameObjectCtor.Dispose`,
`VfxManager.Dispose`, later `EzHookCommon.DisposeAll` (the three `MapEffect` extra hooks) and
finally `Purgatory.Purge` (`Svc.GameNetwork`). Each `Dispose` disables and disposes its hook and
sets it to `null`. Event subscriber lists and the `Init` callbacks are static and are not cleared.

## Gotchas

- `DirectorUpdate`, `MapEffect` and `SendAction` have no events; `Init` takes a single callback
  and throws on a second call. `ActionEffect.Init` exists but is obsolete; use the events.
- Removing the last event handler leaves the hook installed and running.
- A missing signature is reported only in the log (`Could not find ... signature`); the
  subscription or `Init` call appears to succeed. `Svc.GameNetwork` and `ObjectLife` are the
  exceptions: they use `ScanText` and throw.
- `ActionEffectSet.TargetEffects` and every `nint`/pointer argument are valid only during the
  callback.
- `ActionEffectSet.GetSpecificTypeEffect` returns `value`, which truncates damage above 65535; use
  `TargetEffect.GetSpecificTypeEffect(...)` and `entry.Damage`.
- `ActorVfx` create callbacks receive object addresses, `VfxInfo` stores `GameObjectId`s.
- Accessing `Svc.GameNetwork` once is enough to hook all zone packets for the plugin's lifetime.
- `VfxManager` and `ObjectLife` silently or loudly fail without their `Module` flag:
  `VfxManager` returns nothing, `ObjectLife` throws.
- `ObjectLife` is not ready until one framework tick after `ECommonsMain.Init`.
- No `ProcessPacket`/`ProcessActorControlPacket` hook class exists in ECommons at this commit.
