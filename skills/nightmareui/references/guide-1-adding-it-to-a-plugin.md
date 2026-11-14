## 1. Adding it to a plugin

The libraries have no packages; a consumer plugin carries them as git
submodules, one directory per repository:

| Directory | Repository | Needed for |
| --- | --- | --- |
| `ECommons` | `NightmareXIV/ECommons` | always |
| `NightmareUI` | `NightmareXIV/NightmareUI` | always |
| `NightmareUI.OtterGuiWrapper` | `NightmareXIV/NightmareUI.OtterGuiWrapper` | tree selector / GenericFileSystem |
| `OtterGui` | `Limiana/OtterGui` | the wrapper |

Adding a submodule builds that repository's code into the plugin. It is the
plugin owner's decision: propose the directories and the commits to pin, and
add them only when the user asks. A plugin that already declares them needs
nothing added.

All submodules go in the **solution root as siblings**; the paths are fixed by relative references
inside the library csprojs (see below). The plugin csproj then references the projects, as
AutoRetainer does (`AutoRetainer/AutoRetainer.csproj`, plugin project one level below the root):

```xml
<ProjectReference Include="..\ECommons\ECommons\ECommons.csproj" />
<ProjectReference Include="..\NightmareUI\NightmareUI\NightmareUI.csproj" />
<!-- wrapper users only -->
<ProjectReference
  Include="..\NightmareUI.OtterGuiWrapper\NightmareUI.OtterGuiWrapper\NightmareUI.OtterGuiWrapper.csproj" />
<ProjectReference Include="..\OtterGui\OtterGui.csproj" />
```

Splatoon has only the first two lines. Both consumers also add the projects to their `.sln`.

**Build facts** (`NightmareUI/NightmareUI.csproj`): `Microsoft.NET.Sdk` (not `Dalamud.NET.Sdk`),
`TargetFramework net10.0-windows7.0`, `Platforms x64`, `LangVersion preview`. Dalamud assemblies are referenced by `HintPath` from `$(DalamudLibPath)`,
which is `$(appdata)\xivlauncher\Addon\Hooks\dev\` and, on Linux, `$(DALAMUD_HOME)/`. References
include `Dalamud.Bindings.ImGui`, so the code targets the API 13+ bindings; .NET 10 matches
Dalamud API 15. The csproj declares no API level of its own.

**Dependency on ECommons, with evidence.** The csproj carries both

```xml
<PackageReference Include="ECommons" Version="3.2.0.10" />
<ItemGroup Condition="Exists('..\..\ECommons\ECommons\ECommons.csproj')">
    <ProjectReference Include="..\..\ECommons\ECommons\ECommons.csproj" />
</ItemGroup>
```

and 7 of the 14 source files import ECommons namespaces, using `ImGuiEx`,
`ThreadLoadImageHandler`, `PluginLog`, `Svc`, `ExcelWorldHelper`, `Player`, `ResidentalAreas`,
`GenericHelpers` and the `e.Log()` extension. The NuGet reference is unconditional; the project
reference is added when an `ECommons` checkout is a sibling of the `NightmareUI` directory. How
MSBuild resolves the two together was not tested *(unverified)*; AutoRetainer's own csproj handles
the same pair by adding `ExcludeAssets`/`PrivateAssets` to its ECommons `PackageReference`.

The wrapper csproj references `..\..\NightmareUI\NightmareUI\NightmareUI.csproj` and
`..\..\OtterGui\OtterGui.csproj` **unconditionally**, plus the same ECommons pair (NuGet
`3.0.0.2`). Directory names `NightmareUI`, `OtterGui`, `ECommons` are therefore mandatory.

**Initialization.** NightmareUI has no `Init`/`Dispose` of its own. It needs ECommons to be
initialised before the parts that touch `Svc` or `ECommonsMain.Instance` are constructed or drawn:

| Member | Needs |
|---|---|
| `NuiBuilder`, `Nui`, `NuiTools.ButtonTabs` | only `ImGuiEx`/`PluginLog` at draw time |
| `WorldSelector.DrawInternal` | `Svc.Data`, `Player` |
| `NuiTools.RenderResidentialIcon` | `ThreadLoadImageHandler` |
| `new ConfigFileSystem(...)` | `ECommonsMain.Instance` (dereferenced in the constructor, outside its `try`), `Svc.KeyState` |
| `new GenericFileSystem<T>(...)` | `Svc.PluginInterface`, `Svc.KeyState`, `EzConfig.OnSave` |

So: `ECommonsMain.Init(pi, this, ...)` first, then load configuration, then
`NuiTools.SetState(...)` and `Censor.Config = ...` if used, then create windows. AutoRetainer calls
`ECommonsMain.Init` at `AutoRetainer.cs:95` and `PunishLibMain.Init` at `:104`; its `NeoWindow`
(which owns the `ConfigFileSystem`) is held by its service manager as `S.NeoWindow`.

