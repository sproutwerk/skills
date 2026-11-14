# SplatoonAPI: drawing through the Splatoon plugin from a regular plugin

Covers: `ECommons.SplatoonAPI` (`Splatoon`, `Element`, `ElementType`, `RefActorType`,
`RefActorComparisonType`), the reflection bridge that lets a Dalamud plugin hand elements to an
installed Splatoon plugin. Verified against ECommons `bc38dc7` (2026-11-08, v3.2.1.26, Dalamud
API 15); nothing here was compiled. Paths are relative to the skill root.

Pointers: writing Splatoon scripts -> `splatoonscripts` skill; preset/Element JSON semantics (what
each field does when rendered) -> `splatoonpresets` skill;
`DalamudReflector` -> `references/topic-reflection-interop.md`.

What was checked where: every ECommons name and signature below is from the ECommons checkout. The
`Element` field names were cross-checked against the `Splatoon.Element` source copy in the
`splatoonpresets` skill (all 86 proxied fields exist there with the same types). The names ECommons
looks up on the Splatoon plugin object (`AddDynamicElements`, `RemoveDynamicElements`,
`InjectElement`, the `Init` field, the three IPC names) are *(unverified)* on the Splatoon side; no
Splatoon plugin source was available.

No code in the AutoRetainer, PandorasBox or YesAlready checkouts calls this API (PandorasBox
initialises it through `Module.All` and never uses it). Examples are assembled from the signatures.

```csharp
using ECommons;
using ECommons.SplatoonAPI;
```

Inside a namespace that starts with `Splatoon`, or next to a reference to the Splatoon assembly, the
names `Splatoon` and `Element` are ambiguous; qualify them as `ECommons.SplatoonAPI.Splatoon`.

## Requirements and lifecycle

Sources: `references/ecommons/ECommons/ECommonsMain.cs`,
`references/ecommons/ECommons/SplatoonAPI/Splatoon.cs`.

```csharp
ECommonsMain.Init(pluginInterface, this, Module.SplatoonAPI);   // or Module.All
```

- `Module.SplatoonAPI` (or `Module.All`) makes `ECommonsMain.Init` run both
  `DalamudReflector.Init()` and the internal `Splatoon.Init()`. `Module.DalamudReflector` does not
  need to be listed separately.
- The Splatoon plugin must be installed and loaded by the user. ECommons does not install it.
- `Splatoon.Init()`, `Splatoon.Reset()` and `Splatoon.Shutdown()` are `internal`. Plugin code
  cannot call them; they are listed here only to describe behaviour.

| Internal step | When | What it does |
|---|---|---|
| `Init()` | `ECommonsMain.Init` with the module | Invokes IPC func `Splatoon.IsLoaded` (`bool`); if true, connects. Exceptions (Splatoon absent, IPC not registered) are swallowed. Then subscribes to IPC `Splatoon.Loaded` -> connect and `Splatoon.Unloaded` -> reset. |
| connect (private `Connect()`) | At init if Splatoon is loaded, and every time Splatoon (re)loads | `DalamudReflector.TryGetDalamudPlugin("Splatoon", out var plugin, false, true)`, requires the plugin's public `Init` field to be true, sets `Instance`, increments `Version`, invokes the `SetOnConnect` action, logs `Successfully connected to Splatoon.` |
| `Reset()` | IPC `Splatoon.Unloaded`; `ECommonsMain.Dispose()` when `Instance != null` | Sets `Instance = null`, logs `Disconnected from Splatoon`. |
| `Shutdown()` | `ECommonsMain.Dispose()` | Unsubscribes the two IPC handlers. |

`ECommonsMain.Dispose()` does not remove the dynamic elements your plugin added. Call
`Splatoon.RemoveDynamicElements(name)` for each namespace you used before disposing ECommons.

## Splatoon (static class)

`public static class Splatoon`, namespace `ECommons.SplatoonAPI`.

