## Bootstrapping: context, logger, window system

ImSharp needs an `ImSharpContext` before any interpolated string or font
helper works. In a plugin that is `Luna.ImSharpDalamudContext`
(`Luna/Services/ImSharpDalamudContext.cs`), an `ISingletonService` and `IRequiredService`:

```csharp
public ImSharpDalamudContext(IDalamudPluginInterface pluginInterface, IUiBuilder uiBuilder,
                             IFramework framework, ILogger logger)
```

What it does: `ImSharpConfiguration.SetLogger(logger)`; hands the D3D device to
`CustomRenderManager.Instance`; subscribes `ImSharpPerFrame.OnUpdate` to
`IUiBuilder.Draw`; obtains the shared context with
`pluginInterface.GetOrCreateData($"ImSharp.Context.V{ImSharpContext.CurrentVersion}", …)`
(`CurrentVersion = 2`) so all Luna plugins share one set of native buffers;
then `framework.RunOnTick(() => ImSharpConfiguration.SetContext(context, false))`.
So the context becomes valid on a later framework tick, not in the
constructor. `Dispose` relinquishes the shared data. It also exposes
`static AwesomeFont`, `MonoFont`, `DefaultFont` (`Im.Native.ImFont*`) and the
static events `StyleChanged` / `FontChanged`, and marks `CacheManager.Instance`
dirty when Dalamud's default font, scale or style change.

`ImSharpContext` (unmanaged struct) holds four native buffers allocated with
`Marshal.AllocHGlobal`: Hint 128 KiB, Label 128 KiB, Text 4 MiB, Input 8 MiB
(each minus one byte), plus `ImGuiContext`, `MonoFont`, `DefaultFont`,
`ImNodesContext` pointers. `ImSharpConfiguration.IsInitialized` is true once a
non-empty context is set.

`ImSharpPerFrame.OnUpdate()` refreshes the cached `Im.Context`, `Im.Style`,
`Im.Io` pointers once per ImGui frame and resets
`InputStringHandlerBuffer.FrameStorageString`.

Window system: `Luna.WindowSystem` (sealed, derives
`Dalamud.Interface.Windowing.WindowSystem`) registers its own `Draw` on
`IUiBuilder.Draw`, skips drawing while `RequiresImSharp && !ImSharpConfiguration.IsInitialized`,
and logs when the ImGui color/style stack size differs before and after
drawing ("pushed N more colors to the stack than it popped"). Create with
`WindowSystem.Create(IUiBuilder builder, string name, bool requiresImSharp = true)`
or register `WindowSystem.Factory(string name)`; `new ServiceManager(log, name)`
registers that factory for you as a scoped service.

