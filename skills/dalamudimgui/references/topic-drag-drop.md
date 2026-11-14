# Drag and drop

Covers in-UI drag and drop through `ImRaii` and external file drops through
`IDragDropManager`.
Verified against goatcorp/Dalamud `master` @ `43b3411` (2026-11-10, API level 15, Dear ImGui 1.88 docking).
Paths are relative to the skill root (`SKILL.md`, `references/topic-*.md`, `references/...`); a bare
`index/...` or `src/...` path means `references/index/...` or `references/src/...`.

In-UI drag and drop is stock ImGui through `ImRaii`:

```csharp
using (var src = ImRaii.DragDropSource())
    if (src)
    {
        ImGui.SetDragDropPayload("MY_ROW"u8, BitConverter.GetBytes(index));
        ImGui.Text($"Move row {index}");
    }

using (var dst = ImRaii.DragDropTarget())
    if (dst)
    {
        var payload = ImGui.AcceptDragDropPayload("MY_ROW"u8);
        if (!payload.IsNull && payload.IsDelivery())
            this.Move(BitConverter.ToInt32(payload.Handle->DataSpan), index);   // unsafe context
    }
```

Files dragged from Explorer into the game go through `IDragDropManager`
(namespace `Dalamud.Interface.DragDrop` in API 15): `ServiceAvailable`,
`IsDragging`, `Files`, `Directories`, `Extensions`,
`CreateImGuiSource(label, validityCheck[, tooltipBuilder])` called every
frame, and `CreateImGuiTarget(label, out files, out directories)` right
after the item that should accept the drop; it returns true on the frame of
delivery.
