# Dalamud ImGui

Dalamud draws plugin UI with Dear ImGui, but **not** through ImGui.NET. Since
API 13 it ships its own bindings (`Dalamud.Bindings.ImGui`, `.ImPlot`,
`.ImGuizmo`) plus a helper layer in `Dalamud.Interface`. dalamud.dev has
almost no documentation for either; the source is the documentation.

This skill tracks goatcorp/Dalamud `master` at commit `43b3411` (2026-11-10).
Its csproj still reads `DalamudVersion 15.0.3.9`; tag `15.0.4.0` is this commit
plus the version bump. That is **API level 15** (.NET 10, C# 14). Native library:
**Dear ImGui 1.88, docking branch** (`IMGUI_VERSION_NUM 18800`, with
`IMGUI_HAS_DOCK` / `IMGUI_HAS_VIEWPORT`), ImPlot 0.14, via goatcorp's cimgui
forks. Bindings are generated with **HexaGen** from goatcorp's fork of
Hexa.NET.ImGui (`HexaGen.Runtime 1.1.20`). Migration notes come from
goatcorp/dalamud-docs @ `3c7597d` (2026-11-13).

Everything named below was checked against that checkout unless marked
*(unverified)*. This skill assumes you know Dear ImGui itself (immediate mode,
ID stack, Begin/End pairing, tables, draw lists) and documents only what
Dalamud changes or adds.

Reference material, all under `references/`:

| Path | What |
|---|---|
| `src/<upstream path>` | Verbatim copies of 206 load-bearing files (list: `manifest.txt`) |
| `index/imgui-methods.txt` | Every public `ImGui.*` overload (1605 lines), grep-able, with source file |
| `index/imguip-methods.txt` | Every public `ImGuiP.*` overload (internals) |
| `index/imgui-struct-methods.txt` | Methods on `ImDrawListPtr`, `ImFontPtr`, `ImGuiIOPtr`, `ImU8String` ... |
| `index/implot-methods.txt`, `imguizmo-methods.txt` | ImPlot (pointer overloads dropped) / ImGuizmo |
| `index/enums.txt` | Every binding enum with members and values |
| `index/interface-methods.txt`, `interface-files.txt` | Signatures in the copied `Dalamud/Interface` files; all 449 files there |
| `index/fontawesome-icons.txt` | 1381 `FontAwesomeIcon` members (Font Awesome 7.1.0) |
| `upstream-docs/v13.md`, `v14.md`, `v15.md`, `v16/index.md` | API migration notes |
| `upstream-docs/v16/sestring.md` | API 16 guide to replacing `SeString` with `ReadOnlySeString` |

Before using an overload not shown here: `grep ' Name(' references/index/imgui-methods.txt`.

Deep, situational material lives in first-party files under `references/topic-*.md` (see the topic index
below). Each section here ends with a pointer `→ references/topic-<file>.md`; open that file before
writing code in that area. All paths are relative to the skill root.

Out of scope:
- Non-UI Dalamud services (framework, commands, IPC, game data): sibling `dalamud` skill.
- Luna / ImSharp (Ottermandias' successor to OtterGui): sibling `luna` skill.
- Native game UI (Atk addons, agents) through FFXIVClientStructs: `clientstructs` skill.


## Guide sections

Paths written as `references/...` are relative to the skill directory. Read only the section needed for the task.

- [topic index](guide-topic-index.md)
- [project setup](guide-project-setup.md)
- [migrating from imguinet](guide-migrating-from-imguinet.md)
- [the bindings](guide-the-bindings.md)
- [iuibuilder hooking up drawing](guide-iuibuilder-hooking-up-drawing.md)
- [windowsystem and window](guide-windowsystem-and-window.md)
- [imraii](guide-imraii.md)
- [global scale and imguihelpers](guide-global-scale-and-imguihelpers.md)
- [imguicomponents](guide-imguicomponents.md)
- [fonts](guide-fonts.md)
- [textures and icons](guide-textures-and-icons.md)
- [tables and large lists](guide-tables-and-large-lists.md)
- [sestring rendering](guide-sestring-rendering.md)
- [notifications](guide-notifications.md)
- [file dialog](guide-file-dialog.md)
- [drag and drop](guide-drag-and-drop.md)
- [colors and style](guide-colors-and-style.md)
- [implot and imguizmo](guide-implot-and-imguizmo.md)
- [performance in per frame code](guide-performance-in-per-frame-code.md)
- [pitfalls](guide-pitfalls.md)

## Changes since Dalamud 15.0.3.5

The previous pin was `985a56f` (2026-09-28, `15.0.3.5`). A plugin running on
an older Dalamud build lacks the members below; check the target runtime
before relying on them.

- First tagged in `15.0.3.8`: `ImRaii.PushClipRect` with `ClipRectDisposable`,
  and the `Rectangle` and `LineEvaluator` types in `Dalamud.Interface.Utility`
  ([ImRaii catalog](topic-imraii.md#clip-rectangles)).
- First tagged in `15.0.3.8`: `ImGuiHelpers.CenteredSeString`
  ([SeString](topic-sestring.md)).
- First tagged in `15.0.3.8`: the `Header(label, ref visible, flags)`
  constructor of `HeaderDisposable` sets `Alive`, so its `PopID` runs. Earlier
  builds leak one ID push per open header.
- First tagged in `15.0.4.0`: `Easing.Elapsed`. `Easing` now tracks time with
  timestamps instead of a `Stopwatch` field.
- `UiBuilder.FontAtlas` is created on first access. `WindowSystem.Windows`
  returns the system's own list instead of a new copy per call.

The binding projects under `imgui/` and every submodule pin except
FFXIVClientStructs are unchanged between the two commits, so the `ImGui`,
`ImGuiP`, `ImPlot` and `ImGuizmo` indexes are identical to the previous ones.
No build or in-game test was run for this refresh.
