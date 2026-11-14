# Source pins and navigation

Reviewed on 2026-11-13. These pins identify the evidence inspected; they do
not establish exact parity with a user's build. Use the architectural
guidance to locate the contract, then recheck the target revision before
depending on a gate, signature or side effect.

| Source | Revision | Role |
| --- | --- | --- |
| [NightmareXIV/Lifestream](https://github.com/NightmareXIV/Lifestream/tree/b36320170faafc8315f7585edb45cfbcd34e7046) | `b36320170faafc8315f7585edb45cfbcd34e7046` | Head of `main` on 2026-11-13; tags `2.6.0.8` to `2.6.0.13`; main source snapshot |
| [NightmareXIV/ECommons](https://github.com/NightmareXIV/ECommons/tree/958af5c7e735e1633b92070a5b67f8d6c48c30f3) | `958af5c7e735e1633b92070a5b67f8d6c48c30f3` | Plugin's ECommons gitlink; `ErrorCode` and EzIPC registration inspected |
| [NightmareXIV/ECommons.IPC](https://github.com/NightmareXIV/ECommons.IPC/tree/43bbc79935a2454015f9b828452f4ef1e8c51fd7) | `43bbc79935a2454015f9b828452f4ef1e8c51fd7` | Plugin's gitlink; inspected, no source copied |
| Lifestream `ef759e9d3c3cd989b4af569c3f770ee9ba060922` | tags `2.6.0.4` to `2.6.0.7` | Previous commit on `main`; compared, no source copied |

ECommons `main` was at `bc38dc71340fa3b266de58e3d6d876a73313b61e` on
2026-11-13, 63 commits past the gitlink. `ECommons/ErrorCode.cs` and
`LICENSE.md` are unchanged between the two. The gitlink stays the pin because
it is what this Lifestream commit builds against.

## Version mapping and the release gap

The repository's [Lifestream.csproj](source/Lifestream/Lifestream/Lifestream.csproj)
says `<Version>2.5.4.18</Version>` at every commit from 2026-08-31 to the
inspected head, so the csproj does not identify a release. Release versions
come from Git tags and are stamped into the build: the `latest.zip` asset of
each release carries a `Lifestream.json` with the matching `AssemblyVersion`
and `DalamudApiLevel` 15.

Tags do not identify the released source either. `main` has not moved since
`b36320170faafc8315f7585edb45cfbcd34e7046` ("Update Metadata", 2026-10-16),
and the six tags `2.6.0.8` through `2.6.0.13` (releases dated 2026-10-19 to
2026-11-10) all point at it. The release binaries name other commits in
their informational version strings:

| Release | String in `Lifestream.dll` | In the public repository |
| --- | --- | --- |
| `2.6.0.8` | `2.6.0.8+2dcf9668d4548290a06dd482ae8dcbbb7aeb98be` | no |
| `2.6.0.13` | `2.6.0.13+57ffc9c8d9b754db79ab98144746ca46b82a9b9c` | no |

Neither hash resolves through the GitHub API or exists in a full clone, so
the releases were built from commits that have not been pushed. Comparing
the strings in the two binaries shows `2.6.0.13` contains code this snapshot
lacks:

- a `Lifestream.Tasks.CosmicExploration` namespace with `TaskTravelToPlanet`,
  a `WKSPlanetSelect` addon name, `planet` and `cosmicplanet` strings and
  two "Cosmic Exploration Planet Badge" settings (the public source has only
  the `/li cosmic` alias);
- built-in alias JSON named `Crucible` (alias `crucible`) and
  `keyboard brawler` (alias `keyboard`), and an `"Alias":"mb"` value on
  `mbShortcut`; the pinned `StaticAlias.cs` has neither alias and no `Alias`
  value on `mbShortcut`;
- `TaskTpAndGoToWard.Enqueue` with a nullable plot and an extra
  `aetherytePartial` string parameter, plus `TryParseWard` and two address
  patterns that end at the ward number or at a trailing non-numeric word.

The `2.6.0.8` binary differs from the pin in the other direction too. It
holds the `Crucible` alias, and it has localized menu-text literals where
the pinned source already uses `GameText` and `GameTextPattern` (merged on
`main` in September). So the release line is not the public branch plus a
few commits; the two have diverged.

All 62 `[EzIPC]` method names in the pinned
[IPCProvider.cs](source/Lifestream/Lifestream/IPC/IPCProvider.cs) still occur
in the `2.6.0.13` binary's metadata strings. That is a name check only. The
binary was not decompiled, so parameter lists, return types, busy checks and
any gate added after the pin are unverified, and the `/li` grammar in
[commands](commands.md) omits whatever the unpublished commits added. Treat
this skill as a description of the public source, and probe the installed
plugin (`HasFunction`/`HasAction`, a caught first call) before relying on a
gate.

Tags `2.6.0.4` through `2.6.0.7` point at `ef759e9`. `git log
ef759e9..b363201` contains one commit, and `git diff` between them touches
only `CONTRIBUTING.md`; every C# file, including `IPC/IPCProvider.cs`
(SHA-256 `71580c12c025621a5473ae13854316230c00c8eae45f6772df532694f670666c`),
is byte-identical at both. A second remote branch, `tw13`
(`5267764ac2b78aca5870ede33d8dc9903ca3ff38`, 2026-07-31), is older than the
pin and was not reviewed.

## Release notes, 2.6.0.9 to 2.6.0.13

**From release notes, not verified against source.** No public commit
corresponds to any of these releases. The left columns quote upstream; the
last column is string evidence from the `2.6.0.13` binary, which shows a
feature exists but not how it behaves.

Sources read on 2026-11-13: the GitHub release bodies (one line each), the
`Changelog` field for Lifestream in NightmareXIV/MyDalamudPlugins
`pluginmaster.json` (it repeats the `2.6.0.13` line, with
`AssemblyVersion` `2.6.0.13` and `TestingAssemblyVersion` `2.6.0.11`), and the
commit log. The tags are lightweight and carry no annotation, the repository
has no changelog file, and no commit follows the pin.

| Release | Date | Channel | Upstream release note | Binary strings in `2.6.0.13` |
| --- | --- | --- | --- | --- |
| `2.6.0.8` | 2026-10-19 | stable | "Attempt to fix being stuck when trying to interact with limsa's aetheryte from alias" | not checked |
| `2.6.0.9` | 2026-10-26 | prerelease | "Added cosmic exploration quick planet change menu" | `TaskTravelToPlanet`, `SelectPlanet`, `WKSPlanetSelect`, the two Planet Badge settings |
| `2.6.0.10` | 2026-10-26 | prerelease | "Allow travelling to ward without specifying plot" | `TryParseWard`, nullable plot on `TaskTpAndGoToWard.Enqueue`, pattern `(%worlds)%delimiter(%city)%delimiter(W\|ward)%shortDelimiter%numeric` |
| `2.6.0.11` | 2026-10-27 | prerelease | "Read travel menu and confirmation text from game sheets - by ashdam" | `GameText`, `GameTextPattern`; this change is pull request 177, already in the pinned source |
| `2.6.0.12` | 2026-10-29 | stable | "Added `/li world residential ward aetheryte` command syntax support" | `aetherytePartial` parameter, pattern ending `%numeric%shortDelimiter([^0-9]+)`, log text "Ward address parsed:" |
| `2.6.0.13` | 2026-11-10 | stable | "Updated for world visit changes brought by 7.58" | nothing attributable |

What the notes mention, by area:

- **`/li` commands.** Two new address shapes: world, district and ward with
  no plot (`2.6.0.10`), and the same followed by an aetheryte name
  (`2.6.0.12`). The exact delimiters, which residential aethernet names
  match and where the character ends up are undocumented. The help text
  embedded in the binary does not list either form.
- **Aliases.** No note mentions them. The `crucible` and `keyboard` aliases
  and the `mb` alias value are known from binary strings only.
- **IPC.** No note mentions a gate. Whether the ward-only shapes reach
  `GoToHousingAddress` or the address-book tuple gates, which carry a plot
  number, is unknown. `ExecuteCommand` runs the `/li` parser, so it is the
  likely route to the new shapes.
- **World visits.** `2.6.0.13` changed the world-visit flow for game patch
  7.58. The sequences in [travel](travel.md) describe the older flow in the
  pinned source and may not match a 7.58 client.
- **Cosmic Exploration.** The planet menu is an overlay feature. No command
  or gate for it is documented upstream.

What a consumer should probe at runtime:

1. The installed version, from the plugin's `AssemblyVersion` through
   `IDalamudPluginInterface.InstalledPlugins`. Builds before `2.6.0.10`
   and `2.6.0.12` lack the respective ward shapes, and the testing channel
   stopped at `2.6.0.11`.
2. Each gate before first use: subscribe, catch the first call, and treat a
   signature mismatch as "unsupported" instead of retrying.
3. Any ward-only or ward-plus-aetheryte command by its effect: send it only
   when `IsBusy()` is false, confirm `IsBusy()` turns true within a few
   frames, then wait for idle and check the territory and ward. A parser that
   does not know the shape falls through to world and aethernet matching and
   reports in chat only.
4. World changes on a 7.58 client: verify the current world after idle, with
   a deadline, and do not assume the dialog sequence in this reference.

## Dependency context

The project targets `net10.0-windows7.0` with `LangVersion` `preview`,
references Dalamud, `Dalamud.Bindings.ImGui`, Lumina, FFXIVClientStructs and
`InteropGenerator.Runtime` from `$(appdata)\XIVLauncher\addon\Hooks\dev\`,
and builds the submodules ECommons, ECommons.IPC, AutoRetainerAPI,
NightmareUI, NightmareUI.OtterGuiWrapper and OtterGui as project
references; `DalamudPackager 2.1.13` and `NotificationMasterAPI 1.0.0.1` are
NuGet packages. The [manifest](source/Lifestream/Lifestream/Lifestream.json)
declares `DalamudApiLevel` 15 and no version (the build adds it). The
[.gitmodules](source/Lifestream/.gitmodules) file lists eight submodules; the
gitlinks at the pin are ECommons `958af5c7e735e1633b92070a5b67f8d6c48c30f3`,
ECommons.IPC `43bbc79935a2454015f9b828452f4ef1e8c51fd7`, AutoRetainerAPI
`7ccf0f6b4c7923821a43ed1e92456c9d5d7132f2`, ClickLib
`a5f5809dfcc628a92e19b35cb2534efca9075aad`, FFXIVClientStructs (Limiana fork)
`0769d1f180f859688f47a7a99610e9ce10da946c`, NightmareUI
`aae510d183a6114baece32394a9359f002a33a5d`, NightmareUI.OtterGuiWrapper
`371b1f90633394805244b50f1c135e145c6aa98d`, OtterGui
`a08fda67dfb731f007234c78df32a2b3ddccc450`. No package restore or build was
performed for this documentation task; the submodules other than ECommons
and ECommons.IPC were not checked out.

vnavmesh is an optional runtime dependency reached only over IPC
([VnavmeshIPC.cs](source/Lifestream/Lifestream/IPC/VnavmeshIPC.cs)); the
upstream README names it as optional for grand-company pathing.
TextAdvance, Splatoon, Wotsit and AutoRetainer are likewise optional IPC
peers and are not documented here.

## Source map

Paths are relative to this reference directory. Each copied upstream file is
listed in [source-manifest.json](source-manifest.json) with its original
path, repository, full revision and SHA-256. Copies come directly from Git
objects, preserving bytes, line endings and notices. This is a selected
reference set, not a buildable checkout.

| Contract to verify | Source |
| --- | --- |
| Gate names, signatures, busy checks, which helper each gate calls | [IPCProvider.cs](source/Lifestream/Lifestream/IPC/IPCProvider.cs), [IpcUtils.cs](source/Lifestream/Lifestream/IPC/IpcUtils.cs) |
| `/li` parser, `TPAndChangeWorld`, framework update, `ActiveAetheryte` detection | [Lifestream.cs](source/Lifestream/Lifestream/Lifestream.cs) |
| Help text | [Lang.cs](source/Lifestream/Lifestream/Lang.cs) |
| `IsBusy`, `EnqueueTeleport`, `AdjustGateway`, `CanAutoLogin`, `ChangeCharacter`, address parsing, territory and aetheryte helpers | [Utils.cs](source/Lifestream/Lifestream/Utils.cs) |
| `CanTeleport`, `Telepo` call, unlocked-aetheryte scan | [TeleportService.cs](source/Lifestream/Lifestream/Services/TeleportService.cs) |
| `GetRealTerritoryType` interior mapping | [TerritoryWatcher.cs](source/Lifestream/Lifestream/Services/TerritoryWatcher.cs) |
| Instance detection and limits | [InstanceHandler.cs](source/Lifestream/Lifestream/Services/InstanceHandler.cs), [TaskChangeInstance.cs](source/Lifestream/Lifestream/Tasks/SameWorld/TaskChangeInstance.cs) |
| Aetheryte and shard table, 2D positions from `MapMarker`, visitable world lists | [DataStore.cs](source/Lifestream/Lifestream/Systems/Legacy/DataStore.cs), [TinyAetheryte.cs](source/Lifestream/Lifestream/Systems/Legacy/TinyAetheryte.cs) |
| Target, interact, aethernet menu, world list selection, automove | [WorldChange.cs](source/Lifestream/Lifestream/Schedulers/WorldChange.cs) |
| Gateway teleport and arrival waits | [TaskTpToAethernetDestination.cs](source/Lifestream/Lifestream/Tasks/SameWorld/TaskTpToAethernetDestination.cs) |
| Name matching, partial search, teleport fallback | [TaskTryTpToAethernetDestination.cs](source/Lifestream/Lifestream/Tasks/SameWorld/TaskTryTpToAethernetDestination.cs), [TaskAethernetTeleport.cs](source/Lifestream/Lifestream/Tasks/SameWorld/TaskAethernetTeleport.cs), [TaskAetheryteAethernetTeleport.cs](source/Lifestream/Lifestream/Tasks/SameWorld/TaskAetheryteAethernetTeleport.cs) |
| Same-DC world change and retry | [TaskTPAndChangeWorld.cs](source/Lifestream/Lifestream/Tasks/CrossWorld/TaskTPAndChangeWorld.cs), [TaskChangeWorld.cs](source/Lifestream/Lifestream/Tasks/CrossWorld/TaskChangeWorld.cs) |
| Cross-DC sequences and defaults | [CharaSelectVisit.cs](source/Lifestream/Lifestream/Systems/CharaSelectVisit.cs), [TaskConnectAndOpenCharaSelect.cs](source/Lifestream/Lifestream/Tasks/Login/TaskConnectAndOpenCharaSelect.cs) |
| Property shortcut gates, inn table, housing aetheryte IDs, `OnHouseEnterError` | [TaskPropertyShortcut.cs](source/Lifestream/Lifestream/Tasks/Shortcuts/TaskPropertyShortcut.cs), [HousePathData.cs](source/Lifestream/Lifestream/Data/HousePathData.cs), [HouseEnterMode.cs](source/Lifestream/Lifestream/Data/HouseEnterMode.cs) |
| Waypoint following, timeout, vnavmesh collision, shared state | [FollowPath.cs](source/Lifestream/Lifestream/Movement/FollowPath.cs), [MovementConfig.cs](source/Lifestream/Lifestream/Movement/MovementConfig.cs), [Movement/README.md](source/Lifestream/Lifestream/Movement/README.md) |
| vnavmesh bindings and the housing path generator | [VnavmeshIPC.cs](source/Lifestream/Lifestream/IPC/VnavmeshIPC.cs), [TaskGeneratePath.cs](source/Lifestream/Lifestream/Tasks/Utility/TaskGeneratePath.cs) |
| Alias format, command kinds, per-kind behavior, `CanExecute`, shipped alias examples | [CustomAlias.cs](source/Lifestream/Lifestream/Data/CustomAlias.cs), [CustomAliasCommand.cs](source/Lifestream/Lifestream/Data/CustomAliasCommand.cs), [CustomAliasKind.cs](source/Lifestream/Lifestream/Data/CustomAliasKind.cs), [StaticAlias.cs](source/Lifestream/Lifestream/Data/StaticAlias.cs) |
| Configuration defaults | [Config.cs](source/Lifestream/Lifestream/Data/Config.cs), [LiCommandBehavior.cs](source/Lifestream/Lifestream/Data/LiCommandBehavior.cs) |
| Enum values used in signatures | [WorldChangeAetheryte.cs](source/Lifestream/Lifestream/Enums/WorldChangeAetheryte.cs), [ResidentialAetheryteKind.cs](source/Lifestream/Lifestream/Enums/ResidentialAetheryteKind.cs), [AddressBookEntry.cs](source/Lifestream/Lifestream/Data/AddressBookEntry.cs) |
| Task time limits | [TaskSettings.cs](source/Lifestream/Lifestream/Tasks/TaskSettings.cs) |
| `ErrorCode` values | [ErrorCode.cs](source/ECommons/ECommons/ErrorCode.cs) |
| Project, manifest, submodules, upstream README and contribution policy | [Lifestream.csproj](source/Lifestream/Lifestream/Lifestream.csproj), [Lifestream.json](source/Lifestream/Lifestream/Lifestream.json), [.gitmodules](source/Lifestream/.gitmodules), [README.md](source/Lifestream/README.md), [CONTRIBUTING.md](source/Lifestream/CONTRIBUTING.md) |

For unbundled implementations use the matching full checkout. In
particular, the character-select automation lives in
`Lifestream/Tasks/Login/TaskChangeCharacter.cs`,
`Lifestream/GUI/Windows/CharaSelectOverlay.cs` and
`Lifestream/Schedulers/DCChange.cs`; the residential and custom aethernet
systems in `Lifestream/Systems/Residential/` and
`Lifestream/Systems/Custom/`; the address-book travel in
`Lifestream/Tasks/CrossDC/TaskTpAndGoToWard.cs`. An absent file in this set
does not prove the behavior is absent upstream.

## Refresh and verification

Keep the request/wait/verify discipline separate from revision-specific
observations. On refresh, check whether `main` has moved and whether the
newest release binary's informational version names a public commit; then
diff `IPC/IPCProvider.cs`, `Lifestream.cs`, `Utils.cs` and the task files
against the previous pin and follow each changed gate into its
implementation. Validate copied hashes against the pinned Git objects, local
links and frontmatter.

On 2026-11-13 every bundled file was re-hashed against its pinned Git object
and matched. The sequences described here were reviewed against source only;
compilation, game behavior and live travel were not tested. See
[NOTICE.md](../NOTICE.md) for the unresolved Lifestream source grant.
