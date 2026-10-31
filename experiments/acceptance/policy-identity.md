# Policy-content identity revision

The native evidence and experiment code are unchanged from e9eeca1. This subsequent
revision adds metadata for reconstructing the effective policy used by a plan.
It does not add native executions or a practical-value study.

Plans retain the normalized supported policy specification and SHA-256. Generated
runs store compact JSON bytes in `retry.tekton.dev/policy-spec`, and their digest
in `retry.tekton.dev/policy-spec-sha256`. SHA-256 covers those exact UTF-8 bytes,
prefixed `sha256:`. JSON object keys use Go encoding/json ordering; arrays retain
input order. Defaults are explicit as serialized by the supported Policy type.
Metadata and unsupported fields are excluded. This is the effective policy body,
not an archive of arbitrary input YAML or an authenticated certificate.

`internal/planner/policy_identity_test.go` verifies metadata/presentation stability,
content-change detection with unchanged name/generation, and generated snapshot
hash consistency. `results/acceptance/policy-identity-plan.json` and
`policy-identity-manifest.yaml` are a generated offline example from
`testdata/simple`; `policy-identity-audit.json` records an independent byte hash
and parsed-content comparison. `policy-identity-tests.jsonl` records the fresh
full Go test run. The historical native captures retain their original metadata.

Reproduce the example with the README plan/create commands. To audit a generated
manifest, parse YAML, hash the exact policy-spec annotation's UTF-8 bytes, compare
the digest to its policy-spec-sha256 annotation and the plan's policySpecSHA256,
and compare the parsed annotation object to the plan's policySpec. Retaining the
snapshot permits historical rules inspection after an honest local policy edit;
ordinary RBAC and annotation mutability still apply. Other source inputs are not
content-bound by this revision, and the digest says nothing about semantic reuse
safety.
