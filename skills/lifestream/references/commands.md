# Command grammar

`/li` and `/lifestream` both route to `Lifestream.ProcessCommand(command, arguments)`
in [Lifestream.cs](source/Lifestream/Lifestream/Lifestream.cs). The IPC
action `ExecuteCommand(string arguments)` calls
`P.ProcessCommand("/li", arguments)`, so it behaves exactly like typing
`/li <arguments>` except that the `command == "/lifestream"` branch (open the
configuration on empty input) can never match. Every refusal inside the
parser is a chat notification or a log line; nothing is returned to the IPC
caller.

The user-facing help text lives in `Lang.Help`
([Lang.cs](source/Lifestream/Lifestream/Lang.cs)) and is what the
"Help" tab renders. The grammar below is read from the parser, which
accepts more than the help lists.

## Pre-processing

1. Suffix aliases. For every entry of `Utils.LifestreamNativeCommands`
   (`auto home house private fc free company "free company" apartment apt
   shared inn hinn gc gcc hc hcc fcgc gcfc mb market island is sanctuary
   cosmic ardorum moon tp`), every enabled custom alias, and every enabled
   address-book alias: if the arguments end with a space followed by that
   word (case-insensitive), the space becomes a comma. `Odin mb` becomes
   `Odin,mb`.
2. Split on `,`, trimming and dropping empty parts. `primary` is the first
   part (or `""`), `additionalCommand` is the rest re-joined with `,`.
3. Branches are tested in the order below against the whole `arguments`
   string unless noted. The first match wins.

## Branches, in parser order

| Input | Condition | Effect |
| --- | --- | --- |
| `Name Surname@World` or `Name Surname@World DestWorld` | contains `@`; the word after `@` is a public world | `Utils.ChangeCharacter(name, world, destination?)` when `!Utils.IsBusy()`, else "Lifestream is busy". Unknown world logs an error. |
| `debug TaskAetheryteAethernetTeleport <rootId> <aethernetId\|firmament>` | exact prefix | `TaskAetheryteAethernetTeleport.Enqueue(a, b)`; exceptions are logged. |
| `debug WotsitManager clear` / `init` | exact | Wotsit integration maintenance. |
| `stop` | exact | `TaskManager.Abort()`, `followPath?.Stop()`, cancel a pending game shutdown. Reports how many tasks were discarded. |
| any non-empty text | `C.AllowCustomOverrides` (default off) and an address-book alias or custom alias equals it | Runs that alias ahead of every native command. |
| `1` .. `9` | one character | Instance change when `S.InstanceHandler.CanChangeInstance()`; warns when already there, errors when not possible. |
| `open` `select` `window` `w` `world` `travel` | equals, case-insensitive | Opens the world selection window. |
| `auto` | exact | `TaskPropertyShortcut.Enqueue(PropertyType.Auto)`: first available of the configured priority list (`C.PropertyPrio`, per-character overrides). |
| `home` `house` `private` | equals | Private estate. |
| `fc` `free` `company` `free company` | equals | FC estate. |
| `ws` `workshop` | equals | FC estate with `HouseEnterMode.Enter_workshop`. |
| `apartment` `apt` | equals | Apartment. |
| `shared` | equals | Shared estate (preferred one from `C.PreferredSharedEstates`). |
| `inn` / `hinn`, optionally `inn <n>` | equals or prefix with space | Inn shortcut. `<n>` is one-based into `TaskPropertyShortcut.InnData`; out of range prints the valid list. `inn` stays on the current world (`useSameWorld: true`); `hinn` returns to the home world first. The parse is `int.Parse`, so a non-numeric argument throws. |
| `gc` `gcc` `hc` `hcc` `fcgc` `gcfc`, optionally followed by a company | equals or prefix with space | `TaskGCShortcut.Enqueue(company?, isChest, returnHome, fcgc)`. `gcc`/`hcc` target the FC chest, a leading `h` returns home first, `fcgc`/`gcfc` is the FC-member variant. Company words: `TwinAdder`, `Twin Adder`, `Twin`, `Adder`, `TA`, `A`, `serpent`; `Maelstrom`, `Mael`, `S`, `M`, `storm`, `strom`; `ImmortalFlames`, `Immortal Flames`, `Immortal`, `Flames`, `IF`, `F`, `flame`; or any `GrandCompany` enum name. |
| `mb` `market` | equals | `TaskMBShortcut.Enqueue()` when the task manager is idle and the player is interactable; silent otherwise. |
| `island` `is` `sanctuary`, optionally `<npc>` | equals or prefix | `TaskISShortcut.Enqueue(npc?)`; the NPC is matched by substring against `DataStore.IslandNPCs` names and titles. |
| `cosmic` `ardorum` `moon` | equals | When not busy: return to the home world if away, wait for interactable, then run the shipped `StaticAlias.CosmicExploration`. |
| `occult` | equals | `StaticAlias.OccultCrescent` when not busy. |
| `Firmament` | equals, case-insensitive | `StaticAlias.Firmament` when not busy. |
| `tp <text>` | starts with `tp` | `destination` is `primary` after the first `tp`. Empty destination errors. When `P.TaskManager.IsBusy` is false and the player is interactable, `Utils.EnqueueTeleport(destination, additionalCommand)` (see below); on success the additional command runs after arrival. |
| `<world>, <city> W<ward> P<plot>` and the other address shapes | `Utils.TryParseAddressBookEntry` | Builds an address and calls `entry.GoTo()`. Accepted cities: `goblet`, `the goblet`, `lavender beds`, `lavender`, `lb`, `empy`, `empyreum`, `shiro`, `shirogane`, `mist`. Ward and plot accept `W1 P30`, `ward 1 plot 30`, or bare `1 30`; apartments use `A<n>` / `apartment <n>` with optional `s`/`sub`/`subdivision`. When the world is missing the parser retries with the current world prefixed. |
| anything else | fallthrough | See the next section. |

