# Source pins and navigation

Reviewed on 2026-11-13 against
[PunishXIV/Questionable](https://github.com/PunishXIV/Questionable/tree/0a1510bd7be42f3e5063a8bccf398312d2d2f7e9),
branch `new-main`, revision `0a1510bd7be42f3e5063a8bccf398312d2d2f7e9`.
The commit timestamp is `2026-11-06T15:59:41+10:00`. The revision is two
commits after tag `v15.756.3.60`; those commits change quest paths, the
`SinglePlayerDutyOptions` schema, and translations, and leave the version
property alone.

[Directory.Build.targets](source/Questionable/Directory.Build.targets) sets
version `15.756.3.60` for projects other than GatheringPathRenderer.
[Questionable.csproj](source/Questionable/Questionable/Questionable.csproj)
uses Dalamud.NET.Sdk `15.0.0`, ECommons `3.2.1.22`, and PunishLib `2.0.0.1`.
Read that project file for the rest of the dependency graph. The
[model project](source/Questionable/Questionable.Model/Questionable.Model.csproj)
targets `netstandard2.0` with C# 14 and System.Text.Json `10.0.12`.
Dependencies and generated build output are not included in this skill.

## Source map

Paths below are relative to `references/source/Questionable/`.

| Question | Source |
| --- | --- |
| Exported channel names, types, registration and return values | `Questionable/External/QuestionableIpc.cs` |
| Start modes, task execution, global stop, gathering setup | `Questionable/Controller/QuestController.cs` |
| Current, next, simulated and gathering progress | `Questionable/Controller/QuestProgressTracker.cs` |
| ID syntax and Quest row-ID conversion | `Questionable.Model/Questing/ElementId.cs` |
| Eligibility, acceptance, completion, repeatables | `Questionable/Functions/QuestFunctions.cs` |
| Seasonal event list and the active-event filter | `Questionable/Windows/QuestComponents/EventInfoComponent.cs` |
| Priority mutations and accept-only state | `Questionable/Controller/QuestPriorityManager.cs` |
| Priority encoding and legacy import prefix | `Questionable/Windows/PriorityWindow.cs` |
| Registry population, file naming, overrides, reload notification | `Questionable/Controller/QuestRegistry.cs` |
| Downloaded bundle manifest, compatibility and update flow | `Questionable/PathData/` and `Questionable.Model/PathDataFormat.cs` |
| Schema and C# serialization | `QuestPaths/quest-v1.json`, `Questionable.Model/` |
| Runtime schema validation and task skip predicates | `Questionable/Validation/`, `Questionable/Controller/Steps/Shared/SkipCondition.cs` |
| Companion navigation/travel/dialogue calls | `Questionable/External/NavmeshIpc.cs`, `LifestreamIpc.cs`, `TextAdvanceIpc.cs` in that directory |
| Command behavior | `Questionable/Controller/CommandHandler.cs` |
| Upstream contribution policy and attribution | `CONTRIBUTING.md`, `README.md`, `LICENSE` |

Search within the installed skill, for example:

```sh
rg -n 'RegisterFunc|GetIpcProvider' references/source/Questionable/Questionable/External/QuestionableIpc.cs
rg -n 'LoadQuestsFromDownloadedBundle|LoadFromDirectory' references/source/Questionable/Questionable/Controller/QuestRegistry.cs
```

The snapshot includes the model and schemas, selected provider/controller
files, and one quest path. It omits most task implementations, UI, assets,
quest data, and dependencies. It is not a standalone compilable plugin.
Follow missing calls in a full checkout at the pinned revision.

## Refreshing and verification limits

[source-manifest.json](source-manifest.json) lists every copied upstream file
with its original path and SHA-256. Files retain the bytes of a Linux
checkout that applies upstream's `.gitattributes`, so most C# files carry a
BOM and CRLF line endings; hash a checkout, not raw blobs, when comparing.
Notices and upstream examples are untouched. The separate
[license notice](../NOTICE.md) describes their terms.

When refreshing, compare the provider, controller, ID parser, priority
encoding, registry loader, model, and schemas before updating prose. Replace
the selected source files from one revision, regenerate their hashes, and
update every recorded pin, including the hosting repository's `sources.json`
if it has one. Keep any downloaded path bundle
version separate from the plugin revision. Preserve source attribution and
check upstream contribution requirements if a later task includes submitting
changes upstream.

The guidance was checked by reading the source at this revision. The C#
example was not compiled, the plugin was not built, and no live game actions
were performed, so nothing here certifies runtime behavior.
