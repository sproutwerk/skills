## Signature resolution at runtime

`Resolver` (`InteropGenerator.Runtime`) scans the `.text` section once for
every `Address` registered by `Addresses.Register()` (bucketed by first byte,
parallel, optional JSON cache keyed by game version). Unresolved addresses
stay `0`; the generated wrapper then throws `ThrowNullAddress("Struct.Func",
"sig")` on call — that exception is what "signature broke this patch" looks like.

**Dalamud does all of this at startup.** A plugin using Dalamud's bundled
copy never calls the resolver. Only for a standalone/local copy:

```csharp
InteropGenerator.Runtime.Resolver.GetInstance.Setup(sigScanner.SearchBase /*optional clean module copy*/, version, cacheFile);
FFXIVClientStructs.Interop.Generated.Addresses.Register();
InteropGenerator.Runtime.Resolver.GetInstance.Resolve();
```

`ResolverTester` does the same against a file on disk to list broken sigs
after a patch.

