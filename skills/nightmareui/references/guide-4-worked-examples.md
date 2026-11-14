## 4. Worked examples

### 4.1 Settings page with the builder

AutoRetainer, `UI/NeoUI/MultiModeEntries/MultiModeFPSLimiter.cs`, abridged:

```csharp
public class MultiModeFPSLimiter : NeoUIEntry          // NeoUIEntry : ConfigFileSystemEntry
{
    public override string Path => "Multi Mode/FPS Limiter";

    public override NuiBuilder Builder { get; init; } = new NuiBuilder()
        .Section("FPS Limiter")
        .TextWrapped("FPS Limiter is only active when Multi Mode is enabled")
        .Widget("Target frame rate when idling", (x) =>
        {
            ImGui.SetNextItemWidth(100f);
            UIUtils.SliderIntFrameTimeAsFPS(x, ref C.TargetMSPTIdle, C.ExtraFPSLockRange ? 1 : 10);
        })
        .Checkbox("Release FPS lock when game is active", () => ref C.NoFPSLockWhenActive)
        .Checkbox($"Allow extra low FPS limiter values", () => ref C.ExtraFPSLockRange,
            "No support is provided if you enable this and run into ANY errors in Multi Mode");
}
```

`C` is AutoRetainer's static configuration accessor. The builder is created once per entry
instance and drawn every frame. Other patterns from the same code base: width plus help text,
`.Widget(100f, "Additional Interaction Delay, frames", (x) => ImGuiEx.SliderInt(x, ref
C.ExtraFrameDelay.ValidateRange(-10, 100), 0, 50), "help...")` (`MainSettings.cs`); a disabled
block made from gated label-less widgets (`MiscTab.cs`, reformatted):

```csharp
        .If(() => Utils.IsBusy).Widget("", (x) => ImGui.BeginDisabled()).EndIf()
        .Checkbox($"Remove minimized FPS restrictions while plugin is operating", () => ref C.UnlockFPS)
        .If(() => Utils.IsBusy).Widget("", (x) => ImGui.EndDisabled()).EndIf();
```

A throwaway builder inside a draw method also works:

```csharp
new NuiBuilder().Section("Character Data Expunge/Reset", collapsible: true)
    .Widget(() => { /* plain ImGui calls */ }).Draw();   // AutoRetainer, UI/MainWindow/SharedUI.cs

new NuiBuilder()                                        // Splatoon, Gui/CGuiRenderers.cs
    .Section("Attention Color".Loc())
    .TextWrapped("Attention color is used to highlight ...".Loc())
    .EnumCombo(200f, "Attention Color Type", () => ref P.Config.AttentionColorType)
    .Widget(() => { /* ... */ })
    .Draw();
```

### 4.2 Tabbed window with `ButtonTabs`

```csharp
// Splatoon, Gui/LogWindow.cs: whole window body, content in a child
public override void Draw()
{
    NuiTools.ButtonTabs([[new("Standard log", InternalLog.PrintImgui), new("Filtered log", PrintFiltered)]]);
}

// AutoRetainer, UI/MainWindow/AutoRetainerWindow.cs: a conditional tab through null
NuiTools.ButtonTabs([[C.RecordStats ? new("Ventures", S.VentureStats.DrawVentures) : null,
                      new("Gil", S.GilDisplay.Draw), new("FC Data", S.FCData.Draw)]]);

// Splatoon, Gui/Layouts/CGuiLayouts.cs: tabs as a filter selector; stable InternalName,
// localised Name, no child because the content is drawn by the caller afterwards
private readonly NuiTools.ButtonInfo[] ExpansionTabs = [new("All", () => ActiveExpansion = null),
    .. Enum.GetValues<Expansion>().Select(x => new NuiTools.ButtonInfo(
        x.ToString().Replace('_', ' ').Loc(), x.ToString(), () => ActiveExpansion = x))];
NuiTools.ButtonTabs("LayoutsButtonTabs", [shortExpansions ? ExpansionTabsShort : ExpansionTabs], child: false);
```

Persisting the state (Splatoon, `Serializables/Configuration.cs`): the configuration class has
`public NightmareUIState NightmareUIState = new();` and its initialisation method calls
`NuiTools.SetState(NightmareUIState);`.

