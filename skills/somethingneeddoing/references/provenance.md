# Source pins and navigation

Reviewed on 2026-11-13. These pins identify the evidence inspected; they do
not establish parity with a user's installed build. Use the guidance to
locate a behavior, then recheck the target revision, or the plugin's own
Help window, before depending on a command, modifier or Lua member.

| Source | Revision | Role |
| --- | --- | --- |
| [Jaksuhn/SomethingNeedDoing](https://github.com/Jaksuhn/SomethingNeedDoing/tree/527ca83892506c74fea4d5e07b9b9436e50ef95d) | `527ca83892506c74fea4d5e07b9b9436e50ef95d` | Head of `main` on 2026-11-13 (commit dated 2026-10-31); tag `v15.13`; main source snapshot |
| [Jaksuhn/SomethingNeedDoing wiki](https://github.com/Jaksuhn/SomethingNeedDoing/wiki) | `129eccefa9f1f1979569c6835aeee9b4a6700089` | Head of the wiki repository (2026-01-18); all seven pages copied as secondary evidence |
| [PunishXIV/AutoRetainerAPI](https://github.com/PunishXIV/AutoRetainerAPI/tree/b08122fd908fbaf4fa888c042066d219ee9ddbfa) | `b08122fd908fbaf4fa888c042066d219ee9ddbfa` | The plugin's submodule gitlink; noted, no source copied |
| Jaksuhn/SomethingNeedDoing `7a44a440f5927bee304dae6b685a8eb9b8854566` | 2025-04-20 | Last commit inspected that still held the pre-rewrite Lua functions under `SomethingNeedDoing/Old/Macros/LuaFunctions/`; names compared, no source copied |
| Jaksuhn/SomethingNeedDoing `cfb0425` | 2025-04-23 | Commit that deleted that directory, before the first rewrite tag `v12.2` |
| [daemitus/SomethingNeedDoing](https://github.com/daemitus/SomethingNeedDoing/tree/3cd36cc5d893473269f06dae6c973f133b695414) | `3cd36cc5d893473269f06dae6c973f133b695414` | The original plugin's last commit (2024-01-26); native command syntax compared, no source copied |

## Which repository

Several repositories carry this plugin's name. The one documented here was
chosen on these observations, all made on 2026-11-13:

- `Jaksuhn/SomethingNeedDoing` is not marked as a fork by the GitHub API,
  has `main` as its default branch and was last pushed on 2026-10-31. Its
  history has 2257 commits and 128 tags and contains both the pre-rewrite
  code and the rewrite.
- Its project file names it as its own home (`RepoUrl`), names `croizat` as
  author, and its README gives the Puni.sh plugin repository
  `https://puni.sh/api/repository/croizat` as the install source.
- The `PunishXIV` organization listed 22 public repositories and none was
  SomethingNeedDoing. `PunishXIV/SomethingNeedDoing` and
  `Croizat/SomethingNeedDoing` both returned 404.
- `daemitus/SomethingNeedDoing`, the original, was last pushed in May 2024
  and predates the rewrite.

Whether the build served by the Puni.sh repository is produced from this
exact commit was not confirmed; the repository listing could not be read
during the review.

## Version mapping

`SomethingNeedDoing.csproj` says `<Version>0.0.0.0</Version>`, so the
project file does not identify a release. Versions come from Git tags:
`DalamudPackager.targets` sets `VersionComponents="2"` and the release
workflow stamps the tag into the build. The pin is the commit tagged
`v15.13`, and it was also the head of `main`, so there is no gap between the
newest tag and the public source at the review date. No release binary was
downloaded or inspected.

Version numbers quoted from the plugin's changelog (for example the
`<listindex.N>` to `<list.N>` rename in 12.64) come from
`Gui/ChangelogWindow.cs` at the pin.

## Dependency context

The project uses `Dalamud.NET.Sdk/15.0.0` (Dalamud API level 15) and its
lock file targets `net10.0-windows7.0`. NuGet packages at the pin: NLua
1.7.9, ECommons 3.2.1.18, YamlDotNet 18.1.0, AutoCtor 3.1.0, Injectio 6.2.0
and Microsoft.Extensions.Hosting 10.0.11. The repository has two
submodule-style directories: `AutoRetainerAPI` (gitlink above) and a vendored
`DalamudCodeEditor` tree with its own license file. Neither is bundled here.

Much of what a script touches is implemented outside this repository:
`AddonMaster` (the `/click` targets), `Callback.Fire`, `IsAddonReady`,
`Player`, `Chat.SendMessage`, `EzIPC` and `WindowsKeypress` are ECommons.
Statements about them in this skill are limited to how SND calls them. The
`ecommons` skill covers that library at a different ECommons revision.

No package restore or build was performed.

## Source map

Paths are relative to this reference directory. Each copied upstream file is
listed in [source-manifest.json](source-manifest.json) with its original
path, repository, full revision and SHA-256. Copies come directly from Git
objects, preserving bytes, line endings and notices. This is a selected
reference set of 146 plugin files and 7 wiki pages, not a buildable
checkout.

`S/` below stands for `source/SomethingNeedDoing/SomethingNeedDoing/`.

| Behavior to verify | Source |
| --- | --- |
| `/snd` subcommands and aliases | [CommandService.cs](source/SomethingNeedDoing/SomethingNeedDoing/Services/CommandService.cs) |
| Start checks, states, pause and stop, function triggers, SND's own IPC gate | [MacroScheduler.cs](source/SomethingNeedDoing/SomethingNeedDoing/Managers/MacroScheduler.cs), [MacroHierarchyManager.cs](source/SomethingNeedDoing/SomethingNeedDoing/Managers/MacroHierarchyManager.cs), [Enums.cs](source/SomethingNeedDoing/SomethingNeedDoing/Core/Enums.cs) |
| Native line parsing and command table | [MacroParser.cs](source/SomethingNeedDoing/SomethingNeedDoing/NativeMacro/MacroParser.cs), [Commands/](source/SomethingNeedDoing/SomethingNeedDoing/NativeMacro/Commands/) |
| Modifier patterns and meanings | [ModifierDefinition.cs](source/SomethingNeedDoing/SomethingNeedDoing/NativeMacro/ModifierDefinition.cs), [Modifiers.cs](source/SomethingNeedDoing/SomethingNeedDoing/NativeMacro/Modifiers/Modifiers.cs) |
| Native engine loop, error reporting, craft-loop wrapping | `S/NativeMacro/` (engine and `CraftLoopBuilder`) |
| Lua state setup, coroutine driving, yields, cleanup calls | [NLuaMacroEngine.cs](source/SomethingNeedDoing/SomethingNeedDoing/LuaMacro/NLuaMacroEngine.cs), [LuaCodeSnippets.cs](source/SomethingNeedDoing/SomethingNeedDoing/LuaMacro/LuaCodeSnippets.cs), [LuaExtensions.cs](source/SomethingNeedDoing/SomethingNeedDoing/Utils/LuaExtensions.cs) |
| Module registration, how properties and enums reach Lua | [LuaModuleManager.cs](source/SomethingNeedDoing/SomethingNeedDoing/LuaMacro/LuaModuleManager.cs), [LuaModuleBase.cs](source/SomethingNeedDoing/SomethingNeedDoing/LuaMacro/LuaModuleBase.cs), [Attributes/](source/SomethingNeedDoing/SomethingNeedDoing/Attributes/) |
| Module members | [Modules/](source/SomethingNeedDoing/SomethingNeedDoing/LuaMacro/Modules/) |
| Wrapper members | [Wrappers/](source/SomethingNeedDoing/SomethingNeedDoing/LuaMacro/Wrappers/) |
| `IPC.<Name>` tables and gate names | [External/](source/SomethingNeedDoing/SomethingNeedDoing/External/), [IPCModule.cs](source/SomethingNeedDoing/SomethingNeedDoing/LuaMacro/Modules/IPCModule.cs), `S/Core/Interfaces/` |
| Metadata block, keys, config parsing | [MetadataParser.cs](source/SomethingNeedDoing/SomethingNeedDoing/Core/MetadataParser.cs), [MacroMetadata.cs](source/SomethingNeedDoing/SomethingNeedDoing/Core/MacroMetadata.cs), [MacroConfigItem.cs](source/SomethingNeedDoing/SomethingNeedDoing/Core/MacroConfigItem.cs), [ConfigMacroExtensions.cs](source/SomethingNeedDoing/SomethingNeedDoing/Utils/ConfigMacroExtensions.cs) |
| Trigger and cleanup function discovery | [MacroFunctions.cs](source/SomethingNeedDoing/SomethingNeedDoing/Core/MacroFunctions.cs) |
| Trigger events, payloads, chat filters, AutoRetainer post-process | [TriggerEventManager.cs](source/SomethingNeedDoing/SomethingNeedDoing/Managers/TriggerEventManager.cs), [TriggerEventSources.cs](source/SomethingNeedDoing/SomethingNeedDoing/Managers/TriggerEventSources.cs) |
| Script dependencies and their cache | [Dependencies/](source/SomethingNeedDoing/SomethingNeedDoing/Core/Macros/Dependencies/) |
| Git-backed macros | [GitMacroManager.cs](source/SomethingNeedDoing/SomethingNeedDoing/Managers/GitMacroManager.cs), `S/Core/Github/`, [GitInfo.cs](source/SomethingNeedDoing/SomethingNeedDoing/Core/Macros/GitInfo.cs) |
| Settings, defaults, config migration, legacy import | [Config.cs](source/SomethingNeedDoing/SomethingNeedDoing/Config.cs) |
| Chat channels and log prefixes | [ChatExtensions.cs](source/SomethingNeedDoing/SomethingNeedDoing/Utils/ChatExtensions.cs), [FrameworkLogger.cs](source/SomethingNeedDoing/SomethingNeedDoing/Utils/FrameworkLogger.cs) |
| Repair, spiritbond and interact helpers | [Game.cs](source/SomethingNeedDoing/SomethingNeedDoing/Utils/Game.cs) |
| Lua stub generation | [StubGeneratorService.cs](source/SomethingNeedDoing/SomethingNeedDoing/Services/StubGeneratorService.cs) |
| Rewrite notice, in-plugin help text, changelog | [FirstTimeWarningModal.cs](source/SomethingNeedDoing/SomethingNeedDoing/Gui/Modals/FirstTimeWarningModal.cs), [HelpGeneralTab.cs](source/SomethingNeedDoing/SomethingNeedDoing/Gui/Tabs/HelpGeneralTab.cs), [ChangelogWindow.cs](source/SomethingNeedDoing/SomethingNeedDoing/Gui/ChangelogWindow.cs) |
| Project, packaging, submodules, upstream README | [SomethingNeedDoing.csproj](source/SomethingNeedDoing/SomethingNeedDoing/SomethingNeedDoing.csproj), [DalamudPackager.targets](source/SomethingNeedDoing/SomethingNeedDoing/DalamudPackager.targets), [.gitmodules](source/SomethingNeedDoing/.gitmodules), [README.md](source/SomethingNeedDoing/README.md) |
| Upstream's own prose on configs, triggers, dependencies, Git macros | [wiki pages](source/SomethingNeedDoing.wiki/) |

The rest of the plugin's UI (`Gui/` apart from the three files above), most
of `Services/` and the vendored editor are not bundled. An absent file in
this set does not prove a behavior is absent upstream.

## What was not verified

- Nothing was compiled, loaded into the game or run. Every statement is a
  reading of source text.
- Behavior that lives in ECommons, NLua, Dalamud or the game was not
  followed into those codebases. That includes what an `IPC` call does when
  the target plugin is not loaded, which `AddonMaster` entries and methods
  exist, and how NLua converts particular argument types.
- Lua snippets that reach past the modules (`Svc.Condition[...]`,
  `Svc.ClientState.TerritoryType`, `import("System.Numerics")`) follow
  NLua's documented conventions and are not exercised anywhere in the
  pinned tree.
- Several findings are flagged in the references as source readings that
  contradict upstream's help text or wiki: `/wait` with a range, `/equip`
  versus `/equipitem`, the `/click` help examples, config validation at
  start, metadata written back into native macros, and settings that no
  code reads. Each should be confirmed on a running build before it is
  reported upstream or relied on.
- The legacy-to-current table in [migration.md](migration.md) is this
  skill's mapping from the two trees. Upstream publishes none.
- The `IPC` wrappers' signatures were read from SND's `External/` files
  only. They were not compared with the current source of the wrapped
  plugins.

## Refresh

Check whether `main` has moved past the newest tag and whether a newer tag
exists. Diff `NativeMacro/`, `LuaMacro/`, `External/`, `Core/` and
`Managers/` against the previous pin, then follow each changed command,
modifier, module member or trigger into the references. Re-copy changed
files from Git objects, regenerate the manifest hashes, and recheck local
links and frontmatter. The wiki has its own repository and revision.

On 2026-11-13 every bundled file was hashed from its pinned Git object when
the manifest was written. See [NOTICE.md](../NOTICE.md) for the unresolved
source grant.
