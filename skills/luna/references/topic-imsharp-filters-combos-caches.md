# ImSharp filter combos, filters and caches

Covers `SimpleFilterCombo<T>` and the other filter combos, the `IFilter` implementations,
`CacheManager` / `BasicCache`, and clipping a cached list. Verified against Luna `52eb0fe` and
ImSharp `4d09327` (branch `gc1.88`, ImGui 1.88). Paths are relative to the skill root. Grep
`references/api-index.txt` before using a member that is not shown here.

## Filtered combo (worked example)

```csharp
using ImSharp;
using Luna;

// Shape of Penumbra/UI/Classes/ModCombo.cs
public sealed class EntryCombo(IReadOnlyList<Entry> entries)
    : SimpleFilterCombo<Entry>(SimpleFilterType.Regex), IUiService
{
    public override StringU8 DisplayString(in Entry value) => value.Name;
    public override string   FilterString(in Entry value)  => value.Name;
    public override IEnumerable<Entry> GetBaseItems()      => entries;
}

// Draw: label, current value, tooltip, preview width
if (_combo.Draw("##entry"u8, ref _selected, "Pick an entry."u8, 250 * Im.Style.GlobalScale))
    OnSelectionChanged(_selected);
```

`SimpleFilterCombo<T>` overloads:
`Draw(Utf8LabelHandler label, in T current, Utf8TextHandler tooltip, float previewWidth, out T? result)`
and `Draw(label, ref T selection, tooltip, previewWidth)`. Selection compares
with `item.Item!.Equals(Current)`, so `T` needs sensible equality and
`ref _selected` must not be null when drawn. `SimpleFilterType`: `None`,
`Text`, `Partwise`, `Regex`. Optional overrides: `TextColor(in T)`,
`Tooltip(in T)`.

Ready-made:
`GenericCombo<T>(IReadOnlyList<T> values, DisplayDelegate<T>, FilterDelegate<T>, TooltipDelegate<T>?)`,
`EnumCombo<T>`. For full control derive `FilterComboBase<TCacheItem>`
(abstract `ItemHeight`, `GetItems()`, `DrawItem(in item, globalIndex, selected)`,
`IsSelected(item, globalIndex)`; hooks `PreDrawCombo`, `PostDrawCombo`,
`PreDrawList`, `PostDrawList`, `PreDrawFilter`, `PostDrawFilter`,
`EnterPressed`, `OnPopupClosed`; `CreateCache()`), with options
`AllowMouseWheel` (`MouseWheelType`, default `Control`), `MaximumItems` (12),
`ComputeWidth`, `CacheLifetime` (10 s), `DirtyCacheOnClosingPopup`,
`ClearFilterOnSelection`, `ClearFilterOnCacheDisposal`, `PreviewAlignment`,
`Flags`. `DrawBehavior(label, popupWidth, out ret)` attaches the popup to the
previously drawn item.

## Filters and caches

`IFilter` (`DrawFilter(ReadOnlySpan<byte> label, Vector2 availableRegion)`,
`Clear()`, `IsEmpty`, `IsVisible`, event `FilterChanged`) and
`IFilter<T>.WouldBeVisible(in T item, int globalIndex)`. Implementations:
`TextFilterBase<T>` / `TextFilter`, `RegexFilterBase<T>` / `RegexFilter`,
`PartwiseFilterBase<T>` / `PartwiseFilter`, `Utf8FilterBase<T>`,
`NumberFilterBase<TNumber, T>`, `FlagFilterBase<T, TEnum>`,
`TriStateFlagFilterBase<T, TEnum>`, `YesNoFilter<T>`,
`TokenizedFilter<TTokenType, T, TToken>`, `PairFilter<T>`, `MultiFilter<T>`,
`NopFilter<T>.Instance`.

Active filters are highlighted: text-style filters (`TextFilterBase<T>`, `TokenizedFilter<…>`,
`Utf8FilterBase<T>`) push a frame border in `ActiveBorderColor` while they hold text, and the flag
filters push `ActiveBackgroundColor` on their combo while not every flag is set. Both default to
`ImSharpConfiguration.Colors.ActiveFilterBorder` (default: none) and `ActiveFilterBackground`
(default `0x803030A0`); override the property per filter or set the shared values. Luna's
`ColorCache` sets both from Dalamud's attention colors when it refreshes.

`CacheManager.Instance` keys caches by `ImGuiId`:
`GetOrCreateCache<T>(ImGuiId id, Func<T> factory)`, `TryGetCache`,
`SetDirty(id)`, `SetCustomDirty(id)`, `SetFontDirty()`, `SetStyleDirty()`,
`SetColorsDirty()`. A cache derives `BasicCache` (`Update()`, `Dirty`,
`KeepAliveDuration` default 5 s, `Dispose(bool)`) and is dropped when not
requested for that long, so per-frame derived data lives only while its UI is
visible. Typical call, from consumers:
`var cache = CacheManager.Instance.GetOrCreateCache(Im.Id.Current, () => new MyCache(…));`.
`FilterCache<T>` / `BasicFilterCache<T>` add filtered item lists.

```csharp
using var clip = new Im.ListClipper(cache.Count, Im.Style.TextHeightWithSpacing);
foreach (var item in clip.Iterate(cache))
    item.Draw();
```
