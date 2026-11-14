# Compilation requirements

The [recorded Splatoon snapshot](provenance.md) uses the .NET 10 SDK and C#
preview. Official scripts use the `field` keyword (`=> field ??= new()`),
collection expressions (`[.. xs]`, `[]`), primary constructors, and `ref`
locals via `Ref<bool>.Get(...)`.

This skill's source folder is not a build project. For offline compilation,
use the full matching Splatoon checkout: its official csproj references sibling
`ECommons`, `ECommons.IPC`, `Splatoon`, and `WrathCombo.API` projects. It resolves
Dalamud DLLs through `%APPDATA%\XIVLauncher\addon\Hooks\dev\`, so a machine
with the matching XIVLauncher dependencies is needed. Inspect the installed
build's source and project files before applying these requirements to another
revision. A clean compile checks syntax and types; test IPC, addon, and lifecycle
behavior separately in an authorized game session.
