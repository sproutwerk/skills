OtterGui topic: the virtual folder tree `FileSystem<T>` (`Folder`, `Leaf`, `IPath`), its operations
and change event, sort modes, JSON persistence, and the `FileSystemSelector<T, TStateStorage>`
tree UI with its customization points. Verified against `Limiana/OtterGui` `a08fda6` (2026-05-30,
branch `main`). One behavioural difference from upstream `79771ee` in `FileSystem.File.cs` is
flagged below. Worked example read from `NightmareXIV/NightmareUI.OtterGuiWrapper` `371b1f9`.

Paths such as `references/...` and `references/topic-*.md` are relative to the skill root (the directory that
holds `SKILL.md`), not to this file. Nothing was compiled.

## Namespaces (easy to get wrong)

| Types | Namespace | Note the capital |
|---|---|---|
| `FileSystem<T>`, `FileSystemChangeType`, `ISortMode<T>`, `SortMode`, `Extensions` | `OtterGui.Filesystem` | lower-case `s` |
| `FileSystemSelector<T, TStateStorage>`, `ModSelectorSettings` | `OtterGui.FileSystem.Selector` | upper-case `S` |
| `SearchStringSplitter<TEnum, TObj, TEntry>`, `ISplitterEntry<TEnum, TEntry>` | `OtterGui.Filesystem.Selector` | lower-case `s` |

The folder on disk is `Filesystem/`. `NightmareUI.OtterGuiWrapper` imports both
`OtterGui.Filesystem` and `OtterGui.FileSystem.Selector`.

## What it is

An in-memory tree that arranges your own objects (`T : class`) in user-defined folders. It does
not touch the disk by itself; "file system" refers to the folder metaphor. Each object appears as
exactly one `Leaf`; the class implements `IReadOnlyDictionary<T, FileSystem<T>.Leaf>` for the
reverse lookup.

```csharp
public partial class FileSystem<T> : IReadOnlyDictionary<T, FileSystem<T>.Leaf> where T : class
{
    public FileSystem(IComparer<string>? comparer = null);   // sibling name order; default OrdinalIgnoreCase
    public Folder Root;
    public uint   IdCounter = 1;

    public delegate void ChangeDelegate(FileSystemChangeType type, IPath changedObject,
        IPath? previousParent, IPath? newParent);
    public event ChangeDelegate? Changed;                    // fired after the change happened
}
```

## Paths

```csharp
public interface IPath          // FileSystem<T>.IPath
{
    Folder Parent { get; }
    string Name { get; }
    uint   Identifier { get; }      // unique per file system instance, assigned from IdCounter
    ushort IndexInParent { get; }
    byte   Depth { get; }
    bool   IsRoot { get; }
    bool   IsLocked { get; }
    string   FullName();            // names joined with '/', empty for Root
    Folder[] Parents();             // ancestors in order, without Root
}
```

| Type | Extra public members |
|---|---|
| `FileSystem<T>.Folder` | `TotalDescendants`, `TotalLeaves`, `TotalChildren`, `Flags` (`PathFlags`), `GetSubFolders()`, `GetLeaves()`, `GetChildren(ISortMode<T> mode)`, `GetAllDescendants(ISortMode<T> mode)`, `SetLocked(bool value)` |
| `FileSystem<T>.Leaf` (sealed) | `T Value`, `Flags`, `SetLocked(bool value)` |
| `FileSystem<T>.PathFlags` | `[Flags] enum : byte { Locked = 0x01 }` |

