# The bindings in depth

Covers how the binding assembly is built, common signatures and overload families, ImGuiP
internals, and ref / pointer / `*Ptr` struct types. For `ImU8String`, `ImTextureID` and enum rules,
read the [binding examples](guide-the-bindings.md).
Verified against goatcorp/Dalamud `master` @ `43b3411` (2026-11-10, API level 15, Dear ImGui 1.88 docking).
Paths are relative to the skill root (`SKILL.md`, `references/topic-*.md`, `references/...`); a bare
`index/...` or `src/...` path means `references/index/...` or `references/src/...`.

## How the assembly is put together

`imgui/Dalamud.Bindings.ImGui/` contains `Generated/` (raw HexaGen output),
but the csproj has `<Compile Remove>` for `Generated/Functions/**`,
`Generated/Structs/**`, `Internals/Functions/**` and `Manual/Functions/**`.
What actually compiles is:

- `Custom/*.cs`: hand-written overloads (text, scalars, combos, colors, callbacks).
- `Custom/Generated/**/*.gen.cs`: trimmed generator output (non-string
  functions, struct definitions, `*Ptr` wrappers).
- `Generated/Enums`, `Constants`, `Delegates`, `FunctionTable.cs`, plus
  `ImU8String.cs`, `ImTextureID.cs`, `ImVector.cs`.

So do not read `Generated/Functions/Functions.NNN.cs` to learn the API: those
overloads (`string`/`ref byte`/`ReadOnlySpan<byte>` permutations) are **not
in the DLL**. Use `references/index/imgui-methods.txt`.

| Class | Role |
|---|---|
| `ImGui` | Public API. Static class; struct methods are also exposed as instance methods on `*Ptr` |
| `ImGuiP` | `imgui_internal.h`: `GetCurrentWindow`, `ItemAdd`, `ItemSize`, `RenderText`, `ButtonBehavior`, `DockBuilder*`, `TableGetHoveredColumn`, `ImHashStr` ... |
| `ImGuiNative` / `ImGuiPNative` | Thin function-pointer calls with raw pointers |

The static constructor loads `cimgui.dll` from next to the calling assembly;
plugins never initialise or configure that.

## Signatures you will actually use

```csharp
bool Begin(ImU8String name, ImGuiWindowFlags flags = None);
bool Begin(ImU8String name, ref bool open, ImGuiWindowFlags flags = None);
bool BeginChild(ImU8String strId, Vector2 size = default, bool border = false, ImGuiWindowFlags flags = None);
bool Button(ImU8String label, Vector2 size = default);
bool Checkbox(ImU8String label, ref bool v);
bool CheckboxFlags<T>(ImU8String label, ref T flags, T flagsValue) where T : IBinaryInteger<T>;  // integers only, not enum types
bool RadioButton<T>(ImU8String label, ref T v, T vButton);              // any T, enums included
bool Selectable(ImU8String label, bool selected = false, ImGuiSelectableFlags flags = None, Vector2 size = default);
bool InputText(ImU8String label, scoped ref string buf, int maxLength = 512,
               ImGuiInputTextFlags flags = None, ImGuiInputTextCallbackDelegate? callback = null);
bool InputText(ImU8String label, Span<byte> buf, ImGuiInputTextFlags flags = None, ...);   // UTF-8 buffer, no alloc
bool InputTextWithHint(ImU8String label, ImU8String hint, scoped ref string buf, int maxLength = 512, ...);
bool InputTextMultiline(ImU8String label, scoped ref string buf, int maxLength = 512, Vector2 size = default, ...);
bool InputInt(ImU8String label, scoped ref int data, int step = 0, int stepFast = 0, ImU8String format = default, ImGuiInputTextFlags flags = None);
bool SliderFloat(ImU8String label, scoped ref float v, float vMin = 0, float vMax = 0, ImU8String format = default, ImGuiSliderFlags flags = None);
bool DragFloat(ImU8String label, scoped ref float v, float vSpeed = 1, float vMin = 0, float vMax = 0, ImU8String format = default, ImGuiSliderFlags flags = None);
bool DragScalar<T>(ImU8String label, scoped ref T v, float vSpeed, scoped in T vMin, scoped in T vMax, ...) where T : unmanaged, IBinaryNumber<T>;
bool ColorEdit4(ImU8String label, scoped ref Vector4 col, ImGuiColorEditFlags flags = None);
bool Combo(ImU8String label, ref int currentItem, ReadOnlySpan<string> items, int popupMaxHeightInItems = -1);
bool Combo<T>(ImU8String label, ref int currentItem, IReadOnlyList<T> items, Func<T, string> toString, int popupMaxHeightInItems = -1);
bool TreeNodeEx(ImU8String id, ImGuiTreeNodeFlags flags = None, ImU8String label = default);  // first arg is the ID
void Image(ImTextureID userTextureId, Vector2 size /*, uv0, uv1, tintCol, borderCol */);
bool ImageButton(ImTextureID userTextureId, Vector2 size /*, uv0, uv1, framePadding, bgCol, tintCol */);
void TableSetupColumn(ImU8String label, ImGuiTableColumnFlags flags = None, float initWidthOrWeight = 0, uint userId = 0);
Vector2 CalcTextSize(ImU8String text, bool hideTextAfterDoubleHash = false, float wrapWidth = -1);
bool SetDragDropPayload(ImU8String type, ReadOnlySpan<byte> data, ImGuiCond cond = None);
ImGuiPayloadPtr AcceptDragDropPayload(ImU8String type, ImGuiDragDropFlags flags = None);
```

