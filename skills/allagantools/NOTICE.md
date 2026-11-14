# License and source notices

Original guidance and original consumer examples use [MIT](LICENSE), copyright
(c) 2026 sproutwerk. That grant does not replace the terms of
bundled upstream files.

`references/source/InventoryTools/` contains unchanged selected files from
[Critical-Impact/InventoryTools](https://github.com/Critical-Impact/InventoryTools/tree/ad1ff01fede1848dfcd05b76a9ae13b0760d1222).
The plugin identifies itself as Allagan Tools, author `Critical_Impact`, internal
name `InventoryTools`. Its root license is the GNU General Public License,
version 3; the unchanged text is included as
[InventoryTools-GPL-3.0.txt](LICENSES/InventoryTools-GPL-3.0.txt) and with the
snapshot. Upstream authors and contributors retain their rights.

`references/source/CriticalCommonLib/` contains unchanged selected files from
[Critical-Impact/CriticalCommonLib](https://github.com/Critical-Impact/CriticalCommonLib/tree/420f6880b16497a6d6356f6dbeda963364e4d5ef),
the InventoryTools gitlink at the reviewed revision. Its root license is also
the GNU General Public License, version 3, preserved as
[CriticalCommonLib-GPL-3.0.txt](LICENSES/CriticalCommonLib-GPL-3.0.txt) and in the
snapshot. No separate file-level license header was found in the selected
implementation files. This records the observed notices without deciding an
"or later" grant from the license's application examples.

The source copies are GPL material, including their comments and any copied
upstream documentation. They are not MIT examples. Keep the license texts,
source attribution, and [file manifest](references/source-manifest.json)
with them. The snapshots are partial source references, not a complete build
or a claim that all GPL distribution requirements have been satisfied. Anyone
redistributing them takes on the GPL terms, including any corresponding-source
obligations.

CriticalCommonLib's `Enums/InventoryType.cs` was inspected but is not copied.
Its file comments credit Sapphire's `Common.h` at
`c5d63e2eccf483c0e785d373ab9ea9a504c734f4` and Teamcraft's
`container-type.ts` through an unpinned branch link. Follow those file-level
origins and their terms before expanding the snapshot to include that enum;
the CCL root license alone does not resolve inherited attribution. The
inventory reference links to the pinned CCL file so those notices remain
visible at their source.

OtterGui and NuGet dependencies are recorded for build context, not bundled as
source or relicensed here. Dependency references do not make this skill a
redistribution of those complete libraries. A full plugin build has additional
source and dependency notice requirements beyond this selected snapshot.

This skill is independent of Critical_Impact and Square Enix. FFXIV and FINAL
FANTASY are associated with Square Enix.
