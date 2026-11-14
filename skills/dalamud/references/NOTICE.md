# License and source notices

File paths in this notice are relative to the installed skill directory.

Original `SKILL.md`, `references/guide.md`, `references/guide-*.md`, the authored Markdown under `references/topic-*.md`, and `README.md` are Copyright (c) 2026 sproutwerk, licensed under MIT. The full text is in `LICENSE` beside this file.

Unmodified upstream files retain their own copyright notices and licenses:

| Files | Source | License | License text |
| --- | --- | --- | --- |
| `references/dalamud/` | [goatcorp/Dalamud](https://github.com/goatcorp/Dalamud) at `43b34118bcf7213e575ec39517d033191b1acf17` | AGPL-3.0-only | `references/dalamud/LICENSE` |
| `references/docs/` | [goatcorp/dalamud-docs](https://github.com/goatcorp/dalamud-docs) at `3c7597d643d8a8c01ee55930303057a8ce7c6820` | MIT, Copyright (c) 2023 Goat Systems Interactive | `references/docs/LICENSE` |
| `references/sampleplugin/` | [goatcorp/SamplePlugin](https://github.com/goatcorp/SamplePlugin) at `b8477daaa678a5dd72cdcd4a32bc249d3c23a598` | AGPL-3.0-only | `LICENSES/AGPL-3.0-only.txt` |

Dalamud and SamplePlugin each ship the GNU Affero General Public License version 3 text without an "or any later version" statement, so the identifier used here is AGPL-3.0-only. `LICENSES/AGPL-3.0-only.txt` is SamplePlugin's `LICENSE.md` at the pinned commit and is identical to `references/dalamud/LICENSE`.

`references/all-files.txt`, `references/all-public-types.txt`, `references/docs-files.txt`, and `references/source-files.txt` are indexes generated from those trees by the refresh commands in the skill's `README.md`. They hold paths and type names only. Short excerpts of upstream code quoted in `references/guide.md` and `references/guide-*.md` and `references/topic-*.md` remain under the upstream license.

Dalamud is the upstream project's name. This independently maintained reference skill is not an official Dalamud release or endorsement, and it is not affiliated with goatcorp or Square Enix.

The Dalamud copy is a selection of the upstream tree. `Dalamud/Configuration/PluginConfigurations.cs` is listed in `references/all-files.txt` and `references/all-public-types.txt` but is not bundled.
