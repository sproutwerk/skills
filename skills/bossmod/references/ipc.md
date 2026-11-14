# IPC contracts

The provider is [IpcService.cs](source/BossMod/Services/IpcService.cs).
`StartAsync` registers functions under `BossMod.` and `StopAsync` unregisters
them. Use `InvokeFunc`, including for mutating methods that return `bool`.
These are in-process Dalamud call gates, not HTTP endpoints. This provider
registers no API-version or ready/unready event channel. The inspected
registration wrappers and delegates contain no caller-name allowlist or caller
identity check. That is a statement about this provider at this pin, not a
guarantee for other versions or forks.

Consumers should handle a missing or reloaded provider at the call boundary.
A successful read does not guarantee the provider will survive the next call.
The provider wrappers invoke their delegates directly; they do not marshal to
the framework thread. Schedule calls from workers onto that thread in the
consumer. Remove the consumer's own framework subscriptions on disposal; do
not unregister provider functions.

## Queries and configuration

The table gives suffixes after `BossMod.` and C# argument/return types.

| Suffix | Contract | Meaning |
| --- | --- | --- |
| `HasModuleByDataId` | `(uint dataId) -> bool` | Registry lookup by primary actor OID/BNpcBase; ignores disabled-module and active-state checks |
| `Configuration` | `(IReadOnlyList<string> args, bool save) -> List<string>` | Joins args with spaces and calls the config console parser; `save` is deprecated and unused |
| `Configuration.LastModified` | `() -> DateTime` | Local `DateTime.Now` at provider start, then on config `Modified` events; not a revision token |
| `Configuration.DisableModule` | `(string name, bool disable) -> bool` | Adds/removes a string in `BossModuleConfig.DisabledModules`; true only when the list changes |
| `Rotation.ActionQueue.HasEntries` | `() -> bool` | Whether any gathered action entry is non-manual; not confirmation that an action executed |

## Read-only state queries

Version 7.5.8.2 registers these parameterless functions, all returning `bool`
except `Rotation.GetActiveModules`. Older builds lack them, so treat a missing
channel as "unknown", not false.

| Suffix | Return | Meaning |
| --- | --- | --- |
| `Rotation.GetActiveModules` | `List<string>` | Full type names of the modules in every active preset, `VBM Multibox` included; duplicates are not removed |
| `Rotation.IsAutorotationActive` | `bool` | Any active preset holds a module outside the `BossMod.Autorotation.MiscAI` namespace |
| `Rotation.IsAutoTargetActive` | `bool` | An active preset other than `VBM Multibox` holds `BossMod.Autorotation.MiscAI.AutoTarget` |
| `Rotation.ActionTweaks.IsManualQueueEnabled` | `bool` | `ActionTweaksConfig.UseManualQueue` |
| `Rotation.ActionTweaks.IsSmartTargetingEnabled` | `bool` | `UseManualQueue && SmartTargets` |
| `Rotation.ActionTweaks.IsPreventMovingWhileCastingEnabled` | `bool` | `ActionTweaksConfig.PreventMovingWhileCasting` |
| `Rotation.ActionTweaks.IsMoveEscapeHatchHeld` | `bool` | `IMovementOverride.IsForceUnblocked()`: the configured escape-hatch input is held now |
| `AI.IsEnabled` | `bool` | `AIConfig.Enabled` |
| `AI.IsAutoTargetEnabled` | `bool` | `AIConfig.Enabled && !AIConfig.ForbidActions` |

The `Rotation.Is*Active` and `GetActiveModules` queries read the active preset
list only. They report what the selected presets contain, not whether a module
passed the class, level, or PvP filters and ran this frame, and they ignore a
selected encounter plan. The config-backed queries read cached config nodes and
have no side effects. The `AI.*` pair is read-only; use `Configuration` for a
requested AI config write.

## Configuration console

[ConfigRoot.ConsoleCommand](source/BossMod/Config/ConfigRoot.cs) accepts a
config type and field, with an optional value. Use exact names: abbreviations
can be ambiguous. Only fields with `PropertyDisplayAttribute` participate.
Two tokens read the value; three write it. A successful write returns an empty
list and fires the node's `Modified` event. Reads and errors both return text,
so inspect its meaning rather than treating a nonempty list as a failure.

Source-derived read example, uncompiled:

```csharp
var lines = pi.GetIpcSubscriber<
    System.Collections.Generic.IReadOnlyList<string>, bool,
    System.Collections.Generic.List<string>>("BossMod.Configuration")
    .InvokeFunc(new[] { "AIConfig", "Enabled" }, false);
```

The `save` argument does not make a write temporary. Config saves occur every
30 seconds and at hosted-service stop in
[SaveConfigService](source/BossMod/Services/SaveConfigService.cs).
`Configuration.DisableModule` bypasses `Modified.Fire()` and does not validate
the name. The registry compares disabled names to the module type's
`ToString()`, normally its fully qualified name. Do not assume the call unloads
an already loaded module or advances `LastModified`; verify behavior at the
target revision before using it as a stop mechanism.

## Preset database and activation

All suffixes below begin with `Presets.`. JSON is the provider's serialized
preset format, defined by [Preset.cs](source/BossMod/Autorotation/Preset.cs)
and [PlanPresetConverter](source/BossMod/Autorotation/PlanPresetConverter.cs).
Start from `Get` for the target installation rather than inventing module,
track, or option strings.

