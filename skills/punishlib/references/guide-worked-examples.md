## Worked examples

Bootstrap, synchronous plugin (Artisan `Artisan.cs:57-58`; `ECommonsMain.Dispose()` at `:284`):

```csharp
using ECommons;
using PunishLib;

ECommonsMain.Init(pluginInterface, this, Module.DalamudReflector);
PunishLibMain.Init(pluginInterface, "Artisan",
    new AboutPlugin() { Sponsor = "https://ko-fi.com/taurenkey" });
```

Bootstrap, async plugin with a named developer (AutoHook `AutoHook.cs:56-59`, its `CLibMain.Init`
line between the two calls omitted):

```csharp
public async Task LoadAsync(CancellationToken cancellationToken) {
    ECommonsMain.Init(pluginInterface, this, Module.DalamudReflector, Module.ObjectFunctions);
    PunishLibMain.Init(pluginInterface, "AutoHook", new AboutPlugin() {
        Developer = "InitialDet & croizat", Sponsor = "https://ko-fi.com/initialdet" });
```

With `Developer` set the header reads "Developed with (heart) by InitialDet & croizat, published
by Puni.sh"; without it, "Published and developed with (heart) by Puni.sh".

Other forms: `PunishLibMain.Init(pluginInterface, "Saucy", new AboutPlugin());` (no sponsor
button); `PunishLibMain.Init(pi, Name, PunishOption.DefaultKoFi);` (AutoRetainer).

About tab inside a config window. Consumers select it from a sidebar and draw it in the content
child (AutoHook `PluginUI.cs:122-133`, PandorasBox `UI/MainWindow.cs:212-213`, Saucy
`Core/UI/PluginUI.cs:264`, Artisan `UI/PluginUI.cs:261`):

```csharp
using PunishLib.ImGuiMethods;

if (ImGui.Selectable($"{UIStrings.AboutTab}"))          // left column
    _selectedTab = OpenWindow.About;
ImGui.TableNextColumn();                                // right column
using var rightChild = ImRaii.Child($"###AhRight", Vector2.Zero, false);
if (_selectedTab == OpenWindow.About)
    AboutTab.Draw("AutoHook");
```

In a tab bar, put the same call inside the tab item. `Draw` creates no window, child or tab
of its own, apart from the floating API-key window.

Icon button (Artisan `CraftingList/Teamcraft.cs:41`; the last argument puts the icon on the right):

```csharp
if (IconButtons.IconTextButton(Dalamud.Interface.FontAwesomeIcon.Upload, "Export",
        new Vector2(ImGui.GetContentRegionAvail().X, 30), true))
```

Info box and group box. No consumer usage was found, so these are composed from the signatures:

```csharp
InfoBox.DrawBox("General##settings", () => ImGui.Checkbox("Enabled", ref enabled));

if (ImGuiGroup.BeginGroupBox("Advanced", new ImGuiGroup.GroupBoxOptions { Collapsible = true }))
{
    ImGui.TextUnformatted("content");
    ImGuiGroup.EndGroupBox();
}
```

