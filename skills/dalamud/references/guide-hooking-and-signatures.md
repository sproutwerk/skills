## Hooking and signatures

Priority (docs `interaction/index.md`): Dalamud service → ClientStructs member → own signature.

```csharp
using Dalamud.Hooking;
using Dalamud.Utility.Signatures;
using FFXIVClientStructs.FFXIV.Client.UI.Misc;

public sealed unsafe class MacroWatcher : IDisposable
{
    private readonly IPluginLog log;
    private readonly Hook<RaptureMacroModule.Delegates.SetSavePendingFlag> saveHook;

    // Function pointer resolved from a signature; stays null when the sig fails because the type is nullable.
    private delegate byte IsQuestCompletedDelegate(ushort questId);
    [Signature("E8 ?? ?? ?? ?? 41 88 84 2C")]
    private readonly IsQuestCompletedDelegate? isQuestCompleted = null;

    public MacroWatcher(IGameInteropProvider interop, IPluginLog log)
    {
        this.log = log;
        interop.InitializeFromAttributes(this);
        saveHook = interop.HookFromAddress<RaptureMacroModule.Delegates.SetSavePendingFlag>(
            RaptureMacroModule.MemberFunctionPointers.SetSavePendingFlag, Detour);
        saveHook.Enable();                       // creating a hook does not activate it for you
    }

    private void Detour(RaptureMacroModule* self, bool needsSave, uint set)
    {
        try { log.Information("macro save, set {Set}", set); }
        catch (Exception ex) { log.Error(ex, "detour failed"); }
        saveHook.Original(self, needsSave, set); // call through unless you mean to block
    }

    public void Dispose() => saveHook.Dispose(); // disables and frees
}
```

Create hooks only through `IGameInteropProvider` (the static `Hook<T>.FromAddress` etc. are
`internal`). `[Signature]` works on fields **or** properties; with the default `Fallibility` a
nullable member is fallible (warning, member left unset) and a non-nullable one throws
`SignatureException`.

`ScanText` follows the rel32 automatically when the match starts with `E8` or `E9` and throws if the
target is outside `.text`. Hooks created by `[Signature]` are not enabled for you; call `Enable()`
(upstream examples always do). A managed exception escaping a detour crashes the game; wrap the body
in try/catch. Several plugins may hook one function; the last to enable runs first, so call
`Original` and avoid changing arguments. `Hook<T>` members, `AsmHook`, the `[Signature]` property
table → `references/topic-hooking.md`

### ClientStructs interop

- `AddonEventType` is generated from ClientStructs `AtkEventType`
  (`references/dalamud/Dalamud/EnumCloneMap.txt` lists all 32 cloned enums, e.g. `ObjectKind`,
  `GameInventoryType`, `AtkValueType`); member lists are in ClientStructs, not in the Dalamud
  source. The `clientstructs` skill carries the ClientStructs source files at its own pin
  (`references/structs/FFXIV/Component/GUI/AtkEvent.cs` for `AtkEventType`, where `MouseClick = 9`).
- Wrapper ↔ ClientStructs pointer: `AtkUnitBasePtr`, `AgentInterfacePtr`, `UIModulePtr` and
  `AtkValuePtr` all expose `Address` and `IsNull`, convert implicitly to `nint`, and convert
  implicitly from `nint` and `void*`. Cast `(AtkUnitBase*)ptr.Address` to work on the struct.
  `AtkUnitBasePtr.IsReady` is the ClientStructs `IsReady` property, which is not the same test as
  the ECommons `IsAddonReady` extension (`clientstructs` skill, Addon work).