Multi-component scalars use vectors or spans: `DragFloat2(label, ref Vector2)`,
`SliderFloat3(label, ref Vector3)`, `InputFloat4(label, ref Vector4)`,
`DragInt(label, Span<int> v, ...)` for N ints. `InputScalar`/`DragScalar`/
`SliderScalar` are generic and infer `ImGuiDataType` from `T`; typed
shortcuts exist for `InputByte`, `InputLong`, `InputUInt`, `InputDouble` and
more (grep the index).

## ref, pointer and `*Ptr`

- Value outputs are `ref`. Functions that fill a `Vector2` return it:
  `ImGui.GetContentRegionAvail()`, `GetCursorScreenPos()`, `GetWindowPos()`,
  `GetMousePos()`; `(ref Vector2 pOut)` and `(Vector2* pOut)` variants also exist.
- Generated overloads with several outputs exist in every `ref`/pointer
  permutation (`ColorConvertHSVtoRGB(h, s, v, ref r, ref g, ref b)`).
- `ImGui.GetStyleColorVec4(ImGuiCol)` returns `Vector4*`: dereference in
  `unsafe`, or use `ImGui.GetStyle().Colors[(int)ImGuiCol.Text]`.
- Native structs come as `FooPtr` wrappers (`ImGuiIOPtr`, `ImGuiStylePtr`,
  `ImDrawListPtr`, `ImFontPtr`, `ImGuiViewportPtr`, `ImGuiPayloadPtr`,
  `ImGuiListClipperPtr`, `ImGuiTableSortSpecsPtr`): `.Handle` is the raw
  pointer, `.IsNull` checks it, members are `ref` properties
  (`ImGui.GetIO().DeltaTime`, `ImGui.GetStyle().ItemSpacing`), native methods
  are instance methods (`drawList.AddRectFilled(...)`, `clipper.Step()`).
- Drag-drop payload: `payload.IsNull`, `payload.IsDataType("X"u8)`,
  `payload.Handle->DataSpan` (`Span<byte>`), `payload.IsDelivery()`.
- Callbacks: `InputText` takes managed delegates
  (`ImGuiInputTextCallbackDelegate(scoped ref ImGuiInputTextCallbackData)`, or
  the `...InContextDelegate<TContext>` / `...RefContextDelegate<TContext>`
  forms that pass state without a closure).
