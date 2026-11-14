# AddonMaster, AtkReader and addon helpers

Covers: `ECommons/UIHelpers/` (`AddonMaster` and its 117 nested addon wrappers,
`AddonMasterBase<T>`,
`AddonMasterCombined`, `AtkReader` and its readers, `AddonFinder`, `AtkEventDataBuilder`) and
`ECommons/GenericHelpers/AddonHelpers.cs` (`TryGetAddonByName`, `TryGetAddonMaster`, `IsAddonReady`,
`IsReady`). Verified against ECommons `bc38dc7` (2026-11-08, v3.2.1.26, Dalamud
API 15); nothing here was compiled. Paths are relative to the skill root.

Sources: `references/ecommons/ECommons/UIHelpers/**` and
`references/ecommons/ECommons/GenericHelpers/AddonHelpers.cs`. Note two upstream file names contain
shell-hostile characters: `AddonMasterImplementations/!AddonMasterBase.cs`, `.../!AddonMaster.cs`
and
`AtkReaderImplementations/ReaderGCArmy;MemberList.cs` (quote them).

Pointers: `IAddonLifecycle`, `IGameGui`, `AtkUnitBasePtr` -> `dalamud` skill. `AtkUnitBase`,
`AtkValue`, `AddonSelectYesno` and other structs -> `clientstructs` skill. `Callback.Fire` and the
`ClickAddonButton` extensions live in `ECommons.Automation` / `ECommons.Automation.UIInput` (the
automation topic of this skill).

## Namespaces

```csharp
using ECommons;                                        // GenericHelpers (static partial class)
using static ECommons.GenericHelpers;                  // what AutoRetainer does: bare TryGetAddonMaster(...)
using ECommons.UIHelpers;                              // AtkReader, AddonFinder, AtkEventDataBuilder
using ECommons.UIHelpers.AddonMasterImplementations;   // AddonMaster, AddonMasterBase<T>, IAddonMasterBase
using ECommons.UIHelpers.AddonMasterCombinedImplementations; // AddonMasterCombined.ICombinedInventory
using ECommons.UIHelpers.AtkReaderImplementations;     // ReaderXxx
using FFXIVClientStructs.FFXIV.Component.GUI;          // AtkUnitBase, AtkValue, AtkComponentButton
using FFXIVClientStructs.FFXIV.Client.UI;              // AddonSelectYesno, AddonSelectString, ...
```

`GenericHelpers` is `public static unsafe partial class GenericHelpers` in namespace `ECommons` (not
`ECommons.GenericHelpers`; the folder name is not a namespace).

Requirements: nothing in this area needs a `Module` flag, installs a hook, or is touched by
`ECommonsMain.Dispose()`. `Svc` must be initialised (`ECommonsMain.Init`) because lookups go through
`Svc.GameGui`, and some wrappers use `Svc.Data` / `Svc.Log` / `PluginLog`. Everything reads or
writes
game UI memory, so call it from the framework thread (the source has no thread check). All code that
touches `Addon`, `Base` or any `*Button` pointer needs an `unsafe` context.

## AddonHelpers.cs (`ECommons.GenericHelpers`)

| Signature | Notes |
|---|---|
| `public static bool TryGetAddonByName<T>(string Addon, out T* AddonPtr) where T : unmanaged` | See below |
| `public static bool TryGetAddonMaster<T>(out T addonMaster) where T : IAddonMasterBase` | Name from type |
| `public static bool TryGetAddonMaster<T>(string addon, out T addonMaster) where T : IAddonMasterBase` | Explicit name |
| `public static bool IsAddonReady(AtkUnitBase* Addon)` | Not null-safe |
| `public static bool IsReady(this ref AtkUnitBase Addon)` | Extension, same test |
| `public static bool IsAddonReady(AtkComponentNode* Addon)` | Visible and component uld `Loaded` |
| `public static bool IsScreenReady()` | False while `NowLoading`, `FadeMiddle` or `FadeBack` is visible |
| `public static unsafe AtkResNode* GetNodeByIDChain(AtkResNode* node, params int[] ids)` | Walks node ids from root |
| `public static unsafe AtkResNode* GetRootNode(AtkResNode* node)` | Follows `ParentNode` recursively |
| `public static bool IsString(this AtkValueType type)` | `String`, `String8`, `WideString`, `ManagedString` |
| `public static bool IsString(this AtkValue value)` | Same types and `value.String.HasValue` |
| `public static bool IsSelectItemEnabled(AtkTextNode* textNodePtr)` | `[Obsolete("Incompatible with UI mods, use other methods")]`, colour based |

### TryGetAddonByName

```csharp
public static bool TryGetAddonByName<T>(string Addon, out T* AddonPtr) where T : unmanaged
{
    var a = Svc.GameGui.GetAddonByName(Addon, 1);
    if(a == IntPtr.Zero) { AddonPtr = null; return false; }
    else { AddonPtr = (T*)a.Address; return true; }
}
```

- Always index 1 (first instance). There is no index parameter; for further instances use
  `Svc.GameGui.GetAddonByName(name, i)` yourself or `AddonFinder`.
- No visibility, load-state or readiness check. `true` means only "pointer is non-null". An addon
  can
  exist while hidden or half-loaded, so pair it with `IsAddonReady`.
- `T` is an unchecked cast. `TryGetAddonByName<AddonSelectString>("SelectYesno", ...)` compiles and
  returns a wrongly typed pointer.

### IsAddonReady / IsReady: exact code and relation to ClientStructs `IsReady`

```csharp
[MethodImpl(MethodImplOptions.AggressiveInlining)]
public static bool IsAddonReady(AtkUnitBase* Addon)
{
    return Addon->IsVisible && Addon->UldManager.LoadedState == AtkLoadState.Loaded && Addon->IsFullyLoaded();
}

public static bool IsReady(this ref AtkUnitBase Addon)
{
    return Addon.IsVisible && Addon.UldManager.LoadedState == AtkLoadState.Loaded && Addon.IsFullyLoaded();
}
```

Verified: the ECommons test is `IsVisible && UldManager.LoadedState == AtkLoadState.Loaded &&
IsFullyLoaded()`. It does not read the ClientStructs `AtkUnitBase.IsReady` property at all.

- ClientStructs `AtkUnitBase.IsReady` is a generated bitfield property
  (`[BitField<bool>(nameof(IsReady), 0)]`
  on `Flags1A1`, documented "Gets a value indicating whether OnSetup was called"). Dalamud's
  `AtkUnitBasePtr.IsReady` returns that property. `IsFullyLoaded()` is virtual function 66.
- The two differ in practice: a hidden addon that has finished setup is CS-ready but not
  ECommons-ready (visibility is part of the ECommons test).
- Name collision: on an `AtkUnitBase*`, `addon->IsReady` (no parentheses) is the ClientStructs
  property, `addon->IsReady()` (with parentheses) is the ECommons extension. AutoRetainer uses the
  latter (`addon != null && addon->IsReady()`).
- Neither method null-checks. `IsAddonReady(null)` dereferences null.
- There is no `IsAddonReady(this ...)` extension on `AtkUnitBase` in this checkout. YesAlready
  (older NuGet) calls `addon->AtkUnitBase.IsAddonReady()`; where that extension comes from was not
  found in the checkouts *(unverified)*. Use `IsAddonReady(ptr)` or `ptr->IsReady()`.

### TryGetAddonMaster

```csharp
public static bool TryGetAddonMaster<T>(out T addonMaster) where T : IAddonMasterBase
{
    if(TryGetAddonByName<AtkUnitBase>(typeof(T).Name.Split(".")[^1], out var ptr))
    {
        addonMaster = (T)Activator.CreateInstance(typeof(T), (nint)ptr);
        return true;
    }
    addonMaster = default;
    return false;
}
```

- The addon name is `typeof(T).Name` (for a nested class that is the simple class name, e.g.
  `SelectYesno`, `_TitleMenu`). Hence the rule that every wrapper class is named exactly like the
  game addon.
