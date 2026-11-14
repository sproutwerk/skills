Dalamud topic: Dalamud.NET.Sdk project setup details, build requirements, plugin manifest sources and keys, load-control keys. Verified against `goatcorp/Dalamud` `43b3411` (2026-11-10, `DalamudVersion 15.0.3.9`, API 15), `goatcorp/dalamud-docs` `3c7597d` (2026-11-13), `goatcorp/SamplePlugin` `b8477da` (2026-08-15) and `aers/FFXIVClientStructs` `80cb4a7` (2026-11-10). Reviewed 2026-11-13.

Paths such as `references/...` and `references/topic-*.md` are relative to the skill root (the directory that holds `SKILL.md`), not to this file. Part of the `dalamud` skill; items marked *(unverified)* are not confirmed from the pinned checkouts.

## Project setup

The complete minimal csproj is in `references/guide-project-setup.md`.

- `Dalamud.NET.Sdk/<api>.x.y` sets target framework, platform, references to `Dalamud`,
  `Dalamud.Bindings.ImGui/ImPlot/ImGuizmo/ImAnim`, `Lumina`, `Lumina.Excel`, `FFXIVClientStructs`,
  `InteropGenerator.Runtime`, `Newtonsoft.Json`, and runs DalamudPackager. `<Version>` is the only property the SDK
  does not supply. The SDK major version is the API level. SDK internals live in `goatcorp/Dalamud.NET.Sdk` *(not in
  the checkouts; the reference list is from `docs/.../v12-SDK-migration.md` and the v12-v14 notes)*.
- Build needs the .NET 10 SDK (10.0.101+; 10.0.100 has restore problems) and VS 2026 or Rider 2025.3 (v14 notes).
  Dalamud dev files come from a local XIVLauncher install (`%AppData%\XIVLauncher\addon\Hooks\dev\`).
- `AssemblyName` (defaults to the csproj name) is the plugin's **InternalName**: config file/dir name, log tag, DLL
  name, D17 folder. It may never change after release. Display name is the manifest `Name`.
- `packages.lock.json` is expected next to the csproj (project-layout doc).
- Custom ClientStructs build: `<Use_Dalamud_FFXIVClientStructs>false</…>` plus `ProjectReference`s and manual resolver
  init (`references/docs/docs/plugin-development/reverse-engineering/using-custom-cs.md`). Discouraged for shipped
  plugins.
- Legacy setups (`Microsoft.NET.Sdk` + `DalamudPackager` package + `$(DalamudLibPath)` `<Reference>`s, or
  `Dalamud.Plugin.Bootstrap.targets`) are dead: the `targets/` folder was removed in API 14.

### Manifest

Three alternative sources; DalamudPackager (separate repo, not in the checkouts) emits `<InternalName>.json` into the
output and zip:

| Source | Keys |
|---|---|
| csproj properties (SDK 14+) | `Author`, `Name`, `InternalName`, `AssemblyVersion`, `MinimumDalamudVersion`, `Punchline`, `Description`, `ApplicableVersion`, `RepoUrl`, `Tags` / `CategoryTags` / `ImageUrls` (`;`-separated), `DalamudApiLevel`, `LoadRequiredState`, `LoadSync`, `CanUnloadAsync`, `LoadPriority`, `IconUrl`, `Changelog`, `AcceptsFeedback`, `FeedbackMessage` |
| `<InternalName>.json` | same keys, PascalCase |
| `<InternalName>.yaml` | same keys, snake_case (`repo_url`) |

Required: `Name`, `Author`, `Description`, `Punchline`. Auto-filled, do not write by hand: `AssemblyVersion`,
`InternalName`, `DalamudApiLevel`. Never set plumbing fields such as `Dip17Channel`.

Load-control keys (`PluginManifest.cs`):

| Key | Meaning |
|---|---|
| `LoadRequiredState` | `0` = during Framework.Tick when drawing facilities are available (default), `1` = during Framework.Tick, `2` = no requirement. Takes precedence over `LoadPriority`. |
| `LoadSync` | Load not concurrently with other plugins and the game. With `LoadRequiredState` 0 or 1 the `IDalamudPlugin` constructor then runs on the framework thread. |
| `LoadPriority` | Higher loads first, default 0. |
| `CanUnloadAsync` | `Dispose()` may be called off the framework thread. |
| `ApplicableVersion` | Game version gate, default `any`; load fails if the game is newer. |

Since API 15 the manifest inside the zip is used as-is (no longer overwritten from the repository), so it must be
present and accurate.
