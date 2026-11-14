## 3. NightmareUI.OtterGuiWrapper

A second repo that joins NightmareUI to OtterGui's `FileSystem<T>`. Public surface, all under
`NightmareUI.OtterGuiWrapper.FileSystems`:

| Type | What it is |
|---|---|
| `Configuration.ConfigFileSystem` | two-pane settings body: filterable page tree left, selected page right; `Reload()`, `Draw(float? width = null)` |
| `Configuration.ConfigFileSystemEntry` | base class of a page: `Path`, `Builder`, `Draw()`, `DisplayPriority`, `NoFrame`, `ShouldDisplay()` |
| `Configuration.ConfigFileSystemHelpers` | `CreateInstancesOf<T>()`, reflection-based page discovery |
| `Configuration.SimplifiedSelector.FileSystemSelector<T, TStateStorage>` | cut-down copy of OtterGui's selector |
| `Generic.GenericFileSystem<TData>`, `Generic.IFileSystemStorage` | user-editable item tree persisted to JSON, with toolbar |

It exists separately because it is the only part that needs OtterGui (inference from the
project files; Splatoon uses NightmareUI with no OtterGui submodule). Signatures, behaviour, the
AutoRetainer selector example and the wrapper's pitfalls: `references/topic-otterguiwrapper.md`.