- The `string addon` overload is identical except that it looks up the given name. Use it when the
  wrapper is reused for a differently named addon.
- Instance is built by reflection through the `nint` constructor. Source comment: "Slower than
  `TryGetAddonByName`". Avoid per-frame use over many types; `new AddonMaster.X(ptr)` is direct.
- Returns `true` as soon as the addon exists. Always test `m.IsAddonReady` afterwards.
- Do not pass the obsolete `SelectYesnoMaster` to the parameterless-name overload: its type name
  resolves to a non-existent addon `SelectYesnoMaster`.

## AddonMaster structure

`references/ecommons/ECommons/UIHelpers/AddonMasterImplementations/!AddonMaster.cs`:
`public partial class AddonMaster` with a private constructor (pure container). Each addon is a
nested class in its own file. Upstream rules for adding one (from the doc comment): inherit
`AddonMasterBase`, name the class exactly as the addon, nest it in partial `AddonMaster`, implement
both `nint` and `void*` constructors, prefer entry objects over raw indexes, and validate input.

### AddonMasterBase<T>

`public abstract unsafe class AddonMasterBase<T> : IAddonMasterBase where T : unmanaged`.
`public abstract unsafe class AddonMasterBase : AddonMasterBase<AtkUnitBase>` is the non-generic
shortcut. `T` is the ClientStructs addon struct (`AddonSelectYesno`) or `AtkUnitBase` when no struct
is used.

| Member | Notes |
|---|---|
| `protected AddonMasterBase(nint addon)` / `protected AddonMasterBase(void* addon)` | Stores `(T*)addon`. No null check, no name check |
| `public abstract string AddonDescription { get; }` | Human-readable text for settings UI |
| `public T* Addon { get; }` | Typed pointer |
| `public AtkUnitBase* Base` | `(AtkUnitBase*)Addon` |
| `public bool IsVisible` | `Base->IsVisible` |
| `public bool IsAddonReady` | `GenericHelpers.IsAddonReady(Base)` |
| `public bool HasFocus` | `AtkStage.Instance()->GetFocus()` root node equals the root node of any unit in `RaptureAtkUnitManager.Instance()->FocusedUnitsList` (does not compare against this addon) |
| `public bool IsAddonInFocusList` | This addon is in `FocusedUnitsList` |
| `public bool IsAddonOnlyFocusListEntry` | List count is 1 and entry 0 is this addon |
| `public bool IsAddonFocused` | `[Obsolete]`: "For the intended functionality please use HasFocus. For the same functionality please use IsAddonInFocusList." |
| `protected bool ClickButtonIfEnabled(AtkComponentButton* button, bool respectHoldButtons = false)` | Clicks if `button->IsEnabled` and its node is visible, and (when `respectHoldButtons`) the component type is not `ComponentType.HoldButton`. Uses `button->ClickAddonButton(Base)` |
| `protected bool ClickButtonIfEnabled(AtkComponentRadioButton* button)` | Same checks, `ClickRadioButton(Base)` |
| `protected bool ClickCheckboxIfEnabled(AtkComponentCheckBox* checkbox)` | Same checks, `ClickCheckBox(Base)` then `SetChecked(true)` |
| `protected AtkEvent CreateAtkEvent(byte flags = 0)` | `Listener = Base`, `Target = &AtkStage.Instance()->AtkEventTarget`, `State.StateFlags = flags` |
| `protected AtkEventDataBuilder CreateAtkEventData()` | `new AtkEventDataBuilder()` |

There is no `HasAddon` member. The click helpers are `protected`; consumers call the public methods
of the concrete wrapper or use the public `*Button` pointers with the `ClickAddonButton` extension.
Most wrapper methods are declared `void` and discard the `bool` from `ClickButtonIfEnabled`, so a
click on a disabled or hidden button is a silent no-op. Check `XButton->IsEnabled` first when you
need to know.

`IAddonMasterBase` (same file): `AddonDescription`, `Base`, `HasFocus`, `IsAddonInFocusList`,
`IsAddonOnlyFocusListEntry`, `IsAddonReady`, `IsVisible`. It has no `Addon` member.

### Construction and IAddonLifecycle

Every wrapper has exactly two public constructors: `(nint addon)` and `(void* addon)`.
Dalamud API 15 `AddonArgs.Addon` is `AtkUnitBasePtr` (`Dalamud.Game.NativeWrapper`), which has
`Address` (`nint`) and `implicit operator nint`. So `new AddonMaster.SelectYesno(args.Addon)` binds
to the `nint` constructor through the implicit conversion; ECommons itself relies on this in
`AddonFinder` (`new(addon)` with the `AtkUnitBasePtr` from `GetAddonByName`). The explicit form
`args.Addon.Address` is unambiguous.

```csharp
Svc.AddonLifecycle.RegisterListener(AddonEvent.PostSetup, "SelectYesno", OnYesno);

private unsafe void OnYesno(AddonEvent type, AddonArgs args)
{
    var m = new AddonMaster.SelectYesno(args.Addon.Address);
    if(m.Text.Contains("expected prompt")) m.Yes();
}
// Svc.AddonLifecycle.UnregisterListener(OnYesno); on dispose
```

Forms found in consumers:

- AutoRetainer, polling inside a TaskManager step (the dominant pattern):
  `if(TryGetAddonMaster<AddonMaster.SelectYesno>(out var m)) { if(m.IsAddonReady &&
  m.Text.ContainsAny(...)) { if(EzThrottler.Throttle("CeruleumYesNo")) m.Yes(); } }`
- AutoRetainer with a raw pointer: `new AddonMaster.SelectYesno(addon)` (`AtkUnitBase*`, binds to
  `void*`), `new AddonMaster.SelectYesno((nint)yesno).Yes()`, `new AddonMaster.RetainerList(addon)`,
  `new AddonMaster.SelectString(sel).Entries[index].Select()`.
- PandorasBox: `new AddonMaster.SelectYesno((IntPtr)addon).Yes()`,
  `new AddonMaster.Request((IntPtr)addon).HandOver()`,
  `new AddonMaster.ContextMenu(contextMenu).Entries.First().Select()`; lifecycle handlers cast
  `(AtkUnitBase*)args.Addon.Address`.
- YesAlready at the pinned commit does not reference `AddonMaster` in its `.cs` files.

`IsAddonReady` includes `IsVisible`, so it may still be `false` inside a `PostSetup` handler
*(unverified: depends on the addon)*. AutoRetainer polls for `IsAddonReady` instead of acting in
lifecycle handlers.

## Most used wrappers

### SelectYesno (`AddonMasterBase<AddonSelectYesno>`)

| Member | Notes |
|---|---|
| `bool RespectHoldButtons { get; set; } = false` | Documented "Whether to not click the button if it's hold button". Not read anywhere at this commit: `Yes()` does not pass it to `ClickButtonIfEnabled` |
| `bool RespectDisabledButtons { get; set; } = false` | Controls force-enable in `Yes()` |
| `SeString SeString` | From `Addon->PromptText->NodeText` |
| `string Text` | `SeString.GetText()` |
| `SeString SeStringNullTerminated` | From `AtkValues[0]` |
| `string TextLegacy` | `[Obsolete("Please use Text")]`. Text payloads of `SeStringNullTerminated` joined, `\n` replaced by space, trimmed |
| `int ButtonsVisible` | Count of values in `AtkValues[1..3]` typed `String`, `ConstString`, `ManagedString` or `WideString`. Packages up to 3.2.1.22 test `String8` in place of `ConstString` |
| `AtkComponentButton* ThirdButton` | `GetComponentButtonById(14)` |
| `void Yes()`, `void No()`, `void Third()` | |

```csharp
public void Yes()
{
    if(!RespectDisabledButtons)
    {
        if(Addon->YesButton != null && !Addon->YesButton->IsEnabled)
        {
            Svc.Log.Debug($"{nameof(AddonSelectYesno)}: Force enabling yes button");
            var flagsPtr = (ushort*)&Addon->YesButton->AtkComponentBase.OwnerNode->AtkResNode.NodeFlags;
            *flagsPtr ^= 1 << 5;
        }
    }
    ClickButtonIfEnabled(Addon->YesButton);
}
```

