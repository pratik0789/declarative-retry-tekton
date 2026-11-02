I can't tell why `scan` failed. It might have been a brief 503, or a 409 because the epoch moved on. The run was a while ago, so the epoch may have advanced since then anyway.

If it did, the old credential, artifact and signature are rejected. `scan` and then `deploy` would get a 409 unless `acquire`, `publish` and `sign` all run again, in order, in the new epoch.

Re-running those three is still safe if the epoch hasn't changed: they just store fresh records for the current epoch. `warmup` only echoes and touches neither the service nor a workspace, so it can be reused.

RERUN: [acquire, publish, sign]
