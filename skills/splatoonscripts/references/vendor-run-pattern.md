# Vendor-run pattern

Use the [vendor-run state and confirmation contract](vendor-run.md)
for travel, purchase, return and resume workflows.

Patterns the contract rules out: a wait with `abortOnTimeout: false` followed
by steps that assume the travel happened, a `throw` on a bad quantity that
aborts the queue with the shop open and AutoDuty paused, clicking Yes on any
dialog containing "Exchange", and `ValidTerritories = null` with no login
guard in `OnUpdate`.
