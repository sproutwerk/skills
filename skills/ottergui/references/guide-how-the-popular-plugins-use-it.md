## How the popular plugins use it

Artisan uses `ItemSelector<T>` (three selectors), `Table<T>` with `ColumnString` and
`ColumnFlags` (two tables), `OtterGui.Raii.ImRaii`, a few `ImGuiUtil` calls, `TextureCache` and
the `Move` list extension. AutoRetainer has no active `using OtterGui` in its own code; it uses
`NightmareUI.OtterGuiWrapper`, whose two file systems derive from `FileSystem<T>` and whose
generic one uses `FileSystemSelector`. Neither uses `Services/`, `Tasks/`, `Compression/`,
`String/` or the `Widgets/` folder. In this ecosystem OtterGui is a widget donor on top of
ECommons, not an application framework. File-by-file list: `references/topic-usage-in-the-wild.md`.