`Folder.Children` is `internal`. From plugin code enumerate with `GetChildren(mode)`,
`GetSubFolders()`, `GetLeaves()` or `GetAllDescendants(mode)`. `Leaf` has an internal constructor:
create leaves only through the file system. Names are passed through `FixName()`: `/` becomes
`\`, whitespace is trimmed, an empty name becomes `<None>`.

## Operations

"The public facing filesystem methods all throw descriptive exceptions if they are unsuccessful"
(comment in `FileSystem.cs`). They throw plain `System.Exception`.

| Method | Effect | Event |
|---|---|---|
| `bool Find(string fullPath, out IPath child)` | looks up a `/`-separated path; on failure `child` is the deepest existing folder | |
| `(Leaf, int) CreateLeaf(Folder parent, string name, T data)` | throws if the name exists in `parent` | `LeafAdded` |
| `(Leaf, int) CreateDuplicateLeaf(Folder parent, string name, T data)` | appends ` (2)`, ` (3)`... until unique | `LeafAdded` |
| `(Folder, int) CreateFolder(Folder parent, string name)` | throws if the name exists | `FolderAdded` |
| `(Folder, int) FindOrCreateFolder(Folder parent, string name)` | throws only if a non-folder has that name | `FolderAdded` when created |
| `Folder FindOrCreateAllFolders(string path)` | creates every missing folder of the path | `FolderAdded` |
| `void Rename(IPath child, string newName)` | throws for Root or a name clash | `ObjectRenamed` |
| `void RenameWithDuplicates(IPath child, string newName)` | renames, incrementing a duplicate suffix on clash | `ObjectRenamed` |
| `void RenameAndMove(IPath child, string newPath)` | full-path rename, creates folders; throws on empty path or clash | `ObjectMoved` |
| `void Move(IPath child, Folder newParent)` | throws for Root and circular moves; a folder dropped onto a same-named folder is merged | `ObjectMoved` (or merge events) |
| `void Merge(Folder from, Folder to)` | moves all children; `from` is removed when everything moved | `FolderMerged` / `PartialMerge` |
| `void Delete(IPath child)` | throws for Root | `ObjectRemoved` |
| `bool ChangeLockState(IPath path, bool value)` | true when the state changed | `LockChange` |
| `bool Equal(string lhs, string rhs)` | compares two names with the file system's comparer | |
| `bool ContainsKey(T key)`, `bool TryGetValue(T key, out Leaf? value)`, `GetEnumerator()` | dictionary view by object | |

`enum FileSystemChangeType { ObjectRenamed, ObjectRemoved, FolderAdded, LeafAdded, ObjectMoved,
FolderMerged, PartialMerge, Reload, LockChange }`.

The tuple returned by the `Create...` methods is the new node and its index in the parent.
Locking only blocks drag and drop in the selector; every method above works on locked paths.

## Sort modes (`Filesystem/SortMode.cs`)

```csharp
public interface ISortMode<T> where T : class
{
    ReadOnlySpan<byte> Name { get; }
    ReadOnlySpan<byte> Description { get; }
    IEnumerable<FileSystem<T>.IPath> GetChildren(FileSystem<T>.Folder folder);
}
```

Built-in instances are static fields on the interface: `ISortMode<T>.FoldersFirst`,
`Lexicographical`, `InverseFoldersFirst`, `InverseLexicographical`, `FoldersLast`,
`InverseFoldersLast`, `InternalOrder`, `InverseInternalOrder`. The `enum SortMode` lists the same
eight names (for storing the choice in a configuration). `Lexicographical` returns the children in
their stored order, which is kept sorted by the comparer given to the constructor;
`InternalOrder` sorts by `Identifier`, i.e. creation order. Implement `ISortMode<T>` for a custom
order and return it from the selector's `SortMode` property.

## Persistence (`Filesystem/FileSystem.File.cs`)

All three members are `protected`: derive from `FileSystem<T>`.

```csharp
protected void SaveToFile(StreamWriter writer, Func<T, string, (string, bool)> conversion, bool addEmptyFolders);
protected bool Load(FileInfo file, IEnumerable<T> objects, Func<T, string> toIdentifier, Func<T, string> toName);
protected bool Load(JObject? jObject, IEnumerable<T> objects, Func<T, string> toIdentifier, Func<T, string> toName);
```

File format (Newtonsoft.Json, indented):

```json
{
  "Data": { "<identifier>": "Folder/Sub Folder/Leaf Name" },
  "EmptyFolders": [ "Folder/Empty" ],
  "LockedPaths": [ "Folder/Sub Folder" ]
}
```

- `conversion(value, fullPath)` returns the identifier to store and whether to store the leaf at
  all (return `false` for leaves that sit at their default location to keep the file small).
- `EmptyFolders` is written only when `addEmptyFolders` is true; `LockedPaths` only when any path
  is locked.
- `Load` clears the tree, then for every object in `objects` looks up `toIdentifier(obj)` in
  `Data`. Found: the folders are created and the leaf gets the stored name. Not found: the leaf is
  created at `Root` with `toName(obj)`. Name clashes get a duplicate suffix. Objects are the
  source of truth: identifiers in the file without a matching object are dropped.
- The return value is `true` when the loaded tree differs from the file (so you should save).
- `Load` always fires `Changed` with `FileSystemChangeType.Reload`.

**Fork versus upstream.** When the file does not exist or does not parse (`jObject == null`):
upstream `79771ee` still adds every object as a leaf at `Root` (commit `18e62ab`, "Add
auto-leaves even if no saved file exists"). The fork does not have that branch: the tree stays
**empty** and `Load` returns `true`. On the fork, create the leaves yourself on first run.
`NightmareUI.OtterGuiWrapper`'s `GenericFileSystem` only calls `Load` when `info.Exists`.

### Skeleton (pattern of `GenericFileSystem<TData>` in NightmareUI.OtterGuiWrapper)

```csharp
using OtterGui.Filesystem;

