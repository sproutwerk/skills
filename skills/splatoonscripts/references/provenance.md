# Source provenance

Bundled sources were refreshed and the authored guidance re-read against them on 2026-11-13.

| Source | Pin | Version | Used for |
| --- | --- | --- | --- |
| [PunishXIV/Splatoon](https://github.com/PunishXIV/Splatoon/tree/f396ef6c74da85ad5d4d7e1e5ed9401f666bb8e4) | `f396ef6c74da85ad5d4d7e1e5ed9401f666bb8e4` (`main`, 2026-11-13) | `Splatoon.csproj` 3.9.2.32, Dalamud API level 15, `net10.0-windows7.0` | `host/`, `scripts/` |
| [NightmareXIV/ECommons](https://github.com/NightmareXIV/ECommons/tree/bc38dc71340fa3b266de58e3d6d876a73313b61e) | `bc38dc71340fa3b266de58e3d6d876a73313b61e` (`master`, 2026-11-08) | untagged | `ecommons/` |
| [NightmareXIV/Lifestream](https://github.com/NightmareXIV/Lifestream/tree/b36320170faafc8315f7585edb45cfbcd34e7046) | `b36320170faafc8315f7585edb45cfbcd34e7046` (`main`, 2026-10-16) | tag 2.6.0.13 | `lifestream/` |

Upstream license texts are in [licenses/Splatoon-AGPL-3.0.txt](licenses/Splatoon-AGPL-3.0.txt)
and [licenses/ECommons-MIT.txt](licenses/ECommons-MIT.txt). Lifestream has no
license file at the pin. `NOTICE.md` in the skill root lists the terms per file.

Splatoon's git tags stop at 3.8.4.3 and do not track releases; read the version
from `Splatoon/Splatoon.csproj`.

## Which ECommons a script compiles against

Splatoon compiles scripts against the ECommons build it ships, which is its
`ECommons` submodule and not ECommons `master`. At the Splatoon pin the
submodule is `26d30b9ec14c3d3c18e335a9e4c7a5421dd134db`. Every file bundled
under `ecommons/`, and every source file behind the generated listings, is
identical at `26d30b9` and at the `bc38dc7` pin above. The only difference
between the two commits under the bundled paths is
`AddonMasterImplementations/Gathering.cs`, which is not bundled. Check the
submodule again when the installed Splatoon is newer than the pin.

## Upstream paths

- `host/Controller.cs`, `Extensions.cs`, `Metadata.cs`, `SplatoonScript.cs`,
  `ActionEffectResult.cs`: `Splatoon/SplatoonScripting/`.
- `host/BannedSymbols.txt`, `blacklist.csv`, `SplatoonScriptsOfficial.csproj`:
  `SplatoonScripts/`. `host/update.csv.head` is the first 40 lines of
  `SplatoonScripts/update.csv`.
- `scripts/*.cs`: `SplatoonScripts/Generic/`.
  `Duty_Maps_Auto_Target_Numbers.cs` is
  `SplatoonScripts/Duties/Universal/Maps Auto Target Numbers.cs` and
  `Duty_P8S2_Limitless_Desolation.cs` is
  `SplatoonScripts/Duties/Endwalker/P8S2 Limitless Desolation.cs`.
- `ecommons/TaskManagerConfiguration.cs`, `NeoTasks.cs`:
  `ECommons/Automation/NeoTaskManager/` and its `Tasks/` folder.
  `Chat.cs`, `Callback.cs`: `ECommons/Automation/`. `EzThrottler.cs`:
  `ECommons/Throttlers/`. `ProperOnLogin.cs`: `ECommons/Events/`.
  `AddonHelpers.cs`: `ECommons/GenericHelpers/`. `Player.cs`:
  `ECommons/GameHelpers/`. `EzIPC.md`, `EzIPCAttribute.cs`:
  `ECommons/EzIpcManager/`.
- `ecommons/AddonMaster/*.cs`: `ECommons/UIHelpers/AddonMasterImplementations/`.
  `AddonMasterBase.cs` is `!AddonMasterBase.cs` upstream.
- `lifestream/IPCProvider.cs`: `Lifestream/IPC/`. The three `CustomAlias*.cs`
  files: `Lifestream/Data/`.

## Generated listings

The listings are filtered lines of upstream files at the ECommons pin, in the
file order given. They show declarations only; open the source for bodies,
attributes such as `[Obsolete]`, and members declared across several lines.

| Listing | Lines kept | Source files |
| --- | --- | --- |
| `ecommons/TaskManager-public-api.txt` | starting with `public` | `Automation/NeoTaskManager/TaskManager.cs`, `TaskManager.Enqueue.cs`, `TaskManager.Insert.cs`, `TaskManager.Delay.cs`, `TaskManager.Stack.cs` |
| `ecommons/GenericHelpers-public-api.txt` | starting with `public` | `GenericHelpers/GenericHelpers.cs` |
| `ecommons/DalamudReflector-public-api.txt` | starting with `public static` | `Reflection/DalamudReflector.cs`, then `Reflection/ReflectionHelper/ReflectionHelper.cs`, `Call.cs`, `CallStatic.cs`, `Utils.cs` |
| `ecommons/misc-public-api.txt` | containing `public static` | `ImGuiMethods/ThreadLoadImageHandler.cs`, `ExcelServices/ExcelItemHelper.cs`, `ImGuiMethods/Notify.cs`, `Logging/DuoLog.cs`, `Configuration/EzConfig.cs` |
| `ecommons/Svc-services.txt` | containing `public static I`, with indentation and the `[PluginService]` attribute removed | `DalamudServices/Svc.cs` |
| `ecommons/AddonMaster/all-addon-masters.txt` | file names without `.cs` | `UIHelpers/AddonMasterImplementations/` |
| `scripts/generic-scripts-list.txt` | file names | `SplatoonScripts/Generic/` in Splatoon |

## Changes since the previous pins

The previous pins were Splatoon `b6b10c4` (3.9.2.25), ECommons `f1656e8` and
Lifestream `ef759e9`.

- Splatoon adds the `OnActionEffectApplied` hook and the `ActionEffectResult`
  record, makes `InternalData.Overrides` publicly readable, and adds
  `Generic/ChangeTerritoryType.cs`. The host now hooks actor casts through
  `PacketDispatcher.Delegates.HandleActorCastPacket`. `QuestionableQuestQueue.cs`
  drops one unused `using`. `update.csv` is rewritten with new rows.
- ECommons replaces the single TaskManager stack with nested named stacks,
  marks `BeginStack`, `DiscardStack` and the parameterless `EnqueueStack` and
  `InsertStack` obsolete, and makes `TaskManagerTask.Name` settable. The UI
  click helper now calls `ReceiveEvent` through a raw function pointer, which
  fixes a delegate-cast exception on addon clicks. `EzThrottler.Clear()` is
  new. `AddonMaster.SelectYesno.TextLegacy` is obsolete in favor of `Text`.
- Lifestream changes only `CONTRIBUTING.md`. The travel IPC and alias sources
  are unchanged.

## Limits

Nothing here was compiled or run in the game for this refresh, and no travel,
purchase or duty action was performed. The XIVAPI examples were spot-checked
by request on the same day; see [xivapi.md](xivapi.md). Scripts still need the
installed dependency set and in-game validation of lifecycle, IPC and addon
behavior.
