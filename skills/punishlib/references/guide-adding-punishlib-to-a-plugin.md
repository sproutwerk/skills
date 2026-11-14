## Adding PunishLib to a plugin

```xml
<!-- after: git submodule add https://github.com/PunishXIV/PunishLib PunishLib -->
<!-- MyPlugin/MyPlugin.csproj, the form used by AutoHook, Artisan, Saucy, PandorasBox -->
<ItemGroup>
  <PackageReference Include="ECommons" Version="3.2.0.16" />
  <ProjectReference Include="..\PunishLib\PunishLib\PunishLib.csproj" />
</ItemGroup>
```

The submodule sits at the repo root, so the project is `PunishLib/PunishLib/PunishLib.csproj`;
consumers also add it to their `.sln`. Facts about `PunishLib.csproj`:

- `Sdk="Microsoft.NET.Sdk"`, **not** `Dalamud.NET.Sdk`. Dalamud assemblies (`Dalamud`,
  `Dalamud.Common`, `Dalamud.Bindings.ImGui` / `ImPlot` / `ImGuizmo`, `Lumina`, `Lumina.Excel`,
  `FFXIVClientStructs`, `Newtonsoft.Json`) are hint-path references under `DalamudLibPath`: the
  XIVLauncher `Hooks/dev` folder per OS, or `$(DALAMUD_HOME)/` when that is set. A build machine
  without a dev folder must set `DALAMUD_HOME`.
- No explicit API level. `net10.0-windows7.0`, the `Dalamud.Bindings.ImGui` namespace and
  `IDalamudTextureWrap.Handle` place this pin at Dalamud API 15. Consumers on
  `Dalamud.NET.Sdk/15.0.0` (AutoHook, Saucy, PandorasBox) pin `96c0cef`. The previous commit,
  `f8b3e80`, already targets `net10.0-windows7.0`; older plugin branches need an older pin, so
  move the submodule together with the SDK bump.
- `Platforms x64`, `LangVersion preview`, `AllowUnsafeBlocks true`.

### Dependency on ECommons

PunishLib depends on ECommons as a **NuGet package**, not as a submodule:
`<PackageReference Include="ECommons" Version="3.2.0.11" />` (`PunishLib.csproj:30`). It has no
submodules of its own (`.gitmodules` is empty). What it uses from ECommons:

| ECommons symbol | Used in |
|---|---|
| `ECommons.DalamudServices.Svc.Log` | `API.cs`, `EzConfig.cs`, `GenericHelpers.cs` |
| `Svc.PluginInterface` (`Manifest`, `ConfigFile`, `GetPluginConfigDirectory()`) | `AboutTab.cs`, `EzConfig.cs` |
| `ECommons.ImGuiMethods.ThreadLoadImageHandler.TryGetTextureWrap(string url, out IDalamudTextureWrap textureWrap)` | `AboutTab.cs:57` |
| `ECommons.ImGuiMethods.Notify.Success(string s)` | `AboutTab.cs:86` |

The plugin supplies the ECommons that actually loads. Two arrangements exist in consumers:

- Plugin also uses the NuGet package (AutoHook 3.2.0.16, Artisan 3.2.1.17, Saucy 3.2.1.22, PandorasBox 3.2.1.5). NuGet
  resolves to the higher version. This is the simple case.
- Plugin uses ECommons as a submodule `ProjectReference` (AutoRetainer). AutoRetainer then adds
  `<PackageReference Include="ECommons" Version="*">` with
  `<ExcludeAssets>runtime;build;native;contentfiles;analyzers</ExcludeAssets>` and
  `<PrivateAssets>all</PrivateAssets>` next to the project reference, so that the package
  PunishLib pulls in does not ship a second `ECommons.dll`. Copy that block if you mix the two.

