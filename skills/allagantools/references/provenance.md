# Source pins and navigation

Reviewed on 2026-11-13. Upstream can change multiple times daily, including
breaking APIs and behavior. This pinned snapshot inevitably lags and does not
promise exact parity with an installed plugin or continuous synchronization.
Use it to understand the integration architecture and locate the contracts to
verify against the user's target build.

| Project | Revision | Role |
| --- | --- | --- |
| [InventoryTools](https://github.com/Critical-Impact/InventoryTools/tree/ad1ff01fede1848dfcd05b76a9ae13b0760d1222) | `ad1ff01fede1848dfcd05b76a9ae13b0760d1222` | Allagan Tools provider; `main` head at review, tagged `15.1.1` |
| [CriticalCommonLib](https://github.com/Critical-Impact/CriticalCommonLib/tree/420f6880b16497a6d6356f6dbeda963364e4d5ef) | `420f6880b16497a6d6356f6dbeda963364e4d5ef` | Plugin's recorded gitlink; also the `main` head at review |
| [OtterGui](https://github.com/Ottermandias/OtterGui/tree/79771ee5f3d463f02c63bebbedaa0aff49e59718) | `79771ee5f3d463f02c63bebbedaa0aff49e59718` | Plugin gitlink; source not bundled here |

## Identity and version evidence

The pinned [README](https://github.com/Critical-Impact/InventoryTools/blob/ad1ff01fede1848dfcd05b76a9ae13b0760d1222/README.md)
is titled Allagan Tools. The pinned
[plugin manifest](https://github.com/Critical-Impact/InventoryTools/blob/ad1ff01fede1848dfcd05b76a9ae13b0760d1222/InventoryTools/InventoryTools.json)
sets `Name = Allagan Tools`, `Author = Critical_Impact`,
`InternalName = InventoryTools`, and
`RepoUrl = https://github.com/Critical-Impact/InventoryTools`.
The [local copies](source/InventoryTools/README.md) and
[manifest](source/InventoryTools/InventoryTools/InventoryTools.json) preserve
that evidence unchanged.

The latest [changelog entry](source/InventoryTools/InventoryTools/CHANGELOG.md)
at the pin is `15.1.1`, dated 2026-11-12. The changelog explains that release
automation adds a leading `1.` for plugin versions; the JSON manifest does not
itself provide an assembly version. The pinned commit carries the upstream tag `15.1.1` and the same date. These
source dates do not establish compatibility with a later client update.
Changelog `15.0.12` introduces the category count and owner-breakdown IPC
functions and records empty returns for owners with no loaded inventory.
The `15.1.0` and `15.1.1` entries cover the compendium windows and list no IPC
changes; between `15.0.13` and this pin `IPCService.cs` changed only in
whitespace.

## Durable guidance and revision-specific contracts

The durable approach is to query through registered IPC, keep provider absence
separate from zero, retain storage-owner identity, and treat inventories as
observations with limited coverage. Consumer subscriptions belong to the
consumer's lifecycle; shared plugin selections belong to the user.

Channel names, generic signatures, numeric record offsets, enum values, return
quirks, and event timing are revision-specific. Recheck the provider registrations
and method bodies before relying on them. For dependency-backed behavior,
resolve the target plugin's gitlink first; the dependency's latest branch may
describe a different API.

## Dependency context

Both projects declare `Dalamud.NET.Sdk/15.0.0`. The plugin references
CriticalCommonLib and OtterGui projects. Its package declarations include
AllaganLib.Interface `2.3.3`, AllaganLib.Monitors `2.2.9`, DalaMock.Host/Shared
`6.2.2`, and Microsoft.Extensions.Hosting `10.0.10`.
CriticalCommonLib declares AllaganLib.GameSheets `2.3.1`, AllaganLib.Monitors
`2.2.9`, and AllaganLib.Shared `2.2.0`, among other dependencies. The project files and lock
files are bundled for exact requested and resolved versions; they are not an
instruction to restore packages or run analyzers during documentation work.

Primitive IPC consumers need only the matching Dalamud runtime. Direct library
use also needs its dependency graph and native/runtime compatibility. Allagan
Tools owns its monitor instances and DI services; linking a second copy of
CriticalCommonLib does not share those instances.

## Source map

Paths below are relative to this reference. The partial snapshot is for source
inspection and cannot build the plugin on its own.

| Question or claim to recheck | Evidence |
| --- | --- |
| Exact channel strings, generic types, count predicates, return values, teardown | [IPCService.cs](source/InventoryTools/InventoryTools/IPC/IPCService.cs) |
| Hosted-service registration, DI scope, internal filter and craft-source registration | [InventoryToolsPlugin.cs](source/InventoryTools/InventoryTools/InventoryToolsPlugin.cs) |
| Category matching and shared-storage exclusion | [InventoryScopeCalculator.cs](source/InventoryTools/InventoryTools/Logic/Editors/InventoryScopeCalculator.cs), [InventorySearchScope.cs](source/InventoryTools/InventoryTools/Logic/Editors/InventorySearchScope.cs) |
| Key/name resolution, active selections, new craft lists | [ListService.cs](source/InventoryTools/InventoryTools/Lists/ListService.cs) |
| Saved-data loading | [ServiceConfigurator.cs](source/InventoryTools/InventoryTools/Host/ServiceConfigurator.cs) |
| Count regression cases and unknown-owner expectations | [IpcItemCountTests.cs](source/InventoryTools/InventoryToolsTesting/Tests/IpcItemCountTests.cs) (inspected, not executed) |
| Ownership, active character/retainer changes, converted housing IDs | [CharacterMonitor.cs](source/CriticalCommonLib/Services/CharacterMonitor.cs), [Character.cs](source/CriticalCommonLib/Models/Character.cs) |
| Cache loading, raw/aggregate differences, item deltas and threading | [InventoryMonitor.cs](source/CriticalCommonLib/Services/InventoryMonitor.cs), [Inventory.cs](source/CriticalCommonLib/Models/Inventory.cs) |
| Numeric schema, nullable gear sets, owner-width narrowing | [InventoryItem.cs](source/CriticalCommonLib/Models/InventoryItem.cs) |
| Category values | [InventoryCategory.cs](source/CriticalCommonLib/Models/InventoryCategory.cs) |
| Dependency pins and package versions | [.gitmodules](source/InventoryTools/.gitmodules), [plugin project](source/InventoryTools/InventoryTools/InventoryTools.csproj), [plugin lock](source/InventoryTools/InventoryTools/packages.lock.json), [library project](source/CriticalCommonLib/CriticalCommonLib.csproj), [library lock](source/CriticalCommonLib/packages.lock.json) |

Follow unbundled implementation paths, including `Lists/ListFilterService.cs`,
`Services/ConfigurationManagerService.cs`, and `CriticalCommonLib/Crafting/`,
in the full source at the matching revisions when changing those internals.
An absent file in this curated snapshot is not evidence that the upstream
feature or API is absent.

## Refresh and verification limits

[The manifest](source-manifest.json) records upstream paths, full revisions,
and SHA-256 hashes for every copied file, including duplicate license texts.
Copy source from Git objects unchanged when refreshing. Inspect provider
registration, method bodies, callers, and dependency types together, then update
this reference and the file manifest.
Keep [notices](../NOTICE.md) and file-level attribution intact.

Validation is source review and structural checking: frontmatter, reference
paths and anchors, and copied bytes against Git objects. Examples are
source-derived, uncompiled, and untested in game. No game actions or provider
installation checks were used.
