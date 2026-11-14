## Luna or stock Dalamud?

| Situation | Use |
|---|---|
| Small plugin, a settings window, MIT/other license wanted, no submodule appetite | Stock `ImGui` + `ImRaii`. Luna is a submodule with no releases and AGPL-3.0 text. |
| Large tables or lists redrawn every frame, heavy formatted text | ImSharp: no per-call string marshalling allocations, cached filtered tables, list clipper helpers. |
| Filterable/sortable tables, filter combos, folder trees, changelog, tutorial, hotkey pickers | Luna; stock Dalamud has no equivalents of these classes. |
| Reading or contributing to Penumbra / Glamourer | Luna; that code no longer references `ImGui.*` (`Use_Dalamud_ImGui=false`). |
| Existing OtterGui plugin | Port when you need new features; OtterGui's last commit is 2026-04-29. |
| Mixing | Possible: same native library and context. Keep push/pop pairs within one API, and convert enums and texture IDs with the helpers above. |

Luna's DI, events, save service, logging and IPC wrappers are independent of
the UI parts; a plugin can use them and still draw with stock `ImGui`, as long
as `ImSharpDalamudContext` is never required by what it touches.