Verified: by default (`RespectDisabledButtons == false`) `Yes()` force-enables a disabled Yes button
by toggling bit 5 (`NodeFlags.Enabled`, `0x20`) of the button's owner node, then clicks. This
defeats
the countdown on timed confirmations. It is a property, not a parameter: write
`new AddonMaster.SelectYesno(ptr) { RespectDisabledButtons = true }.Yes()` to leave a disabled
button
alone. The click still requires the button node to be visible. `No()` and `Third()` never
force-enable. Hold-to-confirm buttons are clicked regardless of `RespectHoldButtons`.

`public unsafe class SelectYesnoMaster : AddonMaster.SelectYesno` (top level, same namespace) is
`[Obsolete("Please use AddonMaster.SelectYesno")]`.

### SelectString / SelectIconString

`SelectString : AddonMasterBase<AddonSelectString>`, `SelectIconString :
AddonMasterBase<AddonSelectIconString>`.

| Member | Notes |
|---|---|
| `int EntryCount` | `Addon->PopupMenu.PopupMenu.EntryCount` |
| `SeString SeString`, `string Text` | `SelectString` only: prompt, text node id 2 |
| `AtkComponentList* ListComponent` | `GetComponentListById(3)` |
| `List<Pointer<AtkComponentListItemRenderer>> ListItems` | Non-null item renderers |
| `Entry[] Entries` | Rebuilt on every access |
| `Entry.Index`, `Entry.TextNode`, `Entry.SeString`, `Entry.Text` | `Entry` is a struct; text from `PopupMenu.PopupMenu.EntryNames[Index]` |
| `void Entry.Select()` | `Callback.Fire((AtkUnitBase*)Addon, true, Index)` |
| `Entry1()` .. `Entry12()` | `public` on `SelectIconString`, `private` on `SelectString` |

`Entry.Select()` performs no enabled check. Match on `Text`, not on position:

```csharp
if(TryGetAddonMaster<AddonMaster.SelectString>(out var m) && m.IsAddonReady)
{
    foreach(var e in m.Entries)
        if(e.Text.Contains("Entrust", StringComparison.OrdinalIgnoreCase)) { e.Select(); break; }
}
```

### Talk (`AddonMasterBase<AddonTalk>`)

Only `void Click()`: builds an `AtkEvent` with `CreateAtkEvent(132)` and zeroed `AtkEventData`, then
sends `MouseDown`, `MouseClick`, `MouseUp` (event param 0) through `Base->ReceiveEvent`. No text
accessor; read the text from `Addon` (`AddonTalk`) yourself.

### ContextMenu (`AddonMasterBase<AddonContextMenu>`)

`int EntriesCount` (`AtkValues[0].UInt`), `ListComponent` (list id 2), `ListItems`, `Entry[]
Entries`.
`Entry` (readonly struct): `Index` (callback index), `ListIndex` (`Index + 8`, the `AtkValues`
index), `IsNativeEntry`, `TextNode`, `SeString`, `Text`, `Enabled`, `bool Select()`.
`Select()` fires `Callback.Fire(addon, true, 0, Index, 0)` only when `IsNativeEntry && Enabled` and
returns whether it fired. `IsNativeEntry` is true for a `ManagedString` value with exactly one
payload; entries added by Dalamud plugins are rejected (source comment: they have callback index
-1).

### Request (`AddonMasterBase<AddonRequest>`)

`HandOverButton` (14), `CancelButton` (15), `void HandOver()`, `void Cancel()`,
`bool IsHandOverEnabled`, `bool IsFilled`. `IsFilled` is broken at this commit: its loop is
`for(var i = 3u; i >= 7; i++)`, which never runs, so it always returns `true`. Use
`IsHandOverEnabled`.

### RetainerList (`AddonMasterBase<AddonRetainerList>`)

`Entry[] Retainers` built from `ReaderRetainerList`. `Entry` derives from
`ReaderRetainerList.Retainer`, so it has `Name`, `Level`, `Inventory`, `Gil`, `IsActive` plus
`int Index` and `bool Select()`. `Select()` fires
`Callback.Fire(Base, true, 2, (uint)index, Callback.ZeroAtkValue, Callback.ZeroAtkValue)` only when
`IsActive`, and returns whether it fired. `Retainer1()`..`Retainer10()` are `private`. The reader
stops at the first empty slot, so `Retainers.Length` is the number of listed retainers.

### Shop family

| Class | Members |
|---|---|
| `Shop` | `uint NumEntries` (`AtkValues[2]`), `ShopItemInfo[] ShopItems`; `ShopItemInfo`: fields `ItemId`, `CostAmount`, `void Select(int amount = 1)` = `Callback.Fire(Base, true, 0, index, amount)`. Nested `CostInfo` (`itemId`, `cost`) is unused |
| `ShopExchangeCurrency` | `uint CurrencyAmount` (`AtkValues[86]`), `uint NumEntries` (`AtkValues[4]`), `uint CurrencyId` (item whose `Icon` equals `AtkValues[87]`, scans the whole `Item` sheet on every read), `ShopItemInfo[] BasicShopItems`; `ShopItemInfo`: fields `ItemId`, `Index`, `CostAmount`, `void Select(int amount = 1)` = `Callback.Fire(Base, true, 0, Index, amount)` |
| `ShopExchangeCurrencyDialog` | `ExchangeButton` (17), `CancelButton` (18), `Exchange()`, `Cancel()` |
| `ShopExchangeItemDialog` | `ExchangeButton` (18), `CancelButton` (19), `Exchange()`, `Cancel()` |
| `InclusionShop` | `CurrencyAmount` (`AtkValues[297]`), `NumEntries` (`AtkValues[298]`), `ShopItemInfo[] ShopItems`; item fields `ItemId`, `CurrencyId`, `Cost`, `Select(int amount = 1)` = callback `14, index, amount` |

There is no `AddonMaster.ShopExchangeItem` class at this commit (only `ShopExchangeItemDialog`).
`Select` performs no affordability or range check. Rows with item id 0 are skipped, so the array
position is not the shop index; the index used for the callback is stored inside the entry.

### ContentsFinderConfirm, InputNumeric

- `ContentsFinderConfirm : AddonMasterBase<AddonContentsFinderConfirm>`: `Commence()`, `Withdraw()`,
  `Wait()` (buttons from the struct).
- `InputNumeric`: `uint Min` (`AtkValues[2]`), `uint Max` (`AtkValues[3]`), `OkButton` (4),
  `CancelButton` (5), `void Ok(int value)` = `Callback.Fire(Addon, true, value)`, `void Ok()`,
  `void Cancel()`. `Ok(int)` does not clamp or validate against `Min`/`Max`.

### Title and character select (underscore addons)

Class names keep the leading underscore of the addon name, so
`TryGetAddonMaster<AddonMaster._TitleMenu>(out var m)` resolves `_TitleMenu`.

- `_TitleMenu`: `bool IsReady` (`IsScreenReady() && IsAddonReady(Base)`, node list count > 3,
  `NodeList[7]` visible, node 3 alpha `0xFF`, and neither `TitleDCWorldMap` nor `TitleConnect`
  exists); buttons `StartButton` (4), `DataCenterButton` (5), `MoviesAndTitlesButton` (6),
  `OptionsButton` (7), `LicenseButton` (8), `ExitButton` (9) with `Start()`, `DataCenter()`,
  `MoviesAndTitles()`, `Options()`, `License()`, `Exit()`.
- `TitleDCWorldMap`: `static int[] PublicDC`, `bool Select(int dc)` (validates against `PublicDC`
  and
  the `WorldDCGroupType` sheet, then calls `UncheckedSelect`), `void UncheckedSelect(int dc)` =
  `Callback.Fire(Base, true, 17, dc)`.
