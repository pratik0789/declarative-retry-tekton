Rerunning prepare rewrites `source.txt` on the shared PVC, and the content includes the new taskRun name, so the file's hash changes.
If build is reused, `artifact-input.sha` and `artifact.txt` still hold the old hash, and publish fails with STALE-ARTIFACT. So build must run again after prepare.
observe only echoes text and produces no files or results, so reusing it is safe. gate runs again anyway because it failed, and publish never ran.

RERUN: [prepare, build]
