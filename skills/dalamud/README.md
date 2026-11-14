# dalamud

This skill helps an agent create, port, debug and review C# plugins for Dalamud, the FFXIV plugin framework. It covers Dalamud.NET.Sdk setup, the `IDalamudPlugin` and `IAsyncDalamudPlugin` lifecycle, service injection, framework threading, hooks and signatures, IPC, SeString, Lumina sheets, addon events, dev-plugin loading, publishing requirements and API-level migration. ImGui drawing belongs to the `dalamudimgui` skill.

[SKILL.md](SKILL.md) gives the workflow and routes tasks to the references. [references/guide.md](references/guide.md) records the pinned revisions and links the API examples and limitations. The upstream files under `references/dalamud/`, `references/docs/` and `references/sampleplugin/` support exact-member lookups.

Match the bundled revision to the target plugin's SDK and runtime before using a signature. The guidance was checked by reading source. It was not compiled against Dalamud or tested in game, and it says nothing about a newer game patch.

## Refreshing the snapshot

The paths to copy are listed in `references/source-files.txt` (Dalamud) and `references/docs-files.txt` (dalamud-docs). Copy each one unmodified from a checkout at the new pin, then regenerate the two indexes from the Dalamud checkout root:

```sh
find Dalamud -name '*.cs' | LC_ALL=en_US.UTF-8 sort > all-files.txt
LC_ALL=en_US.UTF-8 xargs grep -nE '^\s*public\s+((static|sealed|abstract|unsafe|partial|readonly|ref|new)\s+)*(class|struct|interface|enum|record|delegate)\s' < all-files.txt \
  | sed -E 's/^([^:]+:[0-9]+):\s*/\1: /' > all-public-types.txt
```

Update the pins in `SKILL.md`, both `NOTICE.md` files, `references/guide.md` and the first line of each `references/topic-*.md`, and recheck the authored guidance against the source comparison between the two pins.

## License

Original prose is MIT licensed. Bundled upstream files and quoted excerpts keep the terms listed in [NOTICE.md](NOTICE.md); the license texts ship with the skill.
