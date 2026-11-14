# FFXIVClientStructs

C# explicit-layout structs mapping 1:1 onto the native classes of
`ffxiv_dx11.exe`, plus a Roslyn source generator that turns attributed stub
methods into function-pointer calls. MIT. Repo: `github.com/aers/FFXIVClientStructs`,
branch `main`, no tags/releases — Dalamud pins a commit. No GitHub wiki, no
`docs/`, no CONTRIBUTING; the contributor docs are `README.md`, `ida/README.md`
and `Ghidra/*.md` (copies in `references/upstream-docs/`).

The bundled source tracks `80cb4a7b7e11a6702e2eb8839afd301a918c638e`
(2026-11-10), with `ida/data.yml` executable version
`2026.11.03.0000.0000` and `AssemblyVersion 7.57.0.$(GitCommits)`.
Dalamud `43b34118bcf7213e575ec39517d033191b1acf17` (version 15.0.3.9) pins this revision.
The `dalamud` skill carries its own pin; check each skill's pin separately.

At runtime, `IDalamudPluginInterface.GetDalamudVersion().GitHashClientStructs`
identifies the shipped ClientStructs commit. Follow
[revision selection](guide-conventions-and-version-hygiene.md#select-the-target-revision)
before using a layout or signature. Dalamud services, lifecycle, threading,
and project setup belong to the `dalamud` skill.

Curated source copies live in `references/structs/<upstream path>`;
`references/all-structs.txt` indexes all 1649 `.cs` files at the bundled pin.
The examples were checked against source, without a game or resolver test.
The [provenance record](provenance.md) describes the refresh and its limits.



## Guide sections

Paths written as `references/...` are relative to the skill directory. Read only the section needed for the task.

- [repo layout](guide-repo-layout.md)
- [struct definition and what the generator emits](guide-struct-definition-and-what-the-generator-emits.md)
- [signature resolution at runtime](guide-signature-resolution-at-runtime.md)
- [entry points real members](guide-entry-points-real-members.md)
- [addon ui work](guide-addon-ui-work.md)
- [conventions and version hygiene](guide-conventions-and-version-hygiene.md)
- [contributing a field or function](guide-contributing-a-field-or-function.md)
- [pitfalls](guide-pitfalls.md)
