Dalamud topic: the per-service member tables for every `Dalamud.Plugin.Services` interface (key members and gotchas), the game object model members and the removed services. Verified against `goatcorp/Dalamud` `43b3411` (2026-11-10, `DalamudVersion 15.0.3.9`, API 15), `goatcorp/dalamud-docs` `3c7597d` (2026-11-13), `goatcorp/SamplePlugin` `b8477da` (2026-08-15) and `aers/FFXIVClientStructs` `80cb4a7` (2026-11-10). Reviewed 2026-11-13.

Paths such as `references/...` and `references/topic-*.md` are relative to the skill root (the directory that holds `SKILL.md`), not to this file. Part of the `dalamud` skill; items marked *(unverified)* are not confirmed from the pinned checkouts.

## Services

All in namespace `Dalamud.Plugin.Services` (45 interfaces) except `IDragDropManager` (`Dalamud.Interface.DragDrop`)
and `IDalamudAssetManager` (`Dalamud.Storage.Assets`), which move there in API 16. Full signatures:
`references/dalamud/Dalamud/Plugin/Services/<Name>.cs`.

### Game state

| Service | Key members | Gotchas |
|---|---|---|
| `IClientState` | `IsLoggedIn`, `TerritoryType`, `MapId`, `Instance`, `ClientLanguage`, `IsPvP`, `IsPvPExcludingDen`, `IsGPosing`, `IsClientIdle(out ConditionFlag)`. Events: `Login`, `Logout(int type, int code)`, `TerritoryChanged(uint)`, `MapIdChanged(uint)`, `InstanceChanged(uint)`, `ZoneInit(ZoneInitEventArgs)`, `ClassJobChanged(uint)`, `LevelChanged(uint classJobId, uint level)`, `EnterPvP`, `LeavePvP`, `CfPop(ContentFinderCondition)` | `LocalPlayer` and `LocalContentId` were **removed in API 15**: use `IObjectTable.LocalPlayer` / `IPlayerState.ContentId`. API 16 changes the ids to `RowRef<T>`. |
| `IPlayerState` | `IsLoaded`, `CharacterName`, `EntityId`, `ContentId`, `CurrentWorld`/`HomeWorld` (`RowRef<World>`), `ClassJob`, `Level`, `EffectiveLevel`, `IsLevelSynced`, `Race`, `Tribe`, `Sex`, `GrandCompany`, `HomeAetheryte`, `FavoriteAetherytes`, `IsMentor`, `IsNovice`, `GetAttribute(PlayerAttribute)`, `GetClassJobLevel(ClassJob)`, `GetClassJobExperience(ClassJob)`, `GetGrandCompanyRank(GrandCompany)` | Valid from login to logout, independent of the player object existing. Gate on `IsLoaded`. |
| `IObjectTable` | `LocalPlayer` (`IPlayerCharacter?`), indexer `[int]`, `Length`, `SearchById(ulong gameObjectId)`, `SearchByEntityId(uint)`, `PlayerObjects`, `CharacterManagerObjects`, `ClientObjects`, `EventObjects`, `StandObjects`, `ReactionEventObjects`, `CreateObjectReference(nint)`, enumerable of `IGameObject` | **Main thread only**: throws `InvalidOperationException` elsewhere. Do not keep `IGameObject` across frames; re-fetch, or check `IsValid()`. |
| `ITargetManager` | get/set `Target`, `FocusTarget`, `SoftTarget`, `MouseOverTarget`, `MouseOverNameplateTarget`, `PreviousTarget`, `GPoseTarget` | Setting a target is a game action; mind the publishing rules. |
| `ICondition` | `this[ConditionFlag]`, `Any()`, `Any(params)`, `AnyExcept(params)`, `OnlyAny(params)`, `EqualTo(params)`, `AsReadOnlySet()`, `event ConditionChange(ConditionFlag flag, bool value)` | Common flags: `InCombat`, `Casting`, `Mounted`, `BoundByDuty`, `BetweenAreas`, `BetweenAreas51`, `OccupiedInEvent`, `OccupiedInQuestEvent`, `OccupiedInCutSceneEvent`, `WatchingCutscene`, `Crafting`, `Gathering`, `InFlight`, `LoggingOut`. "Zoning" is conventionally `BetweenAreas` or `BetweenAreas51` *(community convention, unverified)*. |
| `IPartyList` | `Length`, `[int]` → `IPartyMember?` (`Name`, `ContentId`, `EntityId`, `GameObject`, `CurrentHP`, `MaxHP`, `ClassJob`, `Level`, `World`, `Territory`, `Statuses`), `PartyLeaderIndex`, `IsAlliance`, `PartyId` | `IPartyMember.ContentId` is `ulong` (API 15). |
| `IBuddyList` | `CompanionBuddy`, `PetBuddy`, `[int]` | |
| `IFateTable` | `[int]` → `IFate?` (`FateId`, `Name`, `State`, `Progress`, `TimeRemaining`, `Position`, `Radius`, `GameData`) | `Fate.IsValid` was removed in API 14. |
| `IAetheryteList` | `[int]` → `IAetheryteEntry?` (`AetheryteId`, `TerritoryId`, `GilCost`, `IsFavourite`, `Ward`, `Plot`, `AetheryteData`) | The aetherytes available in the Teleport window. |
| `IJobGauges` | `Get<T>() where T : JobGaugeBase` | Types are `ASTGauge`, `BLMGauge`, … `WHMGauge` in `Dalamud.Game.ClientState.JobGauge.Types`. |
| `IDutyState` | `IsDutyStarted`, `ContentFinderCondition` (`RowRef`), events `DutyStarted`, `DutyWiped`, `DutyRecommenced`, `DutyCompleted` (each `IDutyStateEventArgs`) | Args changed in API 15 (previously the TerritoryType row id). |
| `IUnlockState` | `IsMountUnlocked(Mount)`, `IsItemUnlocked(Item)`, `IsItemUnlockable(Item)`, `IsQuestCompleted(Quest)`, `IsEmoteUnlocked(Emote)`, `IsRecipeUnlocked(Recipe)`, `IsUnlockLinkUnlocked(uint)`, `IsRowRefUnlocked<T>(RowRef<T>)`, … , `event Unlock(RowRef)`, `IsAchievementListLoaded`, `IsTitleListLoaded` | Takes Lumina **rows**, not ids. The v14 notes call it experimental (`Dalamud001`); the checkout has no `[Experimental]` on it. |
| `IGameInventory` | `GetInventoryItems(GameInventoryType)` → `ReadOnlySpan<GameInventoryItem>`; events `InventoryChanged`, `InventoryChangedRaw`, `ItemAdded`, `ItemRemoved`, `ItemChanged`, `ItemMoved`, `ItemSplit`, `ItemMerged` and typed `…Explicit` variants | `GameInventoryItem.ItemId` includes HQ/collectable offsets; `BaseItemId` is the plain id (API 12). Also `Quantity`, `IsHq`, `InventorySlot`, `ContainerType`. |
| `IGameConfig` | `TryGet(SystemConfigOption\|UiConfigOption\|UiControlOption, out bool\|uint\|float\|string)`, `Set(option, value)`, sections `System`, `UiConfig`, `UiControl`, events `Changed`, `SystemChanged`, `UiConfigChanged`, `UiControlChanged` | `Set` changes the user's real game settings. |
| `IKeyState` | `this[VirtualKey]` get/set, `IsVirtualKeyValid`, `GetValidVirtualKeys()`, `ClearAll()` | Settable; writes change what the game reads. |
| `IGamepadState` | `Pressed/Repeat/Released/Raw(GamepadButtons)` → `float`, `LeftStick`, `RightStick`, `EnableGamepadNav` (get/set), `GamepadInputAddress` | Setting `EnableGamepadNav` to true is ignored while plugin gamepad input is disabled in Dalamud settings. |
| `IGameLifecycle` | `DalamudUnloadingToken`, `GameShuttingDownToken`, `LogoutToken` | Cancellation tokens to pass to long-running tasks. |

