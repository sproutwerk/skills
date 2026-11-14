# AutoDuty stage reflection

Covers: reading another plugin's state with `DalamudReflector` (the
AutoDuty `Stage` pattern) and the `GetFoP` / `Call` helper family.

Verified against: Splatoon 3.9.2.32 (commit f396ef6, 2026-11-13), ECommons
bc38dc7, Lifestream b363201 — the pins in `references/provenance.md`.

Paths beginning with `references/` are relative to the skill root.

## Poking other plugins by reflection (AutoDuty pattern)

```csharp
using ECommons.Reflection;

public enum Stage : int { Invalid = -1, Stopped = 0, Reading_Path = 1, Action = 2, Looping = 3, Condition = 4,
                          Moving = 5, Waiting_For_Combat = 6, Paused = 7, Dead = 8, Revived = 9, Interactable = 10 }

public Stage GetADStage()
{
    if(DalamudReflector.TryGetDalamudPlugin("AutoDuty", out var instance, suppressErrors: false, ignoreCache: true))
        return instance.GetFoP<Stage>("Stage");   // reads AutoDuty.Stage field/property by name
    return Stage.Invalid;                          // plugin absent → treat as Invalid, never as Stopped
}
void PauseAD()  => Svc.Commands.ProcessCommand("/autoduty pause");
void ResumeAD() => Svc.Commands.ProcessCommand("/autoduty resume");
```

- `TryGetDalamudPlugin(internalName, out object instance, bool
  suppressErrors, bool ignoreCache)`; `GetFoP(name)`, `GetFoP<T>(name)`,
  `SetFoP(name, value)`, `Call(name, object[] args)`, `GetStaticFoP(type,
  name)`, `CallStatic(type, name, args)`. Public, private, static are all
  searched (`AllFlags`), walking base types.
- `GetFoP<T>` is `(T)(object)` — a boxed enum unboxes to any enum with the
  same underlying type, which is why the locally redefined `Stage : int`
  works. Keep the values in sync with AutoDuty's enum; they are not verified
  against AutoDuty's source here.
- AutoDuty also exposes EzIPC: `[EzIPC] Func<bool> IsStopped` with
  `EzIPC.Init(this, "AutoDuty", SafeWrapper.AnyException)`
  (AutoDutyEnableTankStance). Prefer IPC when the plugin offers it.
- `DalamudReflector.GetService("Dalamud.Game.Gui.GameGui")` reaches Dalamud
  internals (the official DisableHandleActionOutHook script does this; it is
  not bundled here).