public sealed class PlanFileSystem : FileSystem<Plan>          // Plan: class with Guid Id, string Name
{
    private readonly string _file;
    private readonly List<Plan> _plans;

    public PlanFileSystem(List<Plan> plans, string file)
    {
        _plans = plans;
        _file  = file;
        var info = new FileInfo(file);
        if (info.Exists)
            Load(info, _plans, p => p.Id.ToString(), p => p.Name);
        else
            foreach (var plan in _plans)                        // needed on the fork, see above
                CreateDuplicateLeaf(Root, plan.Name, plan);
        Changed += (_, _, _, _) => Save();
    }

    public void Add(Plan plan)
    {
        _plans.Add(plan);
        CreateDuplicateLeaf(Root, plan.Name, plan);
    }

    public void Remove(Plan plan)
    {
        _plans.Remove(plan);
        if (TryGetValue(plan, out var leaf))
            Delete(leaf);
    }

    public void Save()
    {
        using var stream = new FileStream(_file, FileMode.Create, FileAccess.ReadWrite, FileShare.ReadWrite);
        using var writer = new StreamWriter(stream);
        SaveToFile(writer, (plan, _) => (plan.Id.ToString(), true), true);
    }
}
```

Saving from the `Changed` handler writes on every change, including each step of a drag and drop
of several items; Penumbra-style code routes this through a delayed save
(`SaveServiceBase<T>.DelaySave`, `references/topic-services-classes.md`).

## Selector UI (`Filesystem/Selector/`)

```csharp
using OtterGui.FileSystem.Selector;

