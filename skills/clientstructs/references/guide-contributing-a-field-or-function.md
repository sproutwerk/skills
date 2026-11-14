## Contributing a field or function

1. Load `ffxiv_dx11.exe` of the matching version in IDA or Ghidra
   (`references/upstream-docs/Ghidra-Getting-Started.md`). Run
   `ida/ffxiv_idarename.py` (needs `pyyaml anytree`; IDA also needs
   `idauser.cfg` in `%AppData%\Hex-Rays\IDA Pro\`) to apply `data.yml`
   names; optionally `ffxiv_exdgetters` and `ffxiv_structimporter.py`
   (imports `ffxiv_structs.yml`; IDA < 9 only).
2. Locate the function/field. Make a signature with Ghidra
   `SigScanner.java` (Ctrl-Alt-S) or `caraxi/sigmaker-x64` for IDA: place the
   cursor on the instruction (or the function start — the tool then sigs a
   caller `E8 …` call, which is what CS prefers for `[MemberFunction]`),
   uppercase hex, `??` for rel32/displacements.
3. Edit the struct: `[FieldOffset(0x…)] public T Name;` for fields (type must
   be unmanaged; fixed arrays via `[FixedSizeArray] internal
   FixedSizeArrayN<T> _name;`), `[MemberFunction("sig")] public partial R
   Name(args);` for functions (omit `this`; `static` if no instance;
   `byte*`/`CStringPointer` for C strings + `[GenerateStringOverloads]`).
   Keep `Size` correct; mark the struct `partial`. Renames: keep the old
   member with `[Obsolete("Renamed to X")]` at the same offset.
4. Add the address to `ida/data.yml` under the class (`funcs:` /
   `vfuncs:` / `instances:` / `vtbls:`) so the renamer and CExporter can
   cross-check. CI (`validate-data.yaml`) runs the validator on `data.yml`
   changes.
5. `dotnet run --project CExporter/CExporter.csproj -c Release` — same as
   `pr-change-check.yaml`. Non-empty `ida/errors.txt` (overlapping fields,
   size mismatches, name collisions between C# and `data.yml`, missing
   `data.yml` entries) fails the PR with a bot comment and the
   `requested changes` label. `CompatChecker` (apicompat main vs PR) adds the
   `breaking change` label for public API breaks. Nightly `struct-update.yaml`
   regenerates `ida/ffxiv_structs.yml` on `main`.
6. PR to `main`; patch-day work goes through the same check
   (`pr-change-check-patch-day.yaml`).

