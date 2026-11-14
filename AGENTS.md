# Maintaining these skills

Skills live in `skills/<name>/SKILL.md`. Skill names have no `-skill` suffix.
Keep each installed skill self-contained, including its license notices and
references. `skills.sh.json` and `sources.json` must list the same skills as
the `skills/` directory. Use a concise entrypoint and link detailed references
where the task needs them.

Verify API claims against the target version's source. Preserve upstream
snapshots and attribution when editing authored guidance. Treat snippets as
examples unless a build or runtime check demonstrates otherwise. Do not run
live game actions merely to validate documentation.

Write authored prose plainly. Keep code, identifiers, source notices, links,
and stated uncertainty intact. Review does not require changing clean text.

## Sources and licenses

Original authored work is MIT, `Copyright (c) 2026 sproutwerk`. Bundled
upstream files keep their own terms: never edit an upstream copyright line or
license text, and record each upstream repository and revision in the skill's
`NOTICE.md` and in `sources.json`. Update both when refreshing imported
content.

## Checks

CI installs every skill with the skills.sh CLI and runs the Go tests for the
helpers under `maintenance/` and `skills/xivapi/scripts/`. Run the same checks locally:

```sh
bash .github/scripts/check-skills-sh.sh .
(cd maintenance/audits && go vet ./... && go test ./...)
(cd maintenance/dalamudimgui && go test ./...)
(cd maintenance/ottergui && go test ./...)
(cd skills/xivapi/scripts/xivapi && go vet ./... && go test ./...)
```

Hosted skills.sh security audits are separate from CLI installation and are
not asserted by CI. The audit table in `README.md` is generated: after
skills.sh has rescanned a change, run `go -C maintenance/audits run .` and
commit the README on its own, apart from the change that caused the scan.

Do not add `cisco-ai-skill-scanner` (`skill-scanner`) or any other tool that
depends on LiteLLM to the checks, directly or transitively.