- `_CharaSelectWorldServer`: `World[] Worlds` (names from string array 1, up to 16); `World` has
  fields `Master`, `Index`, `Name` and `void Select()` = `Callback.Fire(Master.Base, true, 25, 0,
  Index)`.
- `_CharaSelectListMenu`: `WorldButton` (4), `NewCharacterButton` (5),
  `BackUpClientSettingsButton` (6), `World()`, `NewCharacter()`, `BackUpClientSettings()`,
  `bool TemporarilyLocked` (hard-coded `false`), `void SelectWorld()`, `Character[] Characters`
  (from `AgentLobby.Instance()->LobbyData.CharaSelectEntries`). `Character`: `Index`, `Entry`
  (`CharaSelectCharacterEntry*`), `Name`, `HomeWorld`, `CurrentWorld`, `IsVisitingAnotherDC`,
  `CanLoginNormally`, `IsSelected`, `void Select()`, `void Login()`, `void OpenContextMenu()`.
- `Dialogue`: title-screen error window, `OkButton` (4), `Ok()`.

## Complete class table

One row per nested class, 117 in total. The addon name equals the class name in every case (upstream
rule, and what `TryGetAddonMaster<T>(out T)` relies on); the game addon names themselves are not
checkable from the ECommons source. `T` is the generic argument (`-` means `AtkUnitBase`). `XButton`
(n) means `public AtkComponentButton* XButton => GetComponentButtonById(n)`. Methods are `void`
unless a return type is shown. Spellings such as `GcArmyMenberProfile`, `WKSMissionInfomation`,
`StellerMissions`, `Infrastructor` are as in source.

