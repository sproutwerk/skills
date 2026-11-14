## Settings UI

```csharp
public class Config
{
    public bool Active = false;
    public int Threshold = 1500;
    public uint ItemId = 0;
    public string Path = "";
}
Config C => Controller.GetConfig<Config>();

public override void OnSettingsDraw()
{
    ImGui.Checkbox("Active", ref C.Active);
    ImGui.SetNextItemWidth(150f);
    ImGui.InputInt("Threshold", ref C.Threshold);
    ImGuiEx.InputUint("Item id", ref C.ItemId);
    ImGuiEx.Text(EColor.YellowBright, $"State: {state}");
    ImGuiEx.HelpMarker("tooltip");
    if(ImGui.CollapsingHeader("Debug"))
    {
        if(ImGui.Button("Run once")) Enqueue();
        if(TaskManager?.IsBusy == true && ImGui.Button("Abort")) TaskManager.Abort();
    }
}
```

`Dalamud.Bindings.ImGui` is the ImGui namespace (not ImGuiNET).
`ECommons.ImGuiMethods.ImGuiEx` adds `Text(color, s)`, `TextWrapped`,
`TextV`, `TextCentered`, `Checkbox(FontAwesomeIcon, label, ref bool)`,
`Button(label, enabled)`, `IconButton(icon)`, `IconButtonWithText(icon,
text)`, `Tooltip(s)`, `HelpMarker(s)`, `InputUint`, `Combo(id, ref value,
keys, names:)`, `EnumCombo`, `FilteringInputTextWithHint`,
`BeginDefaultTable(headers)`, `CollapsingHeader`, `SetNextItemFullWidth()`,
`InputTextMultilineExpanding`, `RealtimeDragDrop<T>`. Icons:
`ThreadLoadImageHandler.TryGetIconTextureWrap(iconId, hq, out tex)` then
`ImGui.Image(tex.Handle, new(size))`; it returns false until loaded, just
skip that frame. Remove list items with `new TickScheduler(() =>
list.Remove(x))` (ECommons.Schedulers), never inside the loop that draws
them.

