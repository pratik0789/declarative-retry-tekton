I'm rerunning `acquire` and reusing `warmup`.

- `warmup` only runs `echo warmed`. It has no side effects and no outputs anything downstream needs, so it's safe to reuse.
- `publish` is only accepted if the stored credential is from the service's current epoch. The run failed a while ago and the epoch may have moved on since, for example after a credential rotation. The failure itself may have been an epoch rejection (409) rather than a brief outage (503).
- Rerunning `acquire` stores a fresh credential for the current epoch. That costs little and always works, so `publish`, `sign`, `scan` and `deploy` can then go through in one epoch.

RERUN: [acquire]
