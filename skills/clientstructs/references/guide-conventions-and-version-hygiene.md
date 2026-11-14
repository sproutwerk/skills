## Conventions and version hygiene

### Select the target revision

Read the plugin's SDK and dependency pins, then obtain
`IDalamudPluginInterface.GetDalamudVersion().GitHashClientStructs` from its
runtime. If the client is unavailable, inspect `lib/FFXIVClientStructs` at
the intended Dalamud commit and record that runtime matching is unverified.
The API level alone does not identify a ClientStructs revision.

Use the user's existing checkout of FFXIVClientStructs when it matches, or ask
them for one at the exact commit; read the files, do not build or run them.
Verify the commit with `git rev-parse HEAD`. Search
`FFXIVClientStructs/` for the type, then inspect its bases and generated-member
attributes at the same revision. Keep the installed skill's source mirror
unchanged during a consumer task. Current `main` is useful for comparisons;
it does not identify the library shipped by a user's runtime.

### When to refresh this skill

Refresh the bundle when a supported game build changes its reference layouts
or signatures, a documented API or generator contract changes, or recurring
tasks need types missing from the curated sources. Additions to unrelated
types and regenerated export files alone can wait. Recheck the task's exact
runtime revision even when no skill refresh is needed.

For a refresh, choose a full commit ID, compare the curated files and guide
examples, regenerate the complete type index and `data.yml` excerpt, and
update notices, provenance, and the repository's source records together. Preserve upstream
text and licenses; the mirror retains its existing line endings. Verify copied
files against that commit allowing only LF/CRLF differences. Keep source,
build, and in-game verification separate.

### Source conventions

- Struct files carry the C++ name/hierarchy in a header comment
  (`// Client::UI::AddonSelectYesno` / `//   Component::GUI::AtkUnitBase`) and
  often the ctor sig. Official names come from `ida/classinformer.csv` (old
  RTTI); invented names for the rest.
- Unknown fields: `private byte Unk1C;` / `private uint Unk4D00;` named by
  offset, or `Unknown17`; unknown params `a4`, `u3`. Do not build on them.
- File-scoped namespaces, 4-space indent, `var`, CRLF (`.editorconfig`).
  Sigs: 2 hex chars per byte, `??` wildcards, uppercase.
- Offsets and signatures can change with patches, including hotfixes. Check the checkout's
  `ida/data.yml` `version:` header (`2026.11.03.0000.0000`) against the
  running client's `Framework.Instance()->GameVersionString` (or the
  `ffxivgame.ver` file); check `AssemblyVersion` in the csproj (`7.57.0`) and
  the `ida/old/data_<version>.yml` snapshots to see which exe a commit
  targeted. A `ThrowNullAddress` exception at call time = broken sig; a wrong
  value with no exception can indicate a shifted field offset or invalid state.
- Dalamud pins one CS commit per Dalamud release; a plugin built against a
  newer Dalamud API level must compile against that copy. If a member you
  see on `main` is missing, your Dalamud is older than the commit that added
  it.

### Reading `data.yml`

```yaml
version: 2026.11.03.0000.0000
globals:
  0x1428B09A0: g_EmptyUtf8String          # named data addresses (exe base 0x140000000)
functions:
  0x14005E020: WinMain                    # free functions
classes:
  Client::System::Framework::Framework:
    instances:
      - ea: 0x142904928
        pointer: true                     # static Framework* (→ StaticAddress isPointer: true)
      - ea: 0x14290BE90
        name: InstancePointer2
        pointer: true
    vtbls:
      - ea: 0x142129480                   # first vtbl; later ones need `base:` (multiple inheritance)
    vfuncs:
      1: Setup                            # index in vtable → [VirtualFunction(1)]
      4: Tick
    funcs:
      0x1400CEA00: ctor                   # address → name (# comments for args, "static")
      0x1400D1E20: Exit
```

To find a function by name: `grep -nw 'GetInventoryItemCount' ida/data.yml`
gives the address (entries are `      0xADDR: Name # comment` — one space
after the colon, so a fixed-indent pattern like `'  GetInventoryItemCount'`
does not match); the C# `[MemberFunction]` sig in the struct is what
locates it at runtime. To find the C# from an address in a crash log:
`grep -n '0x140875440' ida/data.yml`.

This skill ships only `references/ida/data.yml.excerpt`: the header,
`globals:`/`functions:` samples and the `Framework`, `InventoryManager` and
`AtkUnitBase` class entries (`# ...` marks elided lines). The full
`ida/data.yml` (2637 class entries) exists only in the checkout. When
`references/all-structs.txt` or the excerpt leads beyond the curated files, read
the upstream repository (`aers/FFXIVClientStructs`) at the pinned commit
`80cb4a7b7e11a6702e2eb8839afd301a918c638e` in a checkout the user provides (current `main`
is a newer commit; re-verify offsets against it). `data-validator.js` enforces the
schema (numeric keys, string values, `version` regex, optional numeric vtable
`size`); `data_schema.py` is
the Python model used by the renamer.

