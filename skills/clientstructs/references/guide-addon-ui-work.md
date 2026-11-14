## Addon (UI) work

Vocabulary: an **Addon** is a window — an `AtkUnitBase`-derived struct
(`Client.UI.AddonSelectYesno`, `AddonShop`, … matched to the
runtime name by `[Addon("SelectYesno")]`). An **Agent** is the logic/data
object behind it (`Client.UI.Agent.AgentInventoryContext`, `AgentMap`, …,
`AgentInterface`-derived, keyed by `AgentId`). Agents own state
and open/close addons (`AgentInterface.Show()`, `Hide()`,
`IsAgentActive()`, `GetAddonId()`, `IsAddonReady()`, `IsAddonShown()`);
addons own nodes, `AtkValues` and callbacks. Data flows Agent →
`OpenAddon(values)` / `RefreshAddon` → addon `OnSetup/OnRefresh(AtkValue[])`;
user input flows addon `FireCallback(AtkValue[])` → agent
`ReceiveEvent(returnValue, values, count, eventKind)`.

```csharp
var addon = RaptureAtkUnitManager.Instance()->GetAddonByName("SelectYesno");   // 2nd instance: index: 2
if (addon == null || !addon->IsVisible || !addon->IsReady) return;              // IsReady = OnSetup ran; IsVisible = VisibilityState has Show
// Dalamud/ECommons equivalents: Svc.GameGui.GetAddonByName("SelectYesno", 1).Address (IGameGui),
// GenericHelpers.TryGetAddonByName<AtkUnitBase>("SelectYesno", out var a) && GenericHelpers.IsAddonReady(a)
// ECommons IsAddonReady(a) / a->IsReady() (extension) = IsVisible && UldManager.LoadedState == AtkLoadState.Loaded && IsFullyLoaded();
// it does NOT read the CS IsReady bitfield property — `addon->IsReady` (CS property) and `addon->IsReady()` (ECommons extension) differ.

// Semicolon-separated member lists like the next line are API sketches (what exists), not statements to paste.
addon->NameString; addon->Id; addon->X; addon->Y; addon->Scale; addon->RootNode; addon->WindowNode; addon->ParentId; addon->HostId;

// AtkValues: the addon's argument vector (set by the agent on OnSetup/OnRefresh).
// The payload at 0x8 is a union: read only the member v.Type names. v.String on an Int/Bool value is a bogus pointer. Log = your logger.
for (var i = 0; i < addon->AtkValuesCount; i++) {
    ref var v = ref addon->AtkValues[i];
    switch (v.Type) {
        case AtkValueType.Int:  Log(v.Int); break;
        case AtkValueType.UInt: Log(v.UInt); break;
        case AtkValueType.Bool: Log(v.Bool); break;
        case AtkValueType.String or AtkValueType.ManagedString or AtkValueType.ConstString: Log(v.String.ToString()); break;
        default: Log(v.GetValueAsString()); break;   // type-aware for every AtkValueType (Null/Undefined → "", Vector/Pointer → hex)
    }
}
addon->AtkValuesSpan;   // Span<AtkValue>
var yn = (AddonSelectYesno*)addon; yn->StandardTypedAtkValues->PromptText.String.ToString(); yn->PromptText->NodeText.ToString(); yn->YesButton->IsEnabled;

// Nodes
// Get*ById / SearchNodeById return null for a missing id and GetAsAtkX for a type mismatch (the CS helpers
// AtkComponentBase.GetResNodeById / AtkUnitBase.GetComponentNodeById guard for it); check before ->.
var text = addon->GetTextNodeById(2);                         // AtkTextNode* or null: ->NodeText (Utf8String), SetText("..."), ->TextColor, ->FontSize
var btn  = addon->GetComponentButtonById(8);                  // AtkComponentButton* or null: ->ButtonTextNode, IsEnabled, SetText
var list = addon->GetComponentListById(9);                    // AtkComponentList* or null: ->ListLength, ->SelectedItemIndex, ->ItemRendererList[i].AtkComponentListItemRenderer
var node = addon->GetNodeById(15);                            // AtkResNode* or null: ->Type (NodeType.Text=3, Image=2, Component=10000), ->IsVisible(), ->NodeFlags & NodeFlags.Visible, ->ChildNode/NextSiblingNode/PrevSiblingNode/ParentNode, ->NodeId, ->Width/Height/X/Y/ScreenX/ScreenY
AtkResNode* byIndex = addon->UldManager.NodeList[15];         // raw array, NodeListCount entries; also addon->UldManager.Nodes (Span<Pointer<AtkResNode>>), UldManager.SearchNodeById(id)
if (node == null) return;
if (node->Type == NodeType.Text) node->GetAsAtkTextNode()->NodeText.ToString();
if ((ushort)node->Type >= 1000) { var comp = node->GetAsAtkComponentNode()->Component; comp->GetTextNodeById(3); comp->OwnerNode; comp->UldManager; comp->GetComponentType(); }
```

Crossing to and from Dalamud (verified against Dalamud `43b3411`; details in
the `dalamud` skill):

- Dalamud hands out wrapper structs, not CS pointers: `AtkUnitBasePtr`,
  `AgentInterfacePtr`, `UIModulePtr`, `AtkValuePtr` (`Dalamud.Game.NativeWrapper`).
  Get the CS pointer with `(AtkUnitBase*)wrapper.Address`; going the other way
  `AtkUnitBasePtr` converts implicitly from `nint` and `void*`.
  `IGameGui.GetAddonByName<T>(name, index)` returns `T*` directly.
