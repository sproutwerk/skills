## Identity, install location, loading

- `InternalData.FullName = "{Namespace}@{ClassName}"` (namespace falls
  back to `Default`). This is the key for config, blacklist, update.csv,
  and the disabled-scripts list. Renaming either orphans the config.
- Script folder: `<Dalamud pluginConfigs>/Splatoon/Scripts/` (i.e.
  `Svc.PluginInterface.GetPluginConfigDirectory()` + `Scripts`; on Windows
  `%APPDATA%\XIVLauncher\pluginConfigs\Splatoon\Scripts\`). All `*.cs`
  under it are loaded recursively at plugin start.
- Installing from a URL (`/splatoon` → Scripts tab → install from URL, or
  auto-update) writes the file to `Scripts/<Namespace>/<ClassName>.cs`.
  Loading from a file path keeps the file where it is. If a script with the
  same FullName is already loaded it is disabled, replaced in place, and
  `OnScriptUpdated(previousVersion)` is called.
- Compiled assemblies are cached in `<config>/ScriptCache/{md5}-{splatoonVersion}.bin`.
  Editing the file changes the md5, so a reload recompiles.
- Compile errors go to the Dalamud log (`/xllog`) as
  `Warning: CSxxxx: message`, the script lands in Splatoon's "failed
  scripts" window, and Splatoon tries to force-update it from `update.csv`.
- Per-script files sit **next to the .cs**: config `Foo.cs.json`, element
  overrides `Foo.cs.overrides.json`, alternate named configurations
  `Foo.cs.<8 random chars>.json` (managed from the UI).
- `Metadata.Version` (uint) is what the update system compares.
  `update.csv` rows are `FullName,Version,URL`; if a loaded script's version
  is lower it is downloaded and replaced (users can opt out per script;
  extra update lists can be added in Splatoon config). `blacklist.csv` rows
  are `FullName,Version`; a script whose version is `<=` the listed version
  gets `Blacklisted = true` and `Enable()` refuses to run it.
- Trusted URL prefixes (`ScriptingProcessor.TrustedURLs`):
  `github.com/NightmareXIV/`, `github.com/PunishXIV/`, their
  `raw.githubusercontent.com` forms and `nightmarexiv.org`. Untrusted URLs
  prompt the user.
- Scripts are not `Allowed` (cannot be enabled) until the blacklist download
  has completed once per session.
- The Scripts tab UI details (button labels) are not in the source copy
  used here; the behaviour above is from `ScriptingProcessor.cs`.


## Wine clipboard diagnosis

- **"Install from Clipboard" can silently install nothing.** Under Wine (XIV on Mac, XIVLauncher-RB) `ImGui.GetClipboardText()` may return an empty string; Splatoon then compiles an empty compilation unit successfully, finds no `SplatoonScript` type, and reports nothing. Tell-tale: a ScriptCache entry named `D41D8CD98F00B204E9800998ECF8427E-*.bin` (MD5 of ""). Install by copying the file to `<pluginConfigs>/Splatoon/Scripts/<Namespace>/<Class>.cs` and clicking "Reload and Update" instead. (Observed 2026-09-23 on XIV on Mac, Splatoon 3.9.2.25.)