Object model (`Dalamud.Game.ClientState.Objects.Types` / `.SubKinds`): `IGameObject` (`Name` SeString, `GameObjectId`,
`EntityId`, `BaseId`, `OwnerId`, `ObjectIndex`, `ObjectKind`, `SubKind`, `Position`, `Rotation`, `HitboxRadius`,
`IsDead`, `IsTargetable`, `TargetObject`, `Address`, `IsValid()`; `DataId`, `YalmDistanceX/Z` are `[Obsolete]` →
`BaseId`, `CurrentDistance`) → `ICharacter` (`CurrentHp`, `MaxHp`, `CurrentMp`, `ClassJob`, `Level`, `Customize`
(`Span<byte>`), `CompanyTag`, `NameId`, `OnlineStatus`, `StatusFlags`, `CurrentMount`) → `IBattleChara` (`StatusList`,
`IsCasting`, `CastActionId`, `CastTargetObjectId`, `CurrentCastTime`, `BaseCastTime`, `TotalCastTime`) →
`IPlayerCharacter` (`CurrentWorld`, `HomeWorld`) / `IBattleNpc`. Also `INpc`, `IEventObj`. Pattern-match to narrow:
`if (obj is IBattleChara bc)`. `IStatus`: `StatusId`, `Param`, `RemainingTime`, `SourceId`, `SourceObject`,
`GameData`.

