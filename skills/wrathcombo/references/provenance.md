# Source pins and verification

Reviewed on 2026-11-13. Upstream can change multiple times daily, including
breaking API changes. These snapshots inevitably lag; they do not promise exact
parity with an installed build and are not continuously synchronized. Use this
skill to understand the integration architecture and locate contracts quickly,
then verify the files below against the target version.

## Revision records

| Project | Revision | Metadata |
| --- | --- | --- |
| [WrathCombo](https://github.com/PunishXIV/WrathCombo/tree/a457838183c3cf5a1065ccbd3d70afd1851b3372) | `a457838183c3cf5a1065ccbd3d70afd1851b3372` | `main` head; project version `1.0.4.27`, six commits after tag `1.0.4.27`; Dalamud API 15 |
| [WrathCombo.API](https://github.com/PunishXIV/WrathCombo.API/tree/b40c88cea6e6b62981d9822637a74fd50ed9c44b) | `b40c88cea6e6b62981d9822637a74fd50ed9c44b` | Package `0.5.8` (tag `v0.5.8`, also the `main` head), selected by plugin gitlink |
| [XIVComboPlugin](https://github.com/MKhayle/XIVComboPlugin/tree/a8797120a65354f44433d9d91287c4bc06dab28f) | `a8797120a65354f44433d9d91287c4bc06dab28f` | `master` head; license/README context only; not a claimed ancestor pin |

WrathCombo targets `net10.0-windows`, preview C#, x64, and references local
Dalamud assemblies. Its API project targets `net10.0-windows`, C# 14, and local
Dalamud/FFXIVClientStructs assemblies. Confirm the consumer's compiler, target
framework, and runtime libraries before copying signatures. The main plugin
references the API as a project, not an independently resolved package version.

The three actual gitlinks in the plugin tree are:

| Path | Revision |
| --- | --- |
| `WrathCombo.API` | `b40c88cea6e6b62981d9822637a74fd50ed9c44b` |
| `ECommons` | `26d30b9ec14c3d3c18e335a9e4c7a5421dd134db` |
| `PunishLib` | `96c0cefa8ebf948b3d3d71d85994c61dc8430b5c` |

`.gitmodules` also contains ClientStructs entries, but this tree has no
ClientStructs gitlink. Use the referenced runtime assembly; do not invent a
source pin from those stale entries or borrow another skill's dependency pin.
`Directory.Build.props` replaces PunishLib's ECommons package with the local
project when available. Dependency sources other than WrathCombo.API are omitted.

## Source map

Paths are relative to this reference directory. Read function bodies and callers;
a matching symbol alone does not establish a supported or effective operation.

| Contract to verify against the target | Bundled evidence |
| --- | --- |
| Registered channels, readiness, names, read/write types | [Provider.cs](source/WrathCombo/WrathCombo/Services/IPC/Provider.cs) |
| Wrapper initialization and errors | [Wrapper.cs](source/WrathCombo.API/Wrapper.cs), [IPC/_Helper.cs](source/WrathCombo.API/IPC/_Helper.cs) |
| Consumer call gates | [RawMethods/](source/WrathCombo.API/IPC/RawMethods/) and [IPC/Settings.cs](source/WrathCombo.API/IPC/Settings.cs) |
| Exact enum values and config types | [Enum/](source/WrathCombo.API/Enum/) and [ConfigValueTypeAttribute.cs](source/WrathCombo.API/Attribute/ConfigValueTypeAttribute.cs) |
| Ownership, priority, cancellation, duplicate handling | [Leasing.cs](source/WrathCombo/WrathCombo/Services/IPC/Leasing.cs), [HelpAutoRotConfig.cs](source/WrathCombo/WrathCombo/Services/IPC/HelpAutoRotConfig.cs) |
| Preset discovery, caches, UI effective state | [Search.cs](source/WrathCombo/WrathCombo/Services/IPC/Search.cs), [UIHelper.cs](source/WrathCombo/WrathCombo/Services/IPC/UIHelper.cs) |
| Remote disable, callback dispatch, job setup selection | [Helper.cs](source/WrathCombo/WrathCombo/Services/IPC/Helper.cs) |
| Config conversion and runtime use | [ProvideAutoRotConfig.cs](source/WrathCombo/WrathCombo/Services/IPC/ProvideAutoRotConfig.cs), [AutoRotationConfigIPCWrapper.cs](source/WrathCombo/WrathCombo/AutoRotation/AutoRotationConfigIPCWrapper.cs) |
| Rotation eligibility and execution | [AutoRotationController.cs](source/WrathCombo/WrathCombo/AutoRotation/AutoRotationController.cs) |
| Action request ownership and expiry | [ActionRequestIPCProvider.cs](source/WrathCombo/WrathCombo/Services/ActionRequestIPC/ActionRequestIPCProvider.cs) |
| Hint payload and event | [PositionalHintSnapshot.cs](source/WrathCombo.API/PositionalHintSnapshot.cs), [UpcomingPositionalHintService.cs](source/WrathCombo/WrathCombo/Services/IPC/UpcomingPositionalHintService.cs) |
| Commands and saved-state changes | [Commands.cs](source/WrathCombo/WrathCombo/Commands.cs), [ConfigurationHelper.cs](source/WrathCombo/WrathCombo/Core/ConfigurationHelper.cs) |
| Coexistence diagnostics | [ConflictingPluginsChecks.cs](source/WrathCombo/WrathCombo/Data/Conflicts/ConflictingPluginsChecks.cs) |
| Job changes, initialization, disposal | [WrathCombo.cs](source/WrathCombo/WrathCombo/WrathCombo.cs) |
| API migration history | [Docs/Changes.md](source/WrathCombo.API/Docs/Changes.md) |

The upstream [IPC guide at this pin](https://github.com/PunishXIV/WrathCombo/blob/a457838183c3cf5a1065ccbd3d70afd1851b3372/docs/IPC.md)
and [older consumer example](https://github.com/PunishXIV/WrathCombo/blob/a457838183c3cf5a1065ccbd3d70afd1851b3372/docs/IPCExample.cs)
were also inspected. Their signatures, enum lists, cleanup advice, and comments
can lag the implementation. The older example omits some result handling and
callback invalidation; use the authored lifecycle example with the provider
source instead of treating that sample as a complete consumer.

## Durable guidance and revision-specific details

Keep ownership local to the consumer, respect cancellation, release only your
resources, preserve user configuration, and distinguish accepted work from
observed effects. Those principles survive API changes. Gate names, enum values,
callback timing, cache behavior, job support, and error handling are pinned
observations that need rechecking. In particular, re-read provider, wrapper,
lease manager, and runtime config readers together before relying on a setter.

For a refresh, select the requested plugin revision, resolve its API gitlink,
and inspect the relevant source map entries. Update the authored contract only
where source supports the change, then refresh copied files unchanged and record
new revisions and hashes in the file manifest. Preserve
license and attribution records, including disagreements between upstream files.

## Validation limits

The [manifest](source-manifest.json) records each unchanged copied file's Git
blob and SHA-256. The GNU AGPL license download has a URL and content hash instead
of a Git revision. Original authored guidance is separate from those copies.
Upstream README links retain their original paths and may refer to files outside
this selected mirror; follow their pinned upstream repository for those targets.
All authored skill links resolve within the installed skill or to explicit URLs.

At the 2026-11-13 review, every copied file was compared with its Git object at
the recorded revision, and authored links and anchors were checked. The consumer
example was reviewed against source but not compiled or tested in game. Nothing
was built and no live game actions were run, so there is no runtime guarantee.
