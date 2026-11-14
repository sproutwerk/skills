## Project setup

```xml
<Project Sdk="Dalamud.NET.Sdk/15.0.0">
```

Per the v13 notes, `Dalamud.NET.Sdk` references the three binding assemblies
automatically. Without the SDK, reference `Dalamud.Bindings.ImGui`,
`Dalamud.Bindings.ImPlot`, `Dalamud.Bindings.ImGuizmo` from `$(DalamudLibPath)`
with `<Private>false</Private>` and remove `ImGui.NET` *(HintPath form inferred
from the pre-SDK pattern in the docs; the SDK itself is not in this checkout)*.
The `targets/` folder was removed in API 14. Never ship your own copy of the
binding DLLs or of cimgui.

The SDK reference is switchable: `Dalamud.NET.Sdk` 15.0.0 `Sdk/Sdk.props`
defaults `Use_Dalamud_ImGui` to `true` and adds the three binding references
only under that condition. Plugins built on Luna set
`<Use_Dalamud_ImGui>false</Use_Dalamud_ImGui>` and draw through ImSharp
(`Im.*`), a separate binding to the same native cimgui 1.88, so `ImGui.*`
does not exist in those projects (Penumbra, Glamourer). `ImRaii`,
`ImGuiHelpers`, `ImGuiComponents` and `Window` live in `Dalamud.dll` and take
`Dalamud.Bindings.ImGui` types in their signatures. See the `luna` skill
before mixing the two.

```csharp
using System.Numerics;                          // Vector2/Vector4
using Dalamud.Bindings.ImGui;                   // ImGui, ImGuiP, ImGuiNative, enums, *Ptr structs
using Dalamud.Interface;                        // IUiBuilder/UiBuilder, FontAwesomeIcon, ColorHelpers
using Dalamud.Interface.Colors;                 // ImGuiColors
using Dalamud.Interface.Components;             // ImGuiComponents
using Dalamud.Interface.Utility;                // ImGuiHelpers, ImGuiClip, ImGuiTable, ImGuiId
using Dalamud.Interface.Utility.Raii;           // ImRaii
using Dalamud.Interface.Windowing;              // Window, WindowSystem
```

Also: `Dalamud.Interface.Textures` (+ `.TextureWraps`), `.ManagedFontAtlas`,
`.GameFonts`, `.ImGuiNotification`, `.ImGuiFileDialog`, `.ImGuiSeStringRenderer`.

