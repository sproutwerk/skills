Dalamud topic: API version migration notes, API 9 through 16. Verified against `goatcorp/Dalamud` `43b3411` (2026-11-10, `DalamudVersion 15.0.3.9`, API 15), `goatcorp/dalamud-docs` `3c7597d` (2026-11-13), `goatcorp/SamplePlugin` `b8477da` (2026-08-15) and `aers/FFXIVClientStructs` `80cb4a7` (2026-11-10). Reviewed 2026-11-13.

Paths such as `references/...` and `references/topic-*.md` are relative to the skill root (the directory that holds `SKILL.md`), not to this file. Part of the `dalamud` skill; items marked *(unverified)* are not confirmed from the pinned checkouts.

## API version migration

Upgrade step for every bump: change the csproj header to `Dalamud.NET.Sdk/<N>.0.0`, rebuild, fix errors. ClientStructs
breaks separately each patch (`ffxiv.wildwolf.dev/docs/breaking/<patch>.html`).

- **9** (patch 6.5, .NET 7, 2023-10) — Services only via interfaces in `Dalamud.Plugin.Services` (`DataManager` →
  `IDataManager`). `IPluginLog` added. `Hook<T>.FromX` and `SignatureHelper.Initialize` → `IGameInteropProvider`.
  Texture functions moved to `ITextureProvider`. `IAddonLifecycle`, `IAddonEventManager` added. `IDalamudPlugin.Name`,
  `SeStringManager`, `ChatHandlers`, `Util.HttpClient`, opcodes on `IDataManager` removed. `UiBuilder.OpenMainUi`
  added. `GposeActive` → `IClientState.IsGPosing`. Dev plugins need a manifest.
- **10** (7.0, .NET 8, 2024-07) — `DalamudPluginInterface` → `IDalamudPluginInterface`, `UiBuilder` → `IUiBuilder`,
  object types interfaced (`PlayerCharacter` → `IPlayerCharacter`, `GameObject` → `IGameObject`, `BattleChara` →
  `IBattleChara`, …). `[RequiredVersion]` removed. `ITextureProvider` returns `ISharedImmediateTexture`. `IConsole`,
  `IMarketBoard`, `ITextureReadbackProvider` added. `ClientLanguage` → `Dalamud.Game.ClientLanguage`. Cast time
  renames on `IBattleChara`.
- **11** (7.1, .NET 8, 2024-11) — Lumina 5: rows are structs, `LazyRow` → `RowRef`, namespace
  `Lumina.Excel.GeneratedSheets` → `Lumina.Excel.Sheets` *(old namespace name not stated in the checkout docs)*,
  `GetSheetRaw` → `GetRawSheet`, subrow sheet types. `IClientState.ClassJobChanged`/`LevelChanged` added.
- **12** (7.2, .NET 9, 2025-03) — .NET 9. `IObjectTable` and some `IClientState` members throw off the main thread.
  Corrupted-state exceptions no longer catchable. `InteropGenerator.Runtime.dll` must be referenced (SDK does it).
  `PluginLoadReason` is `[Flags]`. `GameInventoryItem.ItemId` vs `BaseItemId`. Job gauge enums PascalCase.
  Notification/title-screen icons take `ISharedImmediateTexture`. `ISeStringEvaluator` added. Dalamud.NET.Sdk becomes
  the supported setup.
- **13** (7.3, .NET 9, 2025-08) — ImGui bindings: `ImGuiNET` → `Dalamud.Bindings.ImGui` (also ImPlot, ImGuizmo);
  `IDalamudTextureWrap.ImGuiHandle` → `Handle`; enum names PascalCase. `IGameNetwork` removed. `AddonArgs.Addon` is
  `AtkUnitBasePtr`; `IGameGui.GetAddonByName` → `AtkUnitBasePtr`, `GetAgentById` / `FindAgentInterface` →
  `AgentInterfacePtr`, `GetUIModule` → `UIModulePtr`. `AddChatLinkHandler`/`RemoveChatLinkHandler` moved from the
  plugin interface to `IChatGui`. `IDtrBarEntry.OnClick` takes an argument. `ItemKind` → `Dalamud.Utility`.
  `ReadOnlySeString.ToString()` returns text.
