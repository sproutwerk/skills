# Sproutwerk Skills - write better FFXIV plugins

Agent skills for FFXIV plugin development, plugin integration, game-data
queries, and scripting. Each skill pairs task guidance with a pinned snapshot
of the upstream source it describes, so an agent can look up exact signatures
instead of guessing.

## Skills

**Plugin development**

- **[dalamud](skills/dalamud/SKILL.md)**: build, port, debug, and review Dalamud plugins
- **[clientstructs](skills/clientstructs/SKILL.md)**: native structs, offsets, signatures, and unsafe interop
- **[ecommons](skills/ecommons/SKILL.md)**: ECommons helpers, TaskManager, AddonMaster, EzIPC, and config
- **[punishlib](skills/punishlib/SKILL.md)**: the Puni.sh About tab, API-key dialog, and widgets

**Plugin UI**

- **[dalamudimgui](skills/dalamudimgui/SKILL.md)**: stock Dalamud ImGui windows, scopes, tables, fonts, and textures
- **[ottergui](skills/ottergui/SKILL.md)**: OtterGui selectors, tables, combos, and fork differences
- **[luna](skills/luna/SKILL.md)**: Luna services and the ImSharp UI binding
- **[nightmareui](skills/nightmareui/SKILL.md)**: NightmareUI settings pages and its OtterGui wrapper

**Game data**

- **[xivapi](skills/xivapi/SKILL.md)**: sheet queries, game IDs, shops, and map coordinates through XIVAPI v2
- **[allagantools](skills/allagantools/SKILL.md)**: inventory counts, ownership, and craft lists over Allagan Tools IPC

**Splatoon**

- **[splatoonpresets](skills/splatoonpresets/SKILL.md)**: layout JSON, import strings, triggers, and phases
- **[splatoonscripts](skills/splatoonscripts/SKILL.md)**: C# Splatoon scripts and their plugin integrations

**Automation**

- **[somethingneeddoing](skills/somethingneeddoing/SKILL.md)**: SND Lua scripts and native macros
- **[autoretainer](skills/autoretainer/SKILL.md)**: AutoRetainer IPC, venture plans, and MultiMode
- **[questionable](skills/questionable/SKILL.md)**: Questionable IPC and quest-path JSON
- **[lifestream](skills/lifestream/SKILL.md)**: Lifestream travel IPC and the `/li` command grammar
- **[vnavmesh](skills/vnavmesh/SKILL.md)**: navigation IPC, path queries, and arrival checks
- **[bossmod](skills/bossmod/SKILL.md)**: BossMod (VBM) IPC, encounter modules, and replays
- **[wrathcombo](skills/wrathcombo/SKILL.md)**: WrathCombo leases, combo overrides, and auto-rotation

**Doman Mahjong**

- **[doman-mahjong](skills/doman-mahjong/SKILL.md)**: Doman Mahjong rules, scoring, and solver design