```csharp
public static object Instance { get; internal set; }   // the Splatoon plugin object, or null
public static int Version { get; internal set; }       // connection generation counter

public static void SetOnConnect(Action action)
public static bool IsConnected()

public static bool AddDynamicElement(string name, Element e, long[] DestroyCondition)
public static bool AddDynamicElement(string name, Element e, long DestroyCondition)
public static bool AddDynamicElement(string name, Element e, float[] DestroyCondition)
public static bool AddDynamicElement(string name, Element e, float DestroyCondition)
public static bool AddDynamicElements(string name, Element[] e, long DestroyCondition)
public static bool AddDynamicElements(string name, Element[] e, float DestroyCondition)
public static bool AddDynamicElements(string name, Element[] e, float[] DestroyConditionF)
public static bool AddDynamicElements(string name, Element[] e, long[] DestroyCondition)

public static bool RemoveDynamicElements(string name)
public static bool DisplayOnce(Element e)
public static Element DecodeElement(string input)

[Obsolete("Work in progress")]
public static object DecodeLayout(string input)
```

There is one `RemoveDynamicElements` method and no singular `RemoveDynamicElement`.

### Members

| Member | Behaviour |
|---|---|
| `IsConnected()` | `Instance != null`. |
| `SetOnConnect(action)` | Stores the action (one slot; a second call replaces the first). Then invokes IPC `Splatoon.IsLoaded` and, if true, runs the action immediately. The action also runs on every later connect. |
| `AddDynamicElements(name, e, long[])` | The overload all others funnel into. Guards in order: not connected -> warning `Not connected to Splatoon`, false; any element with `IsValid() == false` -> warning `Elements are no longer valid`, false; empty array -> warning `There are no elements`, false. Then builds an array of the Splatoon element type and invokes `AddDynamicElements(name, array, DestroyCondition)` on the plugin. Exceptions are logged (`ex.Log()`) and give false. |
| `RemoveDynamicElements(name)` | Removes every element registered under `name`. Not connected -> warning and false. Exceptions logged, false. |
| `DisplayOnce(e)` | Invokes `InjectElement(element)` on the plugin: the element is shown for the next frame only. Call it every frame while the element should stay visible. Same guards and error handling as above. |
| `DecodeElement(input)` | Calls the static `Splatoon.SplatoonScripting.ScriptingEngine.TryDecodeElement(input, out element)` and wraps the result in an `Element`; returns null when decoding fails. `input` is an element export string copied from Splatoon; the source comment says to make sure the element is enabled. No connection guard: throws `NullReferenceException` when not connected. |
| `DecodeLayout(input)` | Obsolete, "Work in progress", no replacement named. Returns the raw Splatoon layout object or null; nothing in ECommons accepts it. |

### `name`

A non-unique namespace string. Several elements can be added under the same name, in one call or
in several; `RemoveDynamicElements(name)` removes all of them. The name is passed to Splatoon
as-is, so prefix it with your plugin name to avoid collisions with other plugins.

### Destroy conditions

Each entry of the array is one condition (the source comment does not say how several are
combined; by the wording "array of them" any of them applies *(unverified)*).

| Value | Meaning |
|---|---|
| `-2` | destroy on zone change |
| `-1` | destroy on combat end |
| `0` | never destroy |
| other, `long` overloads | absolute tick at which to destroy, on the `Environment.TickCount64` clock |
| other, `float` overloads | seconds from now |

The `float` overloads convert with
`x > 0 ? ((long)(x * 1000f) + Environment.TickCount64) : (long)x`, so `-2f`, `-1f` and `0f` keep
their special meaning.

An `int` argument selects the `long` overload by C# overload resolution (`long` converts implicitly
to `float`, not the reverse, which makes `long` the better target). `AddDynamicElement(name, e, 5)`
therefore means "destroy at tick 5", which is in the past. Write `5f` for five seconds.

## Element

