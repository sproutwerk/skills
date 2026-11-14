# IPC and lifecycle

The contract below comes from the pinned
[IPCService.cs](source/InventoryTools/InventoryTools/IPC/IPCService.cs).
All names have the prefix `AllaganTools.`. These are Dalamud call gates in the
same process, with exact generic argument types. No HTTP server or endpoint
registration was found in this integration surface.

## Read functions

Use `GetIpcSubscriber<arguments..., result>(channel).InvokeFunc(arguments...)`.
The table spells out the types in order; parameter names are explanatory.

| Channel suffix | Arguments | Result | Meaning |
| --- | --- | --- | --- |
| `IsInitialized` | none | `bool` | IPC startup flag |
| `CurrentCharacter` | none | `ulong` | Active character ID; zero when none |
| `InventoryCountByType` | `uint type, ulong? owner` | `uint` | Sum of quantities across all items in that sorted container; null owner includes all cached owners |
| `InventoryCountByTypes` | `uint[] types, ulong? owner` | `uint` | Same sum over the selected types; empty array matches nothing |
| `ItemCount` | `uint itemId, ulong owner, int type` | `uint` | Base item ID, exact owner; `type = -1` includes all sorted containers; NQ and HQ combined |
| `ItemCountHQ` | `uint itemId, ulong owner, int type` | `uint` | Same scope, with flags exactly equal to `ItemFlags.HighQuality` |
| `ItemCountOwned` | `uint itemId, bool currentOnly, uint[] types` | `uint` | Known owners, optional active-character ownership, exact sorted-container list; empty array matches nothing |
| `ItemCountOwnedByCategory` | `uint itemId, bool currentOnly, uint[] categories, bool shared` | `uint` | NQ and HQ combined; empty categories means all; shared=false restricts owner types to character/retainer |
| `GetItemCountsByCharacter` | `uint itemId, bool currentOnly, uint[] categories, bool shared` | `Dictionary<ulong,uint>` | Same category scope, grouped by storage owner; does not fill missing owners with zero |
| `GetCharactersOwnedByActive` | `bool includeOwner` | `HashSet<ulong>` | Inventory dictionary keys belonging to the active character; includeOwner controls whether the active character itself is included |
| `GetCharacterItems` | `ulong owner` | `HashSet<ulong[]>` | Numeric records across that owner's cached inventories |
| `GetCharacterItemsByType` | `ulong owner, uint type` | `HashSet<ulong[]>` | Numeric records for one inventory type |

`InventoryCountByType(s)` measures item quantities, not occupied slots, unique
item IDs, or inventory capacity. `GetCharactersOwnedByActive(false)` is not a
retainer-only query: shared storage can satisfy the ownership predicate. See
[owner semantics](inventory.md#owners-and-identifiers).

The `ItemCountOwned` family requires an entry in the character monitor as well
as a matching item. `ItemCount` and the inventory-quantity functions do not
apply that known-character check. For older builds lacking the category
channels, do not silently substitute `ItemCountOwned`: its empty-array behavior
and shared-storage controls differ.

## Events

Subscribe with `GetIpcSubscriber<payload, bool>(channel).Subscribe(callback)`.
The final `bool` is a call-gate type parameter, not a callback return value.
These providers send messages; they do not register query functions.

| Channel suffix | Payload | Provider behavior |
| --- | --- | --- |
| `Initialized` | `bool` | Sends true after registering the IPC functions and event handlers |
| `RetainerChanged` | `ulong?` | Forwards the monitor's active retainer ID, including zero; nullable channel type must be preserved |
| `ItemAdded` | `(uint, FFXIVClientStructs.FFXIV.Client.Game.InventoryItem.ItemFlags, ulong, uint)` | Base item ID, flags, storage owner, positive quantity delta |
| `ItemRemoved` | same tuple | Same fields, with a positive removed quantity |

[InventoryMonitor](source/CriticalCommonLib/Services/InventoryMonitor.cs)
compares totals by `(itemId, flags, owner)` before and after regeneration.
Moving items between containers of the same owner can leave those totals
unchanged. Transfers between owners can produce a removal and an addition.
The provider omits gil (`ItemId == 1`). Cache loading and some clear/refresh
notifications have no `ItemChanges` payload, so no item IPC events follow.
Use the events to mark a view stale and query again; do not reconstruct a
per-slot inventory or transaction history solely from them.

## Startup, failure, and disposal

`IPCService` is an `IHostedService` registered by
[InventoryToolsPlugin](source/InventoryTools/InventoryTools/InventoryToolsPlugin.cs).
`StartAsync` registers functions, attaches monitor handlers, sets its readiness
flag, and sends `Initialized(true)`. A late consumer misses that message, so
subscribe and then query `IsInitialized`. A provider can unload between the
check and the next call; catch that call's failure too.

`StopAsync` detaches monitor handlers and unregisters query functions. At this
pin it does not send `Initialized(false)` or reset the readiness field. Do not
wait for a shutdown notification that this source does not emit. Treat failed
calls as unavailable and recheck on the next scheduled refresh or provider
startup event. Avoid a per-frame error-log loop.

The inventory monitor generates data in a background task, then schedules its
change event on the framework thread. Retainer changes are also scheduled on
that thread. Startup messages come from hosted-service startup, so do not assume
every callback has the same threading context. Keep event callbacks short and
schedule UI/game-dependent follow-up through the consumer's framework service.
Calls are synchronous; this provider does not promise a transactional snapshot
across multiple queries or a stable enumeration during background changes.

This original, source-derived example manages the lifecycle notification only.
Call `TryReadReady` from scheduled framework work; the event just invalidates
state. Add inventory-change subscriptions only when the view needs them.

```csharp
using System;
using Dalamud.Plugin;
using Dalamud.Plugin.Ipc;

sealed class AllaganToolsAvailability : IDisposable
{
    private readonly ICallGateSubscriber<bool, bool> initialized;
    private readonly ICallGateSubscriber<bool> readiness;
    private volatile bool dirty = true;

    public AllaganToolsAvailability(IDalamudPluginInterface pi)
    {
        readiness = pi.GetIpcSubscriber<bool>("AllaganTools.IsInitialized");
        initialized = pi.GetIpcSubscriber<bool, bool>("AllaganTools.Initialized");
        initialized.Subscribe(OnInitialized);
    }

    public bool Dirty => dirty;
    private void OnInitialized(bool _) => dirty = true;

    public bool TryReadReady(out bool ready)
    {
        ready = false;
        try
        {
            ready = readiness.InvokeFunc();
            dirty = false;
            return true;
        }
        catch (Exception)
        {
            dirty = true;
            return false;
        }
    }

    public void Dispose() => initialized.Unsubscribe(OnInitialized);
}
```

The caller must still perform periodic or user-requested readiness checks;
`Dirty == false` cannot detect unload. Log errors using the consumer's existing
rate-limited diagnostics. This snippet was not compiled or run in game.

## Dependency boundaries

Primitive count/list queries need no reference to InventoryTools or
CriticalCommonLib. Item events use the runtime's FFXIVClientStructs `ItemFlags`
type: replacing it with an integer changes the generic call-gate signature.
Raw records use BCL arrays but encode CriticalCommonLib's layout.

Direct use of `IInventoryMonitor`, `ICharacterMonitor`, `InventoryItem`, or the
crafting classes couples a plugin to the pinned library and its DI setup. A new
monitor instance in a consumer does not give access to Allagan Tools' existing
cache. Avoid reflecting into its container to acquire mutable services. For a
requested upstream implementation change, use the matching gitlink, inspect
service registration and teardown, and assess the GPL terms before copying code.
