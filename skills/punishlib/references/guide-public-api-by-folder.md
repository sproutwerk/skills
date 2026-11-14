## Public API by folder

### Root, namespace `PunishLib`

| Type | Members |
|---|---|
| `PunishLibMain`, `AboutPlugin`, `PunishOption` | see above |
| `PunishConfig` | `string APIKey`, `string FFXIVGameVersion`, `bool UnsupportedConfiguration` (all `{ get; set; }`) |
| `static PunishConfigMethods` | `static string SavedPath { get; }` = `%AppData%/PunishConfig/PunishConfig.json`; `static void Save(this PunishConfig config)`; `static PunishConfig Load()` |
| `static GenericHelpers` | `static uint ToUint(this Vector4 color)`; `static void Safe(System.Action a, bool suppressErrors = false)`; `static bool TryGetFirst<K, V>(this IDictionary<K, V> dictionary, Func<KeyValuePair<K, V>, bool> predicate, out KeyValuePair<K, V> keyValuePair)` |
| `static Motd` | no members; the whole body is commented out |

`PunishConfig` is loaded by `Init` into `PunishLibMain.PunishConfig` and then not used by the
library again; the About tab reads and writes the API key through the internal `SharedConfig`
instead, which is a different file. `Safe` logs `e.Message` and the stack trace through
`Svc.Log.Error` unless `suppressErrors` is true.

### `ImGuiMethods/`, namespace `PunishLib.ImGuiMethods`

```csharp
public static class AboutTab    { public static void Draw(string pluginName); }
public static class IconButtons { public static bool IconTextButton(FontAwesomeIcon icon,
                                    string text, Vector2 size = new(), bool iconOnRight = false); }

public class InfoBox   // properties are all { get; set; }, defaults in brackets
{
    public Vector4 BorderColor;   // [ImGuiColors.DalamudGrey3]
    public Vector4 LabelColor;    // [ImGuiColors.DalamudWhite]
    public Action ContentsAction; // [draws "Action Not Set"]
    public string Label;          // ["Label Not Set"]
    public Vector2 Size, Offset;  // [Vector2.Zero]
    public float CurveRadius, BorderThickness;   // [15.0f], [2.0f]
    public int SegmentResolution; // [10]
    public bool AutoResize, Debug;               // [true], [false]
    public void Draw();
    public void DrawCentered(float percentSize = 0.80f);
    public void DrawStretched();
    public static ReadOnlyDictionary<string, InfoBox> InfoBoxCacheRO { get; }
    public static void DrawBox(string label, Action action);
}

public static partial class ImGuiGroup
{
    public record GroupBoxOptions   // { get; init; } unless noted, defaults in brackets
    {
        public bool Collapsible;          // [false]
        public Action HeaderTextAction;   // [null]
        public uint HeaderTextColor, BorderColor;  // [ImGui.GetColorU32 of ImGuiCol.Text / .Border]
        public Vector2 BorderPadding;     // [ImGui.GetStyle().WindowPadding]
        public float BorderRounding;      // [ImGui.GetStyle().FrameRounding]
        public ImDrawFlags DrawFlags;     // [ImDrawFlags.None]
        public float BorderThickness;     // [2f]
        public float Width, MaxX;         // { get; set; }, written by BeginGroupBox
    }
    public static bool BeginGroupBox(string id = null, float minimumWindowPercent = 1.0f,
                                     GroupBoxOptions options = null);
    public static bool BeginGroupBox(string text, GroupBoxOptions options);
    public static bool BeginGroupBox(uint borderColor, float minimumWindowPercent = 1.0f);
    public static unsafe void EndGroupBox();
    public static unsafe ImGuiWindow* GetCurrentWindow();
    public static unsafe ImGuiWindowFlags GetCurrentWindowFlags();
    public static unsafe bool CurrentWindowHasCloseButton();
}

[StructLayout(LayoutKind.Explicit)]
public struct ImGuiWindow   // PunishLib's own view: Flags @0xC, HasCloseButton @0xD5 (byte),
{ /* ... */ }               // CursorMaxPos @0x130 (Vector2)
```

Behaviour notes:

- `AboutTab.Draw` draws, centered: `"{pluginName} - {Manifest.AssemblyVersion}"`; the header
  line (see Branding); the plugin icon from `Manifest.IconUrl` at an unscaled 200x200; a Discord
  invitation line; then the buttons `Discord`, `Repository` (copies the Puni.sh repo URL to the
  clipboard, opens nothing), `Source Code` (only when `Manifest.RepoUrl` is not null), `Sponsor`
  (only when `About.Sponsor` is not null) and a cog that opens a separate ImGui window titled
  `Puni.sh API Key Settings`.
