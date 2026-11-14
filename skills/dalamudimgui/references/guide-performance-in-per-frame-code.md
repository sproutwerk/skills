## Performance in per-frame code

`Draw` runs every rendered frame on the game's main thread; time spent there
is frame time for the game.

| Do | Instead of |
|---|---|
| `"Label"u8` for constant text | `"Label"` (transcoded each frame) |
| `ImGui.Text($"{a}/{b}")` | `ImGui.Text(a + "/" + b)`, `string.Format`, `.ToString()` |
| Cache filtered/sorted lists; recompute on input change | LINQ `Where/OrderBy/ToList` inside `Draw` |
| `ImGuiClip` / `ImGuiListClipper` for long lists | Emitting thousands of rows |
| Cache Lumina/Excel row lookups and derived strings | Sheet lookups per row per frame |
| One chained `ImRaii.PushColor(...).Push(...)` | Several separate push objects, or pushes inside the row loop |
| `GetFromGameIcon(...)` per frame (cheap, cached) | `CreateFrom*` per frame |
| `Window.Update()` for cheap per-frame state, framework events for game state | Polling expensive game state in `Draw` |
| Background `Task` for IO, results marshalled back into fields | File/network IO in `Draw` |

`ImU8String` interpolation avoids the `string`, but each hole still formats
its value; for thousands of cells per frame, cache the formatted text.

