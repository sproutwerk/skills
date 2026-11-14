# Dalamud plugin development

Dalamud is the C# plugin framework injected into `ffxiv_dx11.exe` by XIVLauncher. A plugin is a .NET
class library whose entry class implements `IDalamudPlugin` (or `IAsyncDalamudPlugin`) and receives
services through dependency injection. Repo `github.com/goatcorp/Dalamud` (AGPL-3.0), docs
`dalamud.dev` (repo `goatcorp/dalamud-docs`, MIT); the API reference at `dalamud.dev/api` is
generated from the XML doc comments in the Dalamud source, so the source is the API truth.

This skill tracks:

| What | Pin |
|---|---|
| `goatcorp/Dalamud` branch `master` | commit `43b3411` (2026-11-10), `Dalamud.csproj` `DalamudVersion 15.0.3.9`, `TargetFramework net10.0-windows`, Lumina 7.7.0. Tag `15.0.4.0` on branch `rc` is this commit plus the version bump. |
| `goatcorp/dalamud-docs` | commit `3c7597d` (2026-11-13) |
| `goatcorp/SamplePlugin` `master` | commit `b8477da` (2026-08-15), `Dalamud.NET.Sdk/15.0.0` |
| `lib/FFXIVClientStructs` submodule of that Dalamud commit | `aers/FFXIVClientStructs` `80cb4a7` (2026-11-10); compare it with the pin the `clientstructs` skill states |

**API level: 15 is current stable** (released 2026-04-29 with patch 7.5, .NET 10). **API 16 is
upcoming** (branch `api16`, "Not finalized", ships with patch 8.0, release date TBA). How this was
determined: `docs/versions/index.md` lists Release = 15; `master` builds assembly version 15.0.3.9 and `PluginManager.DalamudApiLevel =
Assembly.Version.Major`; newest `Dalamud.NET.Sdk` on NuGet is 15.0.0; `docs/versions/v16/index.md` is
marked not finalized and the checkout carries `[Api16ToDo]` markers. Everything below describes
**API 15** unless it says otherwise.

Everything named here is verified from those checkouts unless marked *(unverified)*. Verbatim
copies: `references/dalamud/<upstream path>` (215 files, list in `references/source-files.txt`),
`references/docs/<upstream path>` (51 files, `references/docs-files.txt`). To find a type that is not
covered: `grep -n 'TypeName' references/all-public-types.txt` (765 public type declarations as
`path:line: declaration`) or `references/all-files.txt` (965 files), then read it from a checkout
pinned to `43b3411`. Never invent a member; if it is not in the reference, say so.

Out of scope, one pointer each:

- ImGui drawing, `Dalamud.Interface.Windowing`, `ImRaii`, fonts, textures for UI → sibling skill
  `dalamudimgui`. Only the `IUiBuilder` hookup points are listed here.
- FFXIVClientStructs structs, offsets, addon node trees, `AtkValue`, the generated `Delegates` /
  `MemberFunctionPointers` / `StaticVirtualTablePointer` names used as hook targets, and the member
  lists of the enums Dalamud clones → skill `clientstructs`.
- Ottermandias' Luna UI/helper library (successor to OtterGui) → sibling skill `luna`.
- Looking up game IDs without the client → skill `xivapi`.


## Guide sections

Paths written as `references/...` are relative to the skill directory. Read only the section needed for the task.

- [topic files](guide-topic-files.md)
- [project setup](guide-project-setup.md)
- [entry point and lifecycle](guide-entry-point-and-lifecycle.md)
- [dependency injection](guide-dependency-injection.md)
- [idalamudplugininterface](guide-idalamudplugininterface.md)
- [services](guide-services.md)
- [threading](guide-threading.md)
- [hooking and signatures](guide-hooking-and-signatures.md)
- [commands](guide-commands.md)
- [configuration](guide-configuration.md)
- [ipc](guide-ipc.md)
- [sestring and chat](guide-sestring-and-chat.md)
- [lumina and idatamanager](guide-lumina-and-idatamanager.md)
- [addonlifecycle and addoneventmanager](guide-addonlifecycle-and-addoneventmanager.md)
- [logging](guide-logging.md)
- [dev plugin workflow and debugging](guide-dev-plugin-workflow-and-debugging.md)
- [publishing](guide-publishing.md)
- [api version migration](guide-api-version-migration.md)
- [pitfalls](guide-pitfalls.md)

## Changes within API 15 since the previous pin

Reviewed 2026-11-13 by reading the
[source comparison](https://github.com/goatcorp/Dalamud/compare/985a56f22bb45f42715f13ee62c0acc84b825085...43b34118bcf7213e575ec39517d033191b1acf17)
from the previous pin `985a56f` (15.0.3.5). Nothing was built or run. A runtime older than the
version named beside an item lacks it, so check `IDalamudPluginInterface.GetDalamudVersion()` or the
target SDK before relying on one.

- `IDalamudPluginInterface.Assembly` returns the plugin's `Assembly` (15.0.3.7).
- `AddonArgs.GetAddonPointer<T>()` returns the typed `AtkUnitBase*` (15.0.3.8). `AddonArgs.AddonName`
  is set when the args are created, and `InvalidAddonName` / `InvalidAddonNameSpan` are `[Obsolete]`.
- `IGamepadState.EnableGamepadNav` ignores `true` while plugin gamepad input is disabled in the
  Dalamud settings.
- After 15.0.3.9 (in tag `15.0.4.0`): `Dalamud.Utility` gains the API 16 read-only SeString helpers
  `ReadOnlySePayload(Span)Extensions` (`IsMacro`, `IsLink`, `IsLinkTerminator`, `TryReadLink`,
  `TryReadDalamudLink`), `ReadOnlySeString(Span)Extensions` (`ContainsText(ReadOnlySpan<char>)`,
  `ReplaceText`) and `SeStringBuilderExtensions` for `Lumina.Text.SeStringBuilder` (`PushDalamudLink`,
  `PushItemLink`, `PushPartyFinderLink`, `PushPartyFinderSearchConditionsLink`).
  `SeIconChar.Hydaelyn` replaces the misspelt `Hyadelyn`, which is `[Obsolete]`.
- UI layer, covered by `dalamudimgui`: `ImGuiHelpers.CenteredSeString`, `ImRaii.PushClipRect`, a fix
  to the disposable returned by `ImRaii.Header(label, ref visible, flags)`, and a plugin font atlas
  that is created on first access to `IUiBuilder.FontAtlas`.
