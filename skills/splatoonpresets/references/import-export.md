# Import / export formats

## Layout string (`~Lv2~`)

```
~Lv2~<Layout JSON, single line>
```

Producer: `Utils.Serialize(this Layout l)` =
`"~Lv2~" + JsonConvert.SerializeObject(l, Formatting.None, DefaultValueHandling.Ignore)`.
Consumer: `Utils.ImportLayouts(string)` / `ScriptingEngine.TryDecodeLayout`:

1. Split clipboard on `\n`, trim, drop empty lines. **Each line = one layout.**
2. Line starts with `~Lv2~` → `JsonConvert.DeserializeObject<Layout>(line[5..])`,
   `Name = Name.SanitizeName()` (`,`→`_`, `~`→`_`). No migration is run.
3. Otherwise legacy: `Name~<json or base64 json>` (`Layout.Migrate()` moves
   the old `Elements` dict into `ElementsL` and converts `JobLock` bitmask),
   or a Paisley Park / Waymark-preset JSON `{Name, MapID, A,B,C,D,One..Four:{X,Y,Z,ID,Active}}`
   → one layout `"Legacy preset: <Name>"` zone-locked to the current zone with
   0.7-radius labelled circles (A/1 green, B/2 yellow, C/3 cyan, D/4 magenta).
4. Duplicate `Name` in `P.Config.LayoutsL` → error unless CTRL is held (then
   the duplicate is added anyway; nothing is replaced).
5. Success adds to config, expands the group in the UI and toasts.

Where: `/splatoon` → *Layouts* tab → *Import from clipboard*; or the *Rapid
Import* checkbox (polls the clipboard, imports layouts, and downloads scripts
when the clipboard is a trusted URL).

Newtonsoft specifics: unknown properties are ignored; numeric enums; `Guid`
from string; `HashSet<ushort>` from int array; `bool?`/`float?` from
`null`/value; `InternationalString` from `{"En":"…"}`. `Element` has a
constructor `Element(int t)` — JSON `type` still binds to the field.

## Element string

Elements export as bare JSON (`Copy` button in the element editor, `Copy
Elements` in the layout header = one element per line, Shift = enabled only):

```
{"Name":"In","type":1,"Enabled":false,"radius":10.0,"refActorNPCNameID":13057,"refActorComparisonType":6,"onlyTargetable":true}
```

Consumers: *Paste* into a layout's element list, or
`Controller.RegisterElementFromCode` / `TryRegisterElementFromCode` in scripts,
or `Controller.RegisterElementsFromMultilineCode` (one per line). Element
strings carry no zone/DCond/trigger context.

## Trigger list

Layout header *Copy* → `JsonConvert.SerializeObject(List<Trigger>)`; *Paste
Replace* (CTRL) / *Paste Add*.

## Script configuration string

One JSON object per line (see `scripts-and-configurations.md`):
```
{"TargetScriptName":"SplatoonScriptsOfficial.Duties.Dawntrail@M11S_Weapons","ConfigurationName":"Toxic","Configuration":"<base64>","Overrides":"<base64>"}
```

## Script preset

A raw-file URL, e.g.
`https://github.com/PunishXIV/Splatoon/raw/refs/heads/main/SplatoonScripts/Duties/Dawntrail/M11S%20Weapons.cs`.
Trusted prefixes (`ScriptingProcessor.TrustedURLs`): `github.com/PunishXIV/`,
`github.com/NightmareXIV/`, their `raw.githubusercontent.com` forms,
`nightmarexiv.org`, plus any prefix the user lists in `Config.ExtraTrustedRepos`
(*Trusted Repos* tab, behind a warning; case-sensitive match).
Trusted scripts auto-update; others prompt.

## Web API / HTTP export

*Copy for Web API* gives an URL-encoded layout/element for
`http://127.0.0.1:47774/?elements=…` (ALT = raw JSON for POST body, CTRL =
urlencoded raw). Parameters `enable`, `disable` (`Layout~Element`), `elements`,
`namespace`, `destroyAt` (`COMBAT_EXIT`, `TERRITORY_CHANGE`, ms), `destroy`,
`raw`. See `upstream-docs/WebAPI.md`. Requires *Enable HTTP server* in
settings (port 47774).

HTTP layout payloads have a different prefix from clipboard imports:
`~{"Name":"...","ElementsL":[...]}` for a raw/body HTTP layout, versus
`~Lv2~{"Name":"...","ElementsL":[...]}` for clipboard import. The HTTP
handler removes exactly one initial `~` before parsing JSON; sending a
clipboard `~Lv2~` string directly fails. Use the plugin's Web API export
or deliberately convert the prefix, and keep it at the start of the body.
HTTP responses can report `Error:` with status 200 and earlier toggles may
already be queued; inspect response text and do not blindly retry mutations.


## Presets repository conventions (`upstream-docs/PRESETS-README.md`)

- Markdown per duty under `Presets/<Expansion>/<Category>/<Duty>.md`; each
  preset in a fenced block; multi-layout presets = one `~Lv2~` line each.
- Tags: `[International]`, `[EN]`/`[JP]`/`[DE]`/`[FR]`, `[Untested]`, `[Beta]`,
  `[Script]`.
- Groups named `EX1 - Worqor Lar Dor`, `P1S - Erichthonios`, `97 - Alexandria`;
  layouts named after the mechanic.