- `IconTextButton` draws an empty-label button and overlays icon and text. Its ImGui ID is
  `"###" + icon.ToIconString() + text`, so two buttons with the same icon and text in one ID scope
  collide. With `size == Vector2.Zero` the button is sized to the content.
- `InfoBox.DrawBox(label, action)` creates one `InfoBox` per distinct `label`, caches it in a
  static dictionary and calls `DrawStretched()`. The visible label is the part before `##`.
- `BeginGroupBox` always pushes state and, when it returns false (collapsed), has already called
  `EndGroupBox()` itself. Call `EndGroupBox()` only when it returned true.
  `minimumWindowPercent <= 0` makes the box auto-size to its content.
- `GetCurrentWindow` is a `[LibraryImport("cimgui")]` P/Invoke to `igGetCurrentWindow`; Dalamud's
  binding uses the same library name and export. Whether offsets 0xD5 and 0x130 match the cimgui
  shipped with API 15 is *(unverified)*. `ImGuiP.GetCurrentWindow()` in `Dalamud.Bindings.ImGui`
  is the stock alternative.

Internal to the library (not callable from a plugin): `ImGuiEx` with
`ImGuiLineCentered(string id, Action func)`, `Text(string s)`, `Text(Vector4 col, string s)`;
and `ImageLoadingResult`, which nothing references.

### `Configuration/`, namespace `PunishLib.Configuration`

```csharp
public interface IEzConfig { }
[AttributeUsage(AttributeTargets.Class | AttributeTargets.Struct)]
public class IgnoreDefaultValueAttribute : Attribute { }

public static class EzConfig
{
    public static string DefaultConfigurationFileName { get; }   // <config dir>/DefaultConfig.json
    public static IEzConfig Config { get; private set; }
    public static T Init<T>() where T : IEzConfig, new();
    public static void Migrate<T>() where T : IEzConfig, new();
    public static void Save();
    public static void SaveConfiguration(this IEzConfig Configuration, string path,
                                         bool indented = false, bool appendConfigDirectory = true);
    public static T LoadConfiguration<T>(string path, bool appendConfigDirectory = true)
        where T : IEzConfig, new();
}

public enum SharedConfigKey
{ APIKey, FFXIVGameVersion, UnsupportedConfiguration, InstalledPlugins, ClientLanguage }
public class SharedConfigStorage : IEzConfig
{ public Dictionary<int, string> Strings = new(); public Dictionary<int, bool> Bools = new(); }
```

- Files live in `Svc.PluginInterface.GetPluginConfigDirectory()`; serialization is Newtonsoft.Json.
  `Save()` writes indented JSON. Saving writes `<path>.new` and then moves it over `<path>`; a
  leftover `.new` from a failed save is renamed to `.new.<unix ms>` first.
- `[IgnoreDefaultValue]` on the config class switches `DefaultValueHandling` to `Ignore`.
- `Migrate<T>()` must run **before** `Init<T>()`; if `Config` is already set it throws a
  `NullReferenceException` (sic). It converts Dalamud's `<InternalName>.json` to
  `DefaultConfig.json` and renames the old file to `.old`.
- `LoadConfiguration` returns `new T()` when the file is missing or deserializes to null, and
  uses `ObjectCreationHandling.Replace`.
- This is a frozen subset of ECommons' `ECommons.Configuration.EzConfig`. No consumer read uses
  the PunishLib copy: Saucy's `EzConfig.Migrate` / `EzConfig.Init` calls resolve to
  `using ECommons.Configuration`. Prefer the ECommons one (skill `ecommons`).

### Internal Puni.sh services (for understanding, not for calling)

| Type | What it does |
|---|---|
| `internal class SharedConfig` (`Configuration/SharedConfig.cs`) | Properties `APIKey`, `FFXIVGameVersion`, `UnsupportedConfiguration`, `InstalledPlugins`, `ClientLanguage`. Persists to `%AppData%/PunishConfig/PunishSharedConfig.json` and mirrors each value into Dalamud DataShare under tag `PunishLib.<(int)SharedConfigKey>` as a one-element `List<T>`, so all loaded Puni.sh plugins see one value. Only `APIKey` is read or written by live code. |
| `internal static class API` (`API/API.cs`, namespace `PunishLib.API`) | `Task<bool> ValidateKey()` sends a GET to `https://puni.sh/api/test/auth?authKey=<key>` and returns true on HTTP 200. Called from the `Test Key` button of the API-key window. |

There is no sponsor list, no contributor list, no logging wrapper and no IPC in this pin; logging
goes through ECommons' `Svc.Log`.

