## Compact catalogue

| Area | Entry points | Topic |
|---|---|---|
| UTF-8 ImGui | `ImUtf8.Text`, `Button`, `Checkbox`, `Selectable`, `Combo`, `InputText`, `InputScalar<T>`, `DragScalar<T>`, `Slider<T>`, `ColorEdit`, `Table`, `TabBar`, `TabItem`, `TreeNode`, `Child`, `Popup`, `Modal`, `Menu`, `MenuItem`, `PushId`, `ListClipper`, `CalcTextSize`, style shortcuts (`ImUtf8.FrameHeight`, `ItemSpacing`, `GlobalScale`, ...) | `references/topic-imutf8.md` |
| UTF-8 composite helpers | `ImUtf8.ButtonEx`, `IconButton`, `Icon`, `HoverTooltip`, `LabeledHelpMarker`, `CopyOnClickSelectable`, `InputTextOnDeactivated`, `InputScalarOnDeactivated<T>`, `ResizableChild`, `TextFramed`, `TextRightAligned`, `TextCentered`, `TextShadowed`, `DrawTableColumn`, `ScaledDummy`, `Spinner`, `RotatedText` | `references/topic-imutf8.md` |
| RAII | `ImRaii.PushColor`, `PushStyle`, `PushFont`, `PushId`, `PushIndent`, `PushFrameBorder`, `DefaultColors`, `DefaultStyle`, `DefaultFont`, `Disabled`, `Enabled`, `FramedGroup`, `Child`, `Table`, `TabItem`, `Popup`, ... | `references/topic-raii.md` |
| List selector | `ItemSelector<T>(IList<T> items, Flags flags)`, `Draw(float width)`, hooks `OnDraw`, `Filtered`, `OnAdd`, `OnDelete`, `OnDuplicate`, `OnMove`, `OnClipboardImport`, `OnDrop`; `ItemDetailsWindow.Draw(label, drawHeader, drawDetails)` | `references/topic-helpers.md` |
| Table | `Table<T>(string label, IReadOnlyCollection<T> items, params Column<T>[] headers)`, `Draw(float itemHeight)`; `ColumnString<T>`, `ColumnNumber<T>`, `ColumnSelect<T, TItem>`, `ColumnFlags<T, TItem>`, `TriStateColumnFlags<T, TItem>`, `YesNoColumn<T>`; `ImGuiTable.DrawTable` | `references/topic-table.md` |
| Folder tree | `FileSystem<T>` (`CreateLeaf`, `CreateFolder`, `FindOrCreateAllFolders`, `Move`, `Rename`, `RenameAndMove`, `Merge`, `Delete`, `Find`, event `Changed`), `ISortMode<T>`, protected `Load` / `SaveToFile`, `FileSystemSelector<T, TStateStorage>` | `references/topic-filesystem.md` |
| Combos | `FilterComboBase<T>`, `FilterComboCache<T>`, `FilterComboColors`, `ClippedSelectableCombo<T>`, `ImGuiUtil.GenericEnumCombo<T>` | `references/topic-widgets.md` |
| Onboarding | `Tutorial` (`Register`, `Open`), `Changelog` (`NextVersion`, `RegisterEntry`, `RegisterHighlight`, `RegisterImportant`) | `references/topic-widgets.md` |
| Other widgets | `TagButtons`, `TabBar` + `ITab`, `TwoStateCheckbox`, `XCheckbox`, `TriStateCheckbox`, `FontAwesomeCheckbox`, `MultiStateCheckbox<T>`, `IEditor<T>` (`DragEditor`, `SliderEditor`, `InputEditor`, `ColorEditor`, `EnumEditor`), `Widget.BeginFramedGroup`, `DrawHexViewer`, `ColorPicker`, `PaletteColorPicker`, `KeySelector`, `ModifiableKeySelector`, `ToggleButton.SplitButton` | `references/topic-widgets.md` |
| Clipping, pie, misc | `ImGuiClip.ClippedDraw`, `GetNecessarySkips`, `DrawEndDummy`; `ImGuiPie.Draw`; `Functions.ToCompressedBase64` / `FromCompressedBase64`, `HumanReadableSize`, `IteratePairwise` | `references/topic-helpers.md` |
| DI and services | `ServiceManager`, `IService`, `IRequiredService`, `HookManager`, `FastHook<T>`, `ImGuiCacheService`, `DynamisIpc` | `references/topic-services-classes.md` |
| Classes | `Logger`, `EventWrapper<...>`, `SaveServiceBase<T>` / `ISavable<T>`, `MessageService`, `FrameworkManager`, `KeyboardManager`, `ModifiableHotkey`, `DoubleModifier`, `LowerString`, `UserRegex`, `TextureCache`, `Backup`, `PerformanceTracker<T>`, `StartTimeTracker`, `MultiDictionary`, `IndexSet`, `LazyList<T>`, `TemporaryList<T>` | `references/topic-services-classes.md` |
| Internals | `ImGuiInternal.ItemAdd`, `ItemSize`, `ButtonBehavior`, `RenderFrame`, `RenderTextClipped`; `StringHelpers.ComputeSizeAndId`; `SymbolHelpers` | `references/topic-services-classes.md` |

### Two complete minimal examples

List selector (fork `main` and upstream signature; details and the Artisan variant in
`references/topic-helpers.md`):

```csharp
internal sealed class ProfileList : ItemSelector<Profile>
{
    public ProfileList(List<Profile> profiles)
        : base(profiles, Flags.Add | Flags.Delete | Flags.Filter)
    { }

    protected override bool OnDraw(int idx)
        => ImGui.Selectable(Items[idx].Name, idx == CurrentIdx);

    protected override bool Filtered(int idx)      // true hides the row
        => Filter.Length != 0 && !Items[idx].Name.Contains(Filter, StringComparison.OrdinalIgnoreCase);

    protected override bool OnAdd(string name)
    {
        if (name.Length == 0)
            return false;
        Items.Add(new Profile { Name = name });
        return true;
    }

    protected override bool OnDelete(int idx)
    {
        Items.RemoveAt(idx);
        return true;
    }
}
// _list.Draw(200 * ImGuiHelpers.GlobalScale); then _list.Current is the selected profile
```

Table (complete version with four column types in `references/topic-table.md`):

```csharp
internal sealed class ProfileTable : Table<Profile>
{
    public ProfileTable(IReadOnlyCollection<Profile> items)
        : base("ProfileTable", items, new NameColumn { Label = "Name" })
    { }

    private sealed class NameColumn : ColumnString<Profile>
    {
        public override float Width => 250 * ImGuiHelpers.GlobalScale;
        public override string ToName(Profile item) => item.Name;
    }
}
// _table.Draw(ImGui.GetTextLineHeightWithSpacing());
```

`Profile` is an assumed class with a `Name` property. Neither example was compiled.

