# License and source notices

File paths in this notice are relative to the installed skill directory unless a source-repository path is stated.

Original `SKILL.md`, `references/guide.md`, `references/guide-*.md`, the Markdown files under `references/topic-*.md`, `README.md`, the refresh scripts kept under `maintenance/luna/tools/` in the source repository, and the manifest and version files under `references/` (`manifest-luna.txt`, `manifest-imsharp.txt`, `manifest-targa.txt`, `VERSIONS.txt`) are Copyright (c) 2026 sproutwerk, licensed under MIT. The full text is in `LICENSE` beside this file.

Unmodified upstream files keep their own copyright notices and licenses:

| Files | Source | License | License text |
| --- | --- | --- | --- |
| `references/Luna/` | [Ottermandias/Luna](https://github.com/Ottermandias/Luna) at `52eb0fec6795b10c7e389c268486def9d18a9c36` | AGPL-3.0-only | `references/Luna/LICENSE`, also at `LICENSES/AGPL-3.0-only.txt` |
| `references/ImSharp/` | [Ottermandias/ImSharp](https://github.com/Ottermandias/ImSharp) at `4d09327e3d7ea7d4cfda7fef2a0ca04dddda5cee` | MIT, Copyright (c) 2025 Ottermandias | `references/ImSharp/LICENSE`, also at `LICENSES/MIT-ImSharp.txt` |
| `references/Luna.Targa/` | [Ottermandias/Luna.Targa](https://github.com/Ottermandias/Luna.Targa) at `8f66a2d897b9c5ad9b299a86039f1bac83fe4234` | MIT, Copyright (c) 2026 Ottermandias | `references/Luna.Targa/LICENSE`, also at `LICENSES/MIT-Luna.Targa.txt` |
| `references/api-index.txt`, `references/all-files.txt` | Generated from the three upstream checkouts by `maintenance/luna/tools/refresh.sh` and `maintenance/luna/tools/build-index.sh`; they hold upstream declaration lines and file paths | Lines that start with `Luna/` follow Luna's license; lines that start with `ImSharp/` follow ImSharp's; lines that start with `Luna.Targa/` follow Luna.Targa's | The three texts above |

Luna ships the GNU Affero General Public License version 3 text, and GitHub reports it as `AGPL-3.0`. Upstream does not state that later versions apply, so it is recorded here as `AGPL-3.0-only`. All three upstream license files are byte-identical to upstream at the pinned commits. The commits and their dates are in `references/VERSIONS.txt`, and the mirrored file sets are in the three manifests.

The authored files quote upstream type names, member names and signatures for reference.

Luna, ImSharp and Luna.Targa are the upstream projects' names. This independently maintained reference skill is not an official release or endorsement of any of them.

The bundled trees are selections listed in `references/manifest-luna.txt`, `references/manifest-imsharp.txt` and `references/manifest-targa.txt`. Luna's shader build script, `Luna/Build-Shaders.ps1`, is not bundled.
