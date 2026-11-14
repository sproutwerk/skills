# Source provenance and snapshot limits

`openapi.json` is a saved [XIVAPI v2 API description](https://v2.xivapi.com/api/openapi.json)
with server version `1-48c1413`, fetched 2026-11-13. Curated `schemas/*.yml`
came from [EXDSchema](https://github.com/xivdev/EXDSchema), using its `latest`
branch. The server's default schema revision is older than that branch, so do
not assume the bundled copies match what a response was decoded with; compare
against the `schema` value in the response and refresh against that revision
when exact reproducibility matters.

## Review on 2026-11-13

The [version endpoint](https://v2.xivapi.com/api/version) maps `latest` to
`a72e645591e5bf8b` (`7.58`). A bounded `Item/4?fields=Name` lookup returns
that version with schema
`exdschema@2:rev:e773c41a90aed788cf4c1c48469fa85618ef01fb`
(EXDSchema `ver/2026.09.01.0000.0000`), while `schema=exdschema@2:latest`
resolves to `a1e24a0c72058353f30f23beb996792cdb926868`.

The live OpenAPI reports server `1-48c1413`
([boilmaster source](https://github.com/ackwell/boilmaster/tree/48c1413b1eb8e333799e8b467822a50a2c55a59d),
the head of `main` and `release`). The bundled `openapi.json` is that response
byte for byte. It differs from the earlier `1-f0d5f0a` copy only in
`info.version`; paths, parameters and component schemas are unchanged. All 26
bundled sheet schemas are byte-identical to EXDSchema
[a1e24a0c](https://github.com/xivdev/EXDSchema/tree/a1e24a0c72058353f30f23beb996792cdb926868),
still the `latest` head.

Checked live on this date: every endpoint, parameter, error message and
behaviour stated in `SKILL.md`, a sample of rows in
[lookup-recipes.md](lookup-recipes.md) (item, tomestone slots, shop reverse
lookup, subrow paging, aetheryte count, MapMarker, Map 1025), and the helper's
`search`, `version`, `world2map` and `map2world` subcommands. On 2026-11-14
the helper was rewritten in Go (`scripts/xivapi/`); 36 live read-only
invocations covering every subcommand gave the same output and exit codes as
the Python script it replaces. One behaviour
changed since the September run: `language=chs`, `cht` and `kr` now return
HTTP 400 although the OpenAPI enum still lists them. These checks do not
revalidate every cookbook result. The cookbook's `latest` label is historical;
request its explicit game key and schema when reproducing those results.

The [EXDSchema repository-license endpoint](https://api.github.com/repos/xivdev/EXDSchema/license)
returned 404 on 2026-11-13, as it did on 2026-10-03. This is not proof of
missing permission: a license or grant may exist elsewhere. This skill does
not grant rights to upstream material. The saved API description likewise
keeps its upstream provenance and has not been assigned a license here.

## Lookup snapshot

The cookbook and most lookup-recipe results were recorded live on 2026-09-23
against `version=541c0c12e07da325` (7.56x1, `latest` at the time),
`schema=exdschema@2:rev:e773c41a…` (EXDSchema `ver/2026.09.01.0000.0000`).
EXDSchema's `latest` head was `f3cacf3b…` on that date, which is the revision
the cookbook shows for `schema=exdschema@2:latest`.
Full query transcripts are in [cookbook.md](cookbook.md); the bundled schema files are under `schemas/`.