| Class | T | Public members |
|---|---|---|
| `AirShipExploration` | - | `DeployButton` (85), `Deploy()` |
| `AirShipExplorationDetail` | - | `DeployButton` (26), `CancelButton` (27), `Deploy()`, `Cancel()` |
| `AirShipExplorationResult` | - | `RedeployButton` (47), `FinalizeReportButton` (48), `Redeploy()`, `FinalizeReport()` |
| `BankaCraftworksSupply` | - | `uint CollectableItemId`, `DeliverButton` (71), `CancelButton` (72), `Deliver()`, `Cancel()`, `int RequestedItemNumberAvailable`, `List<int> SlotsFilled`, `int? FirstUnfilledSlot`, `bool? TryHandOver(int slot)` |
| `Bank` | - | `ProceedButton` (35), `CancelButton` (36), `Proceed()`, `Cancel()`, `AtkComponentNumericInput* InputNumeric` (component node 32) |
| `BannerEditor` | `AddonBannerEditor` | `Save()`, `Close()` |
| `BannerList` | - | `EditButton` (2), `DisplayHelpButton` (8), `UseAsInstantPortraitButton` (34), `NumPortraits`, `SelectedPortrait`, `CharacterOption1IconId`, `CharacterOption2IconId`, `BackgroundIconId`, `FrameIconId`, `AccentIconId`, `Portraits[] Portrait`, `Edit()`, `DisplayHelp()`, `UseAsInstantPortrait()` |
| `BannerMIP` | - | `OkButton` (20), `CancelButton` (21), `SettingsButton` (22), `Ok()`, `Cancel()`, `Settings()` |
| `BannerPreview` | - | `UpdateButton` (8), `CancelButton` (9), `AtkComponentCheckBox* DoNotDisplayAgainCheckbox`, `Update()`, `Cancel()` |
| `_CharaSelectListMenu` | - | See above |
| `_CharaSelectWorldServer` | - | See above |
| `CollectablesShop` | - | `TradeButton` (51), eleven `AtkComponentRadioButton*` job tabs (`CarpenterButton` .. `FisherButton`), `Trade()`, `bool SelectDiscipleTab(Job job)`, `bool SelectDiscipleTab(uint classjob)` (throws `ArgumentOutOfRangeException` outside 8..18) |
| `ColorantColoring` | - | `uint ItemId`, `int ItemIconId`, `string ItemName`, `ApplyButton` (68), `SelectAnotherButton` (69), `Apply()`, `SelectAnother()` |
| `CompanyCraftRecipeNoteBook` | - | `BeginButton` (34), `Begin()` |
| `CompanyCraftSupply` | - | `CloseButton` (41), `Close()` |
| `ContentsFinderConfirm` | `AddonContentsFinderConfirm` | `Commence()`, `Withdraw()`, `Wait()` |
| `ContentsFinderSetting` | - | `ConfirmButton` (29), `CloseButton` (30), `bool Confirm()`, `Close()` |
| `ContentsFinderStatus` | - | `WithdrawButton` (40), `Withdraw()` |
| `ContextMenu` | `AddonContextMenu` | See above |
| `DawnStory` | - | `ReaderDawnStory Reader`, `IEnumerable<Entry> Entries`; `Entry`: `ReaderEntry`, `ReaderEntryName`, `SeString Name`, `uint Status`, `int Index`, `Select()` (callback `12, Callback`) |
| `Dialogue` | - | `OkButton` (4), `Ok()` |
| `DifficultySelectYesNo` | - | `ProceedButton` (12), `LeaveButton` (13), radio `NormalButton`, `EasyButton`, `VeryEasyButton`, `Proceed()`, `Leave()`, `SetDifficultyNormal()`, `SetDifficultyEasy()`, `SetDifficultyVeryEasy()` (source: "TODO: needs work") |
| `FreeCompanyChest` | - | `List<Button> ItemsButtons` (visible tabs of 5); `Button`: `AtkComponentRadioButton* Handle`, `bool Accessible`, `Select()` |
| `FreeCompanyCreditShop` | - | `uint FreeCompanyRank`, `bool Unk01`, `uint CompanyCredits`, `bool Unk05`, `uint ItemCount`, `Item[] Items`, `Buy(uint itemId, int quantity)`; `Item`: `Index`, `ItemName`, `ItemId`, `IconId`, `Rank`, `QuantityInInventory`, `MaxPurchaseSize`, `Price`, `Buy(int quantity)` (checks max size and credits, logs error otherwise) |
| `FreeCompanyExchange` | - | `NumberArrayData* Array` (number array 61), `bool IsReady`, `int Count`, `List<Entry> Entries`; `Entry`: `RowRef<CompanyAction> Action`, `RankRequired`, `SealsRequired`, `Buy()` (callback `2, Index`) |
| `Gathering` | `AddonGathering` | `int CurrentIntegrity`, `int TotalIntegrity`, `GatheredItem[] GatheredItems` (always 8); `GatheredItem`: `CheckBox`, `IsEnabled`, `ItemName`, `ItemID`, `IsCollectable`, `ItemLevel`, `GatherChance`, `BoonChance`, `Gather()`; on `master` after the 3.2.1.26 package also `bool IsHidden` (text node 6 visible), `bool IsRare` (text node 7 visible), `int GatherAmount` |
| `GatheringMasterpiece` | `AddonGatheringMasterpiece` | `ItemName`, `ItemID`, `CurrentCollectability`, `MaxCollectability`, `MinCollectability`, `MidCollectability`, `HighCollectability`, `CurrentIntegrity`, `TotalIntegrity`, `GatherChance`, `ScourPower`, `BrazenPowerMin`, `BrazenPowerMax`, `MeticulousPower`, `ScrutinyCheckBox`, `CollectorsIntuitionCheckBox`, `HelpButton` (182), `ReturnButton` (183), `Help()`, `Return()` |
| `GcArmyChangeClass` | - | Nine class buttons (`GladiatorButton` .. `ArcanistButton`) with `Gladiator()`, `Marauder()`, `Pugilist()`, `Lancer()`, `Rogue()`, `Archer()`, `Conjurer()`, `Thaumaturge()`, `Arcanist()` |
| `GcArmyExpedition` | `AddonGcArmyExpedition` | `Deploy()` |
| `GcArmyExpeditionResult` | `AddonGcArmyExpeditionResult` | `Complete()` |
| `GcArmyMenberProfile` | - | `ViewMembersButton` (2), `CloseButton` (3), `QuestionButton` (36), `PostponeButton` (37), `DismissButton` (38), radio `DisplayOrdersButton`, `ChangeClassButton`, `ConfirmChemistryButton`, `OutfitButton`, and one method per button (`ViewMembers()`, `Close()`, `Question()`, `Postpone()`, `Dismiss()`, `DisplayOrders()`, `ChangeClass()`, `ConfirmChemistry()`, `Outfit()`) |
| `GcArmyTraining` | - | `CloseButton` (39), `Close()` |
| `GearSetList` | - | `DisplayHelpButton` (2), `CreateNewGearsetButton` (3), `RefreshButton` (4), `EquipSetButton` (9), `DisplayHelp()`, `CreateNewGearset()`, `Refresh()`, `EquipSet()` |
| `GrandCompanySupplyReward` | `AddonGrandCompanySupplyReward` | `DeliverButton` (38), `bool IsEnabled`, `Deliver()` |
| `InclusionShop` | - | See Shop family |
| `InputNumeric` | - | See above |
| `Inventory` | `AddonInventory` | `OpenArmoryChest()` (button 17); implements `ICombinedInventory` |
| `InventoryExpansion` | `AddonInventoryExpansion` | `OpenArmoryChest()` (button 140); implements `ICombinedInventory` |
| `InventoryLarge` | `AddonInventoryLarge` | `OpenArmoryChest()` (button 71); implements `ICombinedInventory` |
| `ItemDetailCompare` | - | `CloseButton` (148), `Close()` |
| `ItemFinder` | - | `CloseButton` (14), `Close()` |
| `ItemInspectionResult` | `AddonItemInspectionResult` | `NameNode`, `DescNode`, `SeString ItemName`, `SeString Description`, `ItemNameText`, `DescriptionText`, `NextButton` (74), `CloseButton` (73), `Next()`, `Close()` |
| `ItemSearchResult` | `AddonItemSearchResult` | `ListComponent`, `ListItems`, `Entry[] Entries`; `Entry`: `Index`, `HQImageNode`, `MateriaTextNode`, `PriceTextNode`, `QuantityTextNode`, `TotalTextNode`, `RetainerTextNode`. No select method (source TODO) |
| `JournalAccept` | - | `AcceptButton` (44), `DeclineButton` (45), `Accept()`, `Decline()` |
| `JournalDetail` | `AddonJournalDetail` | `bool CanInitiate`, `AcceptMap()`, `Initiate()`, `AbandonDecline()` |
| `JournalResult` | `AddonJournalResult` | `Complete()`, `Decline()` |
| `LetterHistory` | - | `CloseButton` (5), `Close()` |
| `LetterList` | - | `NewButton` (2), `SentLetterHistoryButton` (3), `DeliveryRequestButton` (4), `New()`, `SentLetterHistory()`, `DeliveryRequest()` |
| `LetterViewer` | - | `TakeAllButton` (30), `ReplyButton` (31), `DeleteButton` (32), `TakeAll()`, `Reply()`, `Delete()`, `Item[] Items` (`Item : ReaderLetterViewer.Item`) |
| `LookingForGroup` | - | `RecruitMembersButton` (46), `bool RecruitMembersOrDetails()` |
| `LookingForGroupCondition` | - | `RecruitButton` (113), `CancelButton` (114), `ResetButton` (112), `bool Normal()`, `bool Alliance()`, `bool Recruit()`, `bool Cancel()`, `bool Reset()`, `SelectDutyCategory(byte i)` |
| `LookingForGroupDetail` | `AddonLookingForGroupDetail` | `JoinEditButton` (109), `TellEndButton` (110), `BackButton` (111), `bool JoinEdit()`, `bool TellEnd()`, `bool Back()`, `string PartyLeader`, `string Description`, `string World` |
| `LookingForGroupPrivate` | - | `JoinButton` (6), `CancelButton` (7), `Join(int password)` (callback `0, password`), `Join()`, `Cancel()` |
| `LotteryWeeklyInput` | - | `int Week`, `bool Unk03`, `int Unk04`, `int Unk05`, `PurchaseButton` (31), `RandomButton` (32), `Purchase()`, `Random()` |
| `LotteryWeeklyRewardList` | - | `string Week`, `int WinningNumber`, `CloseButton` (49), `Close()`, `Reward[] Rewards` (5); `Reward`: `Unk01`, `Place`, `MGPReward`, `ItemRewardId`, `int? ItemRewardIconId`, `string? ItemRewardName`, `Requirement` |
| `ManeuversArmorBoarding` | - | `CruiseChaserButton` (4), `OppressorButton` (5), `BruteJusticeButton` (6), `MountButton` (11), `CruiseChaser()`, `Oppressor()`, `BruteJustice()`, `Mount()` |
| `ManeuversRecord` | - | `LeaveButton` (62), `Leave()` |
| `MateriaAttachDialog` | - | `int SuccessRate` (`AtkValues[41]`), `SuccessRateSeString`, `SuccessRateText`, `float SuccessRateFloat`, `MeldButton` (35), `ReturnButton` (36), `Meld()`, `Return()` |
| `MaterializeDialog` | `AddonMaterializeDialog` | `Materialize()` (struct `YesButton`), `No()` |
| `MateriaRetrieveDialog` | `AddonMateriaRetrieveDialog` | `BeginButton` (17), `ReturnButton` (18), `Begin()`, `Return()` |
| `MiragePrismExecute` | - | `CastButton` (23), `ReturnButton` (24), `Cast()`, `Return()` |
| `MiragePrismMiragePlateConfirm` | - | `YesButton` (6), `NoButton` (7), `bool Yes()`, `bool No()` |
| `MiragePrismPrismSetConvert` | - | `StoreAsGlamourButton` (27), `CloseButton` (26), `StoreAsGlamour()`, `Close()`, `Item[] Items` (`Item : ReaderMiragePrismPrismSetConvert.Item`), `uint GlamourPrismsHeld`, `uint GlamourPrismsRequired`, `List<int> SlotsFilled`, `int? FirstUnfilledSlot`, `bool? TryHandOver(int slot)` |
| `MiragePrismPrismSetConvertC` | - | `StoreAsOutfitGlamourCheckBox`, `YesButton` (6), `NoButton` (7), `Yes()`, `No()` |
| `MiragePrismRemove` | - | `DispelButton` (14), `ReturnButton` (15), `Dispel()`, `Return()` |
| `MJIDisposeShop` | - | `uint NumEntries`, `ExportShopInfo[] ExportItems`; `ExportShopInfo`: fields `IconId`, `ItemName`, `Inventory`, `Value`, `SellIconId`, `Allocated`, `Select()` (callback `12, index`) |
| `MJIHud` | - | `SanctuaryRank`, `CurrentIslandXP`, `NextIslandLevelXP`, `IslandersCowrie`, `SeafarersCowrie`, buttons `IsleventoryButton` (20) .. `GuideButton` (27), `Isleventory()`, `SanctuaryCraftingLog()`, `SanctuaryGatheringLog()`, `ManageHideaway()`, `MaterialAllocation()`, `ManageMinions()`, `ManageFurnishing()`, `Guide()` |
| `MJIRecipeNoteBook` | - | `CraftButton` (34), `Craft()` |
| `MoogleCollection` | - | `OpenDutyFinderButton` (42), `OpenAreaMapButton` (43), `RewardsButton` (44), `AccessSpecialSiteButton` (59), `OpenDutyFinder()`, `OpenAreaMap()`, `Rewards()`, `AccessSpecialSite()` |
| `PurifyAutoDialog` | - | `CancelExitButton` (16), `CancelExitButtonSeString`, `CancelExitButtonText`, `bool PurificationActive`, `bool PurificationInactive`, `CancelExit()` |
| `PurifyResult` | - | `BannerSeString`, `BannerText`, `AutomaticButton` (19), `CloseButton` (20), `Automatic()`, `Close()` |
| `PvpProfile` | - | `SeriesMalmstonesButton` (21), `SeriesMalmstones()` |
| `PvpReward` | - | `uint CurrentLevelXp`, `uint TotalLevelXp`, `CloseButton` (124), `Close()` |
| `RecipeNote` | `AddonRecipeNote` | `Synthesize()`, `QuickSynthesis()`, `TrialSynthesis()`, `Material(uint index, bool hq)` (callback `6, index (+0x10000 if hq), 0`, `updateState` false) |
| `RecommendEquip` | - | `EquipButton` (11), `CancelButton` (12), `Equip()`, `Cancel()` |
| `RelicNoteBook` | `AddonRelicNoteBook` | `RelicNoteBookEnemy[] Enemies` (10), `RelicNoteBookDungeon[] Dungeons` (3), `RelicNoteBookFate[] Fates` (3), `RelicNoteBookLeve[] Leves`; each entry: `CheckBox`, `IsEnabled`, `Click()` |
| `Repair` | `AddonRepair` | `RepairAll()` |
| `Request` | `AddonRequest` | See above |
| `RetainerItemTransferList` | `AddonRetainerItemTransferList` | `Confirm()`, `Cancel()` |
| `RetainerItemTransferProgress` | `AddonRetainerItemTransferProgress` | `Close()` |
| `RetainerList` | `AddonRetainerList` | See above |
| `RetainerSell` | `AddonRetainerSell` | `ComparePricesButton` (4), `ConfirmButton` (21), `CancelButton` (22), `int AskingPrice { get; set; }` (set fires callback `2, value`), `int Quantity { get; set; }` (set fires callback `3, value`), `string ItemName`, `ComparePrices()`, `Confirm()`, `Cancel()` |
| `RetainerTaskAsk` | `AddonRetainerTaskAsk` | `AssignButton`, `ReturnButton`, `Assign()`, `Return()` |
| `RetainerTaskResult` | `AddonRetainerTaskResult` | `ReassignButton`, `ConfirmButton`, `Confirm()`, `Reassign()` |
| `ReturnerDialog` | - | `AcceptButton` (4), `DeclineButton` (5), `DecideLaterButton` (6), `Accept()`, `Decline()`, `DecideLater()` |
| `SalvageAutoDialog` | - | `EndDesynthesisButton` (28), `EndDesynthesisButtonSeString`, `EndDesynthesisButtonText`, `bool DesynthesisActive`, `bool DesynthesisInactive`, `EndDesynthesis()` |
| `SalvageDialog` | `AddonSalvageDialog` | `bool BulkDesynthEnabled { get; set; }`, `DesynthesizeButton` (25), `CancelButton` (26), `Desynthesize()`, `Cancel()`, `Checkbox()` (`[Obsolete("Do not use")]`, empty body) |
| `SalvageResult` | - | `CloseButton` (15), `Close()` |
| `SatisfactionSupplyResult` | - | `uint CurrentLevelXp`, `uint TotalLevelXp`, `AcceptButton` (36), `Accept()` |
| `SelectIconString` | `AddonSelectIconString` | See above |
| `SelectOk` | `AddonSelectOk` | `SeString SeString`, `string Text`, `Ok()` |
| `SelectString` | `AddonSelectString` | See above |
| `SelectYesno` | `AddonSelectYesno` | See above |
| `ShopCardDialog` | `AddonShopCardDialog` | `SeString Price`, `int? Quantity { get; set; }`, `int MinQuantity`, `int MaxQuantity`, `SellButton` (16), `CancelButton` (17), `Sell()`, `Cancel()` |
| `Shop` | - | See Shop family |
| `ShopExchangeCurrency` | - | See Shop family |
| `ShopExchangeCurrencyDialog` | - | See Shop family |
| `ShopExchangeItemDialog` | - | See Shop family |
| `SubmarineExplorationMapSelect` | - | `SeString`, `Text`, `ListComponent` (list 7), `int EntryCount`, `ListItems`, `Entry[] Entries`; `Entry`: `Index`, `TextNode`, `SeString`, `Text`, `Select()` (callback `Index`). `Entry.SeString` reads the window's text node 2, not the entry; use `TextNode` |
| `Synthesis` | - | `ReaderSynthesis Reader`, nested `enum Condition` (`Normal`, `Good`, `Excellent`, `Poor`, `Centered`, `Sturdy`, `Pliant`, `Malleable`, `Primed`, `GoodOmen`, `Unknown`) |
| `SynthesisSimpleDialog` | - | `UseHQMaterialsCheckbox`, `SynthesizeButton` (7), `CancelButton` (8), `Synthesize()`, `Cancel()` |
| `Talk` | `AddonTalk` | `Click()` |
| `TitleDCWorldMap` | - | See above |
| `_TitleMenu` | - | See above |
| `TreasureHighLow` | - | `TryLuckButton` (46), `OpenChestButton` (47), `CloseButton` (48), `TryLuck()`, `OpenChest()`, `Close()` |
| `TripleTriadRequest` | - | `string Opponent`, `int CurrentMGP`, `RegionalRule1`, `RegionalRule2`, `MatchRule1`, `MatchRule2`, `int MatchFee`, `uint MGPReward`, `ChallengeButton` (41), `QuitButton` (42), `Challenge()`, `Quit()` |
| `TripleTriadResult` | - | `int WonValue`, `uint MGPReward`, `bool WonGame`, `RematchButton` (21), `QuitButton` (22), `Rematch()`, `Quit()` |
| `VoteMvp` | - | `OkButton` (10), `CancelButton` (11), `SettingsButton` (12), `Ok()`, `Cancel()`, `Settings()` |
| `VVDVoteRoute` | - | `int EntryCount`, `int SelectedEntryNPC`, `SeString`, `Text`, `ListComponent` (list 3), `ListItems`, `Entry[] Entries`; `Entry`: `Index`, `ListIndex`, `TextNode`, `SeString`, `Text`, `Select()` (callback `1, Index`) |
| `WeeklyBingoResult` | - | `AcceptButton` (93), `CancelButton` (94), `Accept()`, `Cancel()` |
| `WKSHud` | - | `StellerMissionsButton` (7), `MechOpsButton` (8), `StellerSuccessButton` (9), `InfrastructorIndexButton` (10), `CosmicResearchButton` (11), `CosmicClassTrackerButton` (12), `int CosmoCredit`, `int LunarCredit`, `Mission()`, `Mech()`, `Steller()`, `Infrastructor()`, `Research()`, `ClassTracker()` |
| `WKSLottery` | - | `WheelLeftButton` (29), `WheelRightButton` (39), `SpinWheelButton` (64), `WheelItems[] LeftWheelItems`, `WheelItems[] RightWheelItems` (`WheelItems`: fields `itemId`, `itemName`, `itemAmount`), `SelectWheelLeft()`, `ConfirmButton()` (a method: clicks `SpinWheelButton`). No right-wheel select (commented out) |
| `WKSMission` | - | `HelpButton` (11), `MissionLogButton` (12), `MissionSelectionButton` (13), `BasicMissionsButton` (17), `ProvisionalMissionsButton` (18), `CriticalMissionsButton` (19), `uint NumEntries`, `uint CurrentTab`, `uint SelectedMissionId`, `string SelectedMissionName`, `StellarMissions[] StellerMissions` (`Name`, `MissionId`, `Select()`, `Initiate()`), `uint NumClasses`, `ClassDropdown[] SelectClass` (`Select()`), `Help()`, `MissionSelection()`, `MissionLog()`, `BasicMissions()`, `ProvisionalMissions()`, `CriticalMissions()` |
| `WKSMissionInfomation` | - | `string Name`, `uint? CurrentScore`, `uint? SilverScore`, `uint? GoldScore`, `uint? CriticalScore`, `CosmoPouchButton` (34), `CosmoCraftingLogButton` (35), `StellerReductionButton` (36), `ReportResultsButton` (37), `AbandonMissionButton` (38), `CosmoPouch()`, `CosmoCraftingLog()`, `StellerReduction()`, `Report()`, `Abandon()` |
| `WKSRecipeNotebook` | - | `NQItemsButton` (39), `HQItemsButton` (40), `SynthesizeButton` (50), `string SelectedCraftingItem`, `CraftItems[] CraftingItems` (`Name`, `Select()`), `NQItemInput()`, `HQItemInput()`, `Synthesize()` |
| `WKSToolCustomize` | - | `ClassSelector[] ClassList` (`ClassName`, `Level`, `Select()`) |
| `WorldTravelFinderStatus` | - | `StartingWorldString`, `DestinationWorldString`, `PositionInQueueString`, `TimeElapsedString`, `TimeRemainingString`, `CancelButton` (13), `Cancel()` |

