# Matched stateful recovery comparisons

This experiment isolates `retryWith` from partial recovery in general. It runs
two author-created counterexamples on Tekton v1.6.0 / Kubernetes v1.31.4 / kind
v0.27.0, with five matched triplets per mechanism. These are feasibility
counterexamples, not a sample of production failures or policies from
independent authors.

## Experimental contract

One failed source PipelineRun and its TaskRuns are exported per triplet. Its
actual Workspace bytes are copied into three isolated persistent claims. The
harness checks equality before every arm. The empty-policy and closure arms
use the same source objects and generated recovery logic; the full-restart arm
uses the referenced Pipeline. Generated recovery manifests are retained before
the name/claimName override used to isolate effects. Arm order is shuffled with
seed 20261011. No timing benefit is claimed.

### Partial write

`observe -> prepare -> mutate -> verify -> publish`

The source `mutate` writes dirty input and fails. Retrying it alone sees that
dirty input and fails before another write. `mutate: retryWith: [prepare]`
reinitializes the input and lets recovery complete. The task's `DIRTY-INPUT`
check blocks the empty-policy arm; the planner itself only issues an advisory
Workspace warning.

### Forward invalidation

`observe -> prepare -> build -> gate -> publish`

The source fails at `gate` after a successful build. The operator selects
`prepare`, whose run-specific input changes. An empty policy inherits the old
build; `publish` detects an input/artifact digest mismatch and fails before
writing a receipt. `prepare: retryWith: [build]` regenerates the artifact.
The planner does not itself certify semantic consistency.

## Output oracle

Native Tekton Succeeded conditions and TaskRuns report execution. A separate
Python check recomputes SHA-256 from captured Workspace bytes, checks the
artifact's stored input digest, checks artifact contents, and checks the
published receipt. This check uses actual bytes rather than expected planner
dispositions. Policies, tasks, fault mechanisms, and assertions nevertheless
share authorship; this is not independent policy adjudication.

Correct manual selection under an empty policy reproduces each declared
closure's task dispositions. The distinction is storing and applying recovery
knowledge, not adding an execution scope unavailable to a knowledgeable
operator.

## Reproduction

Use a new namespace; the harness never resets an existing namespace:

```bash
go build -o /tmp/retryctl ./cmd/retryctl
python3 hack/closure_ablation.py --context kind-retry-paper \
  --namespace retry-closure-new-unique-name --repetitions 5 \
  --output artifacts/closure-ablation
python3 hack/verify_closure_ablation.py \
  --results artifacts/closure-ablation --retryctl /tmp/retryctl
```

The live harness and verifier use the Python standard library. The historical
numerical audit additionally requires PyYAML and SciPy. Retained evidence is in
`results/closure-ablation-20261011/` and
`results/paper-audit-20261011/`. Build metadata records the clean v0.3.3 commit
used for live comparisons. The companion revision adds the explicit fixtures,
later-claim warning, harnesses, and audit.

## Observed results (2026-10-11)

| Mechanism | Arm | Successful / trials | Consistent receipts / trials | TaskRuns per trial |
|---|---|---:|---:|---:|
| Partial write | Empty | 0/5 | 0/5 | 1, then failure |
| Partial write | Closure | 5/5 | 5/5 | 4 |
| Partial write | Full restart | 5/5 | 5/5 | 5 |
| Forward invalidation | Empty | 0/5 | 0/5 | 3, including failed publish |
| Forward invalidation | Closure | 5/5 | 5/5 | 4 |
| Forward invalidation | Full restart | 5/5 | 5/5 | 5 |

Native objects and captured bytes verify all 30 arms. Correct manual selection
reproduces all ten declared plans. Each successful closure avoids 1/5 TaskRuns
(20%) against restart; lower counts in failing arms are not successful savings.

## Limits

- Five repetitions per mechanism on one small Pipeline shape and one host.
- Synthetic local artifacts, not a registry, database, or deployment system.
- No policy-author study, failure-prevalence estimate, CPU/cost measurement,
  or production recovery-rate claim.
- Existing claims were forked for isolation; prior stateful results separately
  exercise mounting the original source claim.
- The new later-claim warning is tested on planning fixtures, not as a
  transaction or a live concurrent/chained recovery guarantee.