public partial class FileSystemSelector<T, TStateStorage> : IDisposable
    where T : class where TStateStorage : struct
{
    public FileSystemSelector(FileSystem<T> fileSystem, IKeyState keyState, Logger log,
        Action<Exception>? exceptionHandler = null, string label = "##FileSystemSelector",
        bool allowMultipleSelection = false);

    public void Draw();
    public T? Selected { get; }
    public IReadOnlySet<FileSystem<T>.IPath> SelectedPaths { get; }
    public delegate void SelectionChangeDelegate(T? oldSelection, T? newSelection, in TStateStorage state);
    public event SelectionChangeDelegate? SelectionChanged;
    public void SelectByValue(T value);              // also expands ancestors and scrolls to it
    public void ClearSelection();
    public void RemovePathFromMultiSelection(FileSystem<T>.IPath path);
    public void SetFilterDirty();
    public virtual ISortMode<T> SortMode { get; }    // default ISortMode<T>.Lexicographical
    public string Label { get; init; }
    public readonly string MoveLabel;                // drag-drop payload name, "<Label>Move"
    public readonly bool AllowMultipleSelection;
    public readonly Action<Exception> ExceptionHandler;
    public virtual void Dispose();                   // unsubscribes from FileSystem.Changed

    public void AddButton(Action<Vector2> action, int priority);
    public void RemoveButton(Action<Vector2> action);
    public int  ButtonCount { get; }
    public void SubscribeRightClickFolder(Action<FileSystem<T>.Folder> action, int priority = 0);
    public void SubscribeRightClickLeaf(Action<FileSystem<T>.Leaf> action, int priority = 0);
    public void SubscribeRightClickMain(Action action, int priority = 0);
    public void UnsubscribeRightClickFolder(Action<FileSystem<T>.Folder> action);
    public void UnsubscribeRightClickLeaf(Action<FileSystem<T>.Leaf> action);
    public void UnsubscribeRightClickMain(Action action);
}
```

What `Draw()` renders: a filter text field on top, a resizable child (`ImUtf8.ResizableChild`)
containing the clipped tree with folder lines, and a row of buttons below. It handles selection
(click; Ctrl-click and Shift-click when `allowMultipleSelection`), drag and drop to move items
between folders, and right-click context menus. `TStateStorage` is a per-row value you compute in
`GetState` (for example a colour) and receive back in `DrawLeafName`.

Defaults installed by the constructor:

| Where | Entry | Priority |
|---|---|---|
| Folder context menu | `ExpandAllDescendants`, `CollapseAllDescendants` | 100 |
| Folder context menu | `SetLocked` | 900 |
| Folder context menu | `DissolveFolder` | 999 |
| Folder context menu | `RenameFolder` | 1000 |
| Leaf context menu | `SetLocked` | 900 |
| Leaf context menu | `RenameLeaf` | 1000 |
| Background context menu | `ExpandAll`, `CollapseAll` | 1 |
| Button row | `FolderAddButton` | 50 |

All of these are `protected` methods, so a subclass can remove them, e.g.
`UnsubscribeRightClickFolder(DissolveFolder)` or `RemoveButton(FolderAddButton)`. Entries are
ordered by priority, then subscription order. Registering the same delegate again with another
priority moves it.

Protected customization points:

| Member | Default | Purpose |
|---|---|---|
| `void DrawLeafName(FileSystem<T>.Leaf leaf, in TStateStorage state, bool selected)` | tree node with `LeafFlags` and the leaf name | how a row looks; should still create a tree node using `LeafFlags` |
| `const ImGuiTreeNodeFlags LeafFlags` | `Leaf \| Bullet \| NoTreePushOnOpen` | |
| `TStateStorage GetState(FileSystem<T>.IPath path)` | `default` | per-row state |
| `bool ApplyFilters(FileSystem<T>.IPath path)` | `FilterValue.Length != 0 && !path.FullName().Contains(FilterValue)` | **return `true` to hide** the path |
| `bool ApplyFiltersAndState(FileSystem<T>.IPath path, out TStateStorage state)` | calls the two above | override when state and filter are computed together |
| `bool ChangeFilter(string filterValue)` | `true` | called when the text changes; return whether to refilter |
| `(float RemainingWidth, bool ClearText) CustomFilters(float width)` | `(width, false)` | draw extra widgets in the filter row |
| `string FilterValue`, `string FilterTooltip` | | current text; tooltip shown on the filter field |
| `void Select(FileSystem<T>.Leaf? leaf, bool clear, in TStateStorage storage = default)` | | selection hook |
| `void DrawPopups()` | nothing | popups that must be drawn outside the pushed styles |
| `uint FolderLineColor`, `ExpandedFolderColor`, `CollapsedFolderColor` | `0xFFFFFFFF` | |
| `bool FoldersDefaultOpen` | `false` | |
| `float CurrentWidth`, `MinimumAbsoluteSize`, `MinimumAbsoluteRemainder`, `MinimumScaling` (0.1), `MaximumScaling` (0.9), `bool UseScaling` (true), `bool Resizable` (true), `void SetSize(Vector2 size)` | | width handling; override `CurrentWidth` and `SetSize` to persist the width |
| `void EnqueueFsAction(Action action)` | | run something after the draw loop; required for anything that modifies the tree while it is being drawn |
| `FolderAddButton(Vector2 size)`, `DeleteSelectionButton(Vector2 size, DoubleModifier modifier, string singular, string plural, Action<T> delete)` | | ready-made buttons for `AddButton` |
| `SetQuickMove(Folder folder, int which, string current, Action<string> onSelect)`, `ClearQuickMove(int which, string current, Action onSelect)`, `QuickMove(Leaf leaf, params string[] folders)` | | optional context menu entries |
| `FileSystem`, `Log`, `SelectedLeaf`, `_selectedPaths` | | fields |

Filter polarity trips people up: `ApplyFilters` answers "is this path filtered **out**".
`ItemSelector<T>.Filtered(int idx)` has the same polarity, while `Column<T>.FilterFunc` and
`FilterComboBase<T>.IsVisible` answer "is it visible".

Exceptions thrown by queued actions go to `ExceptionHandler` (default: `Log.Warning`). An
exception inside `Draw()` is rethrown wrapped with the selector's internal indices.

Arrow-key and letter navigation exists in the source (`HandleKeyNavigation`) but the call is
commented out (`// TODO: do this right.`), so `keyState` is stored and currently unused by the
draw loop.