### Game UI and text

| Service | Key members | Gotchas |
|---|---|---|
| `IChatGui` | `Print(string\|SeString\|ReadOnlySpan<byte> message, string? messageTag = null, ushort? tagColor = null)`, `PrintError(…)`, `Print(XivChatEntry)`. Events `ChatMessage`, `CheckMessageHandled` (`IHandleableChatMessage`), `ChatMessageHandled`, `ChatMessageUnhandled` (`IChatMessage`), `LogMessage` (`ILogMessage`). `AddChatLinkHandler(uint commandId, Action<uint, SeString>)` → `DalamudLinkPayload`, `RemoveChatLinkHandler(uint)` / `()` | Prints to the local log only; it never sends chat to the server. Event signatures changed in API 15. |
| `IGameGui` | `GetAddonByName(string name, int index = 1)` → `AtkUnitBasePtr`, `GetAddonByName<T>(…)` → `T*`, `GetAgentById(int)`, `FindAgentInterface(string\|AtkUnitBasePtr)` → `AgentInterfacePtr`, `GetUIModule()`, `WorldToScreen(Vector3, out Vector2[, out bool inView])`, `ScreenToWorld`, `OpenMapWithMapLink(MapLinkPayload)` / `(uint territory, uint map, Vector3 worldPos)`, `GameUiHidden`, `HoveredItem`, `HoveredAction`; events `UiHideToggled`, `HoveredItemChanged`, `HoveredActionChanged`, `AgentUpdate` | Addon index is 1-based. Wrapper structs have `IsNull`, `Address`, implicit `nint`; cast `Address` to the ClientStructs pointer. |
| `IAddonLifecycle` | see `references/topic-addons.md` | |
| `IAddonEventManager` | see `references/topic-addons.md` | |
| `IAgentLifecycle` | `RegisterListener(AgentEvent, AgentId \| IEnumerable<AgentId>, handler)`, `UnregisterListener(…)`; `AgentArgs.Agent`, `.AgentId`, `GetAgentPointer<T>()`, `PreventOriginal()` | Events: `Pre/PostReceiveEvent`, `…ReceiveEventWithResult`, `Show`, `Hide`, `Update`, `GameEvent`, `LevelChange`, `ClassJobChange`. |
| `IContextMenu` | `event OnMenuOpened(IMenuOpenedArgs)` then `args.AddMenuItem(new MenuItem { Name, OnClicked, PrefixChar, PrefixColor, Priority, IsSubmenu })`; or persistent `AddMenuItem(ContextMenuType, IMenuItem)` / `RemoveMenuItem` | `args.Target` is `MenuTargetDefault` (`TargetName`, `TargetObjectId`, `TargetContentId`, `TargetHomeWorld`) or `MenuTargetInventory` (`TargetItem`). `ContextMenuType.Default` / `Inventory`. |
| `IDtrBar` | `Get(string title, SeString? text = null)` → `IDtrBarEntry` (`Text`, `Tooltip`, `Shown`, `MinimumWidth`, `OnClick` (`Action<DtrInteractionEvent>`), `Remove()`), `Remove(title)`, `Entries` | Server info bar. `Get` returns your existing entry for a title you own; a title owned by another plugin throws `ArgumentException`. |
| `IToastGui` | `ShowNormal(string\|SeString, ToastOptions?)`, `ShowQuest(…, QuestToastOptions?)`, `ShowError(…)`; events `Toast`, `QuestToast`, `ErrorToast` with `ref bool isHandled` | |
| `IFlyTextGui` | `AddFlyText(FlyTextKind kind, uint actorIndex, uint val1, uint val2, SeString text1, SeString text2, uint color, uint icon, uint damageTypeIcon)`, `event FlyTextCreated` (all `ref`) | |
| `INamePlateGui` | events `OnNamePlateUpdate`, `OnPostNamePlateUpdate`, `OnDataUpdate`, `OnPostDataUpdate` `(INamePlateUpdateContext, IReadOnlyList<INamePlateUpdateHandler>)`, `RequestRedraw()` | |
| `IPartyFinderGui` | `event ReceiveListing(IPartyFinderListing, IPartyFinderListingEventArgs)` | |
| `ITitleScreenMenu` | `AddEntry([ulong priority,] string text, ISharedImmediateTexture texture, Action onTriggered)`, `RemoveEntry` | |
| `INotificationManager` | `AddNotification(new Notification { Title, Content, Type = NotificationType.Success, InitialDuration })` → `IActiveNotification` (`DismissNow()`, `ExtendBy`, events `Click`, `Dismiss`, `DrawActions`) | ImGui toast, not the game toast. |
| `ISeStringEvaluator` | `Evaluate(ReadOnlySeString, Span<SeStringParameter> localParameters = default, ClientLanguage? = null)`, `EvaluateMacroString(string)`, `EvaluateFromAddon(uint addonId, …)`, `EvaluateFromLobby`, `EvaluateFromLogMessage`, `EvaluateActStr(ActionKind, uint)`, `EvaluateObjStr(ObjectKind, uint)` | Resolves macros the way the game would display them. |