`TryHandOver(int slot)` (`BankaCraftworksSupply`, `MiragePrismPrismSetConvert`) is a two-step call
meant to be polled: returns `true` if the slot is already filled; otherwise, if `ContextIconMenu` is
not visible it fires the slot callback and returns `false`; if the menu is visible it fires the menu
callback and returns `true`. It never returns `null`.

## AddonMasterCombined.ICombinedInventory

`references/ecommons/ECommons/UIHelpers/AddonMasterCombinedImplementations/ICombinedInventory.cs`:
`public static unsafe partial class AddonMasterCombined` containing
`public interface ICombinedInventory { void OpenArmoryChest(); static ICombinedInventory Get(); }`.
`Get()` returns the first of `Inventory`, `InventoryLarge`, `InventoryExpansion` that exists and
`IsAddonReady`, else `null`.

```csharp
AddonMasterCombined.ICombinedInventory.Get()?.OpenArmoryChest();
```

## AddonFinder

`public unsafe class AddonFinder` (namespace `ECommons.UIHelpers`), two static lazy enumerables over
all instances of a multi-instance addon, starting at index 1 and stopping at the first null:

- `public static IEnumerable<AddonMaster.SelectString> SelectString`
- `public static IEnumerable<AddonMaster.SelectYesno> YesNo`

