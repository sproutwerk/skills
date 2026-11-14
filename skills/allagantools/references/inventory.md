# Inventory data and ownership

Read [IPC contracts](ipc.md) for exact signatures. The source is selected from
InventoryTools and its pinned CriticalCommonLib, not a promise of compatibility
with a later game build.

## Owners and identifiers

`CurrentCharacter` returns `ICharacterMonitor.ActiveCharacterId`. Player
characters use content IDs; retainers have their own IDs. Inventory rows use
`InventoryItem.RetainerId` as the storage-owner ID even when the owner is a
player character, free company, or house. Preserve it as `ulong`. Do not use a
player object ID, character name, world ID, or retainer list position instead.

[CharacterMonitor.BelongsToActiveCharacter](source/CriticalCommonLib/Services/CharacterMonitor.cs)
returns false without an active character. It accepts the active character
itself and records whose `OwnerId` matches that character. Free-company storage
must match the active character's `FreeCompanyId`; housing records use their
`Owners` collection. Housing IDs have a library-specific conversion in
`ConvertHouseId`; do not manufacture them from native pointers or assume they
are player content IDs.

`GetItemCountsByCharacter` retains each storage owner's key. It does not fold a
retainer's count into the player character's entry. To total the selected scope,
sum the returned values once. When querying several player characters with
shared storage enabled, do not add their totals blindly: the same free-company
or housing inventory can belong to more than one character.

`GetCharactersOwnedByActive` only enumerates owners already present in the
inventory dictionary. It is neither a complete account roster nor a character
name lookup. This IPC service has no registered name/world metadata query or
per-container coverage query. A consumer that requires those must obtain them
through a separately verified interface or report that they are unavailable.

## Types, categories, and quality

