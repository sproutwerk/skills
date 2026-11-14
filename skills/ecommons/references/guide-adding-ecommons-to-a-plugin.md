## Adding ECommons to a plugin

Two channels, same source. Details, version table and consumer evidence → `references/topic-setup.md`.

```xml
<!-- NuGet: no Dalamud path setup, works with Dalamud.NET.Sdk. Package has no dependencies. -->
<PackageReference Include="ECommons" Version="3.2.1.26" />

<!-- or git submodule (git submodule add https://github.com/NightmareXIV/ECommons ECommons) -->
<ProjectReference Include="..\ECommons\ECommons\ECommons.csproj" />
```

- Version lines follow the Dalamud API: 3.0.1.x is API 13 (net9), 3.1.0.x is API 14 (net10),
  **3.2.x is API 15**. The mapping is inferred from NuGet publish dates and target frameworks;
  upstream does not state it. Use a package or submodule commit from the line that matches the
  Dalamud you build against.
- A NuGet version is published on every `BaseVersion` bump. There are no tags or releases.
- The submodule project is a plain `Microsoft.NET.Sdk` project that finds Dalamud through its own
  `DalamudLibPath` (Windows: `%appdata%\xivlauncher\Addon\Hooks\dev\`; Linux: `$DALAMUD_HOME`,
  default `~/.xlcore/dalamud/Hooks/dev/`). It needs `LangVersion` preview (C# 14 extension blocks).
- Never reference both the package and the project so that both reach the output.
- An `ECommons` entry in `.gitmodules` does not prove the plugin builds from it (PandorasBox
  lists the submodule but references the package). Read the csproj.

