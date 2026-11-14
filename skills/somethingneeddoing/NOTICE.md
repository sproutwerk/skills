# Source and license notices

Original guidance in this skill uses the accompanying
[MIT license](LICENSE), copyright 2026 sproutwerk.
That license does not relicense upstream source.

Selected files under `references/source/SomethingNeedDoing/` are unchanged
Git objects from
[Jaksuhn/SomethingNeedDoing](https://github.com/Jaksuhn/SomethingNeedDoing/tree/527ca83892506c74fea4d5e07b9b9436e50ef95d),
commit `527ca83892506c74fea4d5e07b9b9436e50ef95d` (tag `v15.13`). The
project file credits `croizat` as author. Preserve upstream authorship and
any file-level notices.

The inspected tree has no LICENSE, COPYING or NOTICE file at its root, and
no such file exists at any commit in the repository's history
(`git log --all --diff-filter=A -- 'LICENSE*' 'LICENCE*' 'COPYING*'` is
empty apart from the vendored editor directory noted below). The README
describes the plugin and how to install it and contains no license grant.
GitHub reported no license for the repository when checked on 2026-11-13.
Redistribution permission for these retained source references therefore
remains unresolved. Public visibility and the licenses of the plugin's
dependencies do not establish a grant for this source.

The seven files under `references/source/SomethingNeedDoing.wiki/` are
unchanged copies of the pages of that repository's wiki at commit
`129eccefa9f1f1979569c6835aeee9b4a6700089`. The wiki carries no license
statement either, and the same unresolved terms apply.

Not bundled: the `DalamudCodeEditor` directory vendored in the upstream
repository (it has its own license file), the `AutoRetainerAPI` submodule,
and the NuGet dependencies NLua, ECommons, YamlDotNet, AutoCtor and
Injectio. Pre-rewrite sources consulted for the migration notes, in the
same repository's history and in `daemitus/SomethingNeedDoing`, were read
and not copied.

No screenshots, game assets, user configuration or character records are
bundled. Game identifiers and names quoted in the references and examples
(addon names, action names, sheet names) are Square Enix data reproduced
for interoperability.

The [provenance record](references/provenance.md) identifies each inspected
revision. The [source manifest](references/source-manifest.json) records
paths and SHA-256 hashes for every copied upstream file. Keep these notices
and the license text with independently installed copies of the skill. This
reference is independent documentation, is not a contribution to the
upstream repository, and is independent of Square Enix, Dalamud and the
upstream plugin maintainers.
