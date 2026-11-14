## Branding-specific versus generally useful

| Part | Verdict | Why |
|---|---|---|
| `AboutTab.Draw` | Puni.sh only | Hard-codes the Puni.sh Discord invite `https://discord.gg/Zzrcc8kmvy`, copies `https://love.puni.sh/ment.json` for `Repository`, and always ends the header with "by Puni.sh" or "published by Puni.sh". None of this is configurable. |
| `AboutPlugin`, `PunishOption.DefaultKoFi` | Puni.sh only | Feed the tab above; `DefaultKoFi` is one fixed Ko-fi URL. |
| API-key window, `API`, `SharedConfig`, `SharedConfigKey`, `SharedConfigStorage`, `PunishConfig`, `PunishConfigMethods` | Puni.sh only | Puni.sh account key, puni.sh endpoint, files under `%AppData%/PunishConfig`. |
| `IconButtons.IconTextButton` | Generally useful | No branding. The most used helper outside the About tab (Artisan calls it in 5 files). |
| `InfoBox`, `ImGuiGroup` group boxes | Generally useful | Plain drawing code. None of the six consumers read call them, so they are lightly exercised. |
| `PunishLib.Configuration.EzConfig`, `IEzConfig`, `IgnoreDefaultValueAttribute`, `GenericHelpers` | Superseded | ECommons has newer versions of each, with the same names. |

A plugin outside the Puni.sh family should not reference PunishLib for the widgets alone:
`PunishLibMain.Init` creates the Puni.sh shared files, and the About tab presents the plugin as
published by Puni.sh. Copy `IconButtons.cs`, `InfoBox.cs` or `ImGuiGroup.cs` instead; BSD-3-Clause
allows it when the copyright notice in `LICENCE.md` is kept. AutoRetainer, inside the family,
shows the other option: it calls `PunishLibMain.Init(pi, Name, PunishOption.DefaultKoFi)` but
draws its own `CustomAboutTab` built on ECommons' `ImGuiEx.LineCentered`.

