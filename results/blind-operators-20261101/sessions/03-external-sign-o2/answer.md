Why acquire and publish must run again:
- `sign` only works if the stored credential and artifact come from the current epoch. We don't know why `sign` failed (a 409 from an epoch change, or a brief 503), and the service may have moved to a new epoch since.
- If the old credential and artifact are reused and the epoch has changed, `sign` will get a 409 again. Re-running `acquire` and `publish` stores fresh records in the current epoch, so `sign`, `scan` and `deploy` all line up.
- `warmup` only runs `echo`, so it has no external state and can safely be reused.

RERUN: [acquire, publish]
