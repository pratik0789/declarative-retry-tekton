# Trusted-artifact expiry in Konflux build Pipelines

Pipelines: `pipelines/*-oci-ta*` in konflux-ci/container-build-catalog at the commit in
`container-build-catalog-commit.txt`. Trusted-artifact tasks take `ociArtifactExpiresAfter`
("Expiration date for the trusted artifacts created in the OCI repository"); Konflux
pull-request PipelineRuns set `image-expires-after: 5d`. For every ordinary task as the
failed one, `internal/planner/konflux_expiry_study_test.go` (run with
`CBC=<checkout> OUT=rows.json`) checks whether a recovery planned after expiry would make an
executing task read a `*_ARTIFACT` result of a reused producer, for: failed task only,
every warning becomes a rerun, all execution ancestors, full restart, a retryWith policy derived from the artifact flows (each consumer retries with its producers), and a policy with
`resumeWithin: 96h` (refuses every recovery requested after expiry). Offline plans only.
