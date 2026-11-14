# ottergui maintenance

These helpers refresh the bundled upstream reference files and regenerate indexes for [`ottergui`](../../skills/ottergui/SKILL.md). They are outside the installable skill. The shell entrypoint writes into `skills/ottergui/references/` in this repository.

Inspect the script and use local checkouts at the intended commits before refreshing. The refresh scripts replace the mirrored source tree, so review the resulting diff and update the pins in the skill guide and topics together. Source notices and license files must remain with copied material.

Run helpers from this directory. Each shell script documents its arguments in its opening comments. Run `go test ./...` here when changing a Go helper.
