OtterGui topic: adding OtterGui to a plugin (submodule, csproj, the nested `OtterGuiInternal`
submodule), how the library is coupled to the Dalamud SDK and API level, and what has to change on
an API bump. Verified against `Limiana/OtterGui` `a08fda6` (2026-05-30, branch `main`), upstream
`Ottermandias/OtterGui` `79771ee` (2026-04-29), `PunishXIV/Artisan` `c8ee6e0`,
`PunishXIV/AutoRetainer` `cc020df` and `goatcorp/Dalamud` `43b3411` (API 15).

Paths such as `references/...` and `references/topic-*.md` are relative to the skill root (the directory that
holds `SKILL.md`), not to this file. Nothing here was compiled.

## What OtterGui is, as a build artifact

- One project, `OtterGui.csproj`, producing assembly `OtterGui` (`RootNamespace` `OtterGui`).
  There is no NuGet package and no release; consumers reference the project from a git submodule.
- There is no README and no docs folder. The XML comments in `Text/` and the `//` comments
  elsewhere are all the documentation that exists.
- `OtterGui.csproj` of the fork, complete (`references/src/OtterGui.csproj`):

```xml
<Project Sdk="Dalamud.NET.Sdk/14.0.1">
    <PropertyGroup>
        <RootNamespace>OtterGui</RootNamespace>
        <AssemblyName>OtterGui</AssemblyName>
        <FileVersion>1.0.0.0</FileVersion>
        <AssemblyVersion>1.0.0.0</AssemblyVersion>
        <Product>OtterGui</Product>
        <Copyright>Copyright © 2025</Copyright>
        <WarningLevel>4</WarningLevel>
        <OutputPath>bin\$(Configuration)\</OutputPath>
        <ContinuousIntegrationBuild>true</ContinuousIntegrationBuild>
    </PropertyGroup>

    <PropertyGroup>
        <DefineConstants>PROFILING</DefineConstants>
        <Use_DalamudPackager>false</Use_DalamudPackager>
        <TargetFramework>net10.0-windows7.0</TargetFramework>
    </PropertyGroup>

    <ItemGroup>
      <Compile Remove="OtterGuiInternal\bin\**" />
      <Compile Remove="OtterGuiInternal\obj\**" />
      <EmbeddedResource Remove="OtterGuiInternal\bin\**" />
      <EmbeddedResource Remove="OtterGuiInternal\obj\**" />
      <None Remove="OtterGuiInternal\bin\**" />
      <None Remove="OtterGuiInternal\obj\**" />
    </ItemGroup>

    <ItemGroup>
        <PackageReference Include="Microsoft.Extensions.DependencyInjection" Version="9.0.2" />
        <PackageReference Include="JetBrains.Annotations" Version="2024.3.0" />
        <Reference Include="Dalamud.Bindings.ImGui" Private="false" />
    </ItemGroup>
</Project>
```

Things to read out of that file:

- It uses the Dalamud SDK, so `Dalamud`, `Dalamud.Bindings.ImGui`, `Newtonsoft.Json`, Lumina and
  FFXIVClientStructs come from the local Dalamud install, not from NuGet. The explicit
  `<Reference Include="Dalamud.Bindings.ImGui" Private="false" />` is the ImGui binding every
  file imports (`using Dalamud.Bindings.ImGui;`). There is no `ImGuiNET` anywhere.
- `Use_DalamudPackager` is `false`: OtterGui is a library, it is not packaged as a plugin.
- `DefineConstants` is `PROFILING`. `Logger.Excessive(...)` is `[Conditional("EXCESSIVE_LOGGING")]`;
  a `[Conditional]` call is dropped at the call site unless the calling project defines that
  symbol, and OtterGui's own project does not.
- **There is no `ProjectReference` to `OtterGuiInternal`.** The submodule directory sits inside
  the project directory, so the SDK's default `**/*.cs` glob compiles its sources into the
  `OtterGui` assembly; the `Compile Remove` lines only exclude its `bin` and `obj`.
  `OtterGuiInternal/OtterGuiInternal.csproj` exists (`Dalamud.NET.Sdk/13.0.0` at pin `49631aa`)
  and is listed in `OtterGui.sln`, but consumers never build it.
- `Microsoft.Extensions.DependencyInjection` 9.0.2 and `JetBrains.Annotations` 2024.3.0 are real
  NuGet dependencies and flow to the consuming plugin.
- `packages.lock.json` is committed and keyed on `net10.0-windows7.0`.

## Adding it to a plugin

```sh
# fork main, the commit AutoRetainer uses
git submodule add https://github.com/Limiana/OtterGui OtterGui
git -C OtterGui checkout a08fda67dfb731f007234c78df32a2b3ddccc450
git submodule update --init --recursive      # pulls OtterGui/OtterGuiInternal
```

`.gitmodules` inside OtterGui (`references/src/.gitmodules`):

```
[submodule "OtterGuiInternal"]
	path = OtterGuiInternal
	url = https://github.com/Ottermandias/OtterGuiInternal.git
	branch = main
```

Project reference, exactly as both consumers write it (the submodule sits next to the plugin
project directory):

