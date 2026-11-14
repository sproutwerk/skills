# schemas/: condition tree JSON Schemas

Covers the five JSON Schema files under `schemas/` for the serialized form of `Luna/Conditions`.
Verified against Luna `52eb0fe` and ImSharp `4d09327` (branch `gc1.88`, ImGui 1.88). Paths are
relative to the skill root. Grep `references/api-index.txt` before using a member that is not shown
here.

`and_condition.json`, `or_condition.json`, `not_condition.json`,
`true_condition.json`, `false_condition.json`: JSON Schema draft 2020-12
fragments for the serialized form of `Luna/Conditions`. Each requires a `Type`
constant (`"And"`, `"True"`, …); list conditions carry `Conditions`, an array
of `{ "$dynamicRef": "#Condition" }`, so a consumer schema supplies the
`Condition` anchor with its own leaf condition types. They are solution items
in `Luna.slnx`; no C# file references them.
