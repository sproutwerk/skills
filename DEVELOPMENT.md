# Development

How to check a change and keep the README's audit table current. The rules
for writing and licensing skills are in [AGENTS.md](AGENTS.md).

## Requirements

- [Node.js](https://nodejs.org) 22.20 or newer, for the skills CLI
- Go, at the version in each helper's `go.mod`
- `jq` and `git`

## Checks

CI runs these on every push. Run them from the repository root before
pushing:

```sh
bash .github/scripts/check-skills-sh.sh .
(cd maintenance/audits && go vet ./... && go test ./...)
(cd maintenance/dalamudimgui && go test ./...)
(cd maintenance/ottergui && go test ./...)
(cd skills/xivapi/scripts/xivapi && go vet ./... && go test ./...)
```

The first one installs every skill into a temporary project with the skills
CLI and compares the result with `skills/`.

## Security audits

[skills.sh](https://skills.sh/sproutwerk/skills) runs three audits on each
skill: Gen Agent Trust Hub, Socket, and Snyk. They are hosted, so CI does not
assert them, and their results arrive some time after a push. The README
shows them in a generated table.

After a change to anything under `skills/`:

1. Push the change.
2. Install from the public repository with telemetry enabled, in a scratch
   directory. skills.sh learns about new skills and new content from
   installs, not from pushes.

   ```sh
   npx skills add sproutwerk/skills --skill '*' --agent codex --copy --yes
   ```

3. Ask whether the rescan has landed:

   ```sh
   go -C maintenance/audits run . check
   ```

   It compares the scan time on each audit's page with the newest commit
   under that skill's directory, lists what is still older, and exits
   non-zero until every audit is newer. How long a rescan takes is not
   documented; a newly listed skill has been scanned within minutes.
4. When it reports ready, rewrite the table and commit the README by itself:

   ```sh
   go -C maintenance/audits run .
   git commit -m "Refresh the audit table" README.md
   ```

Keep that commit apart from the change that caused the scan. The README is
outside every skill directory, so committing it does not start another scan.

To fix a finding, read its detail page, linked from the table. The audits
read bundled upstream source as well as the guidance, so a finding may
describe a plugin's own code. Files left out of a snapshot for that reason
are named in the skill's `NOTICE.md` and, where a refresh script exists, in
`maintenance/<skill>/tools/index-only.txt`.

More on the helper is in [maintenance/audits/README.md](maintenance/audits/README.md).
