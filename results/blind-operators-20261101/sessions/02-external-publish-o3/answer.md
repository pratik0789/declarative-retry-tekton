The credential that `acquire` stored belongs to whatever epoch was current when the run happened. We don't know why `publish` failed (a 409 from a stale epoch, or a 503), or whether the epoch has advanced since.
If we reuse `acquire` and the epoch has moved on, `publish` will get a 409 again. Running `acquire` again is cheap and always stores a credential for the current epoch.
`warmup` only runs `echo`, so it leaves nothing on the service and doesn't need to run again. `sign`, `scan` and `deploy` never ran, so they run anyway.

RERUN: [acquire]
