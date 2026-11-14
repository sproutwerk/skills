## Setup in one screen

```sh
git submodule add https://github.com/Limiana/OtterGui OtterGui
git -C OtterGui checkout a08fda67dfb731f007234c78df32a2b3ddccc450
git submodule update --init --recursive        # OtterGui/OtterGuiInternal is a nested submodule
```

```xml
<!-- YourPlugin/YourPlugin.csproj, as in Artisan and AutoRetainer -->
<ProjectReference Include="..\OtterGui\OtterGui.csproj" />
```

- `OtterGui.csproj` has **no** `ProjectReference` to OtterGuiInternal. The nested submodule's
  `.cs` files are compiled into the `OtterGui` assembly by the default glob. An empty
  `OtterGui/OtterGuiInternal` directory (submodules not recursive) gives CS0246 / CS0234 errors on
  `OtterGuiInternal`.
- The library is coupled to one Dalamud API level through the `Sdk=` version in its csproj, the
  ImGui binding internals it calls (`ImGuiP`, `ImGuiNative`, `ImRect`) and Dalamud enums cloned
  from FFXIVClientStructs. An API bump needs a matching OtterGui commit. Details and the bump
  history: `references/topic-setup.md`.
- It brings `Microsoft.Extensions.DependencyInjection` 9.0.2 and `JetBrains.Annotations` 2024.3.0
  as NuGet dependencies.

