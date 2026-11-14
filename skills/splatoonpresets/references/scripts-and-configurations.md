# Scripts as preset carriers, and Configurations

Source: `Splatoon/SplatoonScripting/{Controller,SplatoonScript,ScriptingEngine,
InternalData,ExportedScriptConfiguration,OverrideData}.cs`,
`Gui/CGuiConfigurations.cs`, `Gui/Layouts/Header/Sections/LayoutConfigurations.cs`,
`SplatoonScripts/Duties/**` (e.g. `R4S Witch Hunt.cs`, `P12S Superchain.cs`).

## Minimal duty script skeleton

Self-contained and compilable as one file (the script compiler adds no
implicit usings). Territory 1232 = R4S, NameId 13057 = Wicked Thunder (real
ids); the layout coordinates are illustrative.

```csharp
using ECommons;                          // GenericHelpers: .Each(), .EqualsAny(), …
using Splatoon.SplatoonScripting;
using System.Collections.Generic;

namespace SplatoonScriptsOfficial.Duties.Dawntrail;   // FullName = "Namespace@ClassName"

public class R4S_Example : SplatoonScript
{
    public override HashSet<uint>? ValidTerritories { get; } = [1232];   // [] = any zone while logged in, null = even logged out
    public override Metadata Metadata { get; } = new(1, "you");          // (version, author[, description[, website]])

    public override void OnSetup()   // once after compile: register static drawing
    {
        Controller.RegisterElementFromCode("In",
            "{\"Name\":\"In\",\"type\":1,\"Enabled\":false,\"radius\":10.0,\"refActorNPCNameID\":13057,\"refActorComparisonType\":6,\"onlyTargetable\":true}");
        Controller.TryRegisterLayoutFromCode("Center",
            "~Lv2~{\"Name\":\"R4S Center\",\"ZoneLockH\":[1232],\"DCond\":1,\"ElementsL\":[{\"Name\":\"Center\",\"type\":0,\"refX\":100.0,\"refY\":100.0,\"radius\":1.0,\"color\":3355508480,\"mechanicType\":2}]}",
            out _);
    }

    public override void OnUpdate()  // every frame while enabled
    {
        Controller.GetRegisteredElements().Each(x => x.Value.Enabled = false);
        if(Controller.InCombat && Controller.CombatSeconds < 30f)            // stand-in for "mechanic active"
            Controller.GetElementByName("In")!.Enabled = true;
    }

    public override void OnReset() { }
}
```

Optional hooks — fragments to paste into the class above, each with the
using it needs (not a standalone file):

```csharp
using ECommons.Configuration;            // IEzConfig
using ECommons.Hooks.ActionEffectTypes;  // ActionEffectSet
using Dalamud.Bindings.ImGui;            // ImGui (Dalamud API 13+; not ImGuiNET)

public override void OnActionEffectEvent(ActionEffectSet set) { /* count resolves */ }

// Per-script config, exported by "Saved Configurations". Either derive from
// SplatoonScript<Config> and use its C property, or:
Config C => Controller.GetConfig<Config>();
public class Config : IEzConfig { public int MyTurn = 1; public bool ShowIdle = false; }
public override void OnSettingsDraw() { ImGui.InputInt("Your turn", ref C.MyTurn); }
```

## Controller API relevant to presets

| Member | Behaviour |
|---|---|
| `RegisterElement(key, Element, overwrite=false)` / `TryRegisterElement` | store under `key`; **bug**: duplicate check looks at `Layouts`, not `Elements` |
| `RegisterElementFromCode(key, json)` → `Element` / `RegisterElementFromCode(json)` (key = `Name`) / `RegisterElementsFromMultilineCode(text)` | decode bare element JSON via `ScriptingEngine.TryDecodeElement` |
| `TryRegisterLayoutFromCode(key, "~Lv2~…", out layout)` / `RegisterLayoutFromCode(key, code)` / `(code)` (key = layout `Name`) | decode via `TryDecodeLayout` (also accepts legacy strings) |
| `TryRegisterLayout(key, Layout)` / `TryRegisterLayout(Layout)` (key `unnamed-N`) | register a built object |
| `GetElementByName`, `TryGetElementByName`, `TryGetLayoutByName`, `GetRegisteredElements()`, `GetRegisteredLayouts()` | read-only dictionaries; mutate the objects |
| `TryUnregisterElement/Layout`, `Clear*` | |
| `Hide(elements, layouts)` | sets `Enabled=false` on all |
| `Schedule(Action, delayMs)`, `CancelSchedulers()`, `ScheduleReset(ms)`, `Reset()` | timers, cleared on reset |
| `GetConfig<T>()`, `SaveConfig()` | per-script config file (`IEzConfig`, `EzConfig`) |
| `InCombat`, `CombatSeconds`, `CombatMiliseconds`, `Phase`, `Scene`, `GetPartyMembers()`, `RolePosition`, `GetMapEffect(idx)` | state helpers |
| `DisplayAttentionWindowLine(text|action)`, `DisplayAttentionWindowRaw` | centred overlay text window |
| `OriginalElements`, `OriginalLayouts` | pristine copies of what `OnSetup` registered |

