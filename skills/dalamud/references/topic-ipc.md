Dalamud topic: IPC call gates (provider and subscriber), IPC errors and DataShare. Verified against `goatcorp/Dalamud` `43b3411` (2026-11-10, `DalamudVersion 15.0.3.9`, API 15), `goatcorp/dalamud-docs` `3c7597d` (2026-11-13), `goatcorp/SamplePlugin` `b8477da` (2026-08-15) and `aers/FFXIVClientStructs` `80cb4a7` (2026-11-10). Reviewed 2026-11-13.

Paths such as `references/...` and `references/topic-*.md` are relative to the skill root (the directory that holds `SKILL.md`), not to this file. Part of the `dalamud` skill; items marked *(unverified)* are not confirmed from the pinned checkouts.

## IPC

Call gates are typed by generic arguments: up to 8 parameters plus `TRet` last. For actions with no return value the
convention is `TRet = object`.

```csharp
// Provider plugin
private readonly ICallGateProvider<string, bool> isKnown;      // Func<string, bool>
private readonly ICallGateProvider<uint, object> changed;      // event with one uint payload

isKnown = pluginInterface.GetIpcProvider<string, bool>("MyPlugin.IsKnown");
isKnown.RegisterFunc(name => known.Contains(name));
changed = pluginInterface.GetIpcProvider<uint, object>("MyPlugin.Changed");
changed.SendMessage(42);                                       // notify subscribers
// Dispose: isKnown.UnregisterFunc(); (UnregisterAction() for RegisterAction)

// Consumer plugin
var sub = pluginInterface.GetIpcSubscriber<string, bool>("MyPlugin.IsKnown");
try { var ok = sub.InvokeFunc("x"); }
catch (IpcNotReadyError) { /* provider not loaded or not registered yet */ }

var ev = pluginInterface.GetIpcSubscriber<uint, object>("MyPlugin.Changed");
ev.Subscribe(OnChanged);     // Action<uint>
ev.Unsubscribe(OnChanged);   // must be the same delegate instance
```

| Side | Members |
|---|---|
| `ICallGateProvider<…>` | `RegisterAction(Action<…>)`, `RegisterFunc(Func<…, TRet>)`, `UnregisterAction()`, `UnregisterFunc()`, `SendMessage(…)`, `SubscriptionCount`, `GetContext()` → `IpcContext?` (`SourcePlugin`), valid only inside a handler |
| `ICallGateSubscriber<…>` | `InvokeAction(…)`, `InvokeFunc(…)`, `Subscribe(Action<…>)`, `Unsubscribe(Action<…>)`, `HasAction`, `HasFunction` |

Errors (`Dalamud.Plugin.Ipc.Exceptions`, all derive `IpcError`): `IpcNotReadyError`, `IpcTypeMismatchError`,
`IpcLengthMismatchError`, `IpcValueNullError`. Calls run synchronously on the **caller's thread**. Primitives pass
as-is; complex types are passed "with appropriate serialization" (doc comment), since each plugin has its own load
context, so share only primitives, strings, tuples/collections of them, or types from assemblies Dalamud itself loads.
Names are global strings: prefix with your InternalName.

**DataShare** (cheaper than IPC for shared reference data): `GetOrCreateData<T>(tag, Func<T> generator)`,
`TryGetData<T>(tag, out T?)`, `GetData<T>(tag)` (throws `KeyNotFoundException`), `RelinquishData(tag)`. `T` must be a
reference type shared through Dalamud itself, not defined in a plugin. Reference counted per consumer; the value is
disposed (if `IDisposable`) when the last consumer relinquishes. Errors: `DataCacheTypeMismatchError`,
`DataCacheValueNullError`, `DataCacheCreationError`.
