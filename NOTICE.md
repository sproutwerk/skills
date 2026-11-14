# License and source notices

These skills are independent references, with no affiliation to Square Enix
or the upstream plugin projects. FFXIV and FINAL FANTASY are associated with
Square Enix. The original guidance, examples, and tooling in
every skill use the [MIT license](LICENSE), Copyright (c) 2026 sproutwerk.

The repository is not uniformly MIT. Each skill bundles upstream files
unchanged, and those keep their own terms:

| Skill | Bundled upstream | License |
| --- | --- | --- |
| dalamud | goatcorp/Dalamud, goatcorp/SamplePlugin | AGPL-3.0-only |
| | goatcorp/dalamud-docs | MIT |
| clientstructs | aers/FFXIVClientStructs; one NightmareXIV/ECommons file | MIT |
| ecommons | NightmareXIV/ECommons | MIT |
| | Two `ActionEffectTypes` enum files | GPL-3.0-only |
| | `Interop/OpenFileName.cs`, `ExtendedBaseParam.cs`, `Automation/Chat.cs` | No recorded license grant |
| punishlib | PunishXIV/PunishLib | BSD-3-Clause |
| dalamudimgui | goatcorp/Dalamud | AGPL-3.0-only |
| | Dalamud ImGui binding projects, goatcorp/dalamud-docs | MIT |
| ottergui | OtterGui (Limiana, Ottermandias, and Taurenkey trees) | Apache-2.0 |
| | Ottermandias/OtterGuiInternal | BSL-1.0 |
| luna | Ottermandias/Luna | AGPL-3.0-only |
| | Ottermandias/ImSharp, Ottermandias/Luna.Targa | MIT |
| nightmareui | NightmareXIV/NightmareUI and NightmareUI.OtterGuiWrapper | MIT; the wrapper's selector files derive from Apache-2.0 OtterGui |
| xivapi | xivdev/EXDSchema files; the saved XIVAPI OpenAPI description | No recorded license grant |
| allagantools | Critical-Impact/InventoryTools and CriticalCommonLib | GPL-3.0 |
| splatoonpresets | PunishXIV/Splatoon | AGPL-3.0 |
| splatoonscripts | PunishXIV/Splatoon | AGPL-3.0-only |
| | NightmareXIV/ECommons files | MIT; part of `Chat.cs` is credited to ChatTwo (EUPL-1.2) |
| | NightmareXIV/Lifestream files | No recorded license grant |
| somethingneeddoing | Jaksuhn/SomethingNeedDoing source and wiki pages | No recorded license grant |
| autoretainer | PunishXIV/AutoRetainer | BSD-3-Clause |
| | PunishXIV/AutoRetainerAPI | MIT |
| questionable | PunishXIV/Questionable | AGPL-3.0 |
| lifestream | NightmareXIV/Lifestream files | No recorded license grant |
| | One NightmareXIV/ECommons file | MIT |
| vnavmesh | awgil/ffxiv_navmesh files | No recorded license grant |
| | xanunderscore/DotRecast license and project file | zlib |
| bossmod | awgil/ffxiv_bossmod | BSD-3-Clause |
| wrathcombo | PunishXIV/WrathCombo | BSD-3-Clause |
| | PunishXIV/WrathCombo.API | AGPL-3.0 |
| | MKhayle/XIVComboPlugin context files | GPL-3.0 |
| doman-mahjong | n/a | n/a |

"No recorded license grant" means the upstream repository publishes no
license for those files. They are bundled unchanged with attribution, and the
MIT license above does not cover them. Each skill's `NOTICE.md` lists its
bundled files, their sources, and the license texts; [sources.json](sources.json)
records the pinned revisions.

## Library licenses

What each library declares, as recorded in the notices bundled with its skill.
This is a summary of those notices, not legal advice.

| Library | Declared license | Exceptions recorded in the notices |
| --- | --- | --- |
| Dalamud | AGPL-3.0-only | Its ImGui binding projects are MIT. Dalamud is not shipped with a plugin. |
| ECommons | MIT | Two enum files carry a GPL-3.0-only header. `Interop/OpenFileName.cs` has a Microsoft all-rights-reserved notice, `ExtendedBaseParam.cs` credits an AGPL-3.0 plugin, and `Automation/Chat.cs` credits an EUPL-1.2 plugin; none of the three records a grant. All are compiled into `ECommons.dll`. |
| OtterGui | Apache-2.0 | The nested OtterGuiInternal is BSL-1.0. |
| Luna | AGPL-3.0-only | The ImSharp and Luna.Targa submodules are MIT. |
| NightmareUI | MIT | The OtterGui wrapper's selector files derive from Apache-2.0 OtterGui. |
| PunishLib | BSD-3-Clause | None. |

The terms are the same whether a library arrives by NuGet or as a submodule,
and a file copied out of a library keeps its header and license.
