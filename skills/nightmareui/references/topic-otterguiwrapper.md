# NightmareUI.OtterGuiWrapper

Covers the companion repo `NightmareXIV/NightmareUI.OtterGuiWrapper`: what it wraps, every public
type with signatures, how AutoRetainer builds its configuration window with it, and its pitfalls.
Setup (submodules, csproj, ECommons ordering) is in [guide section 1](guide-1-adding-it-to-a-plugin.md).

Pins: wrapper `371b1f9` (2026-05-30, MIT), NightmareUI `aae510d` (2026-06-18), `Limiana/OtterGui`
`a08fda6` (2026-05-30, Apache-2.0), ECommons `bc38dc7` (2026-11-08), consumer AutoRetainer
`cc020df` (2026-08-28), which pins exactly these wrapper, NightmareUI and OtterGui commits.
Nothing was compiled or run in game. Sources: `references/NightmareUI.OtterGuiWrapper/`.

## 1. What it is

Three things built on OtterGui's `OtterGui.Filesystem.FileSystem<T>`:

1. `FileSystems/Configuration/` - `ConfigFileSystem`, a ready-made two-pane settings window body:
   a filterable tree of pages on the left, the selected page on the right.
2. `FileSystems/Configuration/SimplifiedSelector/` - `FileSystemSelector<T, TStateStorage>`, a cut
   down copy of OtterGui's selector (same four partial files minus Buttons, Context and DragDrop:
   no toolbar, no context menus, no drag and drop, no key navigation).
3. `FileSystems/Generic/` - `GenericFileSystem<TData>`, a user-editable tree of items persisted to
   JSON, using OtterGui's **full** selector (`OtterGui.FileSystem.Selector`).

**Why a separate repo** (inference from the project files, not stated upstream): it is the only
part that needs OtterGui, so plugins such as Splatoon can take NightmareUI without an OtterGui
submodule. It wraps Limiana's fork of OtterGui, not Ottermandias' upstream.

## 2. Configuration tree (`NightmareUI.OtterGuiWrapper.FileSystems.Configuration`)

```csharp
public class ConfigFileSystemEntry
{
    public virtual string Path { get; }              // MUST override; "Folder/Sub/Leaf"
    public virtual NuiBuilder? Builder { get; init; }
    public virtual void Draw()                       // default: Builder.Draw(); throws without one
    public virtual int DisplayPriority { get; init; } = 0;
    public virtual bool NoFrame { get; set; } = false;
    public virtual bool ShouldDisplay()              // default true
    public virtual bool ShouldHighlight()            // default Builder?.ShouldHighlight
    public virtual Vector4? GetColor()               // leaf text colour while filtering
}

public sealed class ConfigFileSystem : FileSystem<ConfigFileSystemEntry>
{
    public ConfigFileSystem(Func<ICollection<ConfigFileSystemEntry>> getDataAction)
    public FileSystemSelector Selector { get; private set; }
    public ICollection<ConfigFileSystemEntry> DataStorage { get; private set; }
    public Func<ConfigFileSystemEntry> DefaultValueSelector;   // default: DataStorage.First()
    public Action DrawWhenNoSelection { get; set; } = null;
    public void Reload()
    public void Draw(float? width = null)
    public class FileSystemSelector       // : SimplifiedSelector.FileSystemSelector<ConfigFileSystemEntry, State>
    {   public string Filter { get; }  public bool Sorted = false;  public record struct State { } }
}

public static class ConfigFileSystemHelpers
{   public static IEnumerable<T?> CreateInstancesOf<T>() where T : ConfigFileSystemEntry }
```

- `Reload()` calls the data function, drops entries whose `ShouldDisplay()` is false, rebuilds the
  tree from the `Path` strings (slash separated; intermediate segments become folders), creates a
  new `Selector` and reselects the previously selected path. Call it whenever a `ShouldDisplay`
  result may have changed; AutoRetainer calls `S.NeoWindow.Reload()` from `AutoRetainer.cs:285`.
