## OtterGui to Luna

Luna (with its ImGui layer ImSharp) is the successor by the same author and a rewrite, not a
rename: `ImUtf8.*` becomes `Im.*` / `ImEx.*`, `ImRaii.PushColor` / `PushStyle` become
`ImGuiColor.X.Push` / `ImStyleSingle.X.Push`, `Table<T>` becomes `ImSharp.Table` classes,
`FileSystem<T>` becomes `BaseFileSystem`, `EventWrapper` becomes `EventBase`, and `ItemSelector<T>`,
`LowerString` and `ClippedSelectableCombo<T>` have no counterpart. The full table is
`references/topic-ottergui-mapping.md` in the sibling `luna` skill. Luna is AGPL and used by
Penumbra/Glamourer-style plugins only; do not migrate a puni.sh-style plugin to it by default.

