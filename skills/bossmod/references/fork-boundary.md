# Provider identity and porting

This skill covers [awgil/ffxiv_bossmod at c50e67f37aa22f86d064489913ec9887f116accf](https://github.com/awgil/ffxiv_bossmod/tree/c50e67f37aa22f86d064489913ec9887f116accf).
Identify the installed plugin by its manifest, repository and build. A shared
namespace or endpoint name does not establish compatibility with a fork.

Upstream can change several times a day, including breaking changes. Check the
target installation's provider and types before using a pinned contract.

The bundled [provider](source/BossMod/Services/IpcService.cs) registers
`BossMod.*` channels through hosted start/stop. At this revision it has no
`HasActiveModule`, `ActiveModuleName`, `Timeline.*`, `Hints.*` or `Movement.*`
registrations. Its only `AI.*` channels are the read-only `AI.IsEnabled` and
`AI.IsAutoTargetEnabled`; it registers no `AI.SetPreset`, `AI.PauseMovement`
or navigation queries. `HasModuleByDataId` tests registry support; it cannot
identify an active encounter or supply timeline predictions.

Forks can register the same channel name with different delegate types or
behavior. One fork reviewed on 2026-11-13 declares `Configuration` with
`List<string>` and honors its `save` flag, and its `Presets.GetActive` reflects
a single active preset. Read the provider's registration before declaring the
subscriber.

Before porting, verify each channel's generic types, registration kind, return
semantics and state effects. If a consumer needs a capability absent from this
provider, report the limit and design an explicit extension against the
requested VBM source. Do not install another fork as a fallback or substitute
reflection into the active encounter without treating it as an integration
with version-specific internals.

The [module base](source/BossMod/BossModule/BossModule.cs) uses
`BossModule(ModuleInit, WPos, ArenaBounds)`. Recheck `ModuleInit`, state-builder
construction, actor collections, component callbacks and AI hint fields when
porting module code. Matching names do not guarantee matching semantics.

These contracts work without a companion rotation-plugin skill. Source review
does not establish successful plugin loading or runtime IPC behavior.
