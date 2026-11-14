## Threading

The game is single-threaded around `Framework::Tick`. Game memory, `IObjectTable`, `StatusList`,
ClientStructs calls, addon access and most service getters must run on that thread.
`Framework.Update`, hook detours, `IAddonLifecycle`/`IAgentLifecycle` listeners, chat/toast/flytext
events and command handlers already do. `Task.Run`, timers, HTTP continuations, the plugin
constructor and `LoadAsync` do not.

| `IFramework` member | Behaviour |
|---|---|
| `event Update(IFramework framework)` | Every frame. Keep it cheap; throttle with `UpdateDelta` or a timestamp. |
| `IsInFrameworkUpdateThread` | Guard. |
| `RunOnFrameworkThread(Action)` / `<T>(Func<T>)` | Runs **immediately** if already on the framework thread, otherwise next tick. Safe to `.Wait()` on from another thread. The `Func<Task>` overloads are `[Obsolete]` → `RunOnTick`. |
| `RunOnTick(func, TimeSpan delay = default, int delayTicks = default, CancellationToken ct = default)` | Always deferred to a later tick, even when called on the framework thread. Overloads for `Action`, `Func<T>`, `Func<Task>`, `Func<Task<T>>`. |
| `Run(Action \| Func<T> \| Func<Task> \| Func<Task<T>>, CancellationToken ct = default)` | Like `RunOnFrameworkThread`, but `await` continuations inside the callback stay on the framework thread. |
| `DelayTicks(long numTicks, ct)` | Awaitable frame delay; continuation on the framework thread. |
| `GetTaskFactory()` | `TaskFactory` bound to the framework thread. |
| `CreateDebouncer(TimeSpan delay, Action action)` → `IDebouncer` (`Debounce()`, `IsPending`, `Dispose`) | |
| `LastUpdate`, `LastUpdateUTC`, `UpdateDelta`, `IsFrameworkUnloading` | |

```csharp
// Background work, then touch the game.
_ = Task.Run(async () =>
{
    var json = await http.GetStringAsync(url, lifecycle.LogoutToken);
    await framework.RunOnFrameworkThread(() =>
    {
        if (objectTable.LocalPlayer is { } player)
            chat.Print($"{player.Name.TextValue}: {json.Length} bytes");
    });
});
```

Deadlock rule (from the `IFramework` remarks): inside a `Run` callback the current `TaskScheduler`
is the framework thread, so `Task.Factory.StartNew` and `ContinueWith` also land there; calling
`.Wait()` / `.Result` on them locks the game. Use `await`, and `Task.Run` to leave the thread.

Assertions: `IObjectTable` and `StatusList` call `ThreadSafety.AssertMainThread()` (always throws
off-thread); object property getters use a dev-mode-only assert, so off-thread reads may work on
user machines and still be races. `Dalamud.Utility.ThreadSafety.IsMainThread` is public.

