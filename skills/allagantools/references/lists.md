# Saved filters and craft lists

The registered surface is in
[IPCService](source/InventoryTools/InventoryTools/IPC/IPCService.cs), with lookup
and activation behavior in
[ListService](source/InventoryTools/InventoryTools/Lists/ListService.cs).
Each channel below starts with `AllaganTools.` and is a registered function.

## Read existing lists

| Channel suffix | Arguments | Result |
| --- | --- | --- |
| `GetSearchFilters` | none | `Dictionary<string,string>`: key to name, for SearchFilter lists |
| `GetCraftLists` | none | `Dictionary<string,string>`: key to name, excluding default craft lists |
| `GetFilterItems` | `string keyOrName` | `Dictionary<uint,uint>`: item ID to quantity, with semantics dependent on list type |
| `GetCraftItems` | `string keyOrName` | `Dictionary<uint,uint>`: output item IDs to required quantities |
| `GetRetrievalItems` | none | `Dictionary<uint,uint>`: quantities to retrieve for the active craft list |

Use returned keys for subsequent calls. Lookup checks keys first, then exact
names, selecting the first matching name. Duplicate names are ambiguous.
The dictionary getters return new collections; editing those collections does
not modify plugin configuration.

`GetFilterItems` behaves differently for each supported type:

- Craft filters sum `QuantityRequired` for output items. They do not return the
  entire expanded ingredient tree or refresh the craft calculation here.
- Search and sorting filters call `ListFilterService.RefreshList` synchronously
  and sum the returned inventory quantities by base item ID. This can do more
  work than a cached count query, so avoid invoking it every draw frame.
- Game-item filters return matching item IDs with quantity zero. Zero here is
  a search-result convention, not an ownership count.

The provider has no branch for history or curated lists. An unknown list, an
unsupported type, and a supported list with no results can all yield an empty
dictionary. `GetCraftItems` likewise returns empty for a non-craft or unknown
list. Resolve the key through `GetCraftLists` or `GetSearchFilters` when those
registries apply, and preserve uncertainty for other list types.

`GetRetrievalItems` reads the active craft list's existing retrieval calculation.
It omits zero quantities and combines entries with the same item ID across
output and ingredient occurrences. The underlying
[GetQuantityToRetrieveList](https://github.com/Critical-Impact/CriticalCommonLib/blob/420f6880b16497a6d6356f6dbeda963364e4d5ef/Crafting/CraftList.cs#L2801)
keys entries by `(ItemId, IsOutputItem)`. The IPC drops `IsOutputItem` and
returns no quality, owner, or slot detail. It neither moves
items nor proves that storage is accessible. No active craft list also returns
an empty dictionary.

This original read-only example lists saved craft-list keys and names. Apply
the readiness and error handling from [IPC lifecycle](ipc.md#startup-failure-and-disposal)
in the calling plugin:

```csharp
using System.Collections.Generic;
using Dalamud.Plugin;

static Dictionary<string, string> ReadCraftListNames(IDalamudPluginInterface pi)
    => pi.GetIpcSubscriber<Dictionary<string, string>>("AllaganTools.GetCraftLists")
        .InvokeFunc();
```

This example is source-derived and uncompiled. A successful call only describes
saved lists at the time of that call.

## Configuration mutations

Use these only for the user's requested list or selection change. They affect
Allagan Tools' global active selections or saved crafting state, so consumer
unload should not automatically clear a selection it no longer owns.

| Channel suffix | Arguments | Result |
| --- | --- | --- |
| `EnableUiFilter`, `ToggleUiFilter` | `string keyOrName` | `bool` |
| `DisableUiFilter` | none | `bool` |
| `EnableBackgroundFilter`, `ToggleBackgroundFilter` | `string keyOrName` | `bool` |
| `DisableBackgroundFilter` | none | `bool` |
| `EnableCraftList`, `ToggleCraftList` | `string keyOrName` | `bool` |
| `DisableCraftList` | none | `bool` |
| `AddNewCraftList` | `string name, Dictionary<uint,uint> items` | `string` new list key |
| `AddItemToCraftList`, `RemoveItemFromCraftList` | `string keyOrName, uint itemId, uint quantity` | `bool` |

Enable/toggle calls return false when they cannot resolve a list. Their provider
code does not enforce a list-type restriction before activation. Resolve a craft
key from `GetCraftLists` before selecting it as a craft list. Disable calls return
true after calling the clear method, even when no selection changed. These
controls select or clear Allagan Tools lists; they do not enable/disable the
plugin, load/unload it, or execute crafts. This service exposes no active-list
key getter, so a consumer cannot assume it can read back and restore a prior
selection entirely through this IPC surface.

At this pin, `AddItemToCraftList` and `RemoveItemFromCraftList` return false
unconditionally. For a resolved craft list they still call the craft-list
mutation with `ItemFlags.None` and set `NeedsRefresh = true`. Never use false as
a reason to repeat the operation. Read the list before and after the requested
change; a retry can add or remove quantity twice. If an exception interrupts a
mutation, stop dependent work and establish the resulting state before retrying.

`AddNewCraftList` creates the list first, then adds each dictionary entry and
returns its key. Failure partway through can leave a list behind; do not blindly
retry by name. Keep the returned key when available. No delete-list IPC function
is registered in this service.

## Extension boundaries

The plugin registers filter classes and `IGameCraftSource` implementations from
its own assembly through DI. These are implementation hooks for development
inside the plugin, not registered IPC callbacks that another plugin can supply.
This pin exposes saved-list queries, selection controls, and craft-list writes.
It does not register arbitrary search-expression evaluation, external filter
registration, a craft execution API, or item transfers. User-facing search,
craft planning, and inventory highlighting do not establish those contracts.

For requested implementation changes, inspect the target's
`InventoryToolsPlugin.cs` registrations, `Lists/ListFilterService.cs`,
`Logic/Filters/`, and `CriticalCommonLib/Crafting/`. Only selected evidence is
bundled; follow those paths in the full checkout at the recorded pins before
editing filter or crafting internals.
