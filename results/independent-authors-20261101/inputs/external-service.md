# Environment: the service the Pipeline's tasks call

Every task calls `POST http://oracle:8000/execute/<task>` on a shared external service. The service keeps a current epoch and these records:
- acquire: stores a credential for the current epoch.
- publish: accepted only if the stored credential is from the current epoch; stores an artifact for the current epoch.
- sign: accepted only if the stored credential and artifact are from the current epoch; stores a signature for the current epoch.
- scan: accepted only if the stored signature is from the current epoch; stores a scan record for the current epoch.
- deploy: accepted only if the stored scan record is from the current epoch.
A rejected call returns HTTP 409 and the task fails. The service's operators advance the epoch from time to time (for example on credential rotation); after that, records from earlier epochs are rejected. The service can also be briefly unavailable (HTTP 503).
