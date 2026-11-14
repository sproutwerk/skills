# Source and license notices

Original guidance in this skill uses the accompanying
[MIT license](LICENSE), copyright 2026 sproutwerk.
That license does not relicense upstream source.

Selected files under `references/source/Lifestream/` are unchanged Git
objects from
[NightmareXIV/Lifestream](https://github.com/NightmareXIV/Lifestream/tree/b36320170faafc8315f7585edb45cfbcd34e7046),
commit `b36320170faafc8315f7585edb45cfbcd34e7046` (release tags `2.6.0.8`
through `2.6.0.13`).
The plugin manifest and project credit `NightmareXIV` as author. Preserve
upstream authorship and any file-level notices.

The inspected tree has no LICENSE, COPYING or NOTICE file, and no such file
exists at any commit in the repository's history
(`git log --all --diff-filter=A -- 'LICENSE*' 'LICENCE*' 'COPYING*'` is
empty). The README describes features, support links and installation but
contains no license grant. GitHub reported no license for the repository
when checked on 2026-11-13. `CONTRIBUTING.md` at the
pin contains a "Relicensing" clause under which contributors allow
NightmareXIV to relicense their commits, and states a no-unsolicited-pull-
request and no-AI-contribution policy for the plugin itself; neither grants
redistribution rights to third parties. Redistribution permission for these
retained source references therefore remains unresolved. Public
visibility, the MIT license of its ECommons dependency, and the relicensing
clause do not establish a grant for this source. This skill is independent documentation and is not a
contribution to the upstream repository.

`Lifestream/Movement/README.md` credits awgil and
[ffxiv_navmesh](https://github.com/awgil/ffxiv_navmesh) for the movement
code copied into `Movement/FollowPath.cs` and its siblings. That upstream
has no recorded license grant either (see the `vnavmesh` skill notice), so
the follower sources carry two unresolved provenance questions.

`references/source/ECommons/ECommons/ErrorCode.cs` is an unchanged file from
[NightmareXIV/ECommons](https://github.com/NightmareXIV/ECommons/tree/958af5c7e735e1633b92070a5b67f8d6c48c30f3),
the plugin's gitlink at `958af5c7e735e1633b92070a5b67f8d6c48c30f3`. ECommons
is MIT, copyright 2023 NightmareXIV; the unchanged text is in
[LICENSES/ECommons-LICENSE.md](LICENSES/ECommons-LICENSE.md). Only that one
file is bundled from the dependency; its other sources and transitive
dependencies are not distributed here, and the ECommons license does not
cover the Lifestream files.

Other skills in this collection that bundle Lifestream files carry their own
notices under the same unresolved terms; the copies here supersede nothing
there. No screenshots, game assets, user configuration, address
books or live character records are bundled. Game identifiers quoted in the
references (aetheryte, territory and sheet row IDs) are Square Enix data
reproduced for interoperability.

The [provenance record](references/provenance.md) identifies each inspected
revision. The [source manifest](references/source-manifest.json) records
paths and SHA-256 hashes for every copied upstream file. Keep these notices
and license texts with independently installed copies of the skill. This
reference is independent of Square Enix, Dalamud, NightmareXIV and the
upstream plugin maintainers.