- `Draw(width)` draws the selector at `width`, `SameLine`, then a child `"CFSChild"` with the
  page. With `width == null` the width is computed from the widest path segment
  (`CalcTextSize + depth * 25 + scrollbar`). If nothing is selected it selects
  `DefaultValueSelector()`. A page with `NoFrame` or a non-null `Builder` is drawn by its own
  `Draw()`; any other page is wrapped in a `NuiBuilder` section titled with the path segments
  joined by `" - "`.
- Order is the order of the collection (`ISortMode.InternalOrder`); `Selector.Sorted = true`
  switches to `FoldersFirst`. Folders start open and are drawn in `ImGuiColors.DalamudViolet`.
- The filter box **highlights instead of hiding**: `ApplyFilters` always returns false, the text
  is copied into every entry's `Builder.Filter`, and `GetColor()` paints a leaf green when its
  last path segment or any of its builder's widget labels matches, grey otherwise.
- `CreateInstancesOf<T>()` instantiates every non-abstract type in `typeof(T).Assembly` assignable
  to `T` through `Activator.CreateInstance(type, true)` (a non-public parameterless constructor is
  enough) and yields them by descending `DisplayPriority`.

## 3. `SimplifiedSelector.FileSystemSelector<T, TStateStorage>`

`where T : class where TStateStorage : struct`, implements `IDisposable`. Public: constructor
`(FileSystem<T> fileSystem, IKeyState keyState, Logger log, Action<Exception>? exceptionHandler =
null, string label = "##FileSystemSelector", bool allowMultipleSelection = false)`, `T? Selected`,
`IReadOnlySet<FileSystem<T>.IPath> SelectedPaths`, `event SelectionChangeDelegate?
SelectionChanged`, `void ClearSelection()`, `void RemovePathFromMultiSelection(IPath)`,
`void SelectByValue(T value)`, `void Draw(float width)`, `void SetFilterDirty()`,
`bool DrawFilter = true`, `virtual ISortMode<T> SortMode`, `string Label { get; init; }`,
`ExceptionHandler`, `AllowMultipleSelection`, `virtual void Dispose()`. Protected customisation
points: `DrawLeafName`, `ApplyFilters`, `GetState`, `ApplyFiltersAndState`, `ChangeFilter`,
`CustomFilters`, `Select`, `FolderLineColor`, `ExpandedFolderColor`, `CollapsedFolderColor`,
`FoldersDefaultOpen`, `FilterValue`, `FilterTooltip`, `LeafFlags`, `EnqueueFsAction`.

## 4. `GenericFileSystem<TData>` (`...FileSystems.Generic`)

```csharp
public interface IFileSystemStorage
{   Guid GUID { get; set; }  string? GetCustomName();  void SetCustomName(string s); }

public sealed class GenericFileSystem<TData> : FileSystem<TData>
        where TData : class, IFileSystemStorage, new()
{
    public GenericFileSystem(ICollection<TData> dataStorage, string dataFilePrefix)
    public readonly FileSystemSelector Selector;
    public readonly ICollection<TData> DataStorage;
    public bool Create(TData newItem, string newName, out string error)
    public void DoDelete(TData item)
    public bool FindLeaf(TData? item, [NotNullWhen(true)] out Leaf? leaf)
    public IEnumerable<Folder> GetAllFolders()
    public IEnumerable<Leaf> GetAllLeaves()
    public bool TryGetPathByID(Guid id, [NotNullWhen(true)] out string? path)
    public void Save()
    public class FileSystemSelector       // : OtterGui FileSystemSelector<TData, State>
    {
        public float Width = 0f;   public void Draw(float width)
        public event Action<Leaf, State, bool> OnAfterDrawLeafName;
        public event OnBeforeCopyDelegate OnBeforeCopy;            // (TData original, ref TData copy)
        public event OnImportPopupOpenDelegate OnImportPopupOpen;  // (string clipboardText, ref string newName)
        public event OnBeforeItemCreationDelegate OnBeforeItemCreation;  // (ref TData item)
    }
}
```

