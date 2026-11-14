OtterGui topic: the root-level helpers `ItemSelector<T>` / `ItemDetailsWindow`, `ImGuiClip`,
`ImGuiPie`, `ImGuiUtil` (`Util.cs`), `Functions` and the `Custom/` folder. (`ImGuiTable` is in
`references/topic-table.md`.) Verified against `Limiana/OtterGui` `a08fda6` (2026-05-30, branch `main`).
Against upstream `79771ee` only `ImGuiClip.ClippedTableDraw` and
`Custom/IndividualHelpers.ToName` differ; against the Artisan pin `32c4454` `ItemSelector<T>`
differs. All flagged below.

Paths such as `references/...` and `references/topic-*.md` are relative to the skill root (the directory that
holds `SKILL.md`), not to this file. Nothing was compiled.

All types here are in namespace `OtterGui` except the `Custom/` ones (`OtterGui.Custom`).

## ItemSelector<T> (`ItemSelector.cs`)

A vertical list with selection, an optional filter on top and an optional button row below (add,
import from clipboard, duplicate, delete), plus drag and drop reordering. This is the part of
OtterGui that Artisan uses most (`RecipeSelector`, `ListFolders`, `ExpertProfileList`).

```csharp
public class ItemSelector<T>
{
    [Flags]
    public enum Flags : byte
    {
        None = 0x00, Delete = 0x01, Add = 0x02, Import = 0x04, Duplicate = 0x08,
        Filter = 0x10, Move = 0x20, Drop = 0x40,
        ButtonMask = 0x0F, All = 0x7F,
    }

    public ItemSelector(IList<T> items, Flags flags = Flags.None);
    public void Draw(float width);

    public string Label { get; init; }               // default "##ItemSelector"
    public readonly string DragDropLabel;            // "<Label>DragDrop"
    public readonly string MoveLabel;                // "<Label>Move"
    public int CurrentIdx = 0;
    public T?  Current { get; private set; }
    public T?  EnsureCurrent();                      // re-resolve, select index 0 if nothing is selected
    public void TryRestoreCurrent();                 // re-find Current in Items after the list changed
    public void SetFilterDirty();
    public void CreateDropSource<TData>(TData data, string tooltip);
    public bool CreateDropTarget<TData>(Action<TData> action);
    public bool CreateDropTarget<TData>(Func<TData, bool> func);

    protected readonly IList<T>  Items;
    protected readonly List<int> FilteredItems;      // indices into Items
    protected string Filter      = string.Empty;
    protected bool   FilterDirty = true;
    protected void ClearCurrentSelection();
    protected void SetCurrent(T item);

    protected virtual bool OnDraw(int idx);                          // draw one row like a Selectable; true selects it
    protected virtual bool Filtered(int idx);                        // true hides the row
    protected virtual bool OnAdd(string name);
    protected virtual bool OnDelete(int idx);
    protected virtual bool OnDuplicate(string name, int idx);
    protected virtual bool OnClipboardImport(string name, string data);
    protected virtual bool OnMove(int idx1, int idx2);
    protected virtual void OnDrop(object? data, int idx);
    protected virtual bool   DeleteButtonEnabled();                  // default: ImGui.GetIO().KeyCtrl
    protected virtual string DeleteButtonTooltip();
}
```

Rules read from the code:

- Every `On...` hook and `Filtered` defaults to `throw new NotImplementedException()`. Override
  each hook that one of your flags enables. `OnDraw` is always required. `Filtered` is called even
  without `Flags.Filter`, so it is always required too.
- The selector does not modify `Items`. `OnAdd`, `OnDelete`, `OnDuplicate`, `OnMove` must change
  your list themselves and return `true` on success; the selector then fixes the selection and
  marks the filter dirty.
- Add, import and duplicate open a popup asking for a name (`ImGuiUtil.OpenNameField`); the hook
  runs when Enter is pressed. On fork `main` an empty name still calls `OnAdd("")`.
- Delete is a button that is only active while `DeleteButtonEnabled()` returns true.
- With `Flags.Move` each row is a drag source carrying its index and a drop target that calls
  `OnMove(from, to)`. With `Flags.Drop` rows accept payloads created by `CreateDropSource` and
  call `OnDrop(data, idx)`.
- A change in `Items.Count` re-runs the filter automatically; other changes need
  `SetFilterDirty()`.
- `Draw(width)` uses the full remaining height: the list is a child of size
  `(width, -ImGui.GetFrameHeight())` with the buttons below. Rows are clipped with
  `ImGui.GetTextLineHeightWithSpacing()`, so every row must have exactly that height.

