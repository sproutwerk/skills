## The bindings

`ImGui` is the public API, `ImGuiP` is `imgui_internal.h` (unstable across ImGui upgrades),
`ImGuiNative`/`ImGuiPNative` are raw-pointer calls. Do not read
`Generated/Functions/Functions.NNN.cs` to learn the API: those overloads are **not in the
DLL**. Use `references/index/imgui-methods.txt`. Assembly layout → `references/topic-bindings.md`

### ImU8String

`[InterpolatedStringHandler] public ref struct ImU8String`. Every text
parameter in `ImGui` is this type. Implicit conversions exist from `string?`,
`ReadOnlySpan<char>`, `ReadOnlySpan<byte>`, `Span<>`/`Memory<>` of both,
`byte[]`, `char[]`, `byte*`, `char*`.

```csharp
ImGui.Text("Static label"u8);             // zero-copy: points at the literal, already NUL-terminated
ImGui.Text($"HP {hp:N0} / {maxHp:N0}");   // formatted straight to UTF-8, no System.String allocated
ImGui.Text(player.Name);                  // string: transcoded to UTF-8 on every call
ImGui.Button($"Delete##{entry.Id}");
```

- Up to 511 bytes fit in an inline buffer (`AllocFreeBufferSize = 512`);
  longer text rents from `ArrayPool<byte>.Shared`. The binding calls
  `Recycle()` for you after the native call.
- Interpolation uses `Utf8.TryWrite`, so `ISpanFormattable`/`IUtf8SpanFormattable`
  values format without boxing. Format specifiers and alignment work
  (`{x,8:F2}`).
- `default(ImU8String)` is "null" (`IsNull`). Optional parameters such as
  `format`, `overlayText`, `shortcut` use it: omit them rather than passing `""`.
- It is a `ref struct`: you cannot store it in a field, capture it in a
  lambda, or use it across `await`. If you write your own helper taking
  `ImU8String`, call `text.Recycle()` when done (see `ImGuiHelpers.ClickToCopyText`).
- `##`/`###` ID suffixes behave as in C++.

Text functions are unformatted: `ImGui.Text`, `TextColored`, `TextDisabled`,
`TextWrapped`, `TextColoredWrapped` (new, no C++ equivalent), `BulletText`,
`LabelText`, `SetTooltip`. printf formats survive only where ImGui needs
them: the `format` argument of `Drag*`/`Slider*`/`Input*` (`"%.2f"`, `"%d%%"`).

String returns come in pairs: `ImGui.GetClipboardText()` (`string`) and
`GetClipboardTextU8()` (`ReadOnlySpan<byte>` over native memory: consume
it immediately, never store it); same for `GetVersion`, `TableGetColumnName`,
`GetKeyName`, `GetStyleColorName`, `SaveIniSettingsToMemory`.

### Signatures and ref/pointer overloads

- `InputText(label, ref string buf, int maxLength = 512, flags, callback)`; `InputInt` has
  `step = 0` (no +/- buttons unless `step: 1`); `TreeNodeEx(id, flags, label)` takes the ID
  first; `CheckboxFlags<T>` takes integers only, not enum types.
- Value outputs are `ref`; functions that fill a `Vector2` return it
  (`ImGui.GetContentRegionAvail()`). `ImGui.GetStyleColorVec4(ImGuiCol)` returns `Vector4*`.
- Native structs are `FooPtr` wrappers: `.Handle`, `.IsNull`, `ref` properties, instance methods.

Signature list, scalar families, payloads, callbacks → `references/topic-bindings.md`

### ImTextureID

`readonly struct ImTextureID` wrapping a `ulong Handle`. Constructors from
`ulong`, `nint`, `void*`; implicit conversion only **from `ulong`**;
`IsNull`, `ImTextureID.Null`. In plugins the value always comes from
`IDalamudTextureWrap.Handle` (an `ID3D11ShaderResourceView*`). An `nint` you
kept from old code needs `new ImTextureID(ptr)`.

### Enums and keys

Enum members are PascalCase with acronyms lower-cased (`DisplayRgb`,
`NoSavedSettings`, `Hdr`, `Uint8`). 1.88 still has `ImGuiNavInput` and
`ImGuiModFlags`; `ImGuiKey` named keys start at 512 (`Tab = 512`). Things
added to Dear ImGui after 1.88 do not exist: `ImGuiChildFlags`,
`BeginChild` with child flags, `SeparatorText`, `SetNextItemAllowOverlap`, `ImGuiKeyChord`/`Shortcut`, `BeginItemTooltip`,
`TableAngledHeadersRow`, multi-select. `BeginDisabled`/`EndDisabled`, tables,
docking, viewports and `ImGuiListClipper` are present.

