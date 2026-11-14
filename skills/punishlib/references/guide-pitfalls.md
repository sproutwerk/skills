## Pitfalls

- **`AboutTab` before `Init` throws.** `AboutTab` has a static field initializer
  `_inputKey = PunishLibMain.SharedConfig.APIKey`. Touching `AboutTab` before
  `PunishLibMain.Init` raises a `TypeInitializationException` (inner `NullReferenceException`),
  and a failed type initializer stays failed until the plugin is reloaded.
- **`Svc` must be initialized.** `AboutTab.Draw`, `EzConfig.*` with `appendConfigDirectory = true`
  and `GenericHelpers.Safe` dereference ECommons' `Svc`. Without `ECommonsMain.Init` they fail with
  a `NullReferenceException`.
- **Name clashes with ECommons.** Identical type names exist in both libraries:
  `PunishLib.GenericHelpers` and `ECommons.GenericHelpers`; `PunishLib.Configuration.EzConfig`,
  `IEzConfig`, `IgnoreDefaultValueAttribute` and their `ECommons.Configuration` twins. A file
  with both `using ECommons;` and `using PunishLib;` compiles until it names `GenericHelpers`
  or calls `color.ToUint()` / `dict.TryGetFirst(...)` as extension methods, which then become
  ambiguous because both libraries declare the same signatures. Import `PunishLib` only in
  the bootstrap file and `PunishLib.ImGuiMethods` only where widgets are drawn; never import
  `PunishLib.Configuration` next to `ECommons.Configuration`.
- **`ImGuiWindow` clash.** `PunishLib.ImGuiMethods.ImGuiWindow` shares its simple name with
  `Dalamud.Bindings.ImGui.ImGuiWindow`; qualify it when both namespaces are imported.
- **`DefaultKoFi` overrides your sponsor link.** It is applied after `about` is stored.
- **Machine-wide side effects.** `Init` creates `%AppData%/PunishConfig/` and the API-key window
  writes `PunishSharedConfig.json` there. The path comes from
  `Environment.SpecialFolder.ApplicationData`, so it is per OS user (inside the Wine prefix on
  Linux and macOS), not per plugin and not under Dalamud's `pluginConfigs`.
- **DataShare entries are never relinquished.** `SharedConfig` calls `GetOrCreateData` and
  `Dispose()` is empty. The API key is one value shared by all loaded Puni.sh plugins.
- **`Test Key` blocks the frame.** It calls `API.API.ValidateKey().Result` inside the draw
  callback with a fresh `HttpClient`, so the game stalls for the duration of the HTTP request.
  Each distinct key is tested once per plugin load (`testedKeys`).
- **Icon needs `IconUrl`.** Without it the tab passes an empty string to
  `ThreadLoadImageHandler.TryGetTextureWrap`; what ECommons then does is *(unverified)*.
- **`InfoBox.DrawBox` caches by label forever.** The delegate passed on the first call is kept
  and later ones are ignored; capture fields, not per-frame locals, and keep labels unique.
- **`GroupBoxOptions` defaults call ImGui.** Its property initializers call `ImGui.GetColorU32`
  and `ImGui.GetStyle()`, so instances must be created while an ImGui context is current (inside
  a draw callback), not in a static field initializer.
- **Vestigial submodules exist.** YesAlready (`c531fb8`) still lists the `PunishLib` submodule
  but neither its csproj nor its code references it. A `.gitmodules` entry alone does not prove
  usage; check the csproj.
- **Submodule pins differ between plugins.** `96c0cef` in AutoHook, Artisan, Saucy,
  AutoRetainer and PandorasBox; `f8b3e80` in YesAlready's unused submodule. After cloning a consumer run
  `git submodule update --init` and read the library at the pinned commit, not at `master`.