### Minimal selector

```csharp
public sealed class PlanSelector : FileSystemSelector<Plan, PlanSelector.State>
{
    public record struct State(uint Color);

    public PlanSelector(PlanFileSystem fs, IKeyState keys, Logger log)
        : base(fs, keys, log, label: "##PlanSelector")
    {
        SubscribeRightClickLeaf(DeleteEntry, 500);
        UnsubscribeRightClickFolder(DissolveFolder);
    }

    protected override State GetState(FileSystem<Plan>.IPath path)
        => new(path is FileSystem<Plan>.Leaf { Value.Enabled: true } ? 0xFFFFFFFF : 0xFF808080);

    protected override void DrawLeafName(FileSystem<Plan>.Leaf leaf, in State state, bool selected)
    {
        var flags = selected ? ImGuiTreeNodeFlags.Selected | LeafFlags : LeafFlags;
        using var color = ImRaii.PushColor(ImGuiCol.Text, state.Color);
        using var node  = ImRaii.TreeNode(leaf.Name, flags);
    }

    private void DeleteEntry(FileSystem<Plan>.Leaf leaf)
    {
        if (ImGui.MenuItem("Delete"))
            EnqueueFsAction(() => ((PlanFileSystem)FileSystem).Remove(leaf.Value));
    }
}

// Window.Draw():
//   _selector.Draw();
//   ImGui.SameLine();
//   using var child = ImRaii.Child("##details", ImGui.GetContentRegionAvail(), true);
//   if (child && _selector.Selected is { } plan) DrawPlan(plan);
```

`Plan.Enabled` is an assumed property of the example type. Dispose the selector on plugin unload.

## What NightmareUI.OtterGuiWrapper does differently

It ships its own copy of the selector
(`NightmareUI.OtterGuiWrapper.FileSystems.Configuration.SimplifiedSelector.FileSystemSelector<T, TStateStorage>`,
four files mirroring `FileSystemSelector.cs`, `.Draw.cs`, `.State.cs`, `.Utility.cs`) for
`ConfigFileSystem`, and uses OtterGui's own `FileSystemSelector` for `GenericFileSystem<TData>`.
Both of its file systems derive from `OtterGui.Filesystem.FileSystem<T>`. Details belong to the
`nightmareui` skill.

## Search string splitter

`SearchStringSplitter<TEnum, TObj, TEntry>` (`Filesystem/Selector/SearchStringSplitter.cs`) is an
abstract parser for typed filter strings (general, forced, negated and "none" tokens). Public
surface: `void Parse(ReadOnlySpan<char> input)`, `bool IsVisible(TObj haystack)`; implement
`ConvertToken`, `AllowsNone`, `Matches`, `MatchesNone`. The token syntax is defined by `Parse`;
read `references/src/Filesystem/Selector/SearchStringSplitter.cs` before using it.

## List helpers in the same namespace

`OtterGui.Filesystem.Extensions` also holds general helpers, which is why Artisan imports the
namespace without using the tree: `Move<T>(this IList<T> list, int idx1, int idx2)`,
`Move<T>(this IList<T> list, ref int idx1, ref int idx2)`, `AddOrReplace<T>(this List<T> list,
T obj)`, `FixName`, `SplitDirectories`, `IsDuplicateName(out string baseName, out int number)`,
`IncrementDuplicate`, `ObtainUniqueFile(int maxDuplicates = int.MaxValue)`,
`ObtainUniqueString(Predicate<string> isDuplicate, int maxDuplicates = int.MaxValue)`,
`RemoveInvalidPathSymbols`, `IsInvalidInPath(this char c)`, `IsInvalidAscii(this char c)`.