- `AtkUnitBasePtr.IsReady` / `.IsVisible` read the CS `IsReady` / `IsVisible`
  properties (with a null check), so they match `addon->IsReady`, not the
  ECommons `IsAddonReady` extension described above.
- To observe `OnSetup`/`OnRefresh`/`ReceiveEvent`/`Draw`/… use
  `IAddonLifecycle.RegisterListener` (and `IAgentLifecycle` for agents)
  instead of hooking the vfunc. Since API 14 that service replaces the
  vtable of each live addon with its own copy, so `addon->VirtualTable` on a
  live addon is Dalamud's table, not the game's:
  `IAddonLifecycle.GetOriginalVirtualTable(nint)` returns the original.
  `StaticVirtualTablePointer` is still the game's table.
- Several Dalamud enums are generated copies of CS enums (`Dalamud/EnumCloneMap.txt`,
  32 entries): `AddonEventType` = `AtkEventType`, `GameInventoryType` =
  `InventoryType`, `Dalamud…ObjectKind` = `Client.Game.Object.ObjectKind`,
  `AgentId`, `AtkValueType`, the job gauge enums, the Party Finder flags.
  Their member lists live here: every source file is in `references/structs/`.

Component nodes have `Type >= 1000`; `GetNodeType()` reports 10000 for
them. `AtkComponentNode.Component` → `AtkComponentBase` (own `UldManager`,
`OwnerNode`, `GetNodeById/GetTextNodeById/GetImageNodeById`, `SetEnabledState`),
concrete: `AtkComponentButton` (`ButtonTextNode`, `IsEnabled`, `SetText`,
`SetChecked`), `AtkComponentCheckBox`, `AtkComponentList` (`ListLength`,
`SelectedItemIndex`, `HoveredItemIndex`, `ItemRendererList`, `CallBackInterface`),
`AtkComponentDropDownList`, `AtkComponentTextInput`, `AtkComponentNumericInput`,
`AtkComponentScrollBar`, `AtkComponentWindow`, `AtkComponentIcon`,
`AtkComponentDragDrop`, `AtkComponentTreeList`, `AtkComponentRadioButton`,
`AtkComponentGaugeBar`, `AtkComponentSlider`, `AtkComponentTab`, … (enum
`ComponentType`). `AtkResNode.GetAsAtkComponentX()` casts with a type check.

### AtkValue and callbacks

`AtkValue` (0x10): `Type` (`AtkValueType`: `Undefined=0, Null=1, Bool=2,
Int=3, Int64=4, UInt=5, UInt64=6, Float=7, String=8, WideString=9,
ConstString=0xA, Vector=0xB, Pointer=0xC, AtkValues=0xD, Managed=0x20,
ManagedString=0x28, ManagedVector=0x2B`) and a union at 0x8 (`Bool`, `Byte`,
`Int`, `Int64`, `UInt`, `UInt64`, `Float`, `String` (CStringPointer),
`WideString`, `Vector`, `Pointer`, `AtkValues`). Helpers: `SetBool/SetInt/
SetUInt/SetFloat` (release managed memory first), `SetManagedString(string)`,
`SetString(byte*)`, `ChangeType`, `Dtor()`, `GetValueAsString()`.

```csharp
// Fire a callback = what clicking does. Values are addon-specific (dump with an AtkValue/callback logger).
using var values = new RentedAtkValues(2);        // FFXIVClientStructs.Interop; allocates via MemoryPool.Alloc, Dtor()s + frees on Dispose
values[0].SetInt(0); values[1].SetInt(index);
addon->FireCallback((uint)values.Span.Length, values, close: false);   // bool FireCallback(uint valueCount, AtkValue* values, bool close = false)
addon->FireCallbackInt(-1);                                           // single-int shortcut
addon->Close(fireCallback: true);  addon->Hide(disableHideTransition: false, callCloseCallback: true, setShowHideFlags: 0);
```

`FireCallback` routes to the addon's owning `AtkEventInterface.ReceiveEvent`
(the agent) via the addon's event listener; `FireCallbackWithResult` uses
`ReceiveEventWithResult` and returns the agent's `AtkValue*`.
`ShouldFireCallbackAndHideOrClose` + `close: true` is the "callback then
close" path.

ECommons `Callback.Fire(AtkUnitBase* addon, bool updateState, params object[]
values)` (source in `references/ecommons/Callback.cs`, copied from ECommons,
not from the ClientStructs checkout) is a thin wrapper: it `Marshal.AllocHGlobal`s
`AtkValue[values.Length]`, maps `uint→UInt`, `int→Int`, `float→Float`,
`bool→Bool (Byte)`, `string→String (UTF-8, HGlobal)`, passes raw `AtkValue`
through, calls `addon->FireCallback((uint)n, ptr, updateState)`, frees.
`Callback.Fire(addon, true, -1)` closes most addons; the `updateState` bool
is CS's `close` parameter. `Callback.InstallHook()` hooks
`AtkUnitBase.MemberFunctionPointers.FireCallback` with
`Hook<AtkUnitBase.Delegates.FireCallback>` to log all callbacks — the
standard way to discover value layouts.

Addon-specific typed views exist for some addons
(`AddonSelectYesno.StandardTypedAtkValues`, `CollectibleTypedAtkValues`).
`AddonSelectString` exposes `PopupMenu` (`->PopupMenu.EntryNames`,
`EntryCount`).