```xml
<!-- Artisan/Artisan.csproj line 112, AutoRetainer/AutoRetainer.csproj line 44 -->
<ProjectReference Include="..\OtterGui\OtterGui.csproj" />
```

Both consumers use `Sdk="Microsoft.NET.Sdk"` with `$(DalamudLibPath)` references for their own
project and `net10.0-windows7.0` / `net10.0-windows` as target framework, while OtterGui itself
uses `Dalamud.NET.Sdk`. Both plugin manifests declare `"DalamudApiLevel": 15`.

CI: check out with submodules recursively (`actions/checkout` with `submodules: recursive`), or
the build fails on the missing `OtterGuiInternal` namespace.

## SDK and API coupling

| Tree | `Sdk=` in `OtterGui.csproj` | Explicit `TargetFramework` | OtterGuiInternal pin |
|---|---|---|---|
| fork `main` `a08fda6` | `Dalamud.NET.Sdk/14.0.1` | `net10.0-windows7.0` | `49631aa` |
| upstream `79771ee` | `Dalamud.NET.Sdk/15.0.0` | none (SDK default) | `fb9de85` |
| Artisan pin `32c4454` | `Dalamud.NET.Sdk/15.0.0` | `net10.0-windows7.0` | `49631aa` |
| fork branch `Artisan` `26b38ec` | `Dalamud.NET.Sdk/13.0.0` | `net10.0-windows7.0` | `49631aa` |
| fork branch `tw13` `2adc617` | `Microsoft.NET.Sdk` | `net9.0-windows7.0` | same as `main` |

The `dalamud` skill states that the SDK major version is the API level. The fork's csproj still
names SDK 14.0.1 while AutoRetainer, which builds it, targets API 15. How SDK 14.0.1 behaves
against an API 15 Dalamud install is *(unverified)*: nothing was built for this skill. If a build
complains about the SDK, the smallest change that matches upstream is setting the first line to
`Dalamud.NET.Sdk/15.0.0`.

## What breaks on a Dalamud API bump

Read from the commit history of both trees (`git log -- OtterGui.csproj`):

| Date | Tree | Commit | Change |
|---|---|---|---|
| 2025-08-02 | upstream | `27ef850` | "Update for new ImGui bindings" (move to `Dalamud.Bindings.ImGui`) |
| 2025-12-17 | fork | `4106af6` "7.4" | adds `<TargetFramework>net10.0-windows7.0</TargetFramework>` |
| 2025-12-19 | fork | `2fc7cee` "7.4" | `Dalamud.NET.Sdk/13.0.0` to `14.0.1`, lock file regenerated |
| 2026-04-28 | upstream | `398a318` | "Update SDK." (to 15.0.0) |
| 2026-04-29 | fork | `76271df` "7.5" | deletes the `ObjectKind.MountType` arm in `Custom/IndividualHelpers.cs` |
| 2026-04-29 | upstream | `79771ee` | "Fix MountType." (`ObjectKind.MountType` renamed to `ObjectKind.Mount`) |

So per bump, expect to touch:

1. The `Sdk=` version and possibly `TargetFramework` in `OtterGui.csproj`, then regenerate
   `packages.lock.json`.
2. Enum members that Dalamud clones from FFXIVClientStructs. `ObjectKind` is one
   (`Dalamud/EnumCloneMap.txt` maps it to `FFXIVClientStructs.FFXIV.Client.Game.Object.ObjectKind`).
3. Anything that reaches into the ImGui bindings' internals. 14 files use `ImGuiP.` or
   `ImGuiNative.` (23 call sites), and `MultiStateCheckbox`, `ToggleButton` and
   `OtterGuiInternal/ImGuiInternal.cs` use `ImGuiWindowPtr`, `ImRect`, `ItemAdd`,
   `ButtonBehavior`. A cimgui or binding regeneration can change these signatures.
4. Dalamud interfaces used in constructors: `IUiBuilder`, `IChatGui`, `INotificationManager`,
   `IFramework`, `IKeyState`, `IGameInteropProvider`, `IDalamudPluginInterface`, `IDataManager`,
   `ITextureProvider`.

Because you consume a commit of someone else's fork, an API bump means waiting for the fork to
move, or carrying the change yourself in your own fork and repointing the submodule URL (which is
what Artisan effectively does, see `references/topic-fork-delta.md`).

## Two copies in one game process

OtterGui is compiled into each plugin's own output. Every plugin carries its own copy of the
assembly and its own static state: the four static UTF-8 buffers of `ImUtf8` (128 KiB, 128 KiB,
4 MiB, 8 MiB of unmanaged memory, allocated on first use of each buffer type), the `ImRaii` colour
and style stacks, `EventWrapperBase.Logger`. Whether two plugins' copies can ever share that
state depends on Dalamud's plugin load contexts, which the `dalamud` skill covers; it was not
tested here *(unverified)*.

## Logger bootstrap

`new OtterGui.Log.Logger()` reads `Assembly.GetCallingAssembly().GetName().Name` for its prefix and
the `Dalamud.PluginName` Serilog context. Construct it from your plugin assembly, not from a helper
library, or log lines carry the wrong name.