The tree layout is stored in `<plugin config dir>/<dataFilePrefix>FileSystem.json`, keyed by item
`GUID`; the items themselves stay in your own collection. The constructor subscribes `Save` to
`EzConfig.OnSave`. The selector adds four toolbar buttons: new item, import from clipboard, copy
to clipboard (serialised with `EzConfig.DefaultSerializationFactory`, `GUID` reset to
`Guid.Empty`) and delete (OtterGui `DeleteSelectionButton` with a Control modifier). Neither
AutoRetainer nor Splatoon uses this class at the pinned commits, so there is no consumer example.

## 5. Worked example: the selector as AutoRetainer uses it

From `AutoRetainer/UI/NeoUI/!Definitions/` and sibling folders, abridged.

```csharp
public abstract class NeoUIEntry : ConfigFileSystemEntry { }

public abstract class DebugSectionBase : NeoUIEntry
{
    public override string Path => $"Advanced/Debug/{GetType().Name.Replace("Debug", "")}";
    public override bool ShouldDisplay() => C.Verbose;
}

public sealed class NeoWindow : Window
{
    private readonly NeoUIEntry[] Tabs =
    [
        new MainSettings(), new UserInterface(), /* ... */
        ..ConfigFileSystemHelpers.CreateInstancesOf<InventoryManagementBase>(),
        new LoginOverlay(), new MiscTab(),
        ..ConfigFileSystemHelpers.CreateInstancesOf<DebugSectionBase>(),
    ];
    internal ConfigFileSystem FileSystem;

    public NeoWindow() : base("AutoRetainer Configuration")
    {
        P.WindowSystem.AddWindow(this);
        this.SetMinSize();
        FileSystem = new(() => Tabs);
    }
    public void Reload() => FileSystem.Reload();
    public override void Draw() => FileSystem.Draw(null);
    public override void OnClose() => EzConfig.Save();
}
```

A page is either builder based ([guide 4.1](guide-4-worked-examples.md)) or overrides `Draw()` with plain ImGui
(`Experiments/NightMode.cs`), in which case the wrapper frames it in a section. A page that manages
its own layout sets `public override bool NoFrame { get; set; } = true;` (`CharaOrder.cs`).
Ordering inside a `CreateInstancesOf` group uses `public override int DisplayPriority => -10;`.

## 6. Pitfalls

All read from the pinned source; none reproduced in game.

1. **`CreateInstancesOf` constructs more than once.** It enumerates a deferred LINQ query once
   for the priorities and once per distinct priority, so every type's constructor runs several
   times and the returned objects come from different passes. Keep page constructors free of
   side effects.
2. **`Reload()` replaces the selector without disposing the old one**, which stays subscribed to
   `FileSystem.Changed`. Harmless in small numbers; do not call `Reload()` every frame.
3. **`ConfigFileSystemEntry.Path` and `Draw` throw by default** (`NotImplementedException`, after
   a `PluginLog.Error`). Override `Path` always, and either `Draw` or `Builder`. Two entries
   with the same `Path` make the `ToDictionary` in `Reload()` throw.
4. **`GenericFileSystem` import needs a subscriber.** `OnBeforeItemCreation` is invoked without a
   null check; with no handler attached, importing from the clipboard ends in the "Error: ..."
   notification. The class also never unsubscribes from `EzConfig.OnSave` and has no `Dispose`.
5. **Two selectors, two namespaces.** `ConfigFileSystem.FileSystemSelector` derives from the
   wrapper's simplified copy; `GenericFileSystem<T>.FileSystemSelector` derives from OtterGui's.
   OtterGui itself spells the namespaces `OtterGui.Filesystem` (tree) and
   `OtterGui.FileSystem.Selector` (selector).