| Suffix after `Presets.` | Contract | Return and mutation semantics |
| --- | --- | --- |
| `ListAll` | `() -> List<string>` | Default presets followed by user presets, including hidden ones |
| `Get` | `(string name) -> string?` | Serialized preset, or null for no match |
| `Create` | `(string presetSerialized, bool overwrite) -> bool` | Parses one object, wraps it for schema conversion, modifies a user preset or appends one; rejects an exact user-name collision when overwrite is false |
| `Delete` | `(string name) -> bool` | Deletes an exact user-name match; false for no match, including a default-only name |
| `GetActive` | `() -> string?` | Name only when exactly one preset is active; null for zero or multiple |
| `GetActiveList` | `() -> List<string>` | Copy of all active names |
| `SetActive` | `(string name) -> bool` | Resolves name first, then clears all and activates it; false leaves the list unchanged if missing |
| `SetActiveList` | `(List<string> names) -> bool` | Resolves every name before clearing; false for any missing name; an empty list clears presets |
| `Activate` | `(string name) -> bool` | Adds a found preset unless already active; removes the force-disable sentinel through the manager |
| `Deactivate` | `(string name) -> bool` | Removes a found active preset; false when missing or inactive |
| `ClearActive` | `() -> bool` | Clears a nonempty active list; false if already empty |
| `GetForceDisabled` | `() -> bool` | True when the sole active preset is the special empty-name sentinel |
| `SetForceDisabled` | `() -> bool` | Replaces all active presets with that sentinel; false if already force-disabled |

[PresetDatabase](source/BossMod/Autorotation/PresetDatabase.cs) resolves names
using `CurrentCultureIgnoreCase`, searching defaults first. Create/delete use
exact string equality in the user list. Avoid default-name collisions and
case-only duplicates in consumer-owned presets. `Create` can throw for malformed
JSON or invalid deserialization; false is not its only failure mode. Database
saving catches and logs write errors, so true does not prove disk persistence.

The active list is shared state. `SetActive` and `SetActiveList` replace other
consumers' choices. Prefer targeted activation/deactivation of a uniquely
named preset when the requested behavior permits it. Record prior state for a
requested temporary change, but restore it only if it still reflects your
change; IPC provides no compare-and-swap or owner token. On failure, reread the
list and force-disable state before retrying. Do not automatically clear a
force-disable state set by the user.

`ClearActive` permits a selected encounter plan to run again. Use the explicit
force-disable control when the requested goal is to suppress automatic
rotation. Manual actions and other plugin controls have separate paths. See
[rotation and AI behavior](world-and-rotation.md#rotation-selection-and-stop-controls).

## Transient strategies

| Suffix after `Presets.` | Arguments | Return |
| --- | --- | --- |
| `AddTransientStrategy` | `string presetName, string moduleTypeName, string trackName, string value` | `bool` |
| `AddTransientStrategyTargetEnemyOID` | Same four strings, then `int oid` | `bool` |
| `ClearTransientStrategy` | `string presetName, string moduleTypeName, string trackName` | `bool` |
| `ClearTransientModuleStrategies` | `string presetName, string moduleTypeName` | `bool` |
| `ClearTransientPresetStrategies` | `string presetName` | `bool` |

`moduleTypeName` is a fully qualified registry key. The module must already
exist in the chosen preset. Track and option matching uses `InternalName`
with exact string comparisons. For enum tracks `value` names an option; float
and integer tracks parse a number and clamp it to their configured range.
Numeric parsing here uses the process culture, unlike the config console's
invariant float parser. Invalid numeric strings can throw.

An add replaces the transient setting for that track or appends one. It does
not activate the preset. Transient values override serialized values in
`ActiveStrategyOverrides`; the preset JSON writer excludes transient settings.
They are not automatically removed on deactivation, and there is no TTL or
caller token. Coordinate ownership and clear only the preset/module/track the
integration owns. Broad clear methods return true for a found container even
when its list was empty. Single-track clear returns false if no entry existed.

`TargetEnemyOID` selects `StrategyTarget.EnemyByOID`, which resolves the nearest
matching potential target to the player. Its `int oid` is the actor data/OID,
not an actor instance ID. It does not guarantee a target is present.

## Obstacle maps

The registered functions delegate to
[ObstacleMapManager](source/BossMod/Pathfinding/ObstacleMapManager.cs).

| Suffix | Contract |
| --- | --- |
| `ObstacleMap.Generate` | `(System.Numerics.Vector3 centerWorld, float radius, bool writeToFile) -> bool` |
| `ObstacleMap.GetGenerationStatus` | `() -> System.Threading.Tasks.TaskStatus` |
| `ObstacleMap.HasTempMap` | `() -> bool` |
| `ObstacleMap.ClearTempMap` | `() -> bool` |
| `ObstacleMap.EvaluateTempMapQuality` | `() -> BossMod.Pathfinding.ObstacleMapManager.BitmapQuality?` |

`Generate` queues work for a later framework tick, calling
`vnavmesh.Nav.BuildBitmapBounded`. True means queued, not complete. False can
mean an unfinished generation or reset-lock conflict; a previous fault can
be rethrown. Without a task, status is `RanToCompletion`, so that value alone
does not prove a map was generated. For a temporary request, also verify
`HasTempMap` after completion. Persistent requests update the map database;
temporary requests still use a temporary bitmap file while building.

The quality return type is an internal BossMod value type across the plugin
boundary; use its exact assembly/type identity when binding it. The simpler
status calls avoid that dependency. Do not use generation as a read-only
availability probe, clear another caller's temporary map, or assume VBM owns
vnavmesh navigation. This IPC offers map generation, not a general path API.
