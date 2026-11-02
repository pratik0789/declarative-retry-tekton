Sign needs a credential and an artifact from the service's current epoch. The run failed a while ago, so the epoch may have moved on (for example after a credential rotation). If it has, reused acquire and publish records would get sign rejected with HTTP 409.
Running acquire and publish again stores fresh records in the current epoch, and doing so is harmless even if the epoch hasn't changed. Scan and deploy never ran, so they will run after sign and stay in the same epoch.
Warmup only echoes "warmed" and has no state or results, so it is safe to reuse.

RERUN: [acquire, publish]