Container types use CriticalCommonLib's `InventoryType` numeric values.
`ItemCount` accepts `int` so `-1` can mean all types; other type queries use
`uint`/`uint[]`. The provider compares `SortedContainer`, which represents the
library's sorted inventory, not necessarily the original `Container` field.
The [pinned enum](https://github.com/Critical-Impact/CriticalCommonLib/blob/420f6880b16497a6d6356f6dbeda963364e4d5ef/Enums/InventoryType.cs)
includes custom containers such as Armoire and GlamourChest. Do not assume every
value maps directly to a native FFXIVClientStructs enum.

Category queries use the separate
[InventoryCategory enum](source/CriticalCommonLib/Models/InventoryCategory.cs):

| Values | Categories |
| --- | --- |
| 1, 2, 3 | CharacterBags, CharacterSaddleBags, CharacterPremiumSaddleBags |
| 4, 5, 6, 7 | RetainerBags, CharacterArmoryChest, CharacterEquipped, RetainerEquipped |
| 8, 9, 10, 11 | FreeCompanyBags, RetainerMarket, GlamourChest, Armoire |
| 12, 13 | Currency, Crystals |
| 14, 15, 16 | HousingInteriorItems, HousingInteriorStoreroom, HousingInteriorAppearance |
| 17, 18, 19 | HousingExteriorItems, HousingExteriorStoreroom, HousingExteriorAppearance |
| 99 | Other |

For example, categories `[1, 4]` select character bags and retainer bags. Passing
those numbers to `ItemCountOwned` instead selects container types 1 and 4.
An empty category array means all categories; an empty type array means none.
`shared=false` restricts character types to Character and Retainer even if a
shared-storage category was explicitly requested.

These APIs compare base item IDs. All ordinary count functions include NQ and
HQ together. `ItemCountHQ` tests flags for exact equality with `HighQuality`,
not bit containment. It is not a general quality predicate for combined flags.
Do not encode HQ by adding a million to the item ID. The category channels have
no HQ parameter; apply a verified raw-record flag predicate if the task needs
quality-specific category counts, while preserving the coverage limitations.

## Cached data and coverage

[ServiceConfigurator](source/InventoryTools/InventoryTools/Host/ServiceConfigurator.cs)
loads saved character records and inventory data. The
[inventory monitor](source/CriticalCommonLib/Services/InventoryMonitor.cs)
then updates inventories from observed game state. It skips retainer generation
unless the active retainer's required containers are known to the scanner;
housing updates skip containers absent from the scanner's `InMemory` set.
Other characters and closed storage can therefore retain older observations.
An active character ID does not make every stored inventory live.

Keep these consumer outcomes distinct:

| Outcome | What it establishes |
| --- | --- |
| Missing provider, false readiness, call failure | No usable query result |
| Successful count of zero | No matching quantity in the provider's current cached scope |
| Missing key in grouped counts | No matching returned entry for that owner; unseen, excluded, and empty inventories remain possible |
| Empty raw-record set | Unknown owner or absent/empty selected inventory can share this result |
| Positive count | Cached evidence of that quantity; not a guarantee that it is currently accessible |

`GetCharacterItems` returns an empty set for an owner absent from the inventory
dictionary. `GetCharacterItemsByType` also returns an empty set when the selected
inventory is absent. Neither response contains a freshness timestamp or an
explicit loaded/complete flag. Even a known owner's records cannot prove that
all of its storage has been seen. Label results as cached and explain coverage
when a decision depends on a zero. Do not fabricate timestamps or silently
convert unknown state into zero.

The raw methods include every non-null slot, including empty items; aggregate
counts use `AllItems`, which excludes `ItemId == 0`. Filter empty raw records
before computing occupied slots or item counts.

## Numeric records

The raw methods return `HashSet<ulong[]>`. Arrays retain reference equality, so
the set is not a guarantee of distinct rows by value. Treat results as unordered
records and copy any fields retained by the consumer. The pinned
[InventoryItem.ToNumeric](source/CriticalCommonLib/Models/InventoryItem.cs)
writes this layout:

| Index | Field |
| --- | --- |
| 0, 1 | Container, Slot |
| 2, 3 | ItemId, Quantity |
| 4, 5, 6 | Spiritbond, Condition, Flags |
| 7 through 11 | Materia0 through Materia4 |
| 12 through 16 | MateriaLevel0 through MateriaLevel4 |
| 17, 18, 19 | Stain, Stain2, GlamourId |
| 20, 21, 22 | SortedContainer, SortedCategory, SortedSlotIndex |
| 23, 24 | RetainerId (storage owner), RetainerMarketPrice |
| 25 onward | GearSets |

There is no schema version or record length negotiation on the channel. Check
for at least 25 fields before indexing and compare the target source on an
upgrade; a length check alone cannot detect reordered fields.

Two source-level hazards at this pin affect consumers:

- `ToNumeric` allocates `new ulong[25 + GearSets?.Length ?? 0]`. With null
  `GearSets`, the expression yields length zero, so subsequent writes can throw.
  The field defaults to an empty array, but it is nullable. Treat serialization
  failure as unavailable data, not an empty inventory.
- `FromSerializedItem` assigns `RetainerId = (uint)serializedItem[23]`, narrowing
  the owner ID to 32 bits. Avoid that helper when retaining owner identity;
  decode field 23 directly as `ulong` into a consumer-owned record.

Both observations come from source inspection; they were not reproduced in a
running client. Prefer the aggregate query for counts when raw slots are not
needed. If raw data is required, this original parser preserves owner width:

```csharp
static bool TryReadItem(ulong[] row, out ulong owner,
    out uint itemId, out uint quantity)
{
    owner = 0;
    itemId = quantity = 0;
    if (row.Length < 25 || row[2] > uint.MaxValue || row[3] > uint.MaxValue)
        return false;
    owner = row[23];
    itemId = (uint)row[2];
    quantity = (uint)row[3];
    return true;
}
```

Skip records with `itemId == 0` after successful parsing. This example was
checked against the numeric layout; it was not compiled or run in game.
