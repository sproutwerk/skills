Dalamud topic: the dev-plugin workflow, Dalamud dev tools, debugger attachment and Dalamud channels. Verified against `goatcorp/Dalamud` `43b3411` (2026-11-10, `DalamudVersion 15.0.3.9`, API 15), `goatcorp/dalamud-docs` `3c7597d` (2026-11-13), `goatcorp/SamplePlugin` `b8477da` (2026-08-15) and `aers/FFXIVClientStructs` `80cb4a7` (2026-11-10). Reviewed 2026-11-13.

Paths such as `references/...` and `references/topic-*.md` are relative to the skill root (the directory that holds `SKILL.md`), not to this file. Part of the `dalamud` skill; items marked *(unverified)* are not confirmed from the pinned checkouts.

## Dev-plugin workflow and debugging

1. Build. Output folder holds `MyPlugin.dll` and `MyPlugin.json`. A dev plugin without a manifest is not loaded (API
   9).
2. In game `/xlsettings` → Experimental → **Dev Plugin Locations**: add the DLL path or its folder. Enable it in
   `/xlplugins` under the dev plugin category *(menu label unverified)*.
3. Rebuilding reloads the plugin automatically (per-plugin settings `AutomaticReloading`, `StartOnBoot`,
   `NotifyForErrors`). Reload calls `Dispose()` then constructs again, so leaks show up here first. A dev plugin only
   starts on boot when its API level equals the running one.
4. `/xldev` opens the dev bar: Plugins → Open Plugin Stats (`/xlstats`), `/xldata` data widgets (Addon Inspector, Atk
   Array Data, Inventory, Task Scheduler, …), toggles "Enable Addon Lifecycle" / "Enable Agent Lifecycle" for tracing.
5. Debugger: `/xldev` → Dalamud → toggle **Enable AntiDebug** (persisted), then attach to `ffxiv_dx11.exe` with
   Managed (.NET Core, .NET 5+) and Native code types; disable "Just My Code". `AccessViolationException` is not
   catchable since API 12 (.NET 9): it goes to the crash handler.
6. Dalamud channels: Release, Canary, Staging (`stg`), all from `master`; switch via `/xldev` → Dalamud → Branch
   Switcher. Test against `stg` before an API bump lands.
