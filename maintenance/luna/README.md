# luna maintenance

These helpers refresh the bundled upstream reference files and regenerate the indexes for [`luna`](../../skills/luna/SKILL.md). They are kept outside the installable skill.

```sh
git clone https://github.com/Ottermandias/Luna /path/to/Luna
git clone https://github.com/Ottermandias/ImSharp /path/to/ImSharp
git clone https://github.com/Ottermandias/Luna.Targa /path/to/Luna.Targa
git -C /path/to/ImSharp checkout "$(git -C /path/to/Luna ls-tree HEAD ImSharp | awk '{print $3}')"
git -C /path/to/Luna.Targa checkout "$(git -C /path/to/Luna ls-tree HEAD Luna.Targa | awk '{print $3}')"
tools/refresh.sh /path/to/Luna /path/to/ImSharp /path/to/Luna.Targa
```

Check out ImSharp and Luna.Targa at the commits Luna pins, as above, so the three trees match. `refresh.sh` copies exactly the files named in `references/manifest-luna.txt`, `references/manifest-imsharp.txt` and `references/manifest-targa.txt`, replaces the three mirrored trees, recopies the three license texts into both `LICENSES/` directories, and rewrites `api-index.txt`, `all-files.txt` and `VERSIONS.txt`. It prints `MISSING upstream:` for a manifest entry that no longer exists. Before running it, compare the manifests with `git ls-tree -r --name-only HEAD` in each checkout and add, move or drop entries for files upstream added, renamed or removed. The Luna manifest leaves out `Luna/DirectX/`, the shader resources, IDE settings and lock files; the ImSharp manifest leaves out `ImSharp/lib/` and IDE settings; the Luna.Targa manifest leaves out `.editorconfig`, `.gitattributes` and `.gitignore`.

After a refresh, review the diff and update the pins together: the tables in `NOTICE.md` and `references/NOTICE.md`, `references/guide.md`, the "Verified against" line in each `references/topic-*.md`, and `SKILL.md`. License files and source notices stay with the copied material.
