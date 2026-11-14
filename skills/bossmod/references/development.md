# Replay and development workflow

## Build the target project

The snapshot records version `7.5.8.2` (set once in `Common.targets` for all
projects), `Dalamud.NET.Sdk/15.0.0`, and
.NET SDK selection `10.0.0` with `latestFeature` roll-forward. Resolve the
installed SDK's target framework and libraries from the target checkout.
[Common.targets](source/Common.targets) defaults `LocalCS` and `LocalDalamock`
to false; the recorded dependency gitlinks do not mean those sources are used
in a default build. [Provenance](provenance.md) records the pins.

For an encounter change in a full checkout with the target dependencies, a
local compilation command is:

```sh
dotnet build BossMod.Modules/BossMod.Modules.csproj -c Release
```

This command is guidance, not a build result for this skill. The curated
snapshot is incomplete and cannot be built on its own. Review project build
targets before choosing Debug:
[Module.targets](source/Module.targets) copies Debug module DLLs into
`$(AppData)\vbm\modules-dev` after build. Use a separate development environment
when loading changed modules. Do not silently replace a user's live plugin
files to validate documentation.

## Module packs and lifecycle

[ModulePackService](source/BossMod/Services/ModulePackService.cs) loads DLLs
from `Plugin.GetStorageDir()` plus `modules-dev` for dev installs or `modules`
otherwise. The storage root is the platform application-data directory plus
`vbm`. It uses collectible load contexts and reloads config, encounter,
rotation, zone, and replay-analyzer registries on changes. The built-in packs
are `BossMod.Autorotation.dll`, `BossMod.Modules.dll`, and
`BossMod.Ultimate.dll`. At startup the service copies each from the plugin
directory into the module directory under a named global mutex, skipping the
copy when the destination's file version is equal or higher. A pack you built
with a higher file version therefore survives a plugin update, and one with an
equal or lower version is not refreshed by rebuilding the plugin alone.

Registry changes cause the boss manager to unload existing modules and
reconsider world actors. Avoid static references that keep unloaded module
assemblies or actors alive. Match host types and dependencies through the
existing pack build setup; copying a second BossMod assembly into a pack can
create incompatible type identities.

The plugin uses a DalaMock hosted-plugin container. IPC registration belongs
to the hosted service, UI/event hooks to their owning services, and encounter
subscriptions to disposable module managers. When adding a service, follow its
hosted start/stop lifecycle. `TickService.StopAsync` removes UI and game events;
its disposal closes windows. Do not add constructor-only subscriptions without
an equivalent teardown path.

## Replay-based checks

Use an existing user-authorized replay file. `/vbm r` toggles replay management
in a client, and the Debug executable has a DalaMock host. The latter still
needs Dalamud libraries, game data, and local paths: the sample
[Program.cs](source/BossMod/Framework/Program.cs) contains machine-specific
Windows paths and a `cimgui.dll` cleanup workaround. Inspect and adapt it in an
isolated dev setup before running; it is not a portable headless test command.

[ReplayParserLog](source/BossMod/Replay/ReplayParserLog.cs) parses recorded
operations. [ReplayPlayer](source/BossMod/Replay/ReplayPlayer.cs) applies groups
with the same timestamp, and `AdvanceTo(timestamp, update)` calls the update
callback after each group. `Reset()` replaces its `WorldState`; it does not
rewind the existing world in place. The
[replay viewer's MoveTo](source/BossMod/Replay/Visualization/ReplayDetailsWindow.cs)
disposes and reconstructs managers when seeking backward. Follow that pattern
in a harness, including a fresh hint builder and rotation manager where used.
At this pin `AIHintsBuilder` takes an `ObstacleMapManager` as a fourth
constructor argument and no longer owns or disposes it; the viewer passes a new
one built on the replay world. Its update callback runs both the boss module
manager and the active zone module. The parser now stops with a logged warning
on a truncated replay, so a short timeline can mean an incomplete file.

Check observable mechanics at the relevant timestamps:

- A module registers for the intended primary OID, loads, then activates at the
  intended pull condition. A registry hit alone is insufficient.
- A component activated mid-cast handles the seeded cast/status state without
  duplicating historical events. Its AOE origin, rotation, and activation match
  the replay, including interrupted casts or disappearing helpers.
- Damage resolution removes or advances the mechanic. Phase changes and wipe
  resets discard stale entries; a backward seek produces the same state again.
- Player text, arena shapes, and AI forbidden zones agree. Check safe and unsafe
  positions, not only whether something is drawn.

Use replay analysis for cast IDs, status/icon/tether events, and state
transition timings. Expected timeline durations can differ with HP transitions
or early kills. A replay cannot establish native hook compatibility, IPC
availability, or actual action/movement success in the installed client.

## Diagnose a failed integration

| Symptom | Inspect first |
| --- | --- |
| `HasModuleByDataId` is true but no encounter appears | Registry filters, disabled module/category, incomplete flag, actor type, and pull/FATE conditions |
| Older module fails to compile | `ModuleInit`, callback signatures, enum/shape types, and fork identity |
| `GetActive` returns null after activation | `GetActiveList`: multiple presets are valid, and AI may add its compatibility preset |
| Clearing rotation did not stop actions | Selected plan and `GetForceDisabled`; clear allows fallback to plans |
| Transient override returns false | Exact module key, track/option internal names, and whether that module is already in the preset |
| Create returned true but restart lost the preset | Preset database save warnings and filesystem state |
| Hints vanish next frame | Builder clears/rebuilds hints; persist mechanic state in a component, not the shared output |
| Map generation returned true but there is no map | Task status, temp-map existence, vnavmesh provider availability, and generation exceptions |

On an unexpected mutation result, stop dependent automation and read current
state before retrying. Keep logs/replays private unless the user authorizes
sharing; diagnostic commands can include character and encounter details.
