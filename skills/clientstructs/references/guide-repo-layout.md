## Repo layout

| Path | What |
|---|---|
| `FFXIVClientStructs/` | The library. `FFXIV/{Client,Component,Common,Application,Shader}/…` mirror the game's C++ namespaces (`Client::Game::InventoryManager` → `FFXIVClientStructs.FFXIV.Client.Game.InventoryManager`). `STD/` = std::vector/map/set/list/deque/string views. `Havok/` = hk* animation/physics types. `SQEX/CDev` = SE runtime bits. `Interop/` = `Pointer<T>`, `RentedAtkValues`, `SpanExtensions`, `ThreadLocals`. `Attributes/` = `[Addon]`, `[Agent]`, `[InfoProxy]`, `CExporter*`. |
| `InteropGenerator/` | Roslyn incremental generator (netstandard2.0). Emits `Addresses`, `MemberFunctionPointers`, `Delegates`, `*VirtualTable`, string overloads, `Span<T>` accessors, bitfields, flattened inheritance. Diagnostics `CSIG0001..0406` in `Diagnostics/DiagnosticDescriptors.cs`. |
| `InteropGenerator.Runtime/` | Runtime the generated code targets: `Resolver`, `Address`, `CStringPointer`, `BitArray`, `BitOps`, `ThrowHelper`, and all generator attributes (`InteropGenerator.Runtime.Attributes`). |
| `FFXIVClientStructs.Generators/` | Second generator: `AgentGetterGenerator` (`[Agent(id)]` → `AgentX.Instance()` + `AgentModule.GetAgentX()`), `InfoProxyGetterGenerator`. |
| `ida/` | `data.yml` (2637 class entries: instances, vtables, vfuncs, funcs by address for the current exe), `ffxiv_idarename.py` (IDA + Ghidra renamer), `ffxiv_structs.yml` + `ffxiv_structimporter.py` (CExporter output → IDA structs), `ffxiv_sigmaker.py` (unmaintained), `ffxiv_exdgetters.py`, `classinformer.csv` (old RTTI dump = canonical class names), `data_schema.py`, `data-validator.js`, `old/` (previous data.yml snapshots). |
| `Ghidra/` | PyGhidra setup guide, `scripts/SigScanner.java` (Ctrl-Alt-S sig maker). |
| `CExporter/` | Exports structs to C headers + `ida/ffxiv_structs.yml`; also the CI validator (overlap/size/name checks → `ida/errors.txt`). |
| `CompatChecker/` | `apicompat` main vs PR dll → "breaking change" label. |
| `ExcelGenerator/` | Generates `Component/Exd/Sheets` from EXD schemas. |
| `FFXIVClientStructs.ResolverTester/` | Resolves every signature against an on-disk exe, prints `[FAIL]` list. |
| `FFXIVClientStructs.StdContainerTester/` | Unit tests for STD wrappers. |

There is no `Legacy` project. Dalamud ships its own build of the library.
A plugin on `Dalamud.NET.Sdk` gets the `FFXIVClientStructs` reference from
the SDK and must not bundle a second copy. To test a local CS build, set
`<Use_Dalamud_FFXIVClientStructs>false</…>`, add `ProjectReference`s to
`FFXIVClientStructs.csproj` and `InteropGenerator.Runtime.csproj` with
`Private` true or unset, and run the resolver yourself (see "Signature
resolution at runtime"). Upstream discourages shipping that to users
(dalamud-docs `reverse-engineering/using-custom-cs.md`, copy in the
`dalamud` skill).

