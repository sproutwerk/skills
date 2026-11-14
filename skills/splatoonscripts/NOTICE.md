# License and source notices

Original `SKILL.md` and the authored Markdown files directly under `references/` are Copyright (c) 2026 sproutwerk, licensed under MIT. The full text is in `LICENSE` beside this file.

Upstream files are copied unchanged and retain their own copyright and license. Each was written from the upstream blob at the pinned commit on 2026-11-13.

| Files | Source | License | License text |
| --- | --- | --- | --- |
| `references/host/` | [PunishXIV/Splatoon](https://github.com/PunishXIV/Splatoon) at `f396ef6`, from `Splatoon/SplatoonScripting/` and `SplatoonScripts/`. `update.csv.head` is the first 40 lines of `SplatoonScripts/update.csv`. | AGPL-3.0-only | `references/licenses/Splatoon-AGPL-3.0.txt` |
| `references/scripts/*.cs` | PunishXIV/Splatoon at `f396ef6`, from `SplatoonScripts/Generic/`. The two `Duty_*.cs` files come from `SplatoonScripts/Duties/` and are stored under different names. This is a selection: `generic-scripts-list.txt` names every upstream Generic script, bundled or not. | AGPL-3.0-only | `references/licenses/Splatoon-AGPL-3.0.txt` |
| `references/scripts/generic-scripts-list.txt` | File listing of `SplatoonScripts/Generic/` in the same tree. | AGPL-3.0-only | `references/licenses/Splatoon-AGPL-3.0.txt` |
| `references/ecommons/*.cs` except `Chat.cs`, `references/ecommons/EzIPC.md`, `references/ecommons/AddonMaster/*.cs` | [NightmareXIV/ECommons](https://github.com/NightmareXIV/ECommons) at `bc38dc7`. | MIT | `references/licenses/ECommons-MIT.txt` |
| `references/ecommons/*-public-api.txt`, `references/ecommons/Svc-services.txt`, `references/ecommons/AddonMaster/all-addon-masters.txt` | Listings generated from ECommons at the same commit. | MIT | `references/licenses/ECommons-MIT.txt` |

Splatoon ships the GNU Affero General Public License version 3 text and states no "or later" option, so it is recorded here as AGPL-3.0-only. Both license texts match the upstream files at the pinned commits. `references/provenance.md` records the full pins, the upstream paths and how the listings are generated.

## Not covered by these licenses

- `references/lifestream/` (`IPCProvider.cs`, `CustomAlias.cs`, `CustomAliasCommand.cs`, `CustomAliasKind.cs`): copied unchanged from [NightmareXIV/Lifestream](https://github.com/NightmareXIV/Lifestream) at `b363201`. That repository had no license file, and GitHub reported no license for it, when checked on 2026-11-13.
- `references/ecommons/Chat.cs`: copied unchanged from ECommons at `bc38dc7`, which is MIT. The file itself credits [Infiziert90/ChatTwo](https://github.com/Infiziert90/ChatTwo) at `5b58513` for part of its code, and that project is under EUPL-1.2. Which terms govern the credited part has not been established.

The redistribution terms for these files are not settled.

Splatoon, ECommons and Lifestream are the names of their upstream projects. This independently maintained reference skill is not an official release or endorsement of any of them.