- **14** (7.4, .NET 10, 2025-12; `v14.md` still shows v13's release date) — .NET 10 / C# 14. `ISigScanner`,
  `ITargetManager`, `ISelfTestRegistry` moved into `Dalamud.Plugin.Services`. New `IPlayerState`, `IUnlockState`,
  `IReliableFileStorage`. `IClientState.LocalPlayer` / `LocalContentId` obsoleted. AddonLifecycle rework (vtable
  based; new Open/Close/Show/Hide/Move/MouseOver/MouseOut/Focus events; `AddonDrawArgs`, `AddonFinalizeArgs`,
  `AddonUpdateArgs` removed → `AddonArgs`; `AddonReceiveEventArgs.Data` → `AtkEventData`). List services return struct
  enumerators. SharpDX removed → TerraFX. `Util` version helpers → `GetDalamudVersion()`. `targets/` removed. csproj
  manifest properties. `TestingDalamudApiLevel` required for testing builds.
- **15, current stable** (7.5, .NET 10, 2026-04-29) — `IAsyncDalamudPlugin`. `IChatGui` events take
  `IChatMessage`-family objects; `XivChatType` holds real `LogKind` ids. `IClientState.LocalPlayer` / `LocalContentId`
  removed. `ZoneInitEventArgs` uses `RowRef`s and `IReadOnlyList<FestivalEntry> ActiveFestivals`. `IDutyState` events
  take `IDutyStateEventArgs`. `PreventOriginal()` on Addon/Agent lifecycle args. Enums re-synced from ClientStructs
  (32 in `EnumCloneMap.txt`; member names may differ); `HoverActionKind` → `DetailKind`. `ICharacter.Customize` is
  `Span<byte>`. `IPartyMember.ContentId` is `ulong`. 32-bit members removed from `ISigScanner`. MinHook and FASM
  removed. ImRaii end objects are `ref struct`s. Zip manifest must be accurate.
- **16, upcoming, not finalized** (8.0, .NET 10 with a pending move to .NET 11 / C# 15, date TBA) — `SeString` and all `Payload` classes removed;
  every API takes `ReadOnlySeString`/`ReadOnlySeStringSpan`; build with `RentedSeStringBuilder`. `Lumina.Excel.Sheets`
  → `Dalamud.Excel.Sheets` (or reference the Lumina.Excel package yourself). `On` prefix dropped from events and
  delegates (API 15 still has e.g. `IContextMenu.OnMenuOpened`, `INamePlateGui.OnNamePlateUpdate`). `EventHandler<T>`
  events become `Action<T>`. `IClientState` ids become `RowRef<T>` (compare `.RowId`). `IGameGui.HoveredItem` becomes
  `uint`. `HookBackend` parameter and the `AsmHook` `string[]` constructor removed. `Dalamud.Game.ClientLanguage` →
  `Dalamud.Common.ClientLanguage`; `Dalamud.Game.Network.Structures` → `Dalamud.Game.Marketboard.Network.Structures`.
  `IClientState.CfPop` fires on the framework thread instead of inside `Task.Run`, so move slow work off it yourself.
  New `MapCoordinate` / `WorldCoordinate` structs: `MapUtil.GetMapCoordinates(IGameObject, bool)` is removed for
  `IGameObject.GetMapCoordinate()` / `GetWorldCoordinate()`, and `IGameGui.OpenMapWithMapLink` takes the new structs.
  SDK 16: `DalamudPackager` and `DotNet.ReproducibleBuilds` become implicit package references (delete them from the
  csproj), `RuntimeIdentifier` defaults to `win-x64`, `Use_Dalamud_Lumina_Excel` is renamed
  `Use_Dalamud_Excel_Sheets`, a `Dalamud.Common` reference is added, and `<Warn_Outdated_Sdk>true</Warn_Outdated_Sdk>`
  opts in to outdated-SDK warnings. Sources: `references/docs/docs/versions/v16/index.md` and the SeString porting
  guide `references/docs/docs/versions/v16/sestring.md`.
  To stay forward compatible on 15: keep SeString use in one place, never pass `HookBackend`.
