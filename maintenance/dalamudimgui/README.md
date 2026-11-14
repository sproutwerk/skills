# dalamudimgui maintenance

These helpers refresh the bundled upstream reference files and regenerate the indexes for [`dalamudimgui`](../../skills/dalamudimgui/SKILL.md). They are outside the installable skill.

`tools/refresh.sh <Dalamud checkout> [<dalamud-docs checkout>]` copies every file listed in `references/manifest.txt` from the Dalamud checkout, rebuilds `references/index/`, copies the API migration notes when a dalamud-docs checkout is given, and rewrites `references/PROVENANCE.txt`. It needs Go, GNU grep and git. It deletes and recreates `references/src` and `references/upstream-docs`, so check out the intended commits first and review the resulting diff.

To bundle a new upstream file, add its path to `references/manifest.txt` and run the script again. A path that no longer exists upstream is reported as `MISSING`. After a refresh, update the pins in `SKILL.md`, both `NOTICE.md` files, `references/guide.md` and the `Verified against` line of each `references/topic-*.md`. Source notices and license files must stay with the copied material.

Run `go test ./...` in this directory after changing a Go helper.