Release notes for `2.6.0.10` and `2.6.0.12` describe two further address
shapes, world, district and ward without a plot, and the same followed by an
aetheryte name (`/li world residential ward aetheryte`). Neither is in the
pinned parser, so they are not verified against source and are absent from
the table; see
[release notes](provenance.md#release-notes-2609-to-26013).

## The fallthrough: worlds, data centers, aethernet

With `/lifestream` and empty arguments the configuration opens. With `/li`
and empty arguments, `C.LiCommandBehavior` decides: `Return_to_Home_World`
(default) falls through with `primary == ""`, `Open_World_Change_Menu`,
`Open_Configuration`, or `Do_Nothing`.

Then, in order:

1. `ProcessCustomShortcuts(primary)`: an address alias runs `GoTo()`, a
   custom alias runs `Enqueue()` (which refuses when busy). The additional
   command is chained afterwards.
2. `gateway = WorldChangeAetheryte.Uldah` when `additionalCommand == "mb"`,
   else the configured default.
3. A same-DC world whose name starts with `primary` (home world when empty)
   gives `TPAndChangeWorld(w, false, gateway)`.
4. A cross-DC world the same way gives `TPAndChangeWorld(dcw, true, gateway)`.
5. A data-center name prefix picks a random visitable world on it.
6. Otherwise `TaskTryTpToAethernetDestination.Enqueue(primary, allowPartial: true, allowTpFallback: true)`.
   Standing at an aetheryte, the name is matched against the current
   group's shards (and the user's renames). `allowPartial` then searches
   every aetheryte group and teleports to the owning aetheryte before using
   its aethernet. `allowTpFallback` finally tries `Utils.EnqueueTeleport`.

Finally `ProcessAdditionalCommand`: when a second part exists, wait for
`IsScreenReady() && Player.Interactable`, then re-dispatch
`/li <additionalCommand>` through `Svc.Commands.ProcessCommand` on the next
tick. This is how `/li Odin, tp Limsa` and `/li Odin mb` chain.

## `Utils.EnqueueTeleport(destination, additionalCommand)`

Scans `Svc.AetheryteList` (unlocked aetherytes only) three times with a
case-insensitive `Contains`: aethernet name, then place name, then the
territory's place name. The first hit calls
`TeleportService.TeleportToAetheryte(id, wait: additionalCommand != null)`
and prints the destination in green. `wait` inserts two tasks: screen not
ready, then screen ready and interactable. Returns `false` when nothing
matched or the teleport action is unusable; the caller prints nothing in
that case except the TeleportService log line.

## Help text, condensed

```
/li → go to your home world
/li <world> → go to specified world
/li <datacenter> → go to a random world of specified data center
/li <aethernet> → go to specified aethernet destination
/li <world>, tp <location> → go to specified aetheryte destination of specified world
/li <world>, tp <aethernet> → go to specified aethernet destination of specified world
/li mb → go to market board
/li <world> mb → go to market board of specified world
/li auto → private estate, shared estate, FC estate or apartment by configured preference
/li shared, /li home|house|private, /li fc|free|company|free company, /li apt|apartment, /li ws|workshop
/li <district> <ward> <plot> → plot in current world
/li <world> <district> <ward> <plot> → plot of specified world
/li gc [<grandcompany>], /li gcc [<grandcompany>]; hc/hcc return to the home world first
/li cosmic|moon|ardorum, /li island, /li firmament
/li w|world|open|select → open world travel window
/lifestream → open plugin configuration
```

Not in the help but accepted by the parser: `inn`/`hinn [n]`, `occult`,
`stop`, single-digit instance numbers, `fcgc`/`gcfc`, `Name@World`, the
`debug` forms, and any custom or address alias.
