## IPC

Call gates are typed by generic arguments: up to 8 parameters plus `TRet` last (`TRet = object` for
actions). Calls run synchronously on the **caller's thread**; a missing provider throws
`IpcNotReadyError`. Share only primitives, strings and tuples/collections of them. Names are global
strings: prefix with your InternalName. Full API, errors, DataShare → `references/topic-ipc.md`