`public class Element`, namespace `ECommons.SplatoonAPI`.
Source: `references/ecommons/ECommons/SplatoonAPI/Element.cs`.

```csharp
public Element(ElementType type)
public bool IsValid()
public void SetRefCoord(Vector3 v)     // System.Numerics.Vector3
public void SetOffCoord(Vector3 v)
```

- The constructor creates a `Splatoon.Element` in the Splatoon assembly through
  `Assembly.CreateInstance("Splatoon.Element", ..., new object[] { (int)type }, ...)` and records
  the current `Splatoon.Version`. It dereferences `Splatoon.Instance`, so constructing an `Element`
  while not connected throws `NullReferenceException`.
- The wrapped object is held in an `internal object Instance` field. Plugin code cannot reach it,
  so only the properties below can be set from code. Anything else has to come in through
  `Splatoon.DecodeElement` from an export string.
- `IsValid()` is `Version == Splatoon.Version`. Every reconnect increments the version, which
  invalidates all existing `Element` objects. Invalid elements are rejected by
  `AddDynamicElements` and `DisplayOnce`; their properties still read and write the old object.
- `SetRefCoord(v)` sets `refX = v.X`, `refY = v.Z`, `refZ = v.Y`. `SetOffCoord(v)` does the same for
  `offX`, `offY`, `offZ`. Splatoon's `Y` is the game's `Z` and the reverse; pass a game-space
  position (for example an object's `Position`) and let these methods swap.
- Every property is `{ get; set; }` and proxies the public field of the same name and the same
  spelling on `Splatoon.Element`, by `Instance.GetType().GetField("<name>")` on each access. The
  casing is inconsistent because it follows Splatoon's field names. Reads and writes are reflection
  calls; set properties once when building, not per frame.
- There is no wrapper for the element `Name` or for any Splatoon field that is not in this table.

### Properties

All properties proxy the Splatoon field with the identical name. Three are typed as ECommons enums
over an `int` field.

| Property | Type | Splatoon field type |
|---|---|---|
| `type` | `ElementType` | `int` |
| `refActorType` | `RefActorType` | `int` |
| `refActorComparisonType` | `RefActorComparisonType` | `int` |

Shape, position, colour:

| Type | Properties |
|---|---|
| `bool` | `Enabled`, `Filled`, `tether`, `FaceMe` |
| `float` | `refX`, `refY`, `refZ`, `offX`, `offY`, `offZ`, `radius`, `Donut`, `thicc`, `FillStep`, `AdditionalRotation` |
| `int` | `coneAngleMin`, `coneAngleMax` |
| `uint` | `color` |

Overlay text:

| Type | Properties |
|---|---|
| `string` | `overlayText` |
| `uint` | `overlayBGColor`, `overlayTextColor` |
| `float` | `overlayVOffset`, `overlayFScale` |
| `bool` | `overlayPlaceholders` |

Reference actor selection:

| Type | Properties |
|---|---|
| `string` | `refActorName`, `refActorVFXPath` |
| `uint` | `refActorModelID`, `refActorObjectID`, `refActorDataID`, `refActorNPCID`, `refActorNPCNameID`, `refActorObjectEffectData1`, `refActorObjectEffectData2` |
| `List<string>` | `refActorPlaceholder` |
| `bool` | `refActorComparisonAnd`, `onlyTargetable`, `onlyUnTargetable`, `onlyVisible` |
| `int` | `refActorVFXMin`, `refActorVFXMax`, `refActorObjectEffectMin`, `refActorObjectEffectMax` |

Reference actor conditions (cast, status, lifetime):

| Type | Properties |
|---|---|
| `bool` | `refActorRequireCast`, `refActorUseCastTime`, `refActorUseOvercast`, `refActorRequireBuff`, `refActorRequireAllBuffs`, `refActorRequireBuffsInvert`, `refActorUseBuffTime`, `refActorObjectLife` |
| `List<uint>` | `refActorCastId`, `refActorBuffId` |
| `float` | `refActorCastTimeMin`, `refActorCastTimeMax`, `refActorBuffTimeMin`, `refActorBuffTimeMax`, `refActorLifetimeMin`, `refActorLifetimeMax` |

Hitbox and rotation:

| Type | Properties |
|---|---|
| `bool` | `includeHitbox`, `includeOwnHitbox`, `includeRotation`, `LimitRotation` |
| `bool` | `LineAddHitboxLengthX`, `LineAddHitboxLengthY`, `LineAddHitboxLengthZ`, `LineAddHitboxLengthXA`, `LineAddHitboxLengthYA`, `LineAddHitboxLengthZA` |
| `bool` | `LineAddPlayerHitboxLengthX`, `LineAddPlayerHitboxLengthY`, `LineAddPlayerHitboxLengthZ`, `LineAddPlayerHitboxLengthXA`, `LineAddPlayerHitboxLengthYA`, `LineAddPlayerHitboxLengthZA` |
| `float` | `RotationMin`, `RotationMax` |

Distance limit:

| Type | Properties |
|---|---|
| `bool` | `LimitDistance`, `LimitDistanceInvert` |
| `float` | `DistanceSourceX`, `DistanceSourceY`, `DistanceSourceZ`, `DistanceMin`, `DistanceMax` |

Source doc comments on the few documented ones: `refX`/`refY`/`refZ` "When type is fixed coordinate
element: specifies coordinate of that element, otherwise has no effect"; `offX`/`offY`/`offZ`
"Specifies offset from targeted object or fixed coordinate". For what the remaining fields do, use
the `splatoonpresets` skill; the field names are the same as the JSON keys.

`List<>` properties return the live list of the Splatoon object, so
`e.refActorCastId.Add(id)` modifies the element; assigning a new list also works.

## Enums

Sources: `references/ecommons/ECommons/SplatoonAPI/ElementType.cs`, `RefActorType.cs`,
`RefActorComparisonType.cs` in the same folder.

```csharp
public enum ElementType : int
{
    CircleAtFixedCoordinates = 0,
    CircleRelativeToActorPosition = 1,
    LineBetweenTwoFixedCoordinates = 2,
    LineRelativeToObjectPosition = 3,
    ConeRelativeToObjectPosition = 4,
    ConeAtFixedCoordinates = 5
}

public enum RefActorType
{
    IGameObjectWithSpecifiedAttribute = 0,
    Self = 1,
    TargetedEnemy = 2
}

public enum RefActorComparisonType : int
{
    Name = 0,
    ModelID = 1,
    ObjectID = 2,
    DataID = 3,
    NpcID = 4,
    Placeholder = 5,
    NameID = 6,
    VfxPath = 7
}
```

The `Splatoon.Element` copy in the `splatoonpresets` skill lists two more comparison types after
index 7 ("Object Effect", "Icon ID"). The ECommons enum has no members for them although
`Element` exposes the `refActorObjectEffect*` fields; the underlying field is an `int`, so
`(RefActorComparisonType)8` is what reaches Splatoon if you need it.

## Examples

Create elements on connect, keep them for the plugin lifetime, remove them on dispose:

```csharp
public sealed class MyPlugin : IDalamudPlugin
{
    private const string Ns = "MyPlugin.Markers";

    public MyPlugin(IDalamudPluginInterface pi)
    {
        ECommonsMain.Init(pi, this, Module.SplatoonAPI);
        Splatoon.SetOnConnect(OnSplatoonConnect);
    }

    private void OnSplatoonConnect()
    {
        if(!Splatoon.IsConnected()) return;           // see Gotchas
        var circle = new Element(ElementType.CircleRelativeToActorPosition)
        {
            refActorType = RefActorType.Self,
            radius = 5f,
            Filled = false,
            thicc = 3f,
            color = 0xC800FF00,
        };
        Splatoon.AddDynamicElement(Ns, circle, 0L);   // 0 = never destroy
    }

    public void Dispose()
    {
        Splatoon.RemoveDynamicElements(Ns);           // before ECommonsMain.Dispose()
        ECommonsMain.Dispose();
    }
}
```

Timed marker at a world position:

```csharp
if(Splatoon.IsConnected())
{
    var e = new Element(ElementType.CircleAtFixedCoordinates) { radius = 1f };
    e.SetRefCoord(position);                          // game-space Vector3
    Splatoon.AddDynamicElement("MyPlugin.Ping", e, 10f);      // 10 seconds
    // or: new long[] { -2, Environment.TickCount64 + 10_000 }
}
```

Per-frame display of a cached element, rebuilt after a reconnect:

```csharp
private Element? cached;

private void OnFrameworkUpdate(IFramework _)
{
    if(!Splatoon.IsConnected()) return;
    if(cached == null || !cached.IsValid())
    {
        cached = Splatoon.DecodeElement(ExportString);        // null when the string is invalid
        if(cached == null) return;
    }
    Splatoon.DisplayOnce(cached);
}
```

## Behaviour when Splatoon is absent or the module is missing

| Situation | `IsConnected()` | `AddDynamicElement(s)`, `RemoveDynamicElements`, `DisplayOnce` | `new Element(...)`, `DecodeElement`, `DecodeLayout` | `SetOnConnect` action |
|---|---|---|---|---|
| Module requested, Splatoon not installed or not loaded | false | false, plus a `PluginLog.Warning` on every call | `NullReferenceException` | stored, not run until Splatoon loads |
| Module requested, Splatoon loads later | becomes true | work | work | runs on connect |
| Module requested, Splatoon unloaded or reloaded | false until reconnect | false with warning until reconnect; old `Element` objects rejected afterwards | as above while disconnected | runs again on reconnect |
| Module not requested | always false | false with warning | `NullReferenceException` | runs once immediately if Splatoon is loaded, with `Instance` still null |

Guard per-frame calls with `IsConnected()`; an unguarded `DisplayOnce` in a framework update writes
one warning per frame to the log while Splatoon is missing.

## Gotchas

- Create elements inside the `SetOnConnect` action or after an `IsConnected()` check. The
  constructor throws when not connected.
- An `Element` does not survive a Splatoon reload. Rebuild elements in the on-connect action, and
  check `IsValid()` on anything you cache.
- Exceptions thrown by your on-connect action are swallowed silently when the action runs from
  inside `SetOnConnect` (empty `catch`). When it runs from a later connect, the exception is logged
  under the misleading text `Can't find Splatoon plugin: <message>`, and the `Successfully
  connected` line is skipped although `Instance` is already set. Wrap the body in your own
  try/catch.
- `SetOnConnect` decides whether to run the action from the `Splatoon.IsLoaded` IPC, not from
  `IsConnected()`. If the connection attempt failed (or the module was not requested) the action
  still runs with `Instance == null`; hence the `IsConnected()` check in the example.
- Integer literals as destroy condition mean ticks, not seconds. Use an `f` suffix for seconds.
- Dynamic elements outlive your plugin unless you remove them or gave them a destroy condition.
- `name` is shared across all plugins that talk to Splatoon; use a unique prefix.
- Coordinates: `refY`/`offY` are the game's Z axis and `refZ`/`offZ` the game's Y (height). Use
  `SetRefCoord`/`SetOffCoord` instead of assigning the three fields by hand.
- Property names are case-sensitive copies of Splatoon's fields: `thicc`, `radius`, `color`, `type`
  and `tether` are lower case while `Enabled`, `Filled`, `Donut`, `FaceMe` and `FillStep` are not.
- If a Splatoon update renames or removes a field, the matching property throws
  `NullReferenceException` on access (the `GetField` result is null). Nothing checks this up front.
- This API is for plugins. Code running as a Splatoon script uses Splatoon's own types directly
  (`splatoonscripts` skill) and must not go through `ECommons.SplatoonAPI`.
