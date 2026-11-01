# Practitioner-labelled hazards in Konflux release Pipelines

Pipelines: `pipelines/managed/*` in konflux-ci/release-service-catalog at the commit in
`release-service-catalog-commit.txt`. Labels: the ten managed tasks that
konflux-ci/release-service-utils pull request 1109 (open) instruments with the merged
retry-safety helper (pull request 1078) because they write to external systems
(create-github-release, create-pyxis-image, make-repo-public, publish-pyxis-repository,
publish-to-mrrc, push-artifacts-to-storage, push-rpm-data-to-pyxis, push-rpms-to-pulp,
push-snapshot, rh-sign-image-cosign). Thirteen managed Pipelines contain one; none
declares a Workspace.

`internal/planner/konflux_release_study_test.go` (run with `RSC=<checkout> OUT=cells.json`)
plans every ordinary task as the failed task under the upstream and independent
completion assumptions and counts, per rule, reruns of labelled tasks: full restart,
all execution ancestors, every warning becomes a rerun, failed task only, and a policy
derived from the labels (blocksResume on each labelled task, no retryWith). Offline
plans only; no task executed.