No readiness filter. `AddonFinder.YesNo.FirstOrDefault(x => x.IsAddonReady && x.Text.Contains(...))`
replaces a hand-written index loop.

## AtkEventDataBuilder

`public unsafe class AtkEventDataBuilder`: `public readonly AtkEventData Data`, a parameterless
constructor, `public AtkEventDataBuilder Write<T>(int pos, T data) where T : unmanaged` (writes
`data`
at byte offset `pos` inside `Data`, no bounds check, returns `this`), and
`public AtkEventData Build()`. Used by wrappers that synthesise `ReceiveEvent` calls, e.g.
`CreateAtkEventData().Write<byte>(16, i).Write<byte>(22, i).Build()`.

## AtkReader

`public abstract unsafe class AtkReader(AtkUnitBase* UnitBase, int BeginOffset = 0)` with a second
constructor `AtkReader(nint UnitBasePtr, int BeginOffset = 0)`. It reads the addon's `AtkValues`
array; every index is `n + BeginOffset`.

| Member | Behaviour |
|---|---|
| `protected uint? ReadUInt(int n)` | Requires `AtkValueType.UInt` |
| `protected int? ReadInt(int n)` | Requires `AtkValueType.Int` |
| `protected bool? ReadBool(int n)` | Requires `AtkValueType.Bool`; value is `Byte != 0` |
| `protected SeString ReadSeString(int n)` | Requires `String`, `String8`, `WideString` or `ManagedString` |
| `protected string ReadString(int n)` | Same types, `MemoryHelper.ReadStringNullTerminated` |
| `public List<T> Loop<T>(int Offset, int Size, int MaxLength, bool IgnoreNull = false) where T : AtkReader` | Creates up to `MaxLength` sub-readers at offsets `Offset + i * Size` |
| `public bool IsNull` | `AtkValuesCount == 0` or the value at `BeginOffset` has type 0 |
| `public (nint UnitBase, int BeginOffset) AtkReaderParams` | For re-wrapping a reader (see `RetainerList.Entry`) |

- There are no other typed readers (no float, no long).
- All `Read*` return `null` when the value type is 0 (empty), throw `InvalidCastException` when the
  type differs from the requested one, and throw `ArgumentOutOfRangeException` when the index is
  `>= AtkValuesCount`. Readers therefore break when a patch changes the value layout; wrap in
  try/catch where that matters.
- `Read*` are `protected`: you consume a reader through the properties of a subclass.
- `Loop<T>` uses `Activator.CreateInstance(typeof(T), [(nint)UnitBase, Offset + (i * Size)])`, so an
  element type needs a public `(nint, int)` constructor. The offset passed is absolute (the parent's
  `BeginOffset` is not added). It stops at the first element whose `IsNull` is true unless
  `IgnoreNull`.
- Every property access re-reads memory and `Loop` allocates a new list; cache the result in a local
  inside a frame.

Writing one:

```csharp
public unsafe class ReaderMyAddon(AtkUnitBase* addon) : AtkReader(addon)
{
    public uint Count => ReadUInt(0) ?? 0;
    public List<Row> Rows => Loop<Row>(5, 3, (int)Count);

    public class Row(nint addon, int start) : AtkReader(addon, start)
    {
        public string Name => ReadString(0);
        public int Amount => ReadInt(1) ?? 0;
    }
}
```

### Shipped readers (`ECommons.UIHelpers.AtkReaderImplementations`)

The class name suggests the addon; the reader accepts any `AtkUnitBase*` and checks nothing.

| Reader | Members |
|---|---|
| `ReaderBannerList` | `CharacterOption1IconId`, `CharacterOption2IconId`, `BackgroundIconId`, `FrameIconId`, `AccentIconId`, `NumPortraits`, `SelectedPortrait`, `List<Portrait> Portraits`; `Portrait`: `Unk01`, `ClassJobIconId`, `ListIndex`, `GlamourPlateId`, `PortraitBroken`, `Unk06`, `UseAsInstantPortrait`, `GearSetName`, `GearSetILvl`, `IsPortraitBroken`, `IsUseAsInstantPortraitSet` |
| `ReaderContextMenu` | `uint Count`, `List<ContextMenuEntry> Entries`; entry: `string Name`, `SeString NameSeString` |
| `ReaderDawnStory` | `CurrentSelection`, `AvailableExpansions`, `CurrentExpansion`, `EntryCount`, `List<Entry> Entries` (`Unk0`, `Callback`, `Status`), `List<EntryName> EntryNames` (`Name`, `Level`) |
| `ReaderGCArmyCapture` | `PlayerCharName`, `PlayerCharIlvl`, `PlayerCharLvl`, `EntryCount`, `List<DungeonInfo> Entries` (`Unk0`, `Completed`, `Unk2`, `Name`, `Level`, `Synced`) |
| `ReaderGCArmyMemberList` | `EntryCount`, `List<MemberInfo> Entries` (`Unk0`, `SelectionMask`, `Selected`, `Name`, `Class`, `PortraitPath`, `ClassId`, `ClassType`, `Level`, `Unk3`, `Unk4`, `Physical`, `Mental`, `Tactical`, `Chemistry`, `Tactics`), `enum SquadronClassType` |
| `ReaderLetterHistory` | `LetterCount`, `List<Letter> Items`; `Letter`: `Recipient`, `Unk20`, `CanReply`, `SentDate`, `NumItems`, `List<Item> Items`, `Gil`; `Item`: `ItemNameAndQuantity`, `ItemQuantity` |
| `ReaderLetterViewer` | `SenderName`, `List<Item> Items` (`ItemIconId`, `ItemId`, `ItemName`) |
| `ReaderMiragePrismPrismSetConvert` | `Unk00`, `GlamourPrismsHeld`, `Unk02`, `Unk03`, `Unk04`, `OutfitIconId`, `ItemCount`, `List<Item> Items` (`ItemId`, `ItemIconId`, `Unk03`, `Unk04`, `InventoryType`, `InventorySlot`, `Flag`), `enum ItemFlag` (`Missing = 0`, `Unfilled = 2`, `Filled = 3`) |
| `ReaderRetainerList` | `uint VentureCount`, `List<Retainer> Retainers` (`Name`, `Level`, `Inventory`, `Gil`, `IsActive`) |
| `ReaderSelectString` | `string Description`, `int NumEntries`, `List<Entry> Entries` (`Text`) |
| `ReaderSynthesis` | `Unk0`, `IsTrialSynthesis`, `ItemName`, `ItemIconId`, `ItemCount`, `Progress`, `MaxProgress`, `Durability`, `MaxDurability`, `Quality`, `HQChance`, `ConditionName`, `ConditionNameAndTooltip`, `StepCount`, `ResultItemId`, `MaxQuality`, `RequiredQuality`, `Collectability`, `MaxCollectability`, `CollectabilityCheckpoint1..3`, `IsExpertRecipe`, `IsShowingCollectibleInfo`, `Condition` (`AddonMaster.Synthesis.Condition`), `IsCollectible`, `IsMaxProgress`, `IsMaxQuality` |
| `ReaderTripleTriadCoinExchange` | `EntryCount`, `List<CardEntry> Entries` (`Unk0`, `Name`, `Value`, `Count`, `Id`, `InDeck`) |

