## Other ECommons pieces scripts use

If the sibling `ecommons` skill is installed, its TaskManager, AddonMaster,
EzIPC and events/throttlers/schedulers topics go deeper; what is here is
enough on its own.

- `Svc.*` (ECommons.DalamudServices): `ClientState`, `Condition`, `Objects`,
  `Targets`, `Data` (Lumina sheets: `Svc.Data.GetExcelSheet<Item>().GetRowOrDefault(id)`),
  `Commands`, `Chat`, `Toasts`, `GameGui`, `GameInventory`, `Framework`,
  `PluginInterface`, `Log`, `AetheryteList`, `Party`, `PlayerState`. Full
  list: `ecommons/Svc-services.txt`.
- `Player` (ECommons.GameHelpers): `Available`, `Interactable`, `Object`,
  `Position`, `Territory` (RowRef), `Job`, `Level`, `NameWithWorld`,
  `IsInHomeWorld`, `IsInDuty`, `IsMoving`, `Mounted`, `CanMount`, `CanFly`,
  `IsAnimationLocked`, `IsCasting`, `IsDead`, `DistanceTo(Vector3|IGameObject)`,
  `CurrentWorld`, `HomeWorld`. `ECommons.GameHelpers.LegacyPlayer.Player`
  is the obsolete predecessor (QuestionableQuestQueue aliases it).
- `ECommons.GenericHelpers` (also as `using static`): `IsOccupied()`,
  `IsScreenReady()`, `IsAddonReady(AtkUnitBase*)`, `TryGetAddonByName<T>`,
  `TryGetAddonMaster<T>`, `Copy(text)`/`Paste()`, `GetRandom(...)`,
  extensions `.EqualsAny(...)`, `.ContainsAny(...)`, `.Print()`,
  `.TryGetFirst(pred, out x)`, `.InRange(a, b)`, `.Toggle(x)`, `.GetText()`
  on SeString/ReadOnlySeString, `.Log()`.
- `EzThrottler.Throttle(name, miliseconds = 500, rethrottle = false)`
  (upstream spelling; pass it positionally, `ms:` does not compile) returns
  true once per window; `Check(name)` true when the window has expired;
  `Throttle(name, ms, true)` used as "arm a cooldown"; `Reset(name)`;
  `GetRemainingTime(name)`. The static `EzThrottler.Clear()` drops every
  key; it was added to ECommons on 2026-09-29.
  `FrameThrottler` counts frames.
- `ProperOnLogin.RegisterInteractable(Action)` / `RegisterAvailable` /
  `Unregister` (ECommons.Events) — fire once the player is really in the
  world; pair Register in OnEnable with Unregister in OnDisable and use
  `ValidTerritories = null`.
- `ExcelItemHelper.Get(id)`, `.GetName(id)`, `item.GetName()`;
  `ExcelActionHelper.GetActionName(id)`; `ExcelWorldHelper.GetName(id)`;
  `ExcelTerritoryHelper.GetName(id)`; `TerritoryIntendedUseEnum` and
  `OpenAreas.List` (ECommons.ExcelServices.TerritoryEnumeration) for
  building `ValidTerritories` sets.
- `EzHook` + `EzSignatureHelper.Initialize(this)` for signature hooks
  (AutoTeleport); enable in OnEnable, disable in OnDisable.
- `InventoryManager.Instance()->GetInventoryItemCount(itemId)` (+ `, true`
  for HQ), `AgentInventoryContext.Instance()->UseItem(id)`,
  `ActionManager.Instance()->GetActionStatus(ActionType.Item, id) == 0`
  (FFXIVClientStructs) for inventory/items.
- `QuestManager.IsQuestComplete(id)` / `QuestManager.Instance()->IsQuestAccepted(id)`
  (quest row ids are `65536 + questId`).

