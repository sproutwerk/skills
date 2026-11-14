Dalamud topic: `IAddonLifecycle` (events, args types, caveats) and `IAddonEventManager` (custom events on native nodes), including the ClientStructs wrapper pointer notes. Verified against `goatcorp/Dalamud` `43b3411` (2026-11-10, `DalamudVersion 15.0.3.9`, API 15), `goatcorp/dalamud-docs` `3c7597d` (2026-11-13), `goatcorp/SamplePlugin` `b8477da` (2026-08-15) and `aers/FFXIVClientStructs` `80cb4a7` (2026-11-10). Reviewed 2026-11-13.

Paths such as `references/...` and `references/topic-*.md` are relative to the skill root (the directory that holds `SKILL.md`), not to this file. Part of the `dalamud` skill; items marked *(unverified)* are not confirmed from the pinned checkouts.

## AddonLifecycle

Listen to native window events by addon name, no signatures.

```csharp
using Dalamud.Game.Addon.Lifecycle;
using Dalamud.Game.Addon.Lifecycle.AddonArgTypes;
using FFXIVClientStructs.FFXIV.Component.GUI;

addonLifecycle.RegisterListener(AddonEvent.PostSetup, "SelectYesno", OnSetup);
addonLifecycle.RegisterListener(AddonEvent.PreDraw, ["Character", "NamePlate"], OnDraw);
addonLifecycle.UnregisterListener(OnSetup, OnDraw);              // by handler, any event/addon

private unsafe void OnSetup(AddonEvent type, AddonArgs args)
{
    var addon = (AtkUnitBase*)args.Addon.Address;                // args.Addon is AtkUnitBasePtr;
                                                                 // 15.0.3.8+: args.GetAddonPointer<AtkUnitBase>()
    if (args is AddonSetupArgs setup)
        foreach (var value in setup.AtkValueEnumerable)          // AtkValuePtr: ValueType, GetValue(), TryGet<T>
            log.Verbose("{Type} {Value}", value.ValueType, value.GetValue());
}
```

`AddonEvent`: `Pre/PostSetup`, `Pre/PostUpdate`, `Pre/PostDraw`, `PreFinalize` (no Post), `Pre/PostRequestedUpdate`,
`Pre/PostRefresh`, `Pre/PostReceiveEvent`, `Pre/PostOpen`, `Pre/PostClose`, `Pre/PostShow`, `Pre/PostHide`,
`Pre/PostMove`, `Pre/PostMouseOver`, `Pre/PostMouseOut`, `Pre/PostFocus`, `Pre/PostFocusChanged`.

| Args type (`args.Type`) | Extra members |
|---|---|
| `AddonArgs` (`Generic`; Draw, Update, Finalize, Open, …) | `AddonName`, `AddonNameSpan`, `Addon`, `GetAddonPointer<T>()` (15.0.3.8+), `PreventOriginal()`, `PreventOriginalRequested` |
| `AddonSetupArgs`, `AddonRefreshArgs` | `AtkValueCount`, `AtkValues` (nint), `AtkValueEnumerable` |
| `AddonRequestedUpdateArgs` | `NumberArrayData`, `StringArrayData` (nint) |
| `AddonReceiveEventArgs` | `AtkEventType`, `EventParam`, `AtkEvent`, `AtkEventData` |
| `AddonShowArgs` | `SilenceOpenSoundEffect`, `UnsetShowHideFlags` |
| `AddonHideArgs` | `CallHideCallback`, `SetShowHideFlags` |
| `AddonCloseArgs` | `FireCallback` |
| `AddonFocusChangedArgs` | `ShouldFocus` |

- Omitting the addon name listens to **all** addons; avoid on Draw/Update.
- `PreventOriginal()` (API 15) only has an effect in a `Pre*` listener.
- Since API 14 the service swaps addon vtables. Listeners are not recursion-safe (calling `OnRefresh` from a Refresh
  listener re-enters). An addon's vtable pointer is Dalamud's replacement; `GetOriginalVirtualTable(nint)` returns the
  game's table.
- An addon that is already open when you register gets no `PostSetup`; check `gameGui.GetAddonByName(name)` once at
  startup.
- `AtkValueSpan` shown in the upstream how-to is `internal` in the checkout; use `AtkValueEnumerable` or cast
  `AtkValues` to `AtkValue*`.
- `AddonArgs.InvalidAddonName` and `InvalidAddonNameSpan` are `[Obsolete]`: `AddonName` is filled in when the args
  are created, so comparing against the sentinel is dead code.

## AddonEventManager

Attach custom events to native nodes; Dalamud tracks and removes them.

```csharp
using Dalamud.Game.Addon.Events;
using Dalamud.Game.Addon.Events.EventDataTypes;

private IAddonEventHandle? clickHandle;

private unsafe void OnSetup(AddonEvent type, AddonArgs args)
{
    var addon = (AtkUnitBase*)args.Addon.Address;
    var node = addon->GetNodeById(22);
    if (node == null) return;
    node->NodeFlags |= NodeFlags.EmitsEvents | NodeFlags.RespondToMouse | NodeFlags.HasCollision;
    clickHandle = eventManager.AddEvent((nint)addon, (nint)node, AddonEventType.MouseClick, OnClick);
}

private void OnClick(AddonEventType type, AddonEventData data)
{
    if (data is AddonMouseEventData { IsLeftClick: true } mouse)
        log.Debug("clicked at {Pos}", mouse.Position);
}
// eventManager.RemoveEvent(clickHandle); eventManager.SetCursor(AddonCursorType.Clickable); eventManager.ResetCursor();
```

- Delegate is `AddonEventDelegate(AddonEventType atkEventType, AddonEventData data)`. The upstream how-to still shows
  the old three-argument handler; that no longer compiles. `AddonEventData`: `AtkEventType`, `Param`, `AddonPointer`,
  `NodeTargetPointer`, `AtkEventPointer`, `AtkEventDataPointer`. `AddonMouseEventData` adds `IsLeftClick`,
  `IsRightClick`, `IsAltHeld`, `IsControlHeld`, `IsShiftHeld`, `IsDragging`, `IsScrollUp`, `IsScrollDown`, `Position`.
- Pass a **node** (`Atk…Node*`), never a component: a component address crashes. For `AtkComponentButton*` use its
  `OwnerNode`.
- `AddEvent` returns `null` on failure. Events on non-persistent addons die with the addon; on persistent ones
  (`_BagWidget`, `NamePlate`) keep the handle and `RemoveEvent` it.
- `AddonEventType` is generated from ClientStructs `AtkEventType` (`references/dalamud/Dalamud/EnumCloneMap.txt` lists
  all 32 cloned enums, e.g. `ObjectKind`, `GameInventoryType`, `AtkValueType`); member lists are in ClientStructs, not
  in the Dalamud source. The `clientstructs` skill carries the ClientStructs source files at its own pin
  (`references/structs/FFXIV/Component/GUI/AtkEvent.cs` for `AtkEventType`, where `MouseClick = 9`).
- Wrapper ↔ ClientStructs pointer: `AtkUnitBasePtr`, `AgentInterfacePtr`, `UIModulePtr` and `AtkValuePtr` all expose
  `Address` and `IsNull`, convert implicitly to `nint`, and convert implicitly from `nint` and `void*`. Cast
  `(AtkUnitBase*)ptr.Address` to work on the struct. `AtkUnitBasePtr.IsReady` is the ClientStructs `IsReady` property,
  which is not the same test as the ECommons `IsAddonReady` extension (`clientstructs` skill, Addon work).
