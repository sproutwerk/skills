# Vendor-run state and confirmation contract

Use this when writing a travel → purchase → return → resume workflow. This is
an implementation contract, not a claimed runnable or in-game-tested script.

## Successful path

1. Capture the character/content ID, destination, expected shop and currency,
   item ID, unit price, quantity cap, and required return destination. Validate
   these against the installed plugin and current game data.
2. Request travel only while the travel provider is idle. Lifestream's
   `EnqueueCustomAliasFromString` and `EnqueueInnShortcut` return `void`: set
   `travelRequested`, not `travelAccepted`. Require provider idle **and** the
   expected territory/world and shop identity before selecting anything.
3. Re-fetch the ready shop addon and player state inside the purchase task.
   Verify the currency identity, item ID, nonzero price, affordability and
   inventory capacity. `ShopExchangeCurrency.CurrencyId` in the pinned helper
   infers an item from its icon; verify that mapping for the shop rather than
   treating any matching icon as a unique currency identifier.
4. Immediately before `Select(quantity)`, capture item/currency counts and the
   expected transaction values. Set `purchaseIssued` only after selection.
   A missing item, zero price, unaffordable purchase or skipped selection must
   bypass confirmation entirely.
5. Confirm only while `purchaseIssued && !confirmationSent`, with the original
   character, matching live shop and the expected transaction dialog. Validate
   the localized item, quantity and cost using inspected addon fields or a
   complete expected prompt derived from that locale. A substring such as
   `Contains("Exchange")`, or merely the existence of `SelectYesno`, does not
   identify the transaction. If identity cannot be established, stop for review.
   Send Yes once, then wait for the dialog to close and the expected inventory
   and currency deltas. Do not repeatedly click Yes while waiting.
6. Mark `bought` only after verification. Close only the shop/dialog owned by
   this operation. Request return travel only when the provider is idle and
   owned UI is closed. Require provider idle, the configured return destination
   (territory/world and room/location if relevant), player readiness and no
   outstanding transaction before setting `returned`.
7. Resume AutoDuty only when the run's success criteria are met (`bought`,
   `returned`, original character, AutoDuty still paused by this operation).
   A different failure policy must be explicit; never resume just because a
   timeout discarded its wait task.

## Failure and retry boundaries

Every wait has a finite deadline. For dependent work use explicit
`abortOnTimeout: true` / `abortOnError: true`, or a state machine that leaves
success flags false on failure. TaskManager defaults may be overridden by its
owner; `new(timeLimitMS: ...)` alone does not ensure abort behavior.

On timeout, logout, character change, unexpected dialog, uncertain purchase or
return failure: stop dependent actions, report the last verified state and leave
AutoDuty paused. Cleanup is best effort and scoped to this run. Do not cancel
another operation's Lifestream queue or close unrelated dialogs.

A timed-out confirmation may already have committed a purchase. Reconcile item
and currency changes before retrying; uncertainty is a stop condition, not a
reason to repeat the purchase. Retry only a verified no-effect failure, with a
finite attempt budget and delay. Treat logout and IPC exceptions as transitions
to failure even though the script's `OnUpdate` readiness guard skips its own
body: a separately owned TaskManager still needs its queued tasks guarded or
aborted.

## Validation cases before enabling the workflow

Exercise refused travel, arrival in the right territory with the wrong shop,
wrong currency, missing item, zero price, insufficient funds, full inventory,
skipped purchase with an unrelated Exchange dialog, delayed inventory update,
confirmation timeout, logout/character switch, rejected return request and
return timeout. None may confirm an unrelated dialog or resume dependent work.
No live purchase, travel, or AutoDuty action was performed for this reference.

## Sources

- [Lifestream IPCProvider at b363201](https://github.com/NightmareXIV/Lifestream/blob/b363201/Lifestream/IPC/IPCProvider.cs): void enqueue methods and busy handling.
- Bundled `ecommons/TaskManagerConfiguration.cs`: non-aborting timeouts discard a task and continue the queue; unset options inherit manager configuration.
- Bundled `ecommons/AddonMaster/ShopExchangeCurrency.cs`: currency/icon lookup, purchase callback and item fields.