```csharp
internal sealed class ProfileList : ItemSelector<Profile>        // fork main / upstream signature
{
    private readonly List<Profile> _profiles;

    public ProfileList(List<Profile> profiles)
        : base(profiles, Flags.Add | Flags.Delete | Flags.Duplicate | Flags.Move | Flags.Filter)
        => _profiles = profiles;

    protected override bool OnDraw(int idx)
        => ImGui.Selectable(Items[idx].Name, idx == CurrentIdx);

    protected override bool Filtered(int idx)
        => Filter.Length != 0 && !Items[idx].Name.Contains(Filter, StringComparison.OrdinalIgnoreCase);

    protected override bool OnAdd(string name)
    {
        if (name.Length == 0)
            return false;
        _profiles.Add(new Profile { Name = name });
        return true;
    }

    protected override bool OnDelete(int idx)
    {
        _profiles.RemoveAt(idx);
        return true;
    }

    protected override bool OnDuplicate(string name, int idx)
    {
        _profiles.Add(_profiles[idx] with { Name = name });
        return true;
    }

    protected override bool OnMove(int idx1, int idx2)
        => _profiles.Move(idx1, idx2);                    // OtterGui.Filesystem.Extensions
}

// Draw():
//   _list.Draw(200 * ImGuiHelpers.GlobalScale);
//   ImGui.SameLine();
//   ItemDetailsWindow.Draw("##details", DrawHeader, DrawSelected);
```

`Profile` is an assumed `record` with a `Name` property.

