The service only accepts each step if the records it depends on come from the current epoch. scan needs a current-epoch signature, sign needs a current-epoch credential and artifact, and publish needs a current-epoch credential.

The run failed a while ago, and the epoch may have advanced since. The scan failure may even have been that 409, not a 503. Reusing acquire, publish and sign risks stale records and another 409 at scan or deploy. Re-running them is safe, because acquire just writes new records for whatever epoch is current.

warmup only echoes a string and has no results or state, so it can be reused.

RERUN: [acquire, publish, sign]
