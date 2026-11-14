# File dialog

Covers `FileDialogManager`: open/save file and folder dialogs, filter syntax, multi-select.
Verified against goatcorp/Dalamud `master` @ `43b3411` (2026-11-10, API level 15, Dear ImGui 1.88 docking).
Paths are relative to the skill root (`SKILL.md`, `references/topic-*.md`, `references/...`); a bare
`index/...` or `src/...` path means `references/index/...` or `references/src/...`.

```csharp
private readonly FileDialogManager files = new();

// somewhere in Draw
if (ImGui.Button("Import"u8))
    this.files.OpenFileDialog("Import preset", "Presets{.json,.txt},.*", (ok, path) =>
    {
        if (ok) this.Import(path);
    });

this.files.Draw();          // every frame, outside any window scope
```

`FileDialogManager` (`Dalamud.Interface.ImGuiFileDialog`) shows one ImGui
(not OS) dialog at a time: `OpenFileDialog`, `OpenFolderDialog`,
`SaveFileDialog(title, filters, defaultFileName, defaultExtension, callback)`,
`SaveFolderDialog`, each with a `(..., startPath, isModal = false)` overload;
multi-select via `OpenFileDialog(title, filters, Action<bool, List<string>>, selectionCountMax, ...)`.
Filter syntax: `"Source files{.cpp,.h},Image files{.png,.jpg},.md"`.
Callbacks run during `Draw()`.