What each one covers, and what to know before using it, is under
[Skill details](#skill-details).

## Install

Install with the [skills CLI](https://skills.sh/docs/cli). It runs through
`npx`, so the only requirement is [Node.js](https://nodejs.org) 22.20 or newer.

```sh
npx skills add sproutwerk/skills --list                              # see what is available
npx skills add sproutwerk/skills --skill dalamud --skill ecommons    # pick some
npx skills add sproutwerk/skills --skill '*'                         # install everything
```

The CLI asks which agents to install for and whether to install for the
project or for your user. Its [README](https://github.com/vercel-labs/skills)
lists the supported agents and every option.

> [!TIP]
> **Let your agent do it.** Paste this into your agent:
>
> *Install the dalamud and ecommons skills from sproutwerk/skills with
> `npx skills add`, for this project. Check that Node.js 22.20 or newer is
> available first.*

Each folder under `skills/` also works on its own: copy it to
`.agents/skills/<name>` in a plugin repository, keep its references and
notices with it, and record the revision you copied.

## Use

Name the library in your request, or invoke the skill with your agent's skill
command.

| Ask for | Skill |
| --- | --- |
| "Find the shop and currency for this item." | `xivapi` |
| "Fix the table layout in this plugin window." | `dalamudimgui` |
| "Check this layout's actor matching and triggers." | `splatoonpresets` |
| "Port this old SND script to the current Lua API." | `somethingneeddoing` |

> [!IMPORTANT]
> Every skill describes one upstream snapshot, reviewed on 2026-11-13. Check
> your plugin's own dependency pins and the running game build before relying
> on a signature, offset, or IPC channel. Examples were checked by reading
> source, not compiled or run in game.

## Skill details

Pinned revisions are in [sources.json](sources.json), and licenses are in
[NOTICE.md](NOTICE.md).

### Plugin development

| Skill | Use it for | Good to know |
| --- | --- | --- |
| **[dalamud](skills/dalamud/SKILL.md)** | SDK setup, plugin lifecycle, services, threading, hooks, IPC, SeString, Lumina sheets, addon events, publishing rules, and API migrations (9 to 16) | Dalamud 15.0.3.9, API 15. API 16 notes are provisional. Bundles the plugin API source, official docs, and sample plugin. |
| **[clientstructs](skills/clientstructs/SKILL.md)** | Explicit-layout structs, vtables, generator attributes, singleton pointers, `AtkUnitBase` and `AtkValue` callbacks, and `data.yml` | Game build 2026.11.03. Layouts change with every patch, so match the runtime's ClientStructs commit before trusting an offset. |
| **[ecommons](skills/ecommons/SKILL.md)** | `Svc`, `ImGuiEx`, TaskManager, AddonMaster and callbacks, EzIPC, EzHook, EzConfig, throttlers, and game and Excel helpers | `master` at 3.2.1.26. The NuGet package of the same version lacks a few members, and the skill marks them. |
| **[punishlib](skills/punishlib/SKILL.md)** | `PunishLibMain.Init/Dispose`, `AboutTab.Draw`, `IconTextButton`, `InfoBox`, and ECommons init-order problems | 2.0.0.1. The About tab hard-codes Puni.sh branding. |

### Plugin UI

| Skill | Use it for | Good to know |
| --- | --- | --- |
| **[dalamudimgui](skills/dalamudimgui/SKILL.md)** | Plugin windows, `ImRaii` scopes, tables, fonts, textures, SeString rendering, notifications, file dialogs, and porting `ImGuiNET` code | Dalamud 15.0.3.9, Dear ImGui 1.88. A few helpers need 15.0.3.8 or later. |
| **[ottergui](skills/ottergui/SKILL.md)** | Item selectors, filterable tables, the folder-tree selector, filtered combos, and submodule or fork mismatches | Pinned by commit; there are no releases. The Limiana fork, upstream, and Artisan's pin differ in API. |
| **[luna](skills/luna/SKILL.md)** | Luna services and the plugin loader, ImSharp strings, scopes, widgets, tables, and moving from OtterGui | Luna `main` on the day its service rework merged; older pins have different service signatures. Adopting Luna makes a plugin AGPL. |
| **[nightmareui](skills/nightmareui/SKILL.md)** | `NuiBuilder` settings pages, button tabs, `WorldSelector`, `Censor`, and the file-system selector wrapper | Upstream marks both libraries as internal-use. Add them only to a plugin that already uses ECommons. |

### Game data

| Skill | Use it for | Good to know |
| --- | --- | --- |
| **[xivapi](skills/xivapi/SKILL.md)** | Items, actions, quests, NPC shops and currencies, territories, aetherytes, and map coordinates, with no game client | Needs network access. The optional helper CLI needs Go 1.26 or newer. |
| **[allagantools](skills/allagantools/SKILL.md)** | Item counts, character and retainer ownership, cached inventories, change events, saved filters, and craft lists | Allagan Tools 15.1.1. Results are cached, so a zero does not prove absence. No item transfers. |

### Splatoon

| Skill | Use it for | Good to know |
| --- | --- | --- |
| **[splatoonpresets](skills/splatoonpresets/SKILL.md)** | Raid presets, duty overlays, arena and safespot markers, triggers, phases, and presets that fail to import | Splatoon 3.9.2.32. Bundled examples are untested format demonstrations. |
| **[splatoonscripts](skills/splatoonscripts/SKILL.md)** | Duty scripts, vendor runs, auto-teleports, reminders, and plugin integrations written as `SplatoonScript` classes | Scripts compile against the ECommons that the installed Splatoon ships. |

### Automation

| Skill | Use it for | Good to know |
| --- | --- | --- |
| **[somethingneeddoing](skills/somethingneeddoing/SKILL.md)** | Writing and reviewing SND Lua scripts and native macros, porting pre-rewrite scripts, and debugging frozen or silent scripts | SND v15.13. Lua scripts can reach `os`, `io`, and .NET, so read third-party scripts before running them. |
| **[autoretainer](skills/autoretainer/SKILL.md)** | AutoRetainerAPI and `PluginState` IPC, venture overrides, postprocessing, venture plans, MultiMode, and stalled automation | AutoRetainer 4.6.1.26. Upstream publishes non-commercial usage guidelines beside its license. |
| **[questionable](skills/questionable/SKILL.md)** | Quest start and stop, state and completion queries, the priority list, and quest-path JSON | Questionable 15.756.3.60. Forks and downloaded path bundles can differ. |
| **[lifestream](skills/lifestream/SKILL.md)** | Teleports, world and data-center visits, housing, inns, instances, login, and the `/li` grammar | Covers the public source; releases after 2.6.0.8 include unpublished changes, so probe IPC at runtime. |
| **[vnavmesh](skills/vnavmesh/SKILL.md)** | Mesh readiness, path queries, path following, flight, cancellation, and stuck navigation | vnavmesh 1.2.3.15. There is no IPC version handshake. |
| **[bossmod](skills/bossmod/SKILL.md)** | VBM IPC, encounter modules and components, state machines, AI hints, rotation presets, and replay debugging | BossMod 7.5.8.2, awgil's repository. Forks share the `BossMod.*` IPC prefix with different channels. |
| **[wrathcombo](skills/wrathcombo/SKILL.md)** | Revocable leases, PvE combo and option overrides, auto-rotation control, action requests, and positional hints | WrathCombo 1.0.4.27 with API 0.5.8. The lease API covers PvE only. |

### Doman Mahjong

| Skill | Use it for | Good to know |
| --- | --- | --- |
| **[doman-mahjong](skills/doman-mahjong/SKILL.md)** | Waits, shanten, yaku, fu, payments, furiten, calls, and designing a Riichi rules engine | Research from the official guide. Supplies no memory offsets, addon callbacks, or plugin. |

## Choose a framework

Six of the skills cover libraries a plugin can be built on. They overlap, and
each one you add is a pin to keep in step with the Dalamud API level and a
license to carry in every release. Start with what Dalamud provides and add a
library when it replaces code you would otherwise write and maintain yourself.
An agent should follow the plugin's existing references first; the table is for
new plugins and for requests to add a dependency.

| Library | Add it when | Leave it out when | License |
| --- | --- | --- | --- |
| Stock Dalamud ImGui ([dalamudimgui](skills/dalamudimgui/SKILL.md)) | Always. Windows, `ImRaii` scopes, tables, fonts, textures, notifications and file dialogs need nothing else. | Never; Dalamud supplies it and nothing is shipped with the plugin. | Dalamud is AGPL-3.0-only, its ImGui bindings MIT. |
| ECommons ([ecommons](skills/ecommons/SKILL.md)) | The plugin drives game windows or works with other plugins: `AddonMaster` and `Callback` for addons, `EzIPC` for call gates, `TaskManager` for queued steps, throttlers, `EzConfig`, `ImGuiEx`. | The plugin only draws a window and reads game state, or the code is a game-independent library with tests. | MIT. A few files state other terms; see [NOTICE.md](NOTICE.md#library-licenses). |
| OtterGui ([ottergui](skills/ottergui/SKILL.md)) | The plugin already uses it, or it needs the folder tree selector, filtered combo caches or item selectors and can follow a fork's pin. | A new plugin needs scopes and tables only (stock `ImRaii` has them), or would start on OtterGui when Luna is its successor. | Apache-2.0; the nested OtterGuiInternal is BSL-1.0. |
| Luna and ImSharp ([luna](skills/luna/SKILL.md)) | The plugin already uses them, as Penumbra and Glamourer do, or the owner asks for them. ImSharp replaces the stock bindings with its own UTF-8 API; Luna adds services and source generators. | Any other case. It changes how every window in the plugin is written. | Luna is AGPL-3.0-only, ImSharp MIT. |
| NightmareUI ([nightmareui](skills/nightmareui/SKILL.md)) | An ECommons plugin wants `NuiBuilder` settings pages, the world selector or name censoring, and accepts an API that upstream marks as internal and may break between pins. | The plugin does not use ECommons, or has a handful of settings. | MIT. The OtterGui wrapper also needs OtterGui and keeps Apache-2.0 code from it. |
| PunishLib ([punishlib](skills/punishlib/SKILL.md)) | The plugin belongs to the Puni.sh family and needs its About tab or shared API key. | Any plugin outside that family: the About tab has fixed Puni.sh branding. Copy the one widget you want instead. | BSD-3-Clause. |

A plugin that walks an NPC dialogue and calls two other plugins is a typical
case: stock ImGui for its window, ECommons for `AddonMaster` and `EzIPC`, and
nothing else.

### Dependencies between them

- NightmareUI and PunishLib need ECommons. PunishLib takes it from NuGet, so
  check which copy the plugin ends up with.
- A plugin must contain one copy of ECommons: the NuGet package or a pinned
  submodule, never both. Two copies give two sets of static state.
- OtterGui, Luna, ImSharp and NightmareUI are consumed as git submodules, some
  nested. Pin each one, record the commit, and move the pins together with the
  Dalamud API level.
- OtterGui exists as several forks with different APIs. Match the fork the
  plugin already pins before using an example.

## Requests

A request for a new skill should name the upstream project and the
integration work the skill would cover.

## Sources and licenses

These are independent references, with no affiliation to Square Enix or the
upstream plugin projects. Original guidance, examples, and tooling use the
[MIT license](LICENSE). Bundled upstream files keep their own terms, and some
have no recorded license grant: [NOTICE.md](NOTICE.md) lists every skill's
bundled sources and their licenses.
