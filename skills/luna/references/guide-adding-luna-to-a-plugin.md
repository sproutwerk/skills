## Adding Luna to a plugin

Both consumers use a git submodule at repo root plus project references.

```sh
git submodule add https://github.com/Ottermandias/Luna.git Luna
git submodule update --init --recursive     # pulls Luna/ImSharp and Luna/Luna.Targa too
```

Glamourer's `.gitmodules` uses the ssh URL, Penumbra's the https URL. Luna's own
`.gitmodules` points ImSharp and Luna.Targa at `git@github.com:Ottermandias/…` ssh URLs, so a
recursive update needs working GitHub ssh auth (or a `url.…insteadOf` rewrite).

```xml
<!-- Glamourer/Glamourer.csproj, verbatim lines -->
<Project Sdk="Dalamud.NET.Sdk/15.0.0">
  <PropertyGroup>
    <Use_DalamudPackager>false</Use_DalamudPackager>
    <Use_Dalamud_ImGui>false</Use_Dalamud_ImGui>
  </PropertyGroup>
  <ItemGroup>
    <ProjectReference Include="..\Luna\Luna\Luna.csproj" />
    <ProjectReference Include="..\Luna\Luna.Generators\Luna.Generators.csproj"
                      OutputItemType="Analyzer" ReferenceOutputAssembly="false" />
  </ItemGroup>
</Project>
```

- No `ProjectReference` to ImSharp — see above.
- `Use_Dalamud_ImGui=false` is an SDK switch: in `Dalamud.NET.Sdk` 15.0.0
  `Sdk.props` it gates the `Dalamud.Bindings.ImGui`, `.ImPlot` and `.ImGuizmo`
  references (default `true`). Both consumers turn it off so `ImGui.*` is not
  reachable from plugin code. Optional; leave it on while porting.
- `Use_DalamudPackager=false` is what the consumers do; unrelated to Luna.
- The generator reference is per project. Its attributes are emitted `internal`
  into each consuming assembly, so every project that uses them needs the
  analyzer reference (Penumbra.GameData has its own).
- Version coupling: Luna, the plugin and Dalamud must agree on the SDK major
  (15 here) and therefore on the target framework and Dalamud API level. Luna
  has no releases; pin the submodule commit and bump it together with the SDK.
  `Luna.csproj` sets `LangVersion preview`, and ImSharp uses C# 14 `extension`
  blocks and `field`-backed properties, so an older compiler will not build it.
- `Penumbra.csproj` references only the generator project; it gets `Luna.csproj`
  transitively through `Penumbra.GameData`, whose project file references both
  (read through the GitHub API at `27398d9`; that submodule is empty in the
  shallow clone).

License fact: `Luna/LICENSE` is the GNU Affero General Public License v3 text;
ImSharp's `LICENSE` is MIT, copyright 2025 Ottermandias. Penumbra and Glamourer
both link Luna. What AGPL-3.0 requires of a plugin that links Luna is a legal
question this skill does not answer; read `references/Luna/LICENSE`.