Rendering: every frame, for each enabled script, `ProcessLayout` runs on every
registered layout (full layout rules apply: zone lock, DCond, triggers, freeze,
subconfigurations) and `ProcessElement` on every registered element (only the
element's own conditions + `Enabled`). Scripts typically flip `Enabled`,
`refActorObjectID` (comparison type 2), `color`, `refX/refY`, `overlayText`
each frame. Upstream idiom: register elements with `"Enabled":false` and turn
them on in `OnUpdate`; use `refActorComparisonType:2` + `refActorObjectID:0`
for slots that a script fills with EntityIds.

Events the base class offers (all optional overrides): `OnSetup`, `OnEnable`,
`OnDisable`, `OnCombatStart`, `OnCombatEnd`, `OnPhaseChange(int)`,
`OnMapEffect(position, data1, data2)`, `OnObjectEffect(target, data1,
data2)`, `OnTetherCreate(source, target, data2, data3, data5)`,
`OnTetherRemoval(source, data2, data3, data5)`, `OnVFXSpawn(target, path)`,
`OnStartingCast(source, castId)` / `(sourceId, PacketActorCast*)`,
`OnMessage(string)` (same queue as triggers), `OnDirectorUpdate(...)`,
`OnObjectCreation(ptr)`, `OnActorControl(...)`, `OnActionEffect(...)`,
`OnActionEffectEvent(ActionEffectSet)`, `OnActionEffectApplied(ActionEffectSet,
int targetIndex, ActionEffectResult)`, `OnUpdate`, `OnReset`,
`OnScriptUpdated(prev)`, `OnSettingsDraw`, `OnGainBuffEffect`,
`OnRemoveBuffEffect`, `OnUpdateBuffEffect`.

## Per-script user edits: Overrides

Users can *Edit* any registered element/layout in the script's UI. The edited
copy is stored in `OverrideData { Elements: {key: Element}, Layouts: {key:
Layout} }` at `<script path>.overrides[.<confKey>].json` and re-applied by
`Controller.ApplyOverrides()` after `OnSetup` (keys must match registration
keys; unknown keys are rejected on import unless CTRL+SHIFT). "Export
customized settings" (obsolete) copies the raw `OverrideData` JSON; prefer
configuration export below. A script that sets colours itself in `OnUpdate`
will overwrite user edits — upstream warns about this.

## Script "Saved Configurations"

Each script can have named configurations. Identity: an 8-char key from
`[a-z0-9]`; files `<script path>.<key>.json` (config) and
`<script path>.overrides.<key>.json` (overrides); default configuration has
key `""` (files `<script>.json`, `<script>.overrides.json`). Names live in
`Config.ScriptConfigurationNames[FullName][key]`; the active key in
`Config.ActiveScriptConfigurations[FullName]`;
`Config.DefaultScriptConfigurationNames[FullName]` renames the default.
Switching (`ApplyConfiguration`) reloads the script.

Export (`GetExportedConfiguration` → `SerializeConfiguration`), one per line:

```json
{"TargetScriptName":"SplatoonScriptsOfficial.Duties.Dawntrail@M11S_Weapons",
 "ConfigurationName":"Toxic R1/H1",
 "Configuration":"<base64 of Brotli-compressed config json, or null>",
 "Overrides":"<base64 of Brotli-compressed OverrideData json, or null>"}
```

Import (`ApplyExportedConfiguration`): paste in the script's *Saved
Configurations* tab (or a multi-line mass export). Name collisions get
`(2)`, `(3)`…; a fresh key is allocated; bytes are Brotli-decompressed and
written; the config appears in the list — it is not auto-activated.

## Layout Configurations (Subconfigurations) and the Configurations tab

`Layout.Subconfigurations[]` are complete alternative element lists; the
editor's *Add New based on …* clones the current list. `SelectedSubconfigurationID`
selects (Guid.Empty = default `ElementsL`). The *Configurations* tab
(`CGuiConfigurations`) lists, per zone, every layout with subconfigurations and
every script with saved configurations, and offers "Switch all displayed
layouts/scripts to configuration" by **name** — matching `Subconfiguration.Name`
/ `DefaultConfigurationName` for layouts and configuration names for scripts.
Therefore a strategy preset should use one shared name (e.g. `"Toxic"`,
`"Kindred"`) across all its layouts' subconfigurations and its scripts'
exported configurations.

## Preset packaging decision

| Need | Ship |
|---|---|
| Static shapes, cast/status/tether/VFX/map-effect gating, timers relative to combat start | Layout(s) |
| Cross-event state, counting, player sorting/priority, text hints computed from party | Script (elements registered from export strings so users can restyle) |
| Script whose positions/roles depend on strat | Script + exported Configurations named per strat |
| Suppress auto-projection for handled casts | Elements with `refActorRequireCast` (automatic) or `BlacklistedProjectorActions` |