### Data, interop, infrastructure

| Service | Key members | Gotchas |
|---|---|---|
| `IDataManager` | `GetExcelSheet<T>(ClientLanguage? language = null, string? name = null)`, `GetSubrowExcelSheet<T>(…)`, `Excel` (`ExcelModule`), `GameData`, `Language`, `GetFile(path)`, `GetFile<T>(path)`, `GetFileAsync<T>(path, ct)`, `FileExists(path)`, `HasModifiedGameDataFiles` | See `references/guide-lumina-and-idatamanager.md`. |
| `IGameInteropProvider` | `InitializeFromAttributes(object self)`, `HookFromAddress<T>(nint\|nuint\|void*, T detour)`, `HookFromSignature<T>(string, T)`, `HookFromSymbol<T>(module, export, T)`, `HookFromImport<T>(…)`, `HookFromFunctionPointerVariable<T>(nint, T)` | The `HookBackend` parameter is obsolete: never pass it (removed in API 16). |
| `ISigScanner` | `ScanText(sig)`, `TryScanText(sig, out nint)`, `GetStaticAddressFromSig(sig, int offset = 0)`, `TryGetStaticAddressFromSig`, `ScanData`, `ScanModule`, `ScanAllText`, `ResolveRelativeAddress`, `Module`, `SearchBase`, `TextSectionBase/Size`, `DataSectionBase/Size`, `RDataSectionBase/Size` | `Scan*` throw `KeyNotFoundException` on miss; use `Try*`. |
| `IFramework` | see `references/guide-threading.md` | |
| `ICommandManager` | see `references/guide-commands.md` | |
| `IPluginLog` | see `references/guide-logging.md` | |
| `IReliableFileStorage` | `WriteAllTextAsync(path, contents)`, `WriteAllBytesAsync`, `ReadAllTextAsync(path, bool forceBackup = false)`, `ReadAllBytesAsync`, `Exists(path)`, `MaxFileSizeBytes` | Every write is duplicated into a backup store; do not write large files. `SavePluginConfig` already goes through it. |
| `IMarketBoard` | events `OfferingsReceived`, `HistoryReceived`, `ItemPurchased`, `PurchaseRequested`, `TaxRatesReceived` | |
| `IConsole` | `AddCommand(name, description, Func<bool>)` (generic up to 5 args), `AddVariable<T>(name, description, defaultValue)`, `AddAlias`, `GetVariable<T>`, `SetVariable<T>`, `RemoveEntry`, `Prefix` | `[Experimental("Dalamud001")]`: needs `#pragma warning disable Dalamud001`. Console (`/xllog`) commands, not chat commands. |
| `ISelfTestRegistry` | `RegisterTestSteps(IEnumerable<ISelfTestStep>)` | Steps shown in Dalamud's self-test window. |
| `ITextureProvider`, `ITextureReadbackProvider`, `ITextureSubstitutionProvider`, `IDragDropManager`, `IDalamudAssetManager` | | UI/texture services → `dalamudimgui`. `ITextureProvider.GetFromGameIcon(new GameIconLookup(id))` returns `ISharedImmediateTexture`; never cache the wrap. |

Removed services: `IGameNetwork` (API 13; hook the handling functions instead), `ILibcFunction`, static `PluginLog`,
`SeStringManager`, `ChatHandlers` (API 9/10).
