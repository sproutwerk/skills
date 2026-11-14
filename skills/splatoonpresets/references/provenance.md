# Source provenance

Curated serializables, behavior excerpts and upstream docs track
[Splatoon f396ef6](https://github.com/PunishXIV/Splatoon/tree/f396ef6c74da85ad5d4d7e1e5ed9401f666bb8e4),
the `main` head on 2026-11-13. `Splatoon/Splatoon.csproj` at that commit
declares plugin version 3.9.2.32 and `Splatoon/Splatoon.json` declares
Dalamud API level 15. Upstream's newest git tag is the much older `3.8.4.3`,
so the pin is a commit, not a tagged release.

The upstream AGPL-3.0 text is preserved in
[licenses/Splatoon-AGPL-3.0.txt](licenses/Splatoon-AGPL-3.0.txt).
This preserves the source notice; it does not assign a new license to the
whole skill or establish that the curated excerpts alone satisfy all
redistribution requirements.

## Bundled files

| Destination | Upstream path |
| --- | --- |
| `serializables/{CastAnimationKind,DisplayStyle,Element,EnumerationType,InternationalString,Layout,LayoutSubconfiguration,LineEnd,MapEffectData,MechanicType,PairingMode,TargetAlteration,Trigger}.cs` | `Splatoon/Serializables/` |
| `serializables/{CachedTetherInfo,Point2,PointerLineStyle}.cs` | `Splatoon/Structures/` |
| `serializables/{ExportedScriptConfiguration,OverrideData}.cs` | `Splatoon/SplatoonScripting/` |
| `upstream-docs/{CONTRIBUTING,INFORMATION,WALKTHROUGH}.md` | `Presets/` |
| `upstream-docs/PRESETS-README.md` | `Presets/README.md` |
| `upstream-docs/WebAPI.md` | `docs/help/WebAPI.md` |
| `licenses/Splatoon-AGPL-3.0.txt` | `LICENSE` |

Every bundled file is byte-identical to its upstream path at the pin.

Copied upstream docs retain repository-root links such as `/Presets/...` and
`/docs/images/...`. Those assets are not bundled; resolve them against the
[pinned upstream tree](https://github.com/PunishXIV/Splatoon/tree/f396ef6c74da85ad5d4d7e1e5ed9401f666bb8e4).
Their presence is not a claim that the source mirror is a complete offline manual.

## What was checked on 2026-11-13

The authored references were compared with the source at the pin by reading
it. Nothing was compiled, and no import, rendering or HTTP request was run
against a game client.

Between 3.9.2.25 (`b6b10c4`, 2026-09-22) and the pin, none of the bundled
files changed, and neither did the files the authored references cite for
layout evaluation, projection, import and the HTTP server. Changes that touch
this skill's subject:

- `SplatoonScript.OnActionEffectApplied(ActionEffectSet, int, ActionEffectResult)`
  is new; see [scripts and configurations](scripts-and-configurations.md).
- `Config.ExperimentalStatusHooks` (Experimental tab, off by default, applied
  on plugin load) raises status events from hooks instead of a per-frame
  comparison. The `[buff+]`, `[buff-]` and `[buff*]` message formats that
  triggers match are the same in both modes.
- `Config.OverlayFontScale` scales overlay text globally. It is a user
  setting, not a preset field.
- `InternalData.Overrides` is now publicly readable.

## Snapshot behavior

At the pin, `ProjectionState` serializes but has no reader.
`ResetOnTChange` resets triggers on both zone change and combat end;
`ResetOnCombatExit` has no effect. `FreezeResetTerr` clears freeze snapshots
on both events and `FreezeResetCombat` is not read. Recheck those fields in
the installed version before depending on them.

## Refreshing the snapshot

Choose explicit full commit IDs before fetching source. Record each exact
upstream path and its destination; do not search by basename across a checkout.
Fetch those commits, stage replacements in a temporary directory, and compare
hashes and license notices before replacing the snapshot. Missing or ambiguous
paths stop the refresh and leave the existing copy intact. Update source pins,
file mappings, and verification notes together with the accepted replacement.