**Artisan pin only** (`32c4454`, also on the fork's `Artisan` branch): `OnDraw(int idx, out bool
changes)` replaces `OnDraw(int idx)`; `Items`, `ListFlags`, `SetCurrent(int)` and the `Current`
setter are public; events `ItemAdded`, `ItemDeleted`, `ItemSkipTriggered`; `AddButtonTooltip()`.
Full table in `references/topic-fork-delta.md`.

`ItemDetailsWindow`:

```csharp
public static void Draw(string label, Action drawHeader, Action drawDetails)
```

A bordered child filling the remaining content region, with `drawHeader` inside a menu bar and
`drawDetails` in a padded group below. It is the right-hand counterpart of `ItemSelector.Draw`.

## ImGuiClip (`ImGuiClip.cs`)

Clipping helpers for long lists. Two families.

Random-access data, using `ImGuiListClipper` (start and end dummies handled for you):

```csharp
public static void ClippedDraw<T>(IReadOnlyList<T> data, Action<T> draw, float lineHeight)
public static void ClippedDraw<T>(IReadOnlyList<T> data, Action<T, int> draw, float lineHeight)
```

Enumerables without random access, no state kept between frames:

```csharp
public static int  GetNecessarySkips(float height)
public static int  GetNecessarySkipsAtPos(float height, float cursorPosY, int maxCount = int.MaxValue)
public static int  ClippedDraw<T>(IEnumerable<T> data, int skips, Action<T> draw, int? count = null, int startIndex = 0)
public static int  ClippedTableDraw<T>(IEnumerable<T> data, int skips, Action<T> draw, int? count = null, int startIndex = 0)
public static int  FilteredClippedDraw<T>(IEnumerable<T> data, int skips, Func<T, bool> checkFilter, Action<T> draw, int startIndex = 0)
public static void DrawEndDummy(int remainder, float height)
```

Protocol for the second family, from the source comments:

```csharp
var height = ImGui.GetTextLineHeightWithSpacing();          // must include the item spacing
var skips  = ImGuiClip.GetNecessarySkips(height);           // draws the top dummy
var rest   = ImGuiClip.ClippedDraw(items, skips, DrawItem, items.Count);
ImGuiClip.DrawEndDummy(rest, height);                       // draws the bottom dummy
```

- The return value is either the non-negative number of items that were not drawn because they
  are below the visible area, or the **bitwise complement** (`~`) of the next start index when
  the data ran out while there was still room. Pass `~ret` as `startIndex` of the next
  `ClippedDraw` to chain several collections of the same row height; `DrawEndDummy` ignores
  negative values.
- `ClippedDraw(IEnumerable...)` wraps every item in `ImRaii.Group()` and `ImRaii.PushId(idx)`.
  `ClippedTableDraw` pushes only the ID. Upstream `79771ee` additionally wraps each
  `ClippedTableDraw` item in `ImUtf8.Group()` (commit `4a9b71a`, "Fix issue with clipping
  lists"); the fork does not.
- Visibility is decided from the rectangle of the last drawn item, so each `draw` call must
  produce an item or group of exactly `height`.
- All rows need the same height in both families.

## ImGuiPie (`ImGuiPie.cs`)

```csharp
public static void Draw(float radius, IReadOnlyList<(float Percentage, Action DrawTooltip, uint Color)> sections)
```

Draws a filled pie at the cursor. `Percentage` is a fraction of the full circle (`0.25f` is a
quarter); `DrawTooltip` of the section under the mouse is invoked (open the tooltip yourself
inside it). Afterwards the cursor is set to the right of the pie.

## ImGuiUtil (`Util.cs`)

`public static partial class ImGuiUtil`, the `string`-based predecessor of the `ImUtf8` extended
helpers. Several functions carry a `<seealso>` pointing at their `ImUtf8` replacement.

| Group | Signatures |
|---|---|
| Tooltips | `void HoverTooltip(string tooltip, ImGuiHoveredFlags flags = None)`; `void HoverTooltip(HoverTooltipStringHandler tooltip)`; `void HoverTooltip(ImGuiHoveredFlags flags, HoverTooltipStringHandler tooltip)` |
| Help markers | `void LabeledHelpMarker(string label, string tooltip)`; `void SelectableHelpMarker(string tooltip)`; `void HelpPopup(string label, Vector2 size, Action content)` |
| Text | `void TextWrapped(string text)`; `void TextColored(uint color, string text)`; `void BulletTextColored(uint color, string text)`; `void DrawColoredText(params (string, uint)[] data)`; `void TextShadowed(string text, uint foregroundColor, uint shadowColor, byte shadowWidth = 1)`; `void TextShadowed(ImDrawListPtr drawList, Vector2 position, string text, uint foregroundColor, uint shadowColor, byte shadowWidth = 1)`; `void RightAlign(string text, float offset = 0)`; `void RightJustify(string text, uint color)`; `void Center(string text)`; `void PrintIcon(FontAwesomeIcon icon)` |
| Table cells | `void DrawTableColumn(string text)`; `void DrawFrameColumn(string text)` |
| Buttons | `bool DrawDisabledButton(string label, Vector2 size, string description, bool disabled, bool icon = false)`; `bool DrawTextButton(string text, Vector2 size, uint buttonColor)`; `bool DrawTextButton(string text, Vector2 size, uint buttonColor, uint textColor)`; `bool DrawEditButtonText(int id, string current, out string newText, ref bool edit, Vector2 buttonSize, float inputWidth, int maxLength = 256)` |
| Inputs | `bool Checkbox(string label, string description, bool current, Action<bool> setter, ImGuiHoveredFlags flags = None)`; `bool DragFloat(string label, ref float value, float width, float speed, float min, float max, string format)`; `bool DragInt(string label, ref int value, float width, float speed, int min, int max, string format)`; `bool InputUlong(string label, ref ulong value, string format = "%llu", ImGuiInputTextFlags flags = None)`; `bool InputUInt16(string label, ref ushort v, ImGuiInputTextFlags flags)`; `bool GuidInput(string label, string hint, string tooltip, ref Guid? guid, ref string text, float width = 0)`; `bool OpenNameField(string popupName, ref string newName)` |
| Combos | `bool GenericEnumCombo<T>(string label, float width, T current, out T newValue, Func<T, string>? toString = null, int skip = 0, int skipEnd = 0) where T : struct, Enum`; `bool GenericEnumCombo<T>(string label, float width, T current, out T newValue, IEnumerable<T> options, Func<T, string>? toString = null) where T : struct, Enum` |
| Colour | `bool ColorPicker(string label, string tooltip, Vector3 input, Action<Vector3> setter, string letter = "")`; `float ColorIntensity(uint / Vector3 / Vector4 color)`; `ContrastColorBw(uint / Vector3 / Vector4 color)`; `uint HalfTransparent(uint baseColor)`; `uint HalfTransparentText()`; `uint HalfBlend(uint baseColor, uint overlayColor)`; `uint HalfBlendText(uint overlayColor)`; `uint ColorConvertFloat3ToU32(Vector3 color)` |
| Clipboard | `string GetClipboardText()` ("Exception safe clipboard"); `void CopyOnClickSelectable(string text, string copiedText, string tooltip)`; `void CopyOnClickSelectable(string text)` |
| Icons | `void HoverIcon(IDalamudTextureWrap icon, Vector2 iconSize)`; `void HoverIcon(ISharedImmediateTexture icon, Vector2 iconSize)`; `void HoverIcon(ImTextureID id, Vector2 contentSize, Vector2 iconSize)`; `void HoverIconTooltip(IDalamudTextureWrap icon, Vector2 iconSize)`; `(icon, iconSize, string text)`; `(ImTextureID icon, Vector2 iconSize, Vector2 size)` |
| Misc | `void Dummy(float width, float height, bool condition = true)`; `void Dummy(Vector2 size, bool condition = true)`; `bool IsDropping(string name)` |

Behaviour worth knowing:

- `DrawDisabledButton` draws the button at 50% alpha when `disabled`, shows `description` as
  tooltip, and returns `true` only when clicked **and** not disabled. `icon: true` draws the label
  with `UiBuilder.IconFont`, so pass `FontAwesomeIcon.X.ToIconString()`.
- `GenericEnumCombo` returns `true` with `newValue` set when a different value was picked,
  otherwise `newValue = current`. Entries whose name is empty are skipped, so return `""` from
  `toString` to hide a value. `skip` / `skipEnd` drop that many values from the start / end of
  `Enum.GetValues<T>()`.
- `Checkbox` with setter returns `true` and calls `setter` only on an actual change.
- `OpenNameField` is a popup with a single text field (hint "Enter New Name...", max 512, width
  `300 * GlobalScale`); open it with `ImGui.OpenPopup(popupName)`. It returns `true` when Enter
  was pressed; Escape closes it. `newName` is not cleared for you.
- `DragFloat` / `DragInt` return true "if the item was edited but is not active anymore".
- `HoverTooltipStringHandler` (top of `Util.cs`) is the interpolated-string handler for
  `ImGuiUtil.HoverTooltip($"...")`; like its `ImUtf8` counterpart it checks `IsItemHovered` in the
  constructor.

## Functions (`Functions.cs`)

| Function | Comment in source |
|---|---|
| `void IteratePairwise<T>(IReadOnlyList<T> list, Action<T> action1, Action inBetween, Action<T>? action2 = null)` | "Iterate through a list executing actions on each element by its mode." Handles the items in pairs: `action1`, `inBetween`, `action2` (defaults to `action1`) |
| `bool IteratePairwise<T>(IReadOnlyList<T> list, Func<T, bool> func1, Action inBetween, Func<T, bool>? func2 = null)` | same, returns the or'd result |
| `(byte Lowest, byte Second, byte Third, byte Highest) SplitBytes(uint value)` | "Split a uint into four bytes, e.g. for RGBA colors." |
| `string ColorBytes(uint color)` | "Obtain a descriptive hex-string of a RGBA color." |
| `uint ReorderColor(uint seColor)` | "Reorder a ABGR color to RGBA." |
| `uint AverageColor(uint c1, uint c2)` | "Average two given colors." |
| `uint RemoveBit(uint config, int bit)`, `ulong RemoveBit(ulong config, int bit)` | "Remove a single bit, moving all further bits one down." |
| `uint MoveBit(uint config, int bit1, int bit2)`, `ulong MoveBit(ulong config, int bit1, int bit2)` | "Move a bit ... from its position to another, shifting other bits accordingly." |
| `string HumanReadableSize(long size, string format = "{0:0.#} {1}")` | human readable byte size |
| `string ToCompressedBase64<T>(T data, byte version)` | JSON (Newtonsoft) + version byte, GZip, Base64; empty string on failure |
| `byte FromCompressedBase64<T>(string base64, out T? data)` | returns the version byte; `byte.MaxValue` and `data = default` on failure |
| `Vector2 Contain(this Vector2 vec, Vector2 max)` | largest vector of the same aspect ratio that fits into `max` |
| `bool GetQuickAccessFolders(out List<(string Name, string Path)> folders)` | Windows Quick Access folders |
| `bool GetDownloadsFolder(out string folder)` | the Downloads folder |

`ToCompressedBase64` / `FromCompressedBase64` are the share-string format used for clipboard
export and import (pairs with `ItemSelector`'s `Flags.Import`): choose a version byte, bump it
when the type changes, and check the returned version before using `data`. `IteratePairwise`
calls `action1(list[i])`, `inBetween()`, `action2(list[i + 1])` for each pair and `action1` alone
for a trailing odd element.

## Custom (`Custom/`, namespace `OtterGui.Custom`)

Pieces specific to the upstream author's plugins (the guide button's tooltip names Penumbra).
Do not use the fixed link buttons in an unrelated plugin: the URLs are hard-coded.

| Member | Notes |
|---|---|
| `CustomGui.DrawDiscordButton(MessageService message, float width)` | opens `https://discord.gg/kVva7DHV4r` ("the official discord server" in the source comment) |
| `CustomGui.DrawGuideButton(MessageService message, float width)` | opens `https://reniguide.info/` |
| `CustomGui.DrawKofiPatreonButton(MessageService message, Vector2 size)` | split button for Ottermandias' Ko-Fi and Patreon |
| `CustomGui.DrawLinkButton(MessageService message, string text, string address, float width, string? tooltip = null)` | generic: opens `address` with the shell, reports failure through `message`. This one is reusable |
| `CustomGui.DiscordColor`, `ReniColorButton`, `ReniColorHovered`, `ReniColorActive` | `const uint` |
| `IndividualHelpers.DrawObjectKindCombo(float width, ObjectKind current, out ObjectKind result, IEnumerable<ObjectKind> kinds)` | combo over Dalamud `ObjectKind` values |
| `IndividualHelpers.ToName(this ObjectKind kind)` | display names; the fork has no arm for mounts, upstream maps `ObjectKind.Mount => "Mount"` |