| `ReaderXBMContentsItemShop` | `CoinsString`, `int Coins`, `uint StockCount`, `List<StockEntry> StockEntries` (`purchaseIndex`, `Listed`, `Item`, `PriceString`, `Price`, `Discounted`, `Bought`), `List<ItemEntry> ItemEntries` / `ItemEntriesValid` (`Unk0`, `Available`, `IconId`, `Id`, `Name`, `GetIndex(int initialOffset)`), `List<GearEntry> OwnedEntries` / `OwnedEntriesOwned` (`Owned`, `IconId`, `Id`, `Name`, `GetIndex`), constants `ITEM_ENTRY_SIZE`, `ITEM_ENTRY_LENGTH`, `GEAR_ENTRY_SIZE`, `GEAR_ENTRY_LENGTH` |
| `ReaderXBMContentsMainHUD` | `CoinsString`, `Coins`, `ItemEntries`, `ItemEntriesValid`, `OwnedEntries`, `OwnedEntriesOwned` (entry types from `ReaderXBMContentsItemShop`), `int GetItemIndex(ItemEntry item)` |
| `ReaderXBMContentsBooty` | `CoinsString`, `Coins`, `LootCoinsString`, `LootCoins`, `LootCoinsTaken`, `List<LootChoice> LootChoices` (`lootIndex`, `Item`, `Unk4`, `Taken`), `ItemEntries`, `ItemEntriesValid`, `OwnedEntries`, `OwnedEntriesOwned` |
| `ReaderXBMContentsTreasure` | `CoinsString`, `Coins`, `List<TreasureChoice> TreasureChoices` (`treasureIndex`, `Unk0`, `Item`, `Bought`), `ItemEntries`, `ItemEntriesValid`, `OwnedEntries`, `OwnedEntriesOwned` |
| `ReaderXBMMonsterNotebook` | `const int EntryCountPerPage = 25`, `PageCount`, `CurrentPage`, `List<MonsterEntry> CurrentPageEntries` (`Number`, `Caught`, `NumberString`, `Unk1`..`Unk7`), `CurrentNumber`, `SelectedNumber`, `Filtered`, `Selected`, and for the selected monster `CurrentName`, `CurrentClassification`, `CurrentRank`, `CurrentRankCapped`, `CurrentXP`, `CurrentHP`, `CurrentMaxHP`, `CurrentStrength`, `CurrentPhysResistance`, `CurrentConstitution`, `CurrentIntelligence`, `CurrentMagicResistance` (each stat also as a `...String` `SeString`) |
| `ReaderXBMPetParty` | `EntryCount`, `List<MonsterEntry> TeamEntries` (`index`, `Rank`, `IconId`, `Disabled`, `Name`, `HP`, `MaxHP`, the five stats, eleven status-immunity bools from `Slow` to `Poison`, `FedCurrent`, `FedMax`, `SelectionIndex`, `List<FedEntry> FedItems`, `Number`), `TeamSize`, `FedEntry FeedItem` (`IconId`, `ItemId`) |
| `ReaderXBMResult` | `Performance`, `MovesMadeScore`, `EnemyScore`, `EliteScore`, `BossScore`, `RemainingHP`, `RemainingHPScore`, `BonusBoostScore`, `BonusBoostCount`, `List<BonusBoostEntry> BonusBoostEntries` (`Name`, `Score`), `LootCount`, `List<LootEntry> LootEntries` (`ItemId`, `IconId`, `Name`, `Count`), `TotalScore`, `Rank`, `BeastCount`, `List<BeastEntry> BeastEntries` (`IconId`, `Number`, `PrevXP`, `NewXP`, `PrevRank`, `NewRank`) |

The seven `ReaderXBM*` classes and `DigitParsers` are in packages 3.2.1.23 and later. No `AddonMaster`
class wraps them and the source does not name the addons they read; the class names are the only
hint *(addon names unverified)*. Their numeric fields are parsed with
`ECommons.StringHelpers.DigitParsers` (`int FirstNumber(string text)`: first run of digits and
commas, commas removed, `0` when none; `int Digits(string text)`: leading digits only), which is
public and usable on its own.

Constructors: `ReaderBannerList`, `ReaderContextMenu`, `ReaderLetterHistory`, `ReaderLetterViewer`,
`ReaderMiragePrismPrismSetConvert`, `ReaderRetainerList`, `ReaderSelectString`, `ReaderSynthesis`
take `(AtkUnitBase*)`; `ReaderDawnStory`, `ReaderGCArmyCapture`, `ReaderGCArmyMemberList`,
`ReaderTripleTriadCoinExchange` and the `ReaderXBM*` classes take
`(AtkUnitBase* UnitBase, int BeginOffset = 0)`. Pass `master.Base`, e.g.
`new ReaderRetainerList(m.Base)`.

## Gotchas

- `TryGetAddonByName` and `TryGetAddonMaster` succeed for hidden or loading addons. Always add
  `IsAddonReady`.
- The wrapper constructor validates nothing. Wrapping the wrong addon or a stale pointer reads
  garbage or crashes. Do not keep a wrapper across frames; re-acquire it each time.
- `SelectYesno.Yes()` force-enables a disabled button unless `RespectDisabledButtons = true`.
  `RespectHoldButtons` has no effect at this commit.
- ECommons `IsAddonReady` / `IsReady()` is not ClientStructs `IsReady`; see the section above.
- `Entries` / `Retainers` / `ShopItems` and similar arrays are rebuilt on each access, and several
  read item-renderer pointers that are only valid while the list is populated. Take one snapshot per
  call site.
- `SelectString.Entry.Select()`, `SelectIconString.Entry.Select()`, shop `Select` and
  `InputNumeric.Ok(int)` fire callbacks without validating enabled state or range. Only
  `ContextMenu.Entry.Select()`, `RetainerList.Entry.Select()`, `FreeCompanyCreditShop.Item.Buy`,
  `TitleDCWorldMap.Select` and the `ClickButtonIfEnabled` based methods check anything.
- Hard-coded node ids and `AtkValues` offsets are patch-sensitive; the wrappers for rarely used
  addons may lag behind a game patch. Properties that index `AtkValues` directly (not through
  `AtkReader`) have no bounds or type check.
- Bugs visible in source at this commit: `Request.IsFilled` always returns `true`;
  `MJIHud.ManageMinions()` clicks `ManageHideawayButton`, not `ManageMinionButton`;
  `SubmarineExplorationMapSelect.Entry.SeString` returns the window prompt;
  `_CharaSelectListMenu.TemporarilyLocked` is hard-coded `false`.
- `ClickCheckboxIfEnabled` always calls `SetChecked(true)`; it cannot be used to uncheck.
- `AddonMaster` cannot be instantiated and has no registry of wrappers; enumerate the nested types
  by
  reflection if a list is needed (`AddonDescription` exists for that UI purpose, but it is an
  instance property).
- Rapid repeated clicks are the caller's problem: gate with `EzThrottler.Throttle(...)` as
  AutoRetainer does.
