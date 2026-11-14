## Repo layout

| Path | What |
|---|---|
| `Luna/` | Main library, `Sdk="Dalamud.NET.Sdk/15.0.0"`, namespace `Luna`, assembly `Luna`. ~38k lines. |
| `ImSharp/` | Git submodule (`git@github.com:Ottermandias/ImSharp.git`, branch `gc1.88`). **Empty in a non-recursive clone.** ~40k lines, 455 `.cs`. Project at `ImSharp/ImSharp/ImSharp.csproj` (plain `Microsoft.NET.Sdk`, `net10.0`). |
| `Luna.Targa/` | Git submodule (`git@github.com:Ottermandias/Luna.Targa.git`, branch `main`, MIT): a Targa reader and writer in namespace `Luna.Targa`. **Empty in a non-recursive clone.** Mirrored at `references/Luna.Targa/`: `TargaReader` (ref struct; `TryReadTarga(data, out reader)`, `Header`, `CreateRgba32Data()`), `TargaWriter.WriteRgba32` / `WriteColorMapped(stream, in Image, …)`, and `Structs/Image` (`Width`, `Height`, `Pixels`). |
| `Luna.Generators/` | Roslyn incremental generators + analyzers, `netstandard2.0`. |
| `Luna.Tests/` | xunit tests for generators, JSON recovery, UTF-8 transcoding. Good usage samples. |
| `schemas/` | Five JSON Schema (2020-12) files for the serialized condition tree (`Luna/Conditions`). |
| `.github/workflows/build.yml` | Windows build + test, checks out submodules recursively, downloads Dalamud `stg/latest.zip`. |

**ImSharp is compiled into `Luna.dll`, not referenced as a project.**
`Luna.csproj` has `<Compile Include="..\ImSharp\ImSharp\**\*.cs" …/>` (minus
`obj`, `temp`, and ImSharp's own `Resources\ResourceProvider.cs`), and embeds
ImSharp's `Resources\*.dxbc`. `Luna/GlobalUsings.cs` has `global using ImSharp;`
and aliases `ResourceProvider = Luna.ResourceProvider`. `Luna.slnx` lists
`ImSharp.csproj` only for IDE use. Consequences: the ImSharp types live in the
`Luna` assembly under namespace `ImSharp`; a plugin references Luna only; two
plugins each carry their own copy of the ImSharp code and share state through
`ImSharpContext` (below). `Luna.csproj` pulls in `..\Luna.Targa\Luna.Targa\**\*.cs`
the same way, and `Luna.Tests.csproj` compiles the Targa tests.

Also in `Luna.csproj`: `<Compile Remove="Filesystem\Selector\**" />` — the
whole `Luna/Filesystem/Selector/` folder is **not compiled**. Use
`Filesystem/Ui/` instead. `Luna.csproj` removes the `Dalamud.Bindings.ImGuizmo`
and `Dalamud.Bindings.ImPlot` references but keeps `Dalamud.Bindings.ImGui`
(Luna itself needs it for `Window`, textures and the error window).

