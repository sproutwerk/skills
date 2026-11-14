## 5. Pitfalls

All read from the pinned source; none reproduced in game.

1. **Directory layout is part of the API.** `..\..\ECommons\...`, `..\..\NightmareUI\...` and
   `..\..\OtterGui\OtterGui.csproj` are hard-coded. A submodule placed elsewhere (for example
   `lib/`) or renamed breaks the wrapper build and silently drops NightmareUI's ECommons project
   reference, leaving only the NuGet `3.2.0.10`.
2. **Version skew.** The libraries pin no ECommons submodule; they compile against whatever commit
   the plugin pins. An ECommons `ImGuiEx` signature change breaks NightmareUI. Bump the three
   NightmareXIV submodules together.
3. **ECommons first.** `new ConfigFileSystem(...)` before `ECommonsMain.Init` dereferences a null
   `ECommonsMain.Instance`. Do not create it in a static or field initialiser that runs earlier.
4. **`If` does not nest and needs `EndIf`.** A section holds one condition. A second `If` replaces
   the first; `Else` just inverts it. The condition is stored on the section and is not reset at
   the start of a frame, so after a missing `EndIf` the condition also gates, from the second
   frame on, the widgets placed before the `If`.
5. **Builder content is captured at build time.** `TextWrapped(text)`, labels and help strings are
   evaluated once. For a builder held in a field, text that must change (localisation, live
   values) belongs inside a `Widget(() => ...)` lambda.
6. **Exceptions are swallowed.** A throwing widget is logged with `e.Log()` and drawing
   continues. An exception between a `Push`/`Begin` and its `Pop`/`End` inside your lambda
   leaves the ImGui stack unbalanced; use `ImRaii` inside lambdas.
7. **Section header colour is cached.** `Color` is filled from `FrameBg` on first draw and
   overwritten with `FrameBgHovered` when a collapsible header is hovered; the code has no path
   that restores it. A long-lived builder keeps the hovered colour and ignores later style
   changes; a builder recreated each frame does not.
8. **Collapse state is keyed by ImGui ID.** `CollapsedHeaders` stores `ImGui.GetID(name +
   "NuiSection")`, which depends on the ID stack. Renaming a section, or drawing it under a
   different parent ID, loses its state; two sections with one name in one scope share it.
9. **`ButtonTabs` rows must not be empty.** A row that is empty, or all `null`, reaches
   `Math.Clamp(maxButtons, 1, 0)`, which throws outside the `try` (`buttons[0]` would too).
10. **Stale active tab.** If the stored name matches no current button (a conditional tab went
    away, or `Name` was used as the key and the language changed), nothing is highlighted and no
    content is drawn until the user clicks. The default is taken from the first row only.
11. **`ButtonTabs` overload without id** keys its state by call stack. Two tab bars drawn from one
    call site (a loop or shared helper) share an id; pass an explicit `id`. The `scaling`
    parameter is accepted but never read.
12. **`child: true` consumes the rest of the window.** `BeginChild` is called with no size, so
    anything drawn after `ButtonTabs` lands below a child that fills the available space.
13. **Static state is per plugin** under Dalamud's per-plugin load contexts *(unverified)*:
    `NuiTools.State`, `Censor.Config` and `WorldSelector.Instance` are statics of the plugin's
    own copy of `NightmareUI.dll`.
