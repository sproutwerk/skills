## Namespaces

`GenericHelpers`, `ECommonsMain`, `Module` and `EzCmd` are in the root namespace `ECommons`. There
is no `ECommons.GenericHelpers` namespace; `using ECommons;` brings in the extension methods and
`using static ECommons.GenericHelpers;` the plain statics.

```csharp
using ECommons;                                        // GenericHelpers + extensions, EzCmd, Module
using ECommons.DalamudServices;                        // Svc
using ECommons.Logging;                                // PluginLog, DuoLog, InternalLog
using ECommons.ImGuiMethods;                           // ImGuiEx, Notify, EColor
using ECommons.Automation.NeoTaskManager;              // TaskManager (current)
using ECommons.UIHelpers.AddonMasterImplementations;   // AddonMaster
using Callback = ECommons.Automation.Callback;         // Callback is ambiguous otherwise
```

Others follow the folder name (`ECommons.Configuration`, `ECommons.SimpleGui`,
`ECommons.Throttlers`, `ECommons.Automation`, `ECommons.Automation.UIInput`,
`ECommons.GameHelpers`, `ECommons.GameFunctions`, `ECommons.ExcelServices`,
`ECommons.EzIpcManager`, `ECommons.EzHookManager`); full list in `references/topic-setup.md`.

