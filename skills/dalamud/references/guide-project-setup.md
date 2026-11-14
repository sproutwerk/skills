## Project setup

```xml
<!-- MyPlugin/MyPlugin.csproj, complete (from SamplePlugin @ b8477da) -->
<Project Sdk="Dalamud.NET.Sdk/15.0.0">
  <PropertyGroup>
    <Version>0.0.0.1</Version>
    <Author>your name here</Author>
    <Name>Sample Plugin</Name>
    <Punchline>A short one-liner that shows up in /xlplugins.</Punchline>
    <Description>Shown when the user expands the plugin entry.</Description>
    <RepoUrl>https://github.com/you/MyPlugin</RepoUrl>
    <Tags>sample;plugin</Tags>
  </PropertyGroup>
</Project>
```

`<Version>` is the only property the SDK does not supply. The SDK major version is the API level.
`AssemblyName` (defaults to the csproj name) is the plugin's **InternalName** and may never change
after release. Since API 15 the manifest inside the zip is used as-is, so it must be present and
accurate. SDK references, build requirements, manifest keys and load-control keys →
`references/topic-project-setup.md`

