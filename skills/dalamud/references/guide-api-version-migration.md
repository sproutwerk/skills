## API version migration

Upgrade step for every bump: change the csproj header to `Dalamud.NET.Sdk/<N>.0.0`, rebuild, fix
errors. ClientStructs breaks separately each patch. To stay forward compatible on 15: keep SeString
use in one place, never pass `HookBackend`. Per-version notes (9-16) → `references/topic-migration.md`

