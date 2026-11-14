# License and source notices

Original `SKILL.md`, `references/provenance.md`, `references/lookup-recipes.md`, the Go helper under `scripts/xivapi/`, and the commands and commentary in `references/cookbook.md` are Copyright (c) 2026 sproutwerk, licensed under MIT. The full text is in `LICENSE` beside this file.

No bundled upstream file here has settled license terms, so this notice has no table of third-party licenses and the skill has no `LICENSES/` directory.

## Not covered by these licenses

MIT does not apply to the material below. Its redistribution terms are not settled. Nothing in this repository grants rights to it.

| Files | Source | What is known |
| --- | --- | --- |
| `references/schemas/*.yml` | [xivdev/EXDSchema](https://github.com/xivdev/EXDSchema), `latest` branch. Every file is identical to commit `a1e24a0c72058353f30f23beb996792cdb926868`, the branch head on 2026-11-13. | The upstream repository has no license file and GitHub reports no license for it. |
| `references/openapi.json` | Saved from `https://v2.xivapi.com/api/openapi.json` on 2026-11-13, server version `1-48c1413`, which matches [ackwell/boilmaster](https://github.com/ackwell/boilmaster) commit `48c1413b1eb8e333799e8b467822a50a2c55a59d`. | boilmaster is AGPL-3.0-or-later. The description itself carries no license field, and whether the server's license extends to its generated API description is not established. |
| API response output quoted in `references/cookbook.md`, `references/lookup-recipes.md` and the examples in `SKILL.md` | FINAL FANTASY XIV game data returned by the XIVAPI v2 service. | Sheet names, row values and in-game text belong to Square Enix. No grant covering them is recorded here. |

Sources, snapshot dates and their limits are in `references/provenance.md`.

XIVAPI, boilmaster and EXDSchema are the names of their projects. This independently maintained reference skill is not an official release or endorsement of XIVAPI, EXDSchema or Square Enix.
